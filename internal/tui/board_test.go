package tui

import (
	"errors"
	"slices"
	"sshiu/internal/rooms"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// Empate divide a posição, e quem ainda não focou fica sem número (ADR 0010).
func TestPlaces(t *testing.T) {
	if got, want := places([]float64{90, 30.9, 30.1, 12, 0.5}), []int{1, 2, 2, 4, 0}; !slices.Equal(got, want) {
		t.Fatalf("places = %v, queria %v", got, want)
	}
}

func TestHMS(t *testing.T) {
	for min, want := range map[float64]string{0: "0:00:00", 1.5: "0:01:30", 83.25: "1:23:15"} {
		if got := hms(min); got != want {
			t.Errorf("hms(%v) = %q, queria %q", min, got, want)
		}
	}
}

// A semana começa na segunda e o mês no dia 1, à meia-noite do servidor.
func TestPeriodStart(t *testing.T) {
	now := time.Date(2026, 11, 12, 15, 30, 0, 0, time.Local) // uma quinta
	for p, want := range map[int]time.Time{
		1: time.Date(2026, 11, 9, 0, 0, 0, 0, time.Local),
		2: time.Date(2026, 11, 1, 0, 0, 0, 0, time.Local),
		3: time.Unix(0, 0),
	} {
		if got := periodStart(p, now); !got.Equal(want) {
			t.Errorf("periodStart(%d) = %v, queria %v", p, got, want)
		}
	}
}

// Fora de hoje, a tabela soma os dias anteriores do banco aos minutos de hoje do retrato, que já
// trazem a tarefa em andamento (também de quem caiu e está na tolerância de reconexão). Antes de o
// banco responder pelo período certo, só hoje conta.
func TestBoardMin(t *testing.T) {
	x := rooms.Member{ID: "ana", TodayMin: 15, Status: rooms.Offline}
	m := model{bper: 1, bfor: 1, bmins: map[string]float64{"ana": 100}}
	if got := m.boardMin(x); got != 115 {
		t.Fatalf("semana = %v", got)
	}
	if m.bfor = 2; m.boardMin(x) != 15 {
		t.Fatalf("banco de outro período contou: %v", m.boardMin(x))
	}
	if m.bper = 0; m.boardMin(x) != 15 {
		t.Fatalf("hoje = %v", m.boardMin(x))
	}
}

// A resposta atrasada de um período que você já deixou não sobrescreve o atual.
func TestBoardMsgAtrasada(t *testing.T) {
	m := model{bper: 1, bfor: 1, bmins: map[string]float64{"ana": 100}}
	next, _ := m.Update(boardMsg{period: 2, mins: map[string]float64{"ana": 999}})
	if got := next.(model); got.bfor != 1 || got.bmins["ana"] != 100 {
		t.Fatalf("a resposta do mês sobrescreveu a semana: %+v", got.bmins)
	}
}

// A compacta só aparece aberta para a sala, e mostra até sideRows pessoas; o resto fica na tecla o.
func TestBoardSide(t *testing.T) {
	s := rooms.Snapshot{Owner: "eu", Max: 10}
	for _, n := range []string{"a", "b", "c", "d", "e", "f", "g"} {
		s.Members = append(s.Members, rooms.Member{ID: n, Nick: n})
	}
	m := model{id: "eu"}
	if m.boardSide(s) != "" {
		t.Fatal("a tabela fechada apareceu na sala para o dono")
	}
	s.BoardOpen = true
	side := ansi.Strip(m.boardSide(s))
	if lines := strings.Split(side, "\n"); len(lines) != 1+sideRows || !strings.Contains(lines[0], "+4 no o") {
		t.Fatalf("compacta:\n%s", side)
	}
}

// Fechada, a tabela de foco é só do dono; aberto, a sala toda vê; a sala solo não tem.
func TestBoardShown(t *testing.T) {
	m := model{id: "eu"}
	s := rooms.Snapshot{Owner: "ana", Max: 5}
	if m.boardShown(s) {
		t.Fatal("a tabela fechada apareceu para quem não é dono")
	}
	if s.Owner = "eu"; !m.boardShown(s) {
		t.Fatal("o dono não vê a tabela")
	}
	if s.Max = rooms.MinMembers; m.boardShown(s) {
		t.Fatal("a sala solo mostrou a tabela")
	}
	if s = (rooms.Snapshot{Owner: "ana", Max: 5, BoardOpen: true}); !m.boardShown(s) {
		t.Fatal("a tabela aberta não apareceu para a sala")
	}
}

// A tabela tem prioridade sobre a obra: a obra encolhe e sai com a tabela ainda na tela, e só
// depois a tabela sai (ADR 0010).
func TestFitStepsTabelaAntesDaObra(t *testing.T) {
	sala := rooms.Snapshot{Owner: "eu", BoardOpen: true, Obra: "cidade", Meta: 60, Max: rooms.MaxMembers, Members: []rooms.Member{
		{ID: "eu", Nick: "eu"}, {ID: "a", Nick: "ana"}, {ID: "b", Nick: "bruno"},
	}}
	m := model{id: "eu", room: "s", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": sala}}}
	steps := m.fitSteps(sala, 120)
	if steps[0].bd == "" || steps[0].ob == "" {
		t.Fatal("o primeiro passo tem que ter a tabela e a obra")
	}
	semObra, semTabela := -1, -1
	for i, st := range steps {
		if st.bd != "" && st.ob == "" && semObra < 0 {
			semObra = i
		}
		if st.bd == "" && semTabela < 0 {
			semTabela = i
		}
	}
	if semObra < 0 || semTabela < semObra {
		t.Errorf("a tabela saiu antes da obra: sem obra=%d, sem tabela=%d", semObra, semTabela)
	}
	for _, st := range steps[semTabela:] { // sem a tabela, a obra não volta
		if st.ob != "" {
			t.Fatal("a obra voltou depois de a tabela sair")
		}
	}
}

// Quem não vê a tabela não ouve falar dela: a legenda fica só com a obra, ou sem a tecla o.
func TestObraHint(t *testing.T) {
	sala := rooms.Snapshot{Owner: "ana", Obra: "cidade", Max: 5}
	m := model{id: "eu", room: "s", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": sala}}}
	for _, c := range []struct {
		obra       string
		open, dono bool
		want       string
	}{
		{"cidade", false, false, " · o obra"},
		{"", false, false, ""},
		{"cidade", true, false, " · o obra"},
		{"", true, false, " · o tabela"},
		{"", false, true, " · o tabela"},
	} {
		s := sala
		s.Obra, s.BoardOpen = c.obra, c.open
		if c.dono {
			s.Owner = "eu"
		}
		m.w.Rooms["s"] = s
		if got := m.obraHint(); got != c.want {
			t.Errorf("obra %q aberta %v dono %v: %q, queria %q", c.obra, c.open, c.dono, got, c.want)
		}
	}
}

// A resposta do banco vale para o período pedido; se falhar, o dia fica por ler e o tick tenta
// de novo, em vez de a tabela ficar sem ontem até o fim do dia.
func TestBoardMsg(t *testing.T) {
	m := model{bper: 1}
	next, _ := m.Update(boardMsg{period: 1, day: "2026-11-11", mins: map[string]float64{"ana": 50}})
	if got := next.(model); got.bfor != 1 || got.bday != "2026-11-11" || got.bmins["ana"] != 50 {
		t.Fatalf("resposta certa não entrou: %+v", got)
	}
	m.bday = "2026-11-11"
	next, _ = m.Update(boardMsg{period: 1, err: errors.New("disco cheio")})
	if got := next.(model); got.bday != "" || got.err == "" {
		t.Fatalf("a falha marcou o dia como lido: %q %q", got.bday, got.err)
	}
}

// A tela da tecla o cede ao campo da tarefa e some quando não há mais o que mostrar nela.
func TestOScreen(t *testing.T) {
	s := rooms.Snapshot{Owner: "ana", Max: 5, BoardOpen: true}
	m := model{id: "eu", room: "s", obraFull: true, w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}
	if !m.oScreen() {
		t.Fatal("a tabela aberta não ficou na frente")
	}
	if m.entering = true; m.oScreen() {
		t.Fatal("o campo da tarefa ficou escondido atrás da tecla o")
	}
	m.entering, s.BoardOpen = false, false
	m.w.Rooms["s"] = s
	if m.oScreen() {
		t.Fatal("a tecla o ficou só com o nome da sala depois de a tabela fechar")
	}
}

// A largura decide se a tabela compacta fica ao lado do painel, não o estado dele: a 100 colunas
// ela fica ao lado do relógio e sai só enquanto se escreve a tarefa; a 80 ela não cabe nem com a
// sala parada, em vez de aparecer e sumir a cada tecla.
func TestBoardAoLadoDoPainel(t *testing.T) {
	now := time.Now()
	me := rooms.Member{ID: "eu", Nick: "eu", Status: rooms.Focusing, Focus: rooms.Focus{Task: "lista 3",
		Start: now, SegStart: now, Ends: now.Add(25 * time.Minute), Declared: 25}}
	s := rooms.Snapshot{Owner: "eu", Max: 5, BoardOpen: true, BoardRank: true,
		Members: []rooms.Member{me, {ID: "a", Nick: "ana"}, {ID: "b", Nick: "bruno"}, {ID: "c", Nick: "caio"}}}
	m := model{id: "eu", room: "s", now: now, w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s},
		Users: map[string]rooms.Member{"eu": me}}}
	bd := m.boardSide(s)
	if out := ansi.Strip(m.roomBody(s, "sala", 96, 0, true, "", bd, 0)); !strings.Contains(out, "tabela de foco") {
		t.Fatalf("focando a 100 colunas, a tabela saiu:\n%s", out)
	}
	m.entering, m.in = true, newInput("tarefa", 80)
	if out := ansi.Strip(m.roomBody(s, "sala", 96, 0, true, "", bd, 0)); strings.Contains(out, "tabela de foco") {
		t.Fatalf("escrevendo a tarefa, a tabela espremeu o campo:\n%s", out)
	}
	m.entering, me.Status = false, rooms.Idle
	m.w.Users["eu"] = me
	if out := ansi.Strip(m.roomBody(s, "sala", 76, 0, true, "", bd, 0)); strings.Contains(out, "tabela de foco") {
		t.Fatalf("a 80 colunas, a tabela apareceu com a sala parada e sumiria ao focar:\n%s", out)
	}
}

// Na compacta, a coluna do apelido mede só quem aparece: um apelido comprido mais para baixo na
// lista não alarga a tabela que fica ao lado do painel.
func TestBoardSideLargura(t *testing.T) {
	s := rooms.Snapshot{Max: 10, BoardOpen: true, Members: []rooms.Member{
		{ID: "a", Nick: "ana"}, {ID: "b", Nick: "bia"}, {ID: "c", Nick: "caio"}, {ID: "j", Nick: "joaquim_silveira"}}}
	m := model{id: "eu"}
	for _, l := range strings.Split(ansi.Strip(m.board(s, true, 0)), "\n") {
		if w := ansi.StringWidth(strings.TrimRight(l, " ")); w > len("caio")+2+len("0:00:00") {
			t.Fatalf("a coluna mediu quem não aparece: %q (%d colunas)", l, w)
		}
	}
}

// Numa sala cheia e num terminal baixo, quem não cabe vira um "+N" na última linha, em vez de ser
// cortado sem aviso no pé da tela (#119).
func TestBoardLimit(t *testing.T) {
	s := rooms.Snapshot{Max: 10, BoardRank: true}
	for _, n := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} {
		s.Members = append(s.Members, rooms.Member{ID: n, Nick: n})
	}
	m := model{id: "eu"}
	for _, rank := range []bool{true, false} {
		s.BoardRank = rank
		out := ansi.Strip(m.board(s, false, 5))
		if lines := strings.Split(out, "\n"); len(lines) != 5 || !strings.HasPrefix(lines[4], "+") {
			t.Fatalf("rank=%v: queria 5 linhas com o \"+N\" na última:\n%s", rank, out)
		}
		if out := m.board(s, false, 20); strings.Contains(out, "+") {
			t.Fatalf("rank=%v: com espaço, apareceu o \"+N\":\n%s", rank, out)
		}
	}
}
