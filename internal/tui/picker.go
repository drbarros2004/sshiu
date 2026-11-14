package tui

// Menu de salas (dentro da aba "sala"): trocar de sala, criar uma, entrar com código ou sair.
// Quem não está em nenhuma sala cai aqui direto (ADR 0003).

import (
	"fmt"
	"slices"
	"sshiu/internal/rooms"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type form int

const (
	formNone  form = iota
	formNew        // a sala nova inteira: nome, tamanho e obra
	formCode       // código para entrar
	formJoin       // a sala do código, antes de entrar
	formLeave      // confirmando a saída
)

type pickerItem struct{ code, name string } // code "+": criar, "#": entrar com código, "-": sair

func (m model) pickerItems() []pickerItem {
	var items []pickerItem
	mine := m.w.MyRooms(m.id)
	for _, code := range mine {
		items = append(items, pickerItem{code, m.roomName(m.w.Rooms[code])})
	}
	if rooms.Demo && !slices.Contains(mine, rooms.DemoCode) {
		items = append(items, pickerItem{rooms.DemoCode, m.t("sala demo")})
	}
	items = append(items, pickerItem{"+", m.t("criar sala")}, pickerItem{"#", m.t("entrar com código")})
	if m.room != "" && m.room != rooms.DemoCode {
		items = append(items, pickerItem{"-", m.tf("sair de %s", m.w.Rooms[m.room].Name)})
	}
	return items
}

// as linhas da tela de criar sala, na ordem
const (
	newName = iota
	newSize
	newObra
)

// newList: as linhas que aparecem. A sala solo não tem obra (ADR 0008), então a linha dela só
// existe quando a sala comporta mais de uma pessoa.
func newList(size int) []int {
	if size == rooms.MinMembers {
		return []int{newName, newSize}
	}
	return []int{newName, newSize, newObra}
}

func (m model) openPicker() (tea.Model, tea.Cmd) {
	// a tela da tecla o fica para trás: a sala que você escolher abre na tela da sala
	m.picking, m.notice, m.err, m.pcur, m.obraFull = true, "", "", 0, false
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
		if m.form == formNew {
			list := newList(m.size)
			m.ncur = min(m.ncur, len(list)-1)
			switch k {
			case "up":
				m.ncur = max(0, m.ncur-1)
				return m, nil
			case "down":
				m.ncur = min(len(list)-1, m.ncur+1)
				return m, nil
			case "left", "right":
				d := 1
				if k == "left" {
					d = -1
				}
				switch list[m.ncur] {
				case newSize:
					m.size = min(max(m.size+d, rooms.MinMembers), rooms.MaxMembers)
					return m, nil
				case newObra:
					m.nobra = min(max(m.nobra+d, 0), len(rooms.Obras)-1)
					return m, nil
				}
			}
			if list[m.ncur] != newName { // fora da linha do nome, o que se digita não é texto
				return m, nil
			}
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
	case "down":
		m.pcur = min(m.pcur+1, len(items)-1)
	case "esc", "s":
		m.picking = false
	case "enter":
		switch it := items[min(m.pcur, len(items)-1)]; it.code {
		case "+":
			m.form, m.in, m.size = formNew, newInput(m.t("ex.: turma de cálculo"), 30), rooms.MaxMembers
			// a obra já vem escolhida: a sala nasce construindo em vez de nascer vazia
			m.ncur, m.nobra = newName, max(slices.Index(rooms.Obras, "cidade"), 0)
		case "#":
			m.form, m.in = formCode, newInput(m.t("6 letras"), 60)
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
// ("ssh -t -p 23234 sshiu fg7dhm") quer a sala do fim da linha, não uma sala chamada "ssh".
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
	case formNew:
		if v == "" {
			m.err = m.t("dê um nome para a sala")
			return m, nil
		}
		size, obra := m.size, ""
		if size > rooms.MinMembers {
			obra = rooms.Obras[m.nobra]
		}
		m.form = formNone // um enter a mais enquanto a sala nasce criava outra sala igual
		return m, func() tea.Msg {
			code, err := rooms.Create(id, v, size)
			if err != nil {
				return errMsg{err}
			}
			if obra != "" { // a sala nasce com a obra já ligada (ADR 0008)
				if err := rooms.SetObra(id, code, obra); err != nil {
					return errMsg{err}
				}
			}
			return roomMsg{code, true, rooms.SnapshotWorld()}
		}
	case formCode:
		code := roomCode(v)
		if code == "" {
			m.err = m.t("digite o código da sala")
			return m, nil
		}
		if _, ok := m.w.Rooms[code]; !ok {
			m.err = m.tf("nenhuma sala com o código %s", code)
			return m, nil
		}
		m.form, m.joining, m.err = formJoin, code, ""
		return m, nil
	case formJoin:
		if m.joinBlocked(m.w.Rooms[m.joining]) != "" { // a prévia já diz por quê: o enter não repete
			return m, nil
		}
		return m, m.join(m.joining)
	case formLeave:
		code := m.room
		m.form, m.picking, m.room, m.left = formNone, false, "", code
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
	case formNew:
		list := newList(m.size)
		cur := list[min(m.ncur, len(list)-1)]
		name := m.in.View()
		if cur != newName { // fora da linha do nome, o campo vira só o texto, sem cursor
			name = m.st.muted.Render(strings.TrimSpace(m.in.Value()))
		}
		hint := m.t("depois, o dono muda tudo isso em c") // o tamanho também (ADR 0007)
		if m.size == rooms.MinMembers {                   // ADR 0006
			hint = m.t("sala solo: para focar sem companhia, com o seu pet e o modo jardim")
		}
		value := map[int]string{
			newName: name,
			newSize: m.arrows(cur == newSize, fmt.Sprintf("%d", m.size),
				m.size > rooms.MinMembers, m.size < rooms.MaxMembers),
			newObra: m.arrows(cur == newObra, obraName(m.lang, rooms.Obras[m.nobra]),
				m.nobra > 0, m.nobra < len(rooms.Obras)-1),
		}
		label := map[int]string{newName: m.t("nome"), newSize: m.t("quantos membros"), newObra: m.t("obra")}
		var rows []string
		for _, r := range list {
			rows = append(rows, cfgRow(m, r == cur, label[r], value[r]))
		}
		return bold.Render(m.t("criar sala")) + "\n\n" + strings.Join(rows, "\n") +
			"\n\n" + m.st.muted.Render(hint)
	case formCode:
		return bold.Render(m.t("entrar com código")) + "\n\n" + m.in.View() + "\n" +
			m.st.muted.Render(m.t("cole com ctrl+v (cmd+v no mac): o código ou o comando ssh inteiro"))
	case formJoin: // a sala tem nome e tamanho: dá para ver onde você está entrando
		s := m.w.Rooms[m.joining]
		out := bold.Render(m.tf("entrar em %s?", m.roomName(s))) + "\n\n" +
			m.st.muted.Render(m.tf("código %s", m.joining)+" · "+m.roomSummary(s))
		if slices.ContainsFunc(s.Members, func(x rooms.Member) bool { return x.ID == m.id }) {
			out += "\n\n" + m.st.muted.Render(m.t("você já está nesta sala."))
		}
		if why := m.joinBlocked(s); why != "" {
			out += "\n\n" + m.st.bad.Render(why)
		}
		return out
	case formLeave:
		s := m.w.Rooms[m.room]
		aviso := m.t("para voltar, você vai precisar do código de novo. se não sobrar ninguém, a sala some.")
		// a sucessão do ADR 0007: a sala vai para quem entrou primeiro, que é a ordem dos membros
		if heir := slices.IndexFunc(s.Members, func(x rooms.Member) bool { return x.ID != m.id }); s.Owner == m.id && heir >= 0 {
			aviso += " " + m.tf("a sala passa para %s, que está nela há mais tempo.", s.Members[heir].Nick)
		}
		return bold.Render(m.tf("sair de %s?", s.Name)) + "\n\n" + m.st.muted.Width(max(m.width-4, 20)).Render(aviso)
	}
	var b strings.Builder
	if m.room == "" {
		b.WriteString(lipgloss.NewStyle().Width(max(m.width-4, 20)).Render(m.t("você ainda não está em nenhuma sala. crie uma e passe o código para quem vai focar com você, ou entre com o código de alguém.")) + "\n\n")
	}
	b.WriteString(bold.Render(m.t("salas")) + "\n\n")
	items := m.pickerItems()
	col := 28 // a coluna do resumo: depois do nome mais largo, contando a largura na tela (日 vale 2)
	for _, it := range items {
		if _, ok := m.w.Rooms[it.code]; ok {
			col = max(col, ansi.StringWidth(it.name)+2)
		}
	}
	for i, it := range items {
		cursor, name := "  ", it.name
		switch it.code {
		case "+", "#", "-":
			name = it.code + " " + name
		}
		name += strings.Repeat(" ", max(0, col-ansi.StringWidth(name)))
		if i == m.pcur {
			cursor, name = m.st.sel.Render("▸ "), m.st.sel.Render(name)
		}
		line := cursor + name
		if s, ok := m.w.Rooms[it.code]; ok {
			line += m.st.muted.Render(m.roomSummary(s))
			if it.code == m.room {
				line += m.st.ok.Render(m.t("  ← você está aqui"))
			}
		}
		b.WriteString(line + "\n")
	}
	if m.notice != "" { // quem perdeu a vaga ou estava numa sala encerrada cai aqui sem entender
		b.WriteString("\n" + m.noticeStyle().Render(m.notice)) // embaixo: quando ele some, a lista não pula
	}
	return strings.TrimRight(b.String(), "\n")
}

// joinBlocked: por que o Join recusaria a sala, dito na prévia antes do enter ("" = dá para entrar).
func (m model) joinBlocked(s rooms.Snapshot) string {
	switch {
	case len(s.Members) >= s.Max:
		return m.t("a sala está cheia.")
	case s.Closed:
		return m.t("a sala não está aceitando ninguém agora.")
	}
	return ""
}

// roomSummary resume uma sala numa linha: quantos estão focando agora. A sala solo é de uma
// pessoa, então "1/1 membros" não diz nada: ela diz só se você está focando.
func (m model) roomSummary(s rooms.Snapshot) string {
	n := 0
	for _, x := range s.Members {
		if x.Status == rooms.Focusing {
			n++
		}
	}
	if s.Max == rooms.MinMembers {
		if n > 0 {
			return m.t("sala solo · focando")
		}
		return m.t("sala solo")
	}
	return m.tf("%d/%d membros · %d focando", len(s.Members), s.Max, n)
}
