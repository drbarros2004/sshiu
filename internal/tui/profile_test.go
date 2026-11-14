package tui

import (
	"encoding/json"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/store"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Percorrer todos os desenhos e salvar a cor não pode mudar o pet de verdade.
func TestMoodPreviewIsolation(t *testing.T) {
	if err := store.Open(":memory:"); err != nil {
		t.Fatal(err)
	}
	if err := store.PutUser("preview", "ana", pet.EnergyStart); err != nil {
		t.Fatal(err)
	}
	_, original, _, _ := store.GetUser("preview")
	m := model{id: "preview", nick: "ana", tab: tabProfile, now: time.Now(), pet: original,
		st: newStyles(themes[0])}
	press := func(code rune) tea.Cmd {
		t.Helper()
		next, cmd := m.profileKey(tea.KeyPressMsg{Code: code})
		m = next.(model)
		return cmd
	}
	press('c')
	if m.previewMood != pet.MoodOf(original.EnergyAt(m.now)) {
		t.Fatal("a prévia não começou no humor atual")
	}
	seen := map[pet.Mood]bool{}
	for range len(pet.MoodName) {
		seen[m.previewMood] = true
		if cmd := press(tea.KeyDown); cmd != nil || m.pet != original {
			t.Fatal("a prévia alterou o pet ou produziu um efeito")
		}
	}
	if len(seen) != len(pet.MoodName) || !seen[pet.Sleeping] {
		t.Fatal("a prévia não percorreu todos os estados, incluindo dormindo")
	}
	press(tea.KeyUp)
	press(tea.KeyRight)
	if !strings.Contains(m.profileView(80, 100), "prévia do pet") ||
		!strings.Contains(m.profileView(80, 100), "não altera humor ou energia") {
		t.Fatal("a tela não identifica a simulação")
	}
	m.width, m.height = 80, 24
	m.pixel = true
	if view := m.View().Content; !strings.Contains(view, "não altera humor ou energia") ||
		!strings.Contains(view, pet.MoodName[m.previewMood]) {
		t.Fatal("a janela pequena cortou a identificação da prévia")
	}
	cmd := press(tea.KeyEnter)
	if cmd == nil {
		t.Fatal("salvar a cor não produziu um efeito")
	}
	cmd()
	_, saved, _, err := store.GetUser("preview")
	if err != nil || saved.Color != pet.Colors[1].ID {
		t.Fatalf("a cor não foi salva: %+v, %v", saved, err)
	}
	saved.Color = original.Color
	if saved != original || m.custom || strings.Contains(m.profileView(80, 100), "prévia do pet") {
		t.Fatal("salvar a cor alterou os dados de foco ou manteve a simulação")
	}
	press('c')
	if m.previewMood != pet.MoodOf(original.EnergyAt(m.now)) {
		t.Fatal("reabrir manteve o humor simulado")
	}
	press(tea.KeyRight)
	press(tea.KeyEscape)
	if m.custom || m.pcolor != m.pcSaved || m.pet != original {
		t.Fatal("cancelar não descartou a prévia")
	}
	press('c')
	press(tea.KeyDown)
	next, _ := m.switchTab(tabRoom)
	m = next.(model)
	if m.custom || m.pcolor != m.pcSaved || m.pet != original {
		t.Fatal("trocar de aba não descartou a prévia")
	}
}

// O perfil cabe na altura, nos dois layouts: as seções saem de baixo para cima, o estado e o
// calendário ficam até o fim (em 80×24 também), e a seção das tarefas nunca fica só com o título.
func TestProfileFitsHeight(t *testing.T) {
	m := model{nick: "ana", now: time.Now(), st: newStyles(themes[0]), days: map[string]float64{},
		pet: pet.Pet{LastActive: time.Now(), Energy: 50}, totals: store.Totals{Tasks: 3, Done: 3, Bands: [3][2]int{{3, 3}}}}
	m.hours[9] = 60
	m.days[dayKey(m.now)], m.days[dayKey(m.now.AddDate(0, -1, 0))] = 30, 600 // "mês passado −9h 30 até o dia N": o número mais largo
	for range 8 {
		m.tasks = append(m.tasks, store.TaskRow{Start: m.now, Task: "lista 3", Result: pet.Done})
	}
	for _, list := range []bool{false, true} {
		m.statsList = list
		for _, wh := range [][2]int{{76, 17}, {76, 21}, {90, 33}, {116, 33}, {116, 60}} {
			w, h := wh[0], wh[1]
			view := ansi.Strip(m.profileView(w, h))
			lines := strings.Split(view, "\n")
			if len(lines) > h || !strings.Contains(view, "humor") || slices.ContainsFunc(lines, func(l string) bool { return ansi.StringWidth(l) > w }) {
				t.Fatalf("lista %v, %dx%d: %d linhas, sem o estado ou mais largo que a tela:\n%s", list, w, h, len(lines), view)
			}
			if h >= 17 && !strings.Contains(view, m.t(months[m.now.Month()-1])) {
				t.Fatalf("lista %v, %dx%d: sem o calendário:\n%s", list, w, h, view)
			}
			if i := slices.IndexFunc(lines, func(l string) bool { return strings.Contains(l, "últimas tarefas") }); i >= 0 &&
				(i+1 >= len(lines) || !strings.Contains(lines[i+1], "lista 3")) {
				t.Fatalf("lista %v, %dx%d: o título das tarefas sem nenhuma:\n%s", list, w, h, view)
			}
			if h == 60 && !strings.Contains(view, "quando você foca") {
				t.Fatalf("lista %v, %dx%d: sem as stats com altura de sobra:\n%s", list, w, h, view)
			}
		}
	}
}

// O primeiro acesso diz que a cor é o último passo, e o enter leva para a sala.
func TestFirstAccessEndsInRoom(t *testing.T) {
	if err := store.Open(":memory:"); err != nil {
		t.Fatal(err)
	}
	m := model{id: "novo", nick: "ana", tab: tabProfile, custom: true, first: true, now: time.Now(),
		st: newStyles(themes[0]), pet: pet.Pet{LastActive: time.Now()}}
	if !strings.Contains(ansi.Strip(m.profileView(76, 40)), "último passo") {
		t.Fatal("o primeiro acesso não diz que a cor é um passo")
	}
	next, _ := m.profileKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m = next.(model); m.tab != tabRoom || m.first {
		t.Fatalf("o enter não levou para a sala: aba %v", m.tab)
	}
}

func TestFirstAccessMoodPreview(t *testing.T) {
	m := model{}
	next, _ := m.Update(nickMsg{nick: "ana"})
	m = next.(model)
	if !m.custom || m.tab != tabProfile || m.previewMood != pet.MoodOf(pet.EnergyStart) {
		t.Fatal("o primeiro acesso não abriu a prévia no humor inicial")
	}
}

// O traço ocupa a célula que completa a faixa, mesmo quando a energia é fracionária.
func TestEnergyBarThreshold(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		energy  float64
		marker  int
		caption string
	}{
		{0, 3, "próximo humor em 20 · faltam 20"},
		{20, 7, "próximo humor em 40 · faltam 20"},
		{40, 11, "próximo humor em 60 · faltam 20"},
		{60, 15, "próximo humor em 80 · faltam 20"},
		{79.9, 15, "próximo humor em 80 · falta 1"},
		{80, -1, "humor máximo"},
		{100, -1, "humor máximo"},
	} {
		m := model{nick: "ana", st: newStyles(themes[0]), now: now, pet: pet.Pet{Energy: tc.energy, LastActive: now}, statsList: true} // na lista a barra fica sem a borda do painel
		view := ansi.Strip(m.profileView(80, 100))
		_, after, ok := strings.Cut(view, "energia  ")
		if !ok {
			t.Fatalf("energia %v: barra ausente", tc.energy)
		}
		barLine, _, _ := strings.Cut(after, "\n")
		bar := barLine[:strings.LastIndex(strings.TrimRight(barLine, " "), " ")]
		cells := []rune(bar)
		if len(cells) != 20 {
			t.Fatalf("energia %v: barra tem %d células", tc.energy, len(cells))
		}
		marker := slices.Index(cells, energyMarkerGlyph)
		filled := int(tc.energy) / 5
		solidBar := strings.Repeat("█", filled) + strings.Repeat(" ", 20-filled)
		if marker != tc.marker || strings.Replace(bar, string(energyMarkerGlyph), " ", 1) != solidBar || !strings.Contains(view, tc.caption) {
			t.Errorf("energia %v: barra %q e legenda esperada %q", tc.energy, bar, tc.caption)
		}
		if tc.energy == 79.9 && strings.TrimSpace(barLine) != bar+" 79" {
			t.Error("a barra indica 80 antes de o pet ficar radiante")
		}
	}
}

func TestEnergyMarkerColorPriority(t *testing.T) {
	var colors energyPalette
	err := json.Unmarshal([]byte(`{
  "bar": {"radiante": "#FFD700"},
  "unfilled": "xterm:243",
  "markers": {"all": null, "before_radiante": null}
 }`), &colors)
	if err != nil {
		t.Fatal(err)
	}
	if got := colors.markerColor(pet.Radiant); got != 220 {
		t.Errorf("cor do próximo humor = %d, queria 220", got)
	}
	if colors.Unfilled == nil || *colors.Unfilled != 243 {
		t.Error("a cor xterm da parte vazia não foi preservada")
	}
	err = json.Unmarshal([]byte(`{
  "bar": {"all": "xterm:240"},
  "markers": {"all": "#FF8787", "before_radiante": "xterm:221"}
 }`), &colors)
	if err != nil {
		t.Fatal(err)
	}
	if got := colors.barColor(pet.Radiant); got != 240 {
		t.Errorf("bar.all = %d, queria 240", got)
	}
	if got := colors.markerColor(pet.Radiant); got != 210 {
		t.Errorf("markers.all = %d, queria 210", got)
	}
	colors.Markers["all"] = nil
	if got := colors.markerColor(pet.Radiant); got != 221 {
		t.Errorf("exceção do traço = %d, queria 221", got)
	}
}
