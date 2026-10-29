package tui

// Menu de salas (dentro da aba "sala"): trocar de sala, criar uma, entrar com código ou sair.
// Quem não está em nenhuma sala cai aqui direto (ADR 0003).

import (
	"fmt"
	"plantao/internal/rooms"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type form int

const (
	formNone  form = iota
	formName       // nome da sala nova
	formCode       // código para entrar
	formLeave      // confirmando a saída
)

type pickerItem struct{ code, name string } // code "+": criar, "#": entrar com código, "-": sair

func (m model) pickerItems() []pickerItem {
	var items []pickerItem
	mine := m.w.MyRooms(m.id)
	for _, code := range mine {
		items = append(items, pickerItem{code, m.w.Rooms[code].Name})
	}
	if rooms.Demo && !slices.Contains(mine, rooms.DemoCode) {
		items = append(items, pickerItem{rooms.DemoCode, "sala demo"})
	}
	items = append(items, pickerItem{"+", "criar sala"}, pickerItem{"#", "entrar com código"})
	if m.room != "" && m.room != rooms.DemoCode {
		items = append(items, pickerItem{"-", "sair de " + m.w.Rooms[m.room].Name})
	}
	return items
}

func (m model) openPicker() (tea.Model, tea.Cmd) {
	m.picking, m.notice, m.err, m.pcur = true, "", "", 0
	return m, nil
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
		if m.form != formLeave {
			var cmd tea.Cmd
			m.in, cmd = m.in.Update(msg)
			return m, cmd
		}
		return m, nil
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
		switch it := items[min(m.pcur, len(items)-1)]; it.code {
		case "+":
			m.form, m.in = formName, newInput("ex.: Turma de Cálculo", 30)
		case "#":
			m.form, m.in = formCode, newInput("6 letras", 6)
		case "-":
			m.form = formLeave
		case rooms.DemoCode:
			return m, m.join(rooms.DemoCode)
		default:
			m.room, m.picking = it.code, false
		}
	}
	return m, nil
}

func (m model) join(code string) tea.Cmd {
	id := m.id
	return func() tea.Msg {
		if err := rooms.Join(id, code); err != nil {
			return errMsg{err}
		}
		return roomMsg{code, false, rooms.SnapshotWorld()}
	}
}

func (m model) submitForm() (tea.Model, tea.Cmd) {
	v, id := strings.TrimSpace(m.in.Value()), m.id
	switch m.form {
	case formName:
		if v == "" {
			m.err = "dê um nome para a sala"
			return m, nil
		}
		return m, func() tea.Msg {
			code, err := rooms.Create(id, v)
			if err != nil {
				return errMsg{err}
			}
			return roomMsg{code, true, rooms.SnapshotWorld()}
		}
	case formCode:
		return m, m.join(strings.ToLower(v))
	case formLeave:
		code := m.room
		m.form, m.picking, m.room = formNone, false, ""
		return m, func() tea.Msg {
			if err := rooms.Leave(id, code); err != nil {
				return errMsg{err}
			}
			return worldMsg(rooms.SnapshotWorld())
		}
	}
	return m, nil
}

func (m model) pickerView() string {
	switch m.form {
	case formName:
		return bold.Render("Criar sala") + "\n\nNome:\n" + m.in.View()
	case formCode:
		return bold.Render("Entrar com código") + "\n\n" + m.in.View()
	case formLeave:
		return bold.Render("Sair de "+m.w.Rooms[m.room].Name+"?") + "\n\n" +
			m.st.muted.Render("Para voltar, você vai precisar do código de novo. Se você for o último, a sala some.")
	}
	var b strings.Builder
	if m.room == "" {
		b.WriteString("Você ainda não está em nenhuma sala. Crie uma e passe o código para os colegas, ou entre com o código de alguém.\n\n")
	}
	b.WriteString(bold.Render("Salas") + "\n\n")
	for i, it := range m.pickerItems() {
		cursor, name := "  ", it.name
		switch it.code {
		case "+", "#", "-":
			name = it.code + " " + name
		}
		name = fmt.Sprintf("%-28s", name)
		if i == m.pcur {
			cursor, name = m.st.sel.Render("▸ "), m.st.sel.Render(name)
		}
		line := cursor + name
		if s, ok := m.w.Rooms[it.code]; ok {
			line += m.st.muted.Render(roomSummary(s))
			if it.code == m.room {
				line += m.st.ok.Render("  ← você está aqui")
			}
		}
		b.WriteString(line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// roomSummary resume uma sala numa linha: quantos estão focando agora.
func roomSummary(s rooms.Snapshot) string {
	n := 0
	for _, x := range s.Members {
		if x.Status == rooms.Focusing {
			n++
		}
	}
	members := fmt.Sprintf("%d membros", len(s.Members))
	if len(s.Members) == 1 {
		members = "1 membro"
	}
	return fmt.Sprintf("%s · %d focando", members, n)
}
