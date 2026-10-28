package tui

// Seletor de salas (dentro da aba "sala"): trocar de sala, criar uma privada ou entrar com código.

import (
	"fmt"
	"plantao/internal/rooms"
	"plantao/internal/store"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type form int

const (
	formNone  form = iota
	formName       // nome da sala nova
	formSlots      // horários da sala nova
	formCode       // código para entrar
)

func (m model) pickerItems() []store.RoomRow {
	items := append([]store.RoomRow{{Code: rooms.PublicCode, Name: "sala pública"}}, m.myRooms...)
	return append(items, store.RoomRow{Code: "+", Name: "criar sala"}, store.RoomRow{Code: "#", Name: "entrar com código"})
}

func (m model) openPicker() (tea.Model, tea.Cmd) {
	m.picking, m.notice, m.err = true, "", ""
	id := m.id
	return m, func() tea.Msg {
		rs, err := store.MyRooms(id)
		if err != nil {
			return errMsg{err}
		}
		return roomsMsg(rs)
	}
}

func (m model) pickerKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if m.form != formNone {
		switch k {
		case "esc":
			m.form, m.err = formNone, ""
			return m, nil
		case "enter":
			return m.submitForm()
		}
		var cmd tea.Cmd
		m.in, cmd = m.in.Update(msg)
		return m, cmd
	}
	items := m.pickerItems()
	switch k {
	case "up", "k":
		m.pcur = max(m.pcur-1, 0)
	case "down", "j":
		m.pcur = min(m.pcur+1, len(items)-1)
	case "esc", "s":
		m.picking = false
	case "enter":
		switch it := items[min(m.pcur, len(items)-1)]; it.Code {
		case "+":
			m.form, m.in = formName, newInput("ex.: Turma de Cálculo", 30)
		case "#":
			m.form, m.in = formCode, newInput("6 letras", 6)
		default:
			m.room, m.picking = it.Code, false
			m.setWorld(m.w)
		}
	}
	return m, nil
}

func (m model) submitForm() (tea.Model, tea.Cmd) {
	v, id := strings.TrimSpace(m.in.Value()), m.id
	switch m.form {
	case formName:
		if v == "" {
			m.err = "dê um nome para a sala"
			return m, nil
		}
		m.draft, m.form, m.err = v, formSlots, ""
		m.in = newInput("9,14,20,22 (enter usa esses)", 40)
		return m, nil
	case formSlots:
		if v == "" {
			v = "9,14,20,22"
		}
		slots, err := store.ParseSlots(v)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		name := m.draft
		return m, func() tea.Msg {
			r, err := store.CreateRoom(id, name, slots)
			if err != nil {
				return errMsg{err}
			}
			rooms.Open(r.Code, r.Name, r.Slots)
			return roomMsg{r, true, rooms.SnapshotWorld()}
		}
	case formCode:
		code := strings.ToLower(v)
		return m, func() tea.Msg {
			r, err := store.JoinRoom(id, code)
			if err != nil {
				return errMsg{err}
			}
			rooms.Open(r.Code, r.Name, r.Slots)
			return roomMsg{r, false, rooms.SnapshotWorld()}
		}
	}
	return m, nil
}

func (m model) pickerView() string {
	var b strings.Builder
	switch m.form {
	case formName:
		return bold.Render("Criar sala") + "\n\nNome:\n" + m.in.View()
	case formSlots:
		return bold.Render("Criar sala: "+m.draft) + "\n\nEm que horas começam os sprints? " +
			m.st.muted.Render("(25 min de foco + 5 de intervalo; na demo o relógio é comprimido)") + "\n" + m.in.View()
	case formCode:
		return bold.Render("Entrar com código") + "\n\n" + m.in.View()
	}
	b.WriteString(bold.Render("Salas") + "\n\n")
	for i, it := range m.pickerItems() {
		cursor, name := "  ", fmt.Sprintf("%-24s", it.Name)
		if it.Code == "+" || it.Code == "#" {
			name = fmt.Sprintf("%-24s", map[string]string{"+": "+ ", "#": "# "}[it.Code]+it.Name)
		}
		if i == m.pcur {
			cursor, name = m.st.sel.Render("▸ "), m.st.sel.Render(name)
		}
		line := cursor + name
		if s, ok := m.w.Rooms[it.Code]; ok {
			line += m.st.muted.Render(m.roomStatus(s))
			if it.Code != rooms.PublicCode {
				line += m.st.muted.Render(" · código " + it.Code)
			}
			if it.Code == m.room {
				line += m.st.ok.Render("  ← você está aqui")
			}
		}
		b.WriteString(line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// roomStatus resume uma sala numa linha: fase, tempo e quantas pessoas.
func (m model) roomStatus(s rooms.Snapshot) string {
	switch s.Phase {
	case rooms.Focus:
		return fmt.Sprintf("foco · %s · %d na sala", mmss(s.Ends.Sub(m.now)), len(s.Room))
	case rooms.Break:
		return fmt.Sprintf("intervalo · %s · %d na sala", mmss(s.Ends.Sub(m.now)), len(s.Room))
	}
	return fmt.Sprintf("próximo %s · %d na fila", s.Next.Format("15:04"), len(s.Queue))
}
