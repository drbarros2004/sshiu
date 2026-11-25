package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// O conteúdo recente define a largura, sem deslocar a obra ao mudar só a seleção.
// Metade da tela é o teto: uma tarefa enorme não deve expulsar a obra sozinha.
func (m model) preferredFeedWidth(width int) int {
	natural := lipgloss.Width(m.feedBlock(shownEvents, width))
	return min(width, max(feedMin, min(natural, width/2)))
}

func (m model) feedBlock(count, width int) string {
	block := m.feed(count, width)
	if len(m.events()) > 0 {
		block += "\n\n" + m.legend()
	}
	return block
}

// Reserva o recibo e as aspas antes de encurtar a tarefa. Serve ao feed e ao mini log solo.
func fitEventTask(prefix, task, receipt string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(receipt) >= width {
		return ansi.Truncate(strings.TrimLeft(receipt, " "), width, "…")
	}
	available := width - ansi.StringWidth(receipt)
	if task == "" || available < 3 {
		return ansi.Truncate(prefix, available, "…") + receipt
	}
	prefix = ansi.Truncate(prefix, max(0, available-3), "…")
	taskWidth := max(0, available-ansi.StringWidth(prefix)-3)
	ellipsis := "..."
	if taskWidth < 3 {
		ellipsis = "…"
	}
	return prefix + " \"" + ansi.Truncate(task, taskWidth, ellipsis) + "\"" + receipt
}
