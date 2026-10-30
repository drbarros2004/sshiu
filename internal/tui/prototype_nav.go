package tui

// PROTÓTIPO (issue #7): a casca escolhida, com as abas em caixas na barra de cima, que crescem
// com o terminal. O layout anterior está na tag layout-abas-simples.

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// protoScreen monta a tela inteira: abas em cima, rodapé de teclas preso na última linha e,
// no meio, o conteúdo da aba, que recebe a largura e a altura que sobram.
func (m model) protoScreen(content func(w, h int) string, help string) string {
	inner := m.width - 4
	top := m.boxTabs()
	if m.tab == tabRoom && len(m.w.MyRooms(m.id)) > 0 {
		top += "\n" + m.roomStrip()
	}
	nav := "a/d troca de aba · q sai"
	if m.typing() {
		nav = "tab troca de aba · ctrl+c sai"
	}
	if help != "" {
		nav = help + " · " + nav
	}
	foot := m.st.muted.Width(inner).Render(nav)
	h := max(1, m.height-2-lipgloss.Height(top)-2-lipgloss.Height(foot))
	mid := fit(content(inner, h-lipgloss.Height(m.errLine())+1)+m.errLine(), h)
	return lipgloss.NewStyle().Padding(1, 2).Render(top + "\n\n" + mid + "\n\n" + foot)
}

// fit deixa s com exatamente h linhas: corta o que passa, completa com linhas vazias.
func fit(s string, h int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	return strings.Join(lines[:h], "\n")
}

// boxTabs: as abas como caixas que dividem a largura; a selecionada fica aberta embaixo.
func (m model) boxTabs() string {
	labels := []string{"SALAS", "PERFIL", "PERSONALIZAR"}
	line := lipgloss.Color(themes[m.theme].border)
	status := m.status(m.me())
	inner := m.width - 4
	each := min(max((inner-lipgloss.Width(status)-4)/4, 16), 32)
	var tabs []string
	for i, l := range labels {
		on := tab(i) == m.tab
		b := lipgloss.Border{Top: "─", Bottom: "─", Left: "│", Right: "│", TopLeft: "╭", TopRight: "╮", BottomLeft: "┴", BottomRight: "┴"}
		st := lipgloss.NewStyle().BorderForeground(line).Width(each).Align(lipgloss.Center)
		label := m.st.muted.Render(l)
		if on {
			b.Bottom, b.BottomLeft, b.BottomRight = " ", "┘", "└"
			st = st.BorderForeground(m.st.accent)
			label = m.st.title.Render(l)
		}
		if i == 0 {
			b.BottomLeft = map[bool]string{true: "│", false: "└"}[on]
		}
		tabs = append(tabs, st.Border(b).Render(label))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
	gap := max(0, inner-lipgloss.Width(row))
	fill := lipgloss.NewStyle().Border(lipgloss.Border{Bottom: "─"}, false, false, true, false).
		BorderForeground(line).Width(gap).Align(lipgloss.Right).Render(m.salaLabel(gap-1) + "\n" + status + " ")
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, fill)
}

// roomStrip: as suas salas numa linha, embaixo da aba SALAS.
func (m model) roomStrip() string {
	var parts []string
	for _, code := range m.w.MyRooms(m.id) {
		name := m.w.Rooms[code].Name
		if code == m.room {
			parts = append(parts, m.st.sel.Render("● "+name))
		} else {
			parts = append(parts, m.st.muted.Render("○ "+name))
		}
	}
	return " " + strings.Join(parts, "   ") + m.st.muted.Render("     s troca, cria ou entra com código")
}
