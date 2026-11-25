package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"sshiu/internal/rooms"
)

// Reaproveitado entre setas consecutivas e o View. Qualquer outra mensagem o invalida,
// inclusive tick, redimensionamento e atualização da sala: animações continuam andando.
type roomLayoutResult struct {
	plan                  roomBodyPlan
	events, width, height int
}

// As tentativas de encaixe só consultam medidas. Apenas o arranjo escolhido é desenhado.
// O plano não sobrevive ao próximo tick ou atualização da sala.
type roomBodyPlan struct {
	top, band, tail               string
	width, feedWidth, columnWidth int
	topHeight, bandHeight         int
	emptyFeedHeight, legendHeight int
	events                        int
	count                         int // linhas pedidas ao feed; pode mostrar a mensagem vazia sem evento navegável
	feed, beside                  bool
}

func (m model) prepareRoomBody(s rooms.Snapshot, top string, w, mode int, band string, feedOf func(int) string) roomBodyPlan {
	p := roomBodyPlan{top: top, band: band, width: w, feedWidth: w}
	if m.alone() {
		if band != "" {
			p.top += "\n\n" + band
		}
		p.top += "\n\n" + m.invitePanel(s)
		p.band = ""
		p.topHeight = lipgloss.Height(p.top)
		return p
	}
	p.feed, p.events = true, len(m.events())
	if band == "" && mode == bandFora && (m.obraOf(s) != nil || m.boardInRoom(s)) {
		p.top += "\n\n" + m.st.muted.Render(m.bandAway(s))
	}
	if band != "" {
		p.bandHeight = lipgloss.Height(band)
		p.columnWidth = w - lipgloss.Width(band) - roomSectionGap
		if p.columnWidth >= feedMin {
			width := min(p.columnWidth, w/2)
			if lipgloss.Width(feedOf(width)) <= p.columnWidth {
				p.beside, p.feedWidth = true, width
			}
		}
	}
	if p.events == 0 {
		p.emptyFeedHeight = lipgloss.Height(feedOf(p.feedWidth))
	} else {
		p.legendHeight = lipgloss.Height(m.legend())
	}
	if mode == bandSemTabela && m.boardInRoom(s) {
		p.tail = "\n" + m.st.muted.Render(m.t("tabela de foco")+m.t(" · na tecla o"))
	}
	p.topHeight = lipgloss.Height(p.top)
	return p
}

func (p roomBodyPlan) height(n int) int {
	h := p.topHeight + strings.Count(p.tail, "\n")
	feedHeight := 0
	if p.feed && n > 0 {
		feedHeight = p.emptyFeedHeight
		if p.events > 0 {
			// Cada evento ocupa uma linha; a legenda vem depois de uma linha vazia.
			feedHeight = min(n, p.events) + 1 + p.legendHeight
		}
	}
	if p.band != "" {
		if p.beside && feedHeight > 0 {
			return h + 1 + max(p.bandHeight, feedHeight)
		}
		h += 1 + p.bandHeight
	}
	if feedHeight > 0 {
		h += 1 + feedHeight
	}
	return h
}

func (p roomBodyPlan) render(m model, n int) string {
	out, log := p.top, ""
	if p.feed && n > 0 {
		log = m.feedBlock(n, p.feedWidth)
	}
	if p.band != "" {
		if p.beside && log != "" {
			log = lipgloss.NewStyle().Width(p.columnWidth).Render(log)
			out += "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, log,
				m.roomDivider(max(lipgloss.Height(log), p.bandHeight)), p.band)
			log = ""
		} else {
			out += "\n\n" + lipgloss.PlaceHorizontal(p.width, lipgloss.Right, p.band)
		}
	}
	if log != "" {
		out += "\n\n" + log
	}
	return out + p.tail
}
