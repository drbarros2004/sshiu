package tui

// Menu de salas (dentro da aba "sala"): trocar de sala, criar uma, entrar com código ou sair.
// Quem não está em nenhuma sala cai aqui direto (ADR 0003).

import (
	"fmt"
	"plantao/internal/rooms"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type form int

const (
	formNone  form = iota
	formName       // nome da sala nova
	formSize       // quantas pessoas cabem nela
	formCode       // código para entrar
	formJoin       // a sala do código, antes de entrar
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
		if m.form == formSize {
			switch k {
			case "left":
				m.size = max(rooms.MinMembers, m.size-1)
			case "right":
				m.size = min(rooms.MaxMembers, m.size+1)
			}
			return m, nil
		}
		if m.form != formLeave && m.form != formJoin {
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
			m.form, m.in, m.size = formName, newInput("ex.: Turma de Cálculo", 30), rooms.MaxMembers
		case "#":
			m.form, m.in = formCode, newInput("6 letras", 60)
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

// roomCode: o código no que a pessoa escreveu ou colou. Quem cola o convite inteiro
// ("ssh -t -p 23234 plantao fg7dhm") quer a sala do fim da linha, não uma sala chamada "ssh".
func roomCode(s string) string {
	f := strings.Fields(strings.ToLower(s))
	if len(f) == 0 {
		return ""
	}
	return f[len(f)-1]
}

func (m model) submitForm() (tea.Model, tea.Cmd) {
	v, id := strings.TrimSpace(m.in.Value()), m.id
	switch m.form {
	case formName:
		if v == "" {
			m.err = "dê um nome para a sala"
			return m, nil
		}
		m.form, m.err = formSize, "" // o nome fica no campo até a sala nascer
		return m, nil
	case formSize:
		size := m.size
		return m, func() tea.Msg {
			code, err := rooms.Create(id, v, size)
			if err != nil {
				return errMsg{err}
			}
			return roomMsg{code, true, rooms.SnapshotWorld()}
		}
	case formCode:
		code := roomCode(v)
		if code == "" {
			m.err = "digite o código que o colega te passou"
			return m, nil
		}
		if _, ok := m.w.Rooms[code]; !ok {
			m.err = "nenhuma sala com o código " + code
			return m, nil
		}
		m.form, m.joining, m.err = formJoin, code, ""
		return m, nil
	case formJoin:
		return m, m.join(m.joining)
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
		return bold.Render("criar sala") + "\n\nnome:\n" + m.in.View()
	case formSize:
		hint := "dá para mudar de ideia criando outra sala: o tamanho é fixo depois"
		if m.size == rooms.MinMembers { // ADR 0006
			hint = "sala solo: para focar sozinho, com o seu pet e o modo jardim"
		}
		return bold.Render(strings.TrimSpace(m.in.Value())) + "\n\nquantas pessoas cabem?  " +
			m.st.muted.Render("‹ ") + m.st.title.Render(fmt.Sprintf("%d", m.size)) + m.st.muted.Render(" ›") +
			"\n" + m.st.muted.Render(hint)
	case formCode:
		return bold.Render("entrar com código") + "\n\n" + m.in.View() + "\n" +
			m.st.muted.Render("cole com ctrl+v (cmd+v no mac): o código ou o comando ssh inteiro")
	case formJoin: // a sala tem nome e tamanho: dá para ver onde você está entrando
		s := m.w.Rooms[m.joining]
		return bold.Render("entrar em "+s.Name+"?") + "\n\n" +
			m.st.muted.Render("código "+m.joining+" · "+roomSummary(s))
	case formLeave:
		return bold.Render("sair de "+m.w.Rooms[m.room].Name+"?") + "\n\n" +
			m.st.muted.Render("para voltar, você vai precisar do código de novo. se não sobrar ninguém, a sala some.")
	}
	var b strings.Builder
	if m.room == "" {
		b.WriteString(lipgloss.NewStyle().Width(max(m.width-4, 20)).Render("você ainda não está em nenhuma sala. crie uma e passe o código para quem vai focar com você, ou entre com o código de alguém.") + "\n\n")
	}
	if m.notice != "" { // quem perdeu a vaga ou estava numa sala encerrada cai aqui sem entender
		b.WriteString(m.st.ok.Render(m.notice) + "\n\n")
	}
	b.WriteString(bold.Render("salas") + "\n\n")
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
	return fmt.Sprintf("%d/%d membros · %d focando", len(s.Members), s.Max, n)
}
