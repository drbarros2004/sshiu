package tui

// A tabela de foco (ADR 0010): quanto cada membro focou hoje, na semana, no mês ou no total, numa
// tabela com o tempo exato ou num ranking. O dono vê sempre, no menu c e na tecla o; aberta, a
// sala toda vê, na tecla o e compacta ao lado do painel de foco. Os minutos são os da pessoa, em
// qualquer sala, como os da peça da obra. O período é de quem olha; o formato e quem vê são do dono.

import (
	"cmp"
	"fmt"
	"slices"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// periods: os períodos da tabela, na ordem das setas.
var periods = []string{"hoje", "esta semana", "este mês", "total"}

// boardWho: quem vê a tabela, como a página do dono diz.
func (m model) boardWho(open bool) string {
	if open {
		return m.t("a sala toda")
	}
	return m.t("só você")
}

// sideRows: quantas pessoas a tabela compacta mostra; o resto fica na tecla o (docs/regras.md).
// Com o título, ela fica só uma ou duas linhas mais alta que o painel de foco ao lado.
const sideRows = 3

// boardMsg: os minutos das tarefas encerradas antes de hoje no período, lidos do banco no dia day.
type boardMsg struct {
	period int
	day    string
	mins   map[string]float64
	err    error // falhou: o tick tenta de novo
}

// dayStart: o começo do dia de t, no fuso do servidor.
func dayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// periodStart: desde quando o período p conta. O dia, a semana (de segunda) e o mês são os do fuso
// do servidor, como a obra e o calendário do perfil.
func periodStart(p int, now time.Time) time.Time {
	switch p {
	case 1:
		return dayStart(now).AddDate(0, 0, -weekIndex(now))
	case 2:
		return dayStart(now).AddDate(0, 0, 1-now.Day())
	}
	return time.Unix(0, 0)
}

// loadBoard lê do banco os minutos do período escolhido até ontem; hoje vem do retrato. Hoje
// sozinho não precisa do banco.
func (m model) loadBoard() tea.Cmd {
	p, now := m.bper, m.now
	if p == 0 {
		return nil
	}
	return func() tea.Msg {
		mins, err := store.FocusBetween(periodStart(p, now), dayStart(now))
		return boardMsg{p, dayKey(now), mins, err}
	}
}

// shiftPeriod: ←/→ no período da tabela.
func (m model) shiftPeriod(d int) (tea.Model, tea.Cmd) {
	m.bper = min(max(m.bper+d, 0), len(periods)-1)
	return m, m.loadBoard()
}

// boardShown: a tabela de foco aparece para você na tecla o. A sala solo não tem: é uma linha
// só, e o perfil já diz quanto você focou.
func (m model) boardShown(s rooms.Snapshot) bool {
	return s.Max > rooms.MinMembers && (s.BoardOpen || s.Owner == m.id)
}

// boardStale: a tabela está na tela com o banco de outro dia, e precisa reler. Os dias antes de
// hoje só mudam à meia-noite: o resto vem ao vivo no retrato.
func (m model) boardStale() bool {
	return m.bper > 0 && m.bday != dayKey(m.now) && m.tab == tabRoom && !m.inPicker() &&
		m.boardShown(m.w.Rooms[m.room])
}

// boardMin: os minutos de x no período escolhido: os dias antes de hoje, do banco, mais os de hoje
// do retrato, que já trazem a tarefa em andamento. As duas partes contam a tarefa no dia em que
// ela termina, como o pet.
func (m model) boardMin(x rooms.Member) float64 {
	if m.bper == 0 || m.bfor != m.bper {
		return x.TodayMin
	}
	return m.bmins[x.ID] + x.TodayMin
}

// hms: o tempo exato da tabela, em h:mm:ss.
func hms(min float64) string {
	d := time.Duration(min * float64(time.Minute)).Truncate(time.Second)
	return fmt.Sprintf("%d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

// places: a posição de cada um no ranking, com mins já do maior para o menor. Quem focou os
// mesmos minutos divide a posição (1º, 1º, 3º), e quem ainda não focou fica sem número (0): o
// ranking não põe ninguém em último por estar começando.
func places(mins []float64) []int {
	out := make([]int, len(mins))
	for i, v := range mins {
		switch {
		case int(v) == 0:
		case i > 0 && int(v) == int(mins[i-1]):
			out[i] = out[i-1]
		default:
			out[i] = i + 1
		}
	}
	return out
}

// boardTitle: o nome do bloco e, para o dono, quem mais o vê.
func (m model) boardTitle(s rooms.Snapshot) string {
	t := bold.Render(m.t("tabela de foco"))
	if s.Owner == m.id {
		who := m.t(" · só você vê")
		if s.BoardOpen {
			who = m.t(" · a sala toda vê")
		}
		t += m.st.muted.Render(who)
	}
	return t
}

// boardNote: de onde vêm os minutos e, em hoje, quando eles zeram. Com a obra na mesma tela, o
// "reinicia em" fica só no cabeçalho dela.
func (m model) boardNote(reset bool) string {
	note := m.t("minutos focados em qualquer sala")
	if reset && m.bper == 0 {
		note += m.tf(" · reinicia em %s", obraResetIn(m.now))
	}
	return m.st.muted.Render(note)
}

// periodArrows: o período com as setas.
func (m model) periodArrows(sel bool) string {
	return m.arrows(sel, m.t(periods[m.bper]), m.bper > 0, m.bper < len(periods)-1)
}

// boardName: o apelido na coluna, cortado em 16 colunas: um apelido antigo de letras largas passa
// disso e quebraria a linha da pessoa em duas.
func (m model) boardName(x rooms.Member) string { return ansi.Truncate(m.obraName(x), 16, "…") }

// board: a tabela, na ordem em que os membros entraram para as linhas não pularem a cada tick, ou
// o ranking, do mais focado para o menos. Com mini, a versão compacta da sala: sem cabeçalho nem o
// que cada um faz agora (os cards já dizem), a barra pela metade e até sideRows pessoas.
func (m model) board(s rooms.Snapshot, mini bool) string {
	ms := slices.Clone(s.Members)
	if s.BoardRank { // pelos minutos inteiros, como a posição: quem empata fica na ordem de entrada
		slices.SortStableFunc(ms, func(a, b rooms.Member) int { return cmp.Compare(int(m.boardMin(b)), int(m.boardMin(a))) })
	}
	mins := make([]float64, len(ms))
	for i, x := range ms {
		mins[i] = m.boardMin(x)
	}
	pl := places(mins) // as posições contam a sala inteira, mesmo quando só o começo aparece
	name, cols := lipgloss.NewStyle().Width(18), barCols
	if mini { // só quem aparece; a coluna do apelido acompanha o mais comprido deles
		n := min(len(ms), sideRows)
		ms, mins, pl = ms[:n], mins[:n], pl[:n]
		w := 0
		for _, x := range ms {
			w = max(w, lipgloss.Width(m.boardName(x)))
		}
		name, cols = name.Width(w+2), barCols/2
	}
	var lines []string
	if !s.BoardRank {
		if !mini {
			lines = append(lines, m.st.muted.Render(name.Render(m.t("apelido"))+fmt.Sprintf("%-11s", m.t("tempo"))+m.t("agora")))
		}
		for i, x := range ms {
			t := hms(mins[i])
			if !mini {
				t = fmt.Sprintf("%-11s", t) + m.status(x)
			}
			lines = append(lines, name.Render(m.boardName(x))+t)
		}
		return strings.Join(lines, "\n")
	}
	for i, p := range pl {
		pos, n := "", 0
		if p > 0 {
			pos, n = m.tf("%dº", p), int(float64(cols)*mins[i]/mins[0])
		}
		line := fmt.Sprintf("%-4s", pos) + name.Render(m.boardName(ms[i]))
		if p == 1 {
			line = m.st.title.Render(line)
		}
		lines = append(lines, line+m.bar(n, cols)+m.st.muted.Render(fmt.Sprintf("%9s", dur(mins[i]))))
	}
	return strings.Join(lines, "\n")
}

// boardBlock: a tabela como ela aparece na tecla o, com o período que quem olha escolhe. Com tight,
// o título e o período dividem uma linha e a nota sai: são quatro linhas a menos, o que faz uma
// sala de 10 pessoas caber em 80×24.
func (m model) boardBlock(s rooms.Snapshot, reset, tight bool) string {
	if tight {
		return m.boardTitle(s) + "  " + m.periodArrows(true) + "\n" + m.board(s, false)
	}
	return m.boardTitle(s) + "\n" + m.boardNote(reset) + "\n\n" +
		cfgRow(m, false, m.t("período"), m.periodArrows(true)) + "\n\n" + m.board(s, false)
}

// boardSide: a tabela compacta da sala, ao lado do painel de foco, só quando o dono a abriu para
// todo mundo. O período é o que você escolheu na tecla o, onde ela aparece inteira.
func (m model) boardSide(s rooms.Snapshot) string {
	if !s.BoardOpen || len(s.Members) < 2 { // a sala solo tem uma pessoa só
		return ""
	}
	title := bold.Render(m.t("tabela de foco")) + m.st.muted.Render(" · "+m.t(periods[m.bper]))
	if rest := len(s.Members) - sideRows; rest > 0 { // o resto, na tecla o
		title += m.st.muted.Render(m.tf(" · +%d no o", rest))
	}
	return title + "\n" + m.board(s, true)
}

// boardPage: a página da tabela no menu do dono: quem vê, o formato e o período, com a tabela
// embaixo. Se não cabe em h linhas, a nota e as linhas em branco saem, como na tecla o.
func (m model) boardPage(s rooms.Snapshot, h int) string {
	format := m.t("tabela")
	if s.BoardRank {
		format = m.t("ranking")
	}
	rows := cfgRow(m, m.brow == 0, m.t("quem vê"), m.arrows(m.brow == 0, m.boardWho(s.BoardOpen), s.BoardOpen, !s.BoardOpen)) + "\n" +
		cfgRow(m, m.brow == 1, m.t("formato"), m.arrows(m.brow == 1, format, s.BoardRank, !s.BoardRank)) + "\n" +
		cfgRow(m, m.brow == 2, m.t("período"), m.periodArrows(m.brow == 2))
	if out := m.boardTitle(s) + "\n" + m.boardNote(true) + "\n\n" + rows + "\n\n" + m.board(s, false); lipgloss.Height(out) <= h {
		return out
	}
	return m.boardTitle(s) + "\n" + rows + "\n" + m.board(s, false)
}
