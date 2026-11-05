package tui

import (
	"sshiu/internal/pet"
	"sshiu/internal/store"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
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
	if !strings.Contains(m.profileView(), "prévia do pet") ||
		!strings.Contains(m.profileView(), "não altera humor ou energia") {
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
	if saved != original || m.custom || strings.Contains(m.profileView(), "prévia do pet") {
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

func TestFirstAccessMoodPreview(t *testing.T) {
	m := model{}
	next, _ := m.Update(nickMsg{nick: "ana"})
	m = next.(model)
	if !m.custom || m.tab != tabProfile || m.previewMood != pet.MoodOf(pet.EnergyStart) {
		t.Fatal("o primeiro acesso não abriu a prévia no humor inicial")
	}
}
