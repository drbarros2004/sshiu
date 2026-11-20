package tui

// A aba dev, só com a flag -dev: todo mundo do banco numa tabela, com os números, a presença e a
// tarefa editáveis, e atalhos para mexer na sala que você está vendo.

import (
	"fmt"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Dev: flag -dev (main.go). Liga a aba dev; a identidade pelo nome do SSH fica no main.
var Dev bool

// loginHint: o rodapé da tela de apelido. No modo dev quem cai nela entrou com um nome que já é o
// apelido de alguém (um bot da demo, por exemplo), e a chave não tem nada a ver.
func (m model) loginHint() string {
	if Dev {
		return m.t("enter confirma · o nome do SSH já é o apelido de outra pessoa")
	}
	return m.t("enter confirma · sua chave SSH é o seu login")
}

// tabs: quantas abas aparecem. A dev é a última e só existe no modo dev.
func tabs() tab {
	if Dev {
		return tab(len(tabNames))
	}
	return tabDev
}

type devPanel struct {
	rows     []store.DevRow
	row, col int       // pessoa e coluna escolhidas (col indexa devCols)
	input    string    // o campo aberto: "nova" (apelido) ou "tarefa"; "" = nenhum
	at       time.Time // quando as linhas foram lidas
}

type devMsg struct {
	rows []store.DevRow
	w    rooms.World // lido junto: online, tarefa e hoje vêm da memória, e o tick está até 500 ms atrás
	at   time.Time   // os Cmd rodam em paralelo: uma leitura velha pode chegar depois da nova
	sel  string      // a pessoa a escolher (a recém-criada); "" = fica onde está
	err  error
}

// As colunas que mudam com +/-, na ordem da tabela.
var devCols = []struct {
	name  string
	field rooms.DevField
}{{"energ.", rooms.DevEnergy}, {"parado", rooms.DevIdle}, {"hoje", rooms.DevToday}, {"antes", rooms.DevBefore},
	{"streak", rooms.DevStreak}, {"cor", rooms.DevColor}, {"online", rooms.DevOnline}, {"tarefa", rooms.DevFocus}}

// devDo roda f (se houver) e relê a tabela. O erro de f aparece na tela.
func devDo(f func() error) tea.Cmd {
	return func() tea.Msg {
		var err error
		if f != nil {
			err = f()
		}
		at := time.Now()
		rows, lerr := store.DevUsers(at)
		if err == nil {
			err = lerr
		}
		return devMsg{rows: rows, w: rooms.SnapshotWorld(), at: at, err: err}
	}
}

func (m model) devKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	d := &m.dev
	k := msg.String()
	if d.input == "assume" {
		if k == "esc" {
			d.input = ""
		}
		if k == "enter" {
			d.input = ""
			if len(d.rows) > 0 {
				return m, m.assume(d.rows[d.row].ID)
			}
		}
		return m, nil
	}
	if d.input != "" {
		switch k {
		case "esc":
			d.input, m.err = "", ""
		case "enter":
			text, input := strings.TrimSpace(m.in.Value()), d.input
			d.input, m.err = "", ""
			if input == "nova" && text == "" { // enter no campo vazio desiste, como o esc
				return m, nil
			}
			if input == "tarefa" {
				id, code := d.rows[d.row].ID, m.room
				if text == "" {
					text = store.DevTask
				}
				return m, devDo(func() error { rooms.DevStart(id, code, text); return nil })
			}
			return m, func() tea.Msg { // o cursor vai para quem nasceu, que pode estar fora da tela
				id, err := rooms.DevCreate(text)
				msg := devDo(func() error { return err })().(devMsg)
				msg.sel = id
				return msg
			}
		default:
			var cmd tea.Cmd
			m.in, cmd = m.in.Update(msg)
			return m, cmd
		}
		return m, nil
	}
	m.err = ""
	if len(d.rows) == 0 && k != "n" {
		return m, nil
	}
	// segurar a tecla manda vários sinais num evento só ("+++"): cada um é um passo
	if k != "" && (strings.Trim(k, "+=") == "" || strings.Trim(k, "-") == "") {
		sign := 1
		if k[0] == '-' {
			sign = -1
		}
		id, f := d.rows[d.row].ID, devCols[d.col].field
		return m, devDo(func() error {
			for range len(k) {
				if err := rooms.DevAdjust(id, f, sign, time.Now()); err != nil {
					return err
				}
			}
			return nil
		})
	}
	switch k {
	case "ctrl+u":
		d.input = "assume"
		return m, nil
	case "up", "k":
		d.row = max(0, d.row-1)
	case "down", "j":
		d.row = min(len(d.rows)-1, d.row+1)
	case "left", "h":
		d.col = max(0, d.col-1)
	case "right", "l":
		d.col = min(len(devCols)-1, d.col+1)
	case "space":
		if m.room == "" {
			m.err = m.t("abra uma sala na aba sala para pôr gente nela")
			return m, nil
		}
		id, code := d.rows[d.row].ID, m.room
		in := slices.Contains(m.w.MyRooms(id), code)
		return m, tea.Batch(devDo(func() error {
			if in {
				return rooms.Leave(id, code)
			}
			return rooms.Join(id, code)
		}), act(func() {}))
	case "n":
		d.input = "nova"
		m.in = newInput(m.t("apelido da pessoa nova"), 16)
	case "t":
		d.input = "tarefa"
		m.in = newInput(m.t("ex.: fazer o exercício 1 da lista 3"), 80)
		if x := m.w.Users[d.rows[d.row].ID]; x.Focus.Task != "" {
			m.in.SetValue(x.Focus.Task)
		}
	case "x":
		id := d.rows[d.row].ID
		return m, devDo(func() error { rooms.Finish(id, pet.Done); return nil })
	case "o":
		s, ok := m.w.Rooms[m.room]
		if !ok {
			m.err = m.t("abra uma sala na aba sala para passá-la a alguém")
			return m, nil
		}
		id, code, owner := d.rows[d.row].ID, s.Code, s.Owner
		return m, devDo(func() error { return rooms.Transfer(owner, code, id) })
	}
	return m, nil
}

// setRows troca as linhas sem trocar de pessoa: a lista vai por apelido, e quem nasce no meio dela
// (outra sessão, um ssh com nome novo) empurraria o cursor para a pessoa ao lado no meio de um ajuste.
// sel, se não vazio, escolhe outra pessoa.
func (d *devPanel) setRows(rows []store.DevRow, at time.Time, sel string) bool {
	if !at.After(d.at) {
		return false
	}
	id := sel
	if id == "" && d.row < len(d.rows) {
		id = d.rows[d.row].ID
	}
	d.rows, d.at = rows, at
	d.row = max(0, min(d.row, len(rows)-1))
	if i := slices.IndexFunc(rows, func(r store.DevRow) bool { return r.ID == id }); i >= 0 {
		d.row = i
	}
	return true
}

func (m model) devHelp() string {
	if m.dev.input == "assume" {
		return m.t("enter assume identidade e descarta rascunhos · esc cancela")
	}
	if m.dev.input != "" {
		return m.t("enter confirma · esc desiste")
	}
	return m.t("ctrl+u assume usuário") + " · " + m.t("setas escolhem · +/- muda · espaço põe ou tira da sala · o passa a sala · t tarefa · x cumpre · n pessoa nova")
}

// devView cabe em 80 colunas (os números em 6, o apelido em 10) e rola as pessoas na altura h.
func (m model) devView(h int) string {
	d := m.dev
	var b strings.Builder
	if s, ok := m.w.Rooms[m.room]; ok {
		b.WriteString(m.t("espaço põe ou tira de ") + m.st.sel.Render(s.Name) + m.st.muted.Render(" ("+s.Code+")"+m.t(" · * a sala é da pessoa")))
	} else {
		b.WriteString(m.st.muted.Render(m.t("abra uma sala na aba sala para pôr gente nela")))
	}
	// as linhas fixas: título, linha vazia, cabeçalho, linha vazia e as 4 da legenda (+2 do campo)
	fixed := 8
	if d.input != "" {
		fixed += 2
	}
	n := max(1, h-fixed)
	first := max(0, min(d.row-n+1, len(d.rows)-n)) // o cursor fica sempre à vista
	last := min(len(d.rows), first+n)
	if len(d.rows) > n {
		b.WriteString(m.st.muted.Render(m.tf(" · %d a %d de %d", first+1, last, len(d.rows))))
	}
	b.WriteString("\n\n")
	head := fmt.Sprintf("  %-10s", m.t("apelido"))
	for _, c := range devCols {
		head += fmt.Sprintf(" %6s", m.t(c.name))
	}
	b.WriteString(m.st.muted.Render(head+"  "+m.t("salas")) + "\n")
	today := m.now.Format(time.DateOnly)
	for i := first; i < last; i++ {
		r := d.rows[i]
		p := r.Pet
		todayMin := 0.0
		if p.Today == today {
			todayMin = p.TodayMin
		}
		x, loaded := m.w.Users[r.ID] // na memória: o status e a tarefa em andamento
		if loaded {
			todayMin = x.TodayMin
		}
		online, task := m.t("não"), "—"
		if x.Status != rooms.Offline {
			online = m.t("sim")
		}
		if x.Focus.Task != "" { // offline também pode ter tarefa, durante a tolerância; esgotada, 00:00
			task = mmss(x.Focus.Ends.Sub(m.now))
		}
		vals := []string{fmt.Sprint(int(p.EnergyAt(m.now))), fmt.Sprint(int(m.now.Sub(p.LastActive).Hours() / 24)),
			fmt.Sprint(int(todayMin)), fmt.Sprint(int(r.Before)), fmt.Sprint(p.StreakAt(m.now)),
			"██", online, task}
		swatch := pet.Pets.Swatch(p.Look()) // a cor é o bloco pintado: o nome não cabe
		cur := "  "
		if i == d.row {
			cur = m.st.sel.Render("› ")
		}
		nick := r.Nick
		if r.ID == m.id {
			nick += m.t(" (você)")
		}
		line := cur + fmt.Sprintf("%-10s", ansi.Truncate(nick, 10, "…"))
		for j, v := range vals {
			st := lipgloss.NewStyle()
			if i == d.row && j == d.col {
				st = m.st.tabOn.Padding(0)
			}
			if devCols[j].field == rooms.DevColor {
				st = st.Foreground(swatch)
			}
			line += " " + st.Render(fmt.Sprintf("%6s", v))
		}
		var names []string
		for _, code := range m.w.MyRooms(r.ID) {
			if s := m.w.Rooms[code]; s.Owner == r.ID {
				names = append(names, s.Name+"*")
			} else {
				names = append(names, s.Name)
			}
		}
		b.WriteString(line + "  " + m.st.muted.Render(strings.Join(names, ", ")) + "\n")
	}
	switch d.input {
	case "assume":
		if len(d.rows) > 0 {
			b.WriteString("\n" + m.tf("assumir identidade: %s", d.rows[d.row].Nick) + "\n")
		}
	case "nova":
		b.WriteString("\n" + m.t("pessoa nova: ") + m.in.View() + "\n")
	case "tarefa":
		b.WriteString("\n" + m.tf("tarefa de %s: ", d.rows[d.row].Nick) + m.in.View() + "\n")
	}
	// cada linha da legenda cabe em 76 colunas, o miolo de um terminal de 80
	b.WriteString("\n" + m.st.muted.Render(m.tf("hoje: tarefas de %d min · antes: tarefas de %d min, uma por dia para trás",
		rooms.DevTodayStep, rooms.DevBeforeStep)))
	b.WriteString("\n" + m.st.muted.Render(m.tf("energ.: de %d em %d · parado: dias sem focar · online: sessão de mentira",
		rooms.DevEnergyStep, rooms.DevEnergyStep)))
	b.WriteString("\n" + m.st.muted.Render(m.tf("tarefa: + começa uma de %d min ou dá mais %d, - passa %d min",
		rooms.DevFocusMin, rooms.DevFocusStep, rooms.DevFocusStep)))
	b.WriteString("\n" + m.st.muted.Render(m.t("outra pessoa: ssh -p <porta> <apelido>@localhost")))
	return b.String()
}
