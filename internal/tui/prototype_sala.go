package tui

// PROTÓTIPO (issue #7): jeitos de arrumar a sala no espaço do terminal, trocados com "v".
//   atual: a sala de hoje, alinhada à esquerda, sem se ajustar à altura
//   1 coluna central: o conteúdo de hoje numa coluna centralizada, com largura máxima; quando
//     falta altura, os cards perdem o desenho do pet e a lista de eventos encurta
//   2 dois painéis: os cards à esquerda e os eventos à direita (empilha em terminal estreito)
// A variante "foco em destaque" (sala em lista) foi descartada.

import (
	"fmt"
	"plantao/internal/rooms"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	salaNames = []string{"atual", "1 · coluna central", "2 · dois painéis"}
	salaShort = []string{"atual", "1", "2"}
)

// salaLabel: a etiqueta do protótipo, no canto da barra de abas, curta se não couber em w.
func (m model) salaLabel(w int) string {
	if m.tab != tabRoom || m.picking || m.room == "" {
		return ""
	}
	st := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#ffd75f")).Padding(0, 1)
	if l := st.Render("protótipo: sala " + salaNames[m.salaV] + " · v troca"); lipgloss.Width(l) < w {
		return l
	}
	return st.Render("sala " + salaShort[m.salaV] + " · v")
}

func (m model) salaVariant(w, h int) string {
	switch m.salaV {
	case 1:
		return m.salaColumn(w, h)
	case 2:
		return m.salaPanels(w, h)
	}
	s := m.w.Rooms[m.room] // atual: como no master
	return m.roomTitle() + "\n\n" + m.focusPanel() + "\n\n" + m.cards(s, w/18, false) + "\n\n" + m.feed(shownEvents) + "\n" + m.legend()
}

// fitting tenta do mais completo ao mais enxuto (pet desenhado e 6 eventos → ponto colorido e
// nenhum evento) e fica com o primeiro que cabe em h linhas.
func fitting(h int, build func(compact bool, n int) string) string {
	var s string
	for _, compact := range []bool{false, true} {
		for n := shownEvents; n >= 0; n -= 2 {
			if s = build(compact, n); lipgloss.Height(s) <= h {
				return s
			}
		}
	}
	return s
}

func (m model) salaColumn(w, h int) string {
	s := m.w.Rooms[m.room]
	cw := min(w, 110)
	block := fitting(h, func(compact bool, n int) string {
		out := m.roomTitle() + "\n\n" + m.focusPanel() + "\n\n" + m.cards(s, cw/18, compact)
		if n > 0 {
			out += "\n\n" + m.feed(n) + "\n" + m.legend()
		}
		return out
	})
	block = lipgloss.NewStyle().Width(lipgloss.Width(block)).Render(block) // centraliza o bloco, não cada linha
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, block)
}

func (m model) salaPanels(w, h int) string {
	if w < 120 {
		return m.salaColumn(w, h)
	}
	s := m.w.Rooms[m.room]
	rw := min(max(w*2/5, 44), 64)
	lw := w - rw - 2
	left := fitting(h, func(compact bool, _ int) string {
		return m.roomTitle() + "\n\n" + m.focusPanel() + "\n\n" + m.cards(s, lw/18, compact)
	})
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(lw).Render(left), "  ", m.eventsPanel(rw, h))
}

// eventsPanel: os eventos numa caixa da altura da tela, cada um em duas ou três linhas.
func (m model) eventsPanel(w, h int) string {
	tw := w - 4
	var b strings.Builder
	b.WriteString(bold.Render("Eventos") + "\n\n")
	ev, sel := m.events(), m.selIndex()
	if len(ev) == 0 {
		b.WriteString(m.st.muted.Width(tw).Render("nada por aqui ainda: quando alguém começar ou cumprir uma tarefa, aparece aqui"))
	}
	for i, e := range ev {
		if lipgloss.Height(b.String())+4 > h-3 {
			break
		}
		nick := e.Nick
		if e.Who == m.id {
			nick = "você"
		}
		mark := "  "
		if i == sel {
			mark = m.st.title.Render("> ")
		}
		head, at := mark+bold.Render(nick)+" "+e.Verb, m.st.muted.Render(e.At.Format("15:04"))
		b.WriteString(head + strings.Repeat(" ", max(1, tw-lipgloss.Width(head)-lipgloss.Width(at))) + at + "\n")
		b.WriteString("  " + lipgloss.NewStyle().Italic(true).MaxWidth(tw-2).Render(e.Task) + "\n")
		if r := m.reacts(e); r != "" {
			b.WriteString("  " + r + "\n")
		}
		b.WriteString("\n")
	}
	legend := m.st.muted.Render("reagir:") + " 1 👏  2 🔥  3 💪  4 ☕"
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(themes[m.theme].border)).
		Padding(0, 1).Width(w).Render(fit(strings.TrimRight(b.String(), "\n"), max(1, h-3)) + "\n" + legend)
}

// reacts: as reações de um evento numa linha, as suas em destaque.
func (m model) reacts(e rooms.Event) string {
	var parts []string
	for k, who := range e.Reacts {
		if len(who) == 0 {
			continue
		}
		r := fmt.Sprintf("%s %d", reactionGlyphs[k], len(who))
		if slices.Contains(who, m.id) {
			r = m.st.title.Render(r)
		}
		parts = append(parts, r)
	}
	return strings.Join(parts, "  ")
}
