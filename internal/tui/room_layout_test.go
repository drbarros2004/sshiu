package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"sshiu/internal/rooms"
)

func TestFeedKeepsCoinReceiptAndQuotes(t *testing.T) {
	coins := 34
	s := rooms.Snapshot{Max: 10, Events: []rooms.Event{{ID: 1, Who: "eu", Nick: "eu", Verb: "cumpriu",
		Task: strings.Repeat("revisar ", 30), Coins: &coins, Reacts: [rooms.Reactions][]string{0: {"ana"}}}}}
	m := model{id: "eu", room: "s", lang: "en", st: newStyles(themes[0]), w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}
	got := ansi.Strip(m.feed(1, 90))
	if !strings.Contains(got, "...\" (+34 coins gained)") || !strings.Contains(got, "👏 1") || ansi.StringWidth(got) > 90 {
		t.Fatalf("tarefa longa perdeu o recibo, aspas ou reação: %s", got)
	}
	s.Events[0].Who = "ana"
	m.w.Rooms["s"] = s
	if strings.Contains(ansi.Strip(m.feed(1, 90)), "coins gained") {
		t.Fatal("mostrou o recibo de outra identidade")
	}
}

func TestRoomFeedUsesAvailableHeight(t *testing.T) {
	s := rooms.Snapshot{Code: "s", Name: "turma", Max: 10, Members: []rooms.Member{{ID: "eu", Nick: "eu"}, {ID: "ana", Nick: "ana"}}}
	for i := range 20 {
		s.Events = append(s.Events, rooms.Event{ID: i + 1, Who: "eu", Nick: "eu", Verb: "cumpriu", Task: fmt.Sprintf("atividade-%02d", i)})
	}
	m := model{id: "eu", nick: "eu", room: "s", st: newStyles(themes[0]), now: time.Now(), w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}
	large, n := m.roomLayout(200, 90)
	if n != 20 || !strings.Contains(ansi.Strip(large), "atividade-19") {
		t.Fatal("com altura sobrando, o feed deve mostrar os 20 eventos disponíveis")
	}
	small, n := m.roomLayout(200, 35)
	if n >= 20 || strings.Count(small, "\n")+1 > 35 || strings.Contains(ansi.Strip(small), "atividade-19") {
		t.Fatal("o feed deve encolher com a tela sem forçar os 20 eventos")
	}
}

func TestRoomSectionsUseRightEdgeAndDividers(t *testing.T) {
	s := rooms.Snapshot{Code: "s", Name: "turma", Max: 10, Obra: "cidade", Meta: 60, BoardOpen: true,
		Members: []rooms.Member{{ID: "eu", Nick: "eu"}, {ID: "ana", Nick: "ana"}},
		Events:  []rooms.Event{{ID: 1, Who: "eu", Nick: "eu", Verb: "cumpriu", Task: "revisar"}}}
	m := model{id: "eu", room: "s", now: time.Now(), st: newStyles(themes[0]), w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}
	band := m.roomBand(s, 250, bandArranjo)
	out := ansi.Strip(m.roomBody(s, "turma", 250, 1, "cards", bandArranjo, band))
	if !strings.Contains(out, "│") {
		t.Fatal("faltaram divisores entre as seções")
	}
	right := 0
	for _, line := range strings.Split(out, "\n") {
		right = max(right, ansi.StringWidth(strings.TrimRight(line, " ")))
		if ansi.StringWidth(line) > 250 {
			t.Fatal("seção passou da largura da sala")
		}
	}
	if right != 250 {
		t.Fatalf("a tabela deveria alcançar a margem direita: %d de 250", right)
	}
	if strings.Contains(ansi.Strip(m.roomBand(s, 250, bandSemTabela)), "tabela de foco") {
		t.Fatal("o estágio sem tabela deve manter só a obra")
	}
}

func TestEmptyFeedStaysBesideArtwork(t *testing.T) {
	s := rooms.Snapshot{Code: "s", Name: "turma", Max: 10, Obra: "cidade", Meta: 60, BoardOpen: true,
		Members: []rooms.Member{{ID: "eu", Nick: "eu"}, {ID: "ana", Nick: "ana"}}}
	m := model{id: "eu", room: "s", lang: "en", now: time.Now(), st: newStyles(themes[0]), w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}
	out := ansi.Strip(m.roomView(240, 80))
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "quiet in here.") {
			if !strings.Contains(line, "city") || !strings.Contains(line, "│") {
				t.Fatal("a mensagem vazia deve ficar à esquerda da obra, na mesma faixa")
			}
			if strings.Contains(out, "react to the selected event:") {
				t.Fatal("feed vazio não deve oferecer reações sem evento")
			}
			return
		}
	}
	t.Fatal("a mensagem de feed vazio desapareceu")
}

func TestLongEventsRequestMoreWidth(t *testing.T) {
	s := rooms.Snapshot{Code: "s", Max: 10, Events: []rooms.Event{{ID: 1, Who: "eu", Nick: "eu", Verb: "cumpriu", Task: "ler"}}}
	m := model{id: "eu", room: "s", lang: "en", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}
	short := m.preferredFeedWidth(240)
	s.Events[0].Task = strings.Repeat("uma tarefa longa ", 100)
	m.w.Rooms["s"] = s
	long := m.preferredFeedWidth(240)
	if long <= short || long > 120 {
		t.Fatalf("a largura deve crescer com a tarefa até metade da tela: curta=%d, longa=%d", short, long)
	}
}
