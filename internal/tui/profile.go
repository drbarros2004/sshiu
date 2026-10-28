package tui

// Abas "perfil" (o pet grande + histórico) e "tema".

import (
	"fmt"
	"plantao/internal/pet"
	"plantao/internal/store"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
)

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
		fmt.Sprintf("streak   %d dia(s)", m.pet.StreakAt(m.now)) + "\n\n" +
		bold.Render("últimos sprints") + "\n"
	if len(m.hist) == 0 {
		info += m.st.muted.Render("nenhum ainda")
	}
	for _, r := range m.hist {
		label := resultLabel[r.Result]
		switch r.Result {
		case pet.Done:
			label = m.st.ok.Render(fmt.Sprintf("%-14s", label))
		default:
			label = m.st.bad.Render(fmt.Sprintf("%-14s", label))
		}
		info += fmt.Sprintf("%s  %s %s\n", r.At.Format("02/01 15:04"), label, m.st.muted.Render(r.Lock))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, pet.RenderBig(mood, m.tick, m.pixel), "    ", strings.TrimRight(info, "\n"))
}

func (m model) themeKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.theme = (m.theme + len(themes) - 1) % len(themes)
	case "down", "j":
		m.theme = (m.theme + 1) % len(themes)
	case "enter":
		id, name := m.id, themes[m.theme].name
		m.saved = m.theme
		return m, act(func() {
			if err := store.SetTheme(id, name); err != nil {
				log.Error("setTheme", "err", err)
			}
		})
	}
	m.st = newStyles(themes[m.theme]) // prévia ao vivo
	return m, nil
}

func (m model) themeView() string {
	var b strings.Builder
	for i, t := range themes {
		cursor, name := "  ", fmt.Sprintf("%-16s", t.name)
		if i == m.theme {
			cursor, name = m.st.sel.Render("▸ "), m.st.sel.Render(name)
		}
		var sw string
		for _, c := range []string{t.accent, t.good, t.bad, t.muted} {
			sw += lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render("██") + " "
		}
		if i == m.saved {
			sw += m.st.muted.Render(" (atual)")
		}
		b.WriteString(cursor + name + sw + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
