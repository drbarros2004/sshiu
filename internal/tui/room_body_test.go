package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"sshiu/internal/rooms"
)

func TestRoomBodyPlanHeightMatchesRender(t *testing.T) {
	for _, language := range []lang{"pt", "en"} {
		for _, width := range []int{40, 80, 160} {
			for _, count := range []int{0, 1, shownEvents} {
				s := rooms.Snapshot{Code: "s", Max: 10, BoardOpen: true,
					Members: []rooms.Member{{ID: "eu"}, {ID: "ana"}}}
				for i := range count {
					s.Events = append(s.Events, rooms.Event{ID: i + 1, Who: "eu", Nick: "eu", Verb: "cumpriu", Task: strings.Repeat("ler ", 40),
						Reacts: [rooms.Reactions][]string{0: {"ana"}}})
				}
				m := model{id: "eu", room: "s", lang: language, st: newStyles(themes[0]), w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}
				for _, mode := range []int{bandFora, bandArranjo, bandSemTabela} {
					band := ""
					if mode != bandFora {
						band = "obra\nlinha\nlinha\nlinha"
					}
					p := m.prepareRoomBody(s, "título\n\npainel\n\ncards", width, mode, band,
						func(w int) string { return m.feedBlock(shownEvents, w) })
					for _, n := range []int{0, 1, 2, shownEvents} {
						t.Run(fmt.Sprintf("%s/%d/%d/%d/%d", language, width, count, mode, n), func(t *testing.T) {
							if got, want := p.height(n), lipgloss.Height(p.render(m, n)); got != want {
								t.Fatalf("a medição deve refletir o desenho: altura %d, desenho %d", got, want)
							}
						})
					}
				}
			}
		}
	}
}

func TestRoomNavigationReusesPlanAndMovesSelection(t *testing.T) {
	s := rooms.Snapshot{Code: "s", Name: "turma", Max: 10,
		Members: []rooms.Member{{ID: "eu", Nick: "eu"}, {ID: "ana", Nick: "ana"}},
		Events:  []rooms.Event{{ID: 1, Who: "eu", Verb: "cumpriu", Task: "ler"}, {ID: 2, Who: "ana", Verb: "começou", Task: "escrever"}}}
	m := model{id: "eu", nick: "eu", room: "s", tab: tabRoom, width: 160, height: 60,
		st: newStyles(themes[0]), now: time.Now(), w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}, Users: map[string]rooms.Member{"eu": s.Members[0]}}}
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = next.(model)
	plan := m.roomPlan
	if plan == nil || m.sel != 2 {
		t.Fatal("a primeira seta deve preparar as medidas e selecionar o segundo evento")
	}
	before := m.View().Content
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	m = next.(model)
	if m.roomPlan != plan || m.sel != 1 || m.View().Content == before {
		t.Fatal("a segunda seta deve reaproveitar as medidas, mas redesenhar a seleção")
	}
	for _, msg := range []tea.Msg{tea.WindowSizeMsg{Width: 80, Height: 24}, tickMsg{now: m.now.Add(time.Second), w: m.w}, tea.KeyPressMsg{Code: 'o'}} {
		next, _ = m.Update(msg)
		if next.(model).roomPlan != nil {
			t.Fatalf("%T deve invalidar o plano para não congelar o layout ou a animação", msg)
		}
	}
}
