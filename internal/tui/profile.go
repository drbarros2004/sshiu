package tui

// Abas "perfil" (o pet grande + as últimas tarefas) e "visual" (tema e modo jardim).

import (
	"fmt"
	"plantao/internal/pet"
	"plantao/internal/store"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
)

var resultLabel = map[pet.Result]string{pet.Done: "✓ cumprida", pet.NotDone: "não cumprida", pet.Interrupted: "interrompida"}

func (m model) profileView() string {
	if m.pet.LastActive.IsZero() {
		return m.st.muted.Render("carregando…")
	}
	e := m.pet.EnergyAt(m.now)
	mood := pet.MoodOf(e)
	n := int(e / 5)
	moodSt := lipgloss.NewStyle().Foreground(lipgloss.Color(pet.MoodColor[mood]))
	info := bold.Render(m.nick) + "\n\n" +
		"humor    " + moodSt.Render(pet.MoodName[mood]) + "\n" +
		"energia  " + moodSt.Render(strings.Repeat("█", n)) + m.st.muted.Render(strings.Repeat("░", 20-n)) + fmt.Sprintf(" %.0f", e) + "\n" +
		fmt.Sprintf("streak   %d %s", m.pet.StreakAt(m.now), plural(m.pet.StreakAt(m.now), "dia seguido", "dias seguidos")) + "\n\n" +
		m.statsBlock(false) + "\n\n" +
		bold.Render("últimas tarefas") + "\n"
	if len(m.tasks) == 0 {
		info += m.st.muted.Render("nenhuma ainda")
	}
	for _, t := range m.tasks {
		label := fmt.Sprintf("%-13s", resultLabel[t.Result])
		if t.Result == pet.Done {
			label = m.st.ok.Render(label)
		} else {
			label = m.st.muted.Render(label)
		}
		info += fmt.Sprintf("%s  %s %3d min  %s", t.Start.Format("02/01 15:04"), label, int(t.Focused.Minutes()), t.Task)
		if r, ok := m.w.Rooms[t.Room]; ok { // a sala pode ter sumido desde então: aí não há nome a mostrar
			info += m.st.muted.Render(" · " + r.Name)
		}
		info += "\n"
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, pet.RenderBig(mood, m.tick, m.pixel), "    ", strings.TrimRight(info, "\n"))
}

// visualRows: as linhas da aba visual, cada uma com o valor em prévia.
func (m model) visualRows() [][2]string {
	wp := "sorteado a cada tarefa"
	if i := wallpaperIndex(m.vis.Wallpaper); i >= 0 {
		wp = wallpapers[i].name
	}
	shown := map[bool]string{true: "aparece", false: "não aparece"}
	return [][2]string{
		{"tema", themes[m.theme].name},
		{"wallpaper", wp},
		{"relógio", clockModes[m.vis.Clock]},
		{"tarefa", shown[m.vis.Task]},
		{"quem mais está focando", shown[m.vis.Room]},
	}
}

func (m model) visualKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	step := 0
	switch msg.String() {
	case "up", "k":
		m.vrow = max(0, m.vrow-1)
	case "down":
		m.vrow = min(len(m.visualRows())-1, m.vrow+1)
	case "left":
		step = -1
	case "right":
		step = 1
	case "j":
		return m.openGarden(true)
	case "enter":
		id, name, g := m.id, themes[m.theme].name, m.vis
		m.saved, m.visSaved, m.notice = m.theme, m.vis, "salvo"
		return m, act(func() {
			if err := store.SetTheme(id, name); err != nil {
				log.Error("setTheme", "err", err)
			}
			if err := store.SetGarden(id, g); err != nil {
				log.Error("setGarden", "err", err)
			}
		})
	}
	if step == 0 {
		return m, nil
	}
	m.notice = ""
	cycle := func(i, n int) int { return (i + step + n) % n }
	switch m.vrow {
	case 0:
		m.theme = cycle(m.theme, len(themes))
		m.st = newStyles(themes[m.theme]) // prévia ao vivo
	case 1:
		m.vis.Wallpaper = cycleWallpaper(m.vis.Wallpaper, step)
	case 2:
		m.vis.Clock = cycle(m.vis.Clock, len(clockModes))
	case 3:
		m.vis.Task = !m.vis.Task
	case 4:
		m.vis.Room = !m.vis.Room
	}
	return m, nil
}

func (m model) visualView() string {
	var b strings.Builder
	for i, r := range m.visualRows() {
		if i == 1 {
			b.WriteString("\n" + bold.Render("modo jardim") + m.st.muted.Render(" · a tela de descanso durante o foco") + "\n")
		}
		cursor, label, value := "  ", fmt.Sprintf("%-24s", r[0]), m.st.muted.Render("‹ ")+r[1]+m.st.muted.Render(" ›")
		if i == m.vrow {
			cursor, label = m.st.sel.Render("▸ "), m.st.sel.Render(label)
		}
		if i == 0 {
			for _, c := range []string{themes[m.theme].accent, themes[m.theme].good, themes[m.theme].bad, themes[m.theme].muted} {
				value += " " + lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render("██")
			}
		}
		b.WriteString(cursor + label + value + "\n")
	}
	if m.notice != "" {
		b.WriteString("\n" + m.st.ok.Render(m.notice))
	}
	return strings.TrimRight(b.String(), "\n")
}
