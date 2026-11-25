package tui

import (
	"slices"
	"sshiu/internal/rooms"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestRoomCode(t *testing.T) {
	for in, want := range map[string]string{
		"fg7dhm":                       "fg7dhm",
		"  FG7DHM\n":                   "fg7dhm", // colado com espaço e quebra de linha
		"ssh -t -p 23234 sshiu fg7dhm": "fg7dhm", // o convite inteiro
		"":                             "",
	} {
		if got := roomCode(in); got != want {
			t.Errorf("roomCode(%q) = %q, queria %q", in, got, want)
		}
	}
}

// O card conta as reações que o membro recebeu, menos as que ele mesmo deu no próprio evento.
func TestReceived(t *testing.T) {
	s := rooms.Snapshot{Events: []rooms.Event{{Who: "ana", Reacts: [rooms.Reactions][]string{
		0: {"bia", "ana"}, 1: {"bia"}}}}}
	if got := received(s)["ana"]; got != 2 {
		t.Fatalf("recebidas: %d", got)
	}
}

func TestSoloGardenWidth(t *testing.T) {
	if got := soloGardenWidth(116); got != 60 { // terminal de 120 colunas
		t.Fatalf("120 colunas: %d", got)
	}
	if got := soloGardenWidth(76); got != 0 { // 80 colunas: o jardim não cabe
		t.Fatalf("80 colunas: %d", got)
	}
}

// Passar a sala e remover alguém listam os outros, na ordem dos cards: ninguém passa a sala nem
// a própria vaga para si mesmo (ADR 0007).
func TestOthers(t *testing.T) {
	m := model{id: "eu", room: "s", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": {Members: []rooms.Member{
		{ID: "sumiu", Status: rooms.Offline},
		{ID: "eu", Status: rooms.Idle},
		{ID: "foca", Status: rooms.Focusing},
	}}}}}
	var got []string
	for _, x := range m.others() {
		got = append(got, x.ID)
	}
	if want := []string{"foca", "sumiu"}; !slices.Equal(got, want) {
		t.Fatalf("outros %v, queria %v", got, want)
	}
}

// Passar a sala pede um "sim": o enter na pessoa só abre a pergunta, que começa no "não", e
// nada vai para o servidor antes de a resposta ser "sim" (ADR 0007).
func TestGiveConfirm(t *testing.T) {
	m := model{id: "eu", room: "s", cfg: true, cpage: pageGive, w: rooms.World{Rooms: map[string]rooms.Snapshot{
		"s": {Name: "turma", Owner: "eu", Members: []rooms.Member{{ID: "eu"}, {ID: "bia", Nick: "bia"}}}}}}
	press := func(code rune) tea.Cmd {
		t.Helper()
		next, cmd := m.cfgKey(tea.KeyPressMsg{Code: code})
		m = next.(model)
		return cmd
	}
	if cmd := press(tea.KeyEnter); cmd != nil || m.cpage != pageGiveYes || m.cwho != "bia" || m.cyes {
		t.Fatalf("escolher a pessoa já passou a sala: page %v, cmd %v", m.cpage, cmd != nil)
	}
	if cmd := press(tea.KeyEnter); cmd != nil || m.cpage != pageGive { // enter no "não"
		t.Fatalf("o não passou a sala: page %v", m.cpage)
	}
	press(tea.KeyEnter)
	press(tea.KeyRight)
	if cmd := press(tea.KeyEnter); cmd == nil || m.cfg {
		t.Fatalf("o sim não passou a sala: cfg %v", m.cfg)
	}
}

func TestOrderedMembers(t *testing.T) {
	s := rooms.Snapshot{Members: []rooms.Member{
		{ID: "sumiu", Status: rooms.Offline},
		{ID: "foca1", Status: rooms.Focusing},
		{ID: "eu", Status: rooms.Idle},
		{ID: "parado", Status: rooms.Idle},
		{ID: "foca2", Status: rooms.Focusing},
	}}
	var got []string
	for _, x := range orderedMembers("eu", s) {
		got = append(got, x.ID)
	}
	if want := []string{"eu", "foca1", "foca2", "parado", "sumiu"}; !slices.Equal(got, want) {
		t.Fatalf("ordem %v, queria %v", got, want)
	}
}

// Quem sai por vontade própria não vê o aviso de quem foi removida, e um tick tirado antes da
// saída não traz a sala de volta. Quem perde a vaga vê o aviso, fora da cor de sucesso.
func TestLeaveDoesNotFlashBack(t *testing.T) {
	member := func(ids ...string) []rooms.Member {
		var ms []rooms.Member
		for _, id := range ids {
			ms = append(ms, rooms.Member{ID: id})
		}
		return ms
	}
	before := rooms.World{Rooms: map[string]rooms.Snapshot{
		"a": {Code: "a", Name: "antes", Members: member("eu", "bia")},
		"b": {Code: "b", Name: "depois", Members: member("eu")}}}
	after := rooms.World{Rooms: map[string]rooms.Snapshot{
		"a": {Code: "a", Name: "antes", Members: member("bia")},
		"b": {Code: "b", Name: "depois", Members: member("eu")}}}
	m := model{id: "eu", room: "", left: "a", w: before, st: newStyles(themes[0])} // logo depois do enter em "sair"
	m.setWorld(before)                                                             // o tick de antes da saída
	if m.room != "b" || m.notice != "" {
		t.Fatalf("o tick velho trouxe a sala de volta: room %q, aviso %q", m.room, m.notice)
	}
	m.setWorld(after)
	if m.room != "b" || m.notice != "" {
		t.Fatalf("a saída virou remoção: room %q, aviso %q", m.room, m.notice)
	}

	m = model{id: "eu", room: "a", w: before, st: newStyles(themes[0])}
	m.setWorld(after) // removida pelo dono
	if m.notice != "você não está mais em antes" || m.noticeStyle().GetForeground() != m.st.bad.GetForeground() {
		t.Fatalf("a remoção não avisou, ou avisou em verde: %q", m.notice)
	}
}

// O cursor do menu de salas acompanha a lista quando ela encolhe.
func TestPickerCursorFollowsList(t *testing.T) {
	m := model{id: "eu", pcur: 5, w: rooms.World{Rooms: map[string]rooms.Snapshot{}}}
	m.setWorld(rooms.World{Rooms: map[string]rooms.Snapshot{}})
	if n := len(m.pickerItems()); m.pcur != n-1 {
		t.Fatalf("cursor %d numa lista de %d", m.pcur, n)
	}
}

// A tarefa comprida encolhe no feed e as reações continuam inteiras na tela.
func TestFeedKeepsReactions(t *testing.T) {
	m := model{id: "eu", room: "s", st: newStyles(themes[0]), w: rooms.World{Rooms: map[string]rooms.Snapshot{
		"s": {Max: 4, Members: []rooms.Member{{ID: "eu"}, {ID: "bia"}}, Events: []rooms.Event{{ID: 1, Who: "bia", Nick: "bia",
			Verb: "começou", Task: strings.Repeat("revisar derivadas ", 8), Reacts: [rooms.Reactions][]string{0: {"eu"}, 1: {"eu"}}}}}}}}
	for _, l := range strings.Split(m.feed(6, 96), "\n") {
		if w := lipgloss.Width(l); w > 96 || !strings.Contains(l, "🔥 1") {
			t.Fatalf("largura %d, linha %q", w, ansi.Strip(l))
		}
	}
}

// Sem tarefa a coluna da direita são os números, e o jardim (m.side) é nil: a tela não pode
// depender dele para se desenhar.
func TestSoloViewWithoutGarden(t *testing.T) {
	m := model{id: "a", room: "s", width: 120, height: 36, st: newStyles(themes[0]), now: time.Now(),
		w: rooms.World{Users: map[string]rooms.Member{"a": {ID: "a"}},
			Rooms: map[string]rooms.Snapshot{"s": {Code: "s", Name: "solo", Max: rooms.MinMembers}}}}
	if got := m.soloView(); !strings.Contains(got, "│") {
		t.Fatalf("sem as duas colunas: %q", got)
	}
}

// Com apelido largo, emoji ou comprido, todos os cards de uma fileira têm a mesma altura e
// largura, e nenhuma fileira passa da tela: o nome comprido quebrava a linha e deslocava os pets (#84).
func TestCardsAlinhados(t *testing.T) {
	nicks := []string{"ana", "田中さくら東京大", "🐱🐶🐭🐹🐰🦊🐻🐼", "abcdefghijklmnop", "ＡＢＣＤＥＦＧＨ",
		"um_apelido_antigo_bem_mais_comprido_que_o_limite"} // de antes do limite: termina em "…"
	s := rooms.Snapshot{Max: rooms.MaxMembers}
	for i, n := range nicks {
		s.Members = append(s.Members, rooms.Member{ID: n, Nick: n, Energy: float64(i * 18),
			Status: rooms.Focusing, Focus: rooms.Focus{Task: "uma tarefa bem comprida para quebrar a linha", Ends: time.Now().Add(time.Minute)}})
	}
	m := model{id: "ana", room: "s", st: newStyles(themes[0]), now: time.Now(),
		w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}
	for _, size := range []int{cardFull, cardMid} { // o médio também: as reações no status não quebram a linha
		for _, w := range []int{40, 76, 116, 156} {
			var cards []string
			for i, x := range s.Members {
				cards = append(cards, m.card(x, i*3, size, m.cardInner(s.Members)))
			}
			for _, c := range cards {
				if lipgloss.Width(c) != lipgloss.Width(cards[0]) || lipgloss.Height(c) != lipgloss.Height(cards[0]) {
					t.Fatalf("tamanho %d: card %d×%d, o primeiro %d×%d:\n%s", size, lipgloss.Width(c), lipgloss.Height(c),
						lipgloss.Width(cards[0]), lipgloss.Height(cards[0]), c)
				}
			}
			for _, l := range strings.Split(m.cards(s, w, size, 0), "\n") {
				if lipgloss.Width(l) > max(w, lipgloss.Width(cards[0])) {
					t.Fatalf("%d colunas: linha de %d", w, lipgloss.Width(l))
				}
			}
		}
	}
	if got := m.cardInner(s.Members[:1]); got != cardMin {
		t.Errorf("sala de apelidos curtos: miolo %d, queria %d", got, cardMin)
	}
}

// Cada fileira recebe todos os cards que cabem antes de abrir a próxima.
func TestCardsFillRows(t *testing.T) {
	s := rooms.Snapshot{Max: rooms.MaxMembers}
	for _, n := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"} {
		s.Members = append(s.Members, rooms.Member{ID: n, Nick: n})
	}
	m := model{id: "a", room: "s", st: newStyles(themes[0]), w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}
	w := 8 * (m.cardInner(s.Members) + 2) // cabem oito por fileira
	var per []int
	for _, l := range strings.Split(m.cards(s, w, cardDot, 0), "\n") {
		if n := strings.Count(l, "╭"); n > 0 {
			per = append(per, n)
		}
	}
	if !slices.Equal(per, []int{8, 1}) {
		t.Errorf("fileiras %v, queria [8 1]", per)
	}
	// dez em fileiras de quatro: 4, 4 e 2
	s10 := s
	s10.Members = append(slices.Clone(s.Members), rooms.Member{ID: "j", Nick: "j"})
	per = nil
	for _, l := range strings.Split(m.cards(s10, 4*(m.cardInner(s10.Members)+2), cardDot, 0), "\n") {
		if n := strings.Count(l, "╭"); n > 0 {
			per = append(per, n)
		}
	}
	if !slices.Equal(per, []int{4, 4, 2}) {
		t.Errorf("dez em fileiras de quatro: %v, queria [4 4 2]", per)
	}
	// no "+N" do fim, a fileira que aparece vem cheia: 8 e "+1", não 5 e "+4"
	if out := ansi.Strip(m.cards(s, w, cardDot, 1)); strings.Count(out, "╭") != 8 || !strings.Contains(out, "+1:") {
		t.Errorf("com uma fileira só, o \"+N\" escondeu quem cabia:\n%s", out)
	}
}

// A prévia de entrar e o menu de salas dizem o que a sala aceita antes do enter.
func TestJoinBlockedAndSummary(t *testing.T) {
	m := model{}
	full := rooms.Snapshot{Max: 2, Members: []rooms.Member{{ID: "a"}, {ID: "b"}}}
	closed := rooms.Snapshot{Max: 4, Closed: true, Members: []rooms.Member{{ID: "a"}}}
	if m.joinBlocked(full) == "" || m.joinBlocked(closed) == "" || m.joinBlocked(rooms.Snapshot{Max: 4}) != "" {
		t.Fatal("joinBlocked errou a sala cheia, a fechada ou a aberta")
	}
	solo := rooms.Snapshot{Max: rooms.MinMembers, Members: []rooms.Member{{ID: "a", Status: rooms.Focusing}}}
	if got := m.roomSummary(solo); got != "sala solo · focando" {
		t.Fatalf("resumo da sala solo: %q", got)
	}
}

// Uma sala de 10 em 80×24 não cabe: as fileiras de cards que sobram viram um "+N", e a tela não
// passa da altura (#119).
func TestRoomViewCheiaCabe(t *testing.T) {
	s := rooms.Snapshot{Code: "abc", Name: "turma", Owner: "eu", Max: 10}
	for _, n := range []string{"eu", "ana", "bia", "caio", "davi", "eva", "fabio", "gil", "heitor", "iris"} {
		s.Members = append(s.Members, rooms.Member{ID: n, Nick: n, Status: rooms.Idle})
	}
	m := model{id: "eu", nick: "eu", room: "abc", st: newStyles(themes[0]), now: time.Now(),
		w: rooms.World{Rooms: map[string]rooms.Snapshot{"abc": s}, Users: map[string]rooms.Member{"eu": s.Members[0]}}}
	out := ansi.Strip(m.roomView(76, 20))
	if h := strings.Count(out, "\n") + 1; h > 20 || !strings.Contains(out, "aumente a janela") {
		t.Fatalf("a sala cheia passou da altura (%d linhas) ou não disse quem ficou de fora:\n%s", h, out)
	}
}
