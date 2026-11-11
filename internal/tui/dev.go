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
	"github.com/charmbracelet/x/ansi"
)

// Dev: flag -dev (main.go). Liga a aba dev; a identidade pelo nome do SSH fica no main.
var Dev bool

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
	at   time.Time // os Cmd rodam em paralelo: uma leitura velha pode chegar depois da nova
	err  error
}

// As colunas que mudam com +/-, na ordem da tabela.
var devCols = []struct {
	name  string
	field rooms.DevField
}{{"energia", rooms.DevEnergy}, {"parado", rooms.DevIdle}, {"hoje", rooms.DevToday}, {"antes", rooms.DevBefore},
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
		return devMsg{rows, at, err}
	}
}

func (m model) devKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	d := &m.dev
	k := msg.String()
	if d.input != "" {
		switch k {
		case "esc":
			d.input, m.err = "", ""
		case "enter":
			text, input := strings.TrimSpace(m.in.Value()), d.input
			d.input, m.err = "", ""
			if input == "tarefa" {
				id, code := d.rows[d.row].ID, m.room
				if text == "" {
					text = store.DevTask
				}
				return m, devDo(func() error { rooms.DevStart(id, code, text); return nil })
			}
			return m, devDo(func() error { _, err := rooms.DevCreate(text); return err })
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
			m.err = "abra uma sala na aba sala para pôr gente nela"
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
		m.in = newInput("apelido da pessoa nova", 16)
	case "t":
		d.input = "tarefa"
		m.in = newInput("ex.: fazer o exercício 1 da lista 3", 80)
		if x := m.w.Users[d.rows[d.row].ID]; x.Focus.Task != "" {
			m.in.SetValue(x.Focus.Task)
		}
	case "x":
		id := d.rows[d.row].ID
		return m, devDo(func() error { rooms.Finish(id, pet.Done); return nil })
	case "o":
		s, ok := m.w.Rooms[m.room]
		if !ok {
			m.err = "abra uma sala na aba sala para passá-la a alguém"
			return m, nil
		}
		id, code, owner := d.rows[d.row].ID, s.Code, s.Owner
		return m, devDo(func() error { return rooms.Transfer(owner, code, id) })
	}
	return m, nil
}

// setRows troca as linhas sem trocar de pessoa: a lista vai por apelido, e quem nasce no meio dela
// (outra sessão, um ssh com nome novo) empurraria o cursor para a pessoa ao lado no meio de um ajuste.
func (d *devPanel) setRows(rows []store.DevRow, at time.Time) {
	if !at.After(d.at) {
		return
	}
	id := ""
	if d.row < len(d.rows) {
		id = d.rows[d.row].ID
	}
	d.rows, d.at = rows, at
	d.row = max(0, min(d.row, len(rows)-1))
	if i := slices.IndexFunc(rows, func(r store.DevRow) bool { return r.ID == id }); i >= 0 {
		d.row = i
	}
}

func (m model) devHelp() string {
	if m.dev.input != "" {
		return "enter confirma · esc desiste"
	}
	return "setas escolhem · +/- muda · espaço põe ou tira da sala · o passa a sala · t tarefa · x cumpre · n pessoa nova"
}

func (m model) devView() string {
	d := m.dev
	var b strings.Builder
	if s, ok := m.w.Rooms[m.room]; ok {
		b.WriteString("espaço põe ou tira de " + m.st.sel.Render(s.Name) + m.st.muted.Render(" ("+s.Code+")"))
	} else {
		b.WriteString(m.st.muted.Render("abra uma sala na aba sala para pôr gente nela"))
	}
	b.WriteString("\n\n")
	head := fmt.Sprintf("  %-14s", "apelido")
	for _, c := range devCols {
		head += fmt.Sprintf(" %7s", c.name)
	}
	b.WriteString(m.st.muted.Render(head+"   salas") + "\n")
	today := m.now.Format(time.DateOnly)
	for i, r := range d.rows {
		p := r.Pet
		todayMin := 0.0
		if p.Today == today {
			todayMin = p.TodayMin
		}
		x, loaded := m.w.Users[r.ID] // na memória: o status e a tarefa em andamento
		if loaded {
			todayMin = x.TodayMin
		}
		online, task := "não", "—"
		if x.Status != rooms.Offline {
			online = "sim"
		}
		if x.Focus.Task != "" { // offline também pode ter tarefa, durante a tolerância
			task = mmss(x.Focus.Ends.Sub(m.now))
			if x.Focus.Expired(m.now) {
				task = "esgotou"
			}
		}
		vals := []string{fmt.Sprint(int(p.EnergyAt(m.now))), fmt.Sprint(int(m.now.Sub(p.LastActive).Hours() / 24)),
			fmt.Sprint(int(todayMin)), fmt.Sprint(int(r.Before)), fmt.Sprint(p.StreakAt(m.now)),
			pet.Colors[pet.ColorIndex(p.Color)].Name, online, task}
		cur := "  "
		if i == d.row {
			cur = m.st.sel.Render("› ")
		}
		nick := r.Nick
		if r.ID == m.id {
			nick += " (você)"
		}
		line := cur + fmt.Sprintf("%-14s", ansi.Truncate(nick, 14, "…"))
		for j, v := range vals {
			cell := fmt.Sprintf(" %7s", v)
			if i == d.row && j == d.col {
				cell = " " + m.st.tabOn.Padding(0, 0, 0, 1).Render(fmt.Sprintf("%6s", v)) // o número fica na ponta, como nas outras linhas
			}
			line += cell
		}
		var names []string
		for _, code := range m.w.MyRooms(r.ID) {
			if s := m.w.Rooms[code]; s.Owner == r.ID {
				names = append(names, s.Name+"*")
			} else {
				names = append(names, s.Name)
			}
		}
		b.WriteString(line + "   " + m.st.muted.Render(strings.Join(names, ", ")) + "\n")
	}
	switch d.input {
	case "nova":
		b.WriteString("\npessoa nova: " + m.in.View() + "\n")
	case "tarefa":
		b.WriteString("\ntarefa de " + d.rows[d.row].Nick + ": " + m.in.View() + "\n")
	}
	b.WriteString("\n" + m.st.muted.Render(fmt.Sprintf("hoje: tarefas de %d min · antes: tarefas de %d min, uma por dia para trás · energia de %d em %d · parado: dias sem focar",
		rooms.DevTodayStep, rooms.DevBeforeStep, rooms.DevEnergyStep, rooms.DevEnergyStep)))
	b.WriteString("\n" + m.st.muted.Render(fmt.Sprintf("tarefa: + começa uma de %d min ou dá mais %d, - passa %d min · online: uma sessão de mentira · * a sala é da pessoa",
		rooms.DevFocusMin, rooms.DevFocusStep, rooms.DevFocusStep)))
	b.WriteString("\n" + m.st.muted.Render("entre como outra pessoa: ssh -p <porta> <apelido>@localhost"))
	return b.String()
}
