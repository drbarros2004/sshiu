package tui

// Menu do dono da sala, na tecla c (ADR 0007). O que se configura muda na própria página, com ←/→
// na linha escolhida; passar a sala, remover alguém e encerrar abrem uma página própria. Passar e
// encerrar não têm volta pelas mãos de quem fez, então pedem um "sim" antes.

import (
	"fmt"
	"plantao/internal/rooms"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// as linhas do menu, na ordem da tela
const (
	cfgName = iota
	cfgSize
	cfgEntries
	cfgObra
	cfgMeta
	cfgGive
	cfgKick
	cfgEnd
)

// cfgList: as linhas que aparecem nesta sala. A sala solo não tem obra (ADR 0008), e a meta só
// existe depois de haver um tema: configuração que não faz nada não ocupa linha.
func cfgList(s rooms.Snapshot) []int {
	rows := []int{cfgName, cfgSize, cfgEntries}
	if s.Max > rooms.MinMembers {
		rows = append(rows, cfgObra)
		if s.Obra != "" {
			rows = append(rows, cfgMeta)
		}
	}
	return append(rows, cfgGive, cfgKick, cfgEnd)
}

// obraName: o tema como ele aparece no menu.
func obraName(obra string) string {
	if obra == "" {
		return "nenhuma"
	}
	return obra
}

type cfgPage int

const (
	pageMenu cfgPage = iota
	pageGive
	pageKick
	pageEnd
	pageGiveYes // a confirmação de passar a sala para cwho
)

// owns: a sala que você está vendo é sua (ADR 0007). A demo não tem dono.
func (m model) owns() bool {
	s, ok := m.w.Rooms[m.room]
	return ok && s.Owner == m.id
}

// others: os outros membros, que é quem recebe a sala ou perde a vaga.
func (m model) others() []rooms.Member {
	var out []rooms.Member
	for _, x := range orderedMembers(m.id, m.w.Rooms[m.room]) {
		if x.ID != m.id {
			out = append(out, x)
		}
	}
	return out
}

// do roda uma mudança da sala e mostra o erro, se o servidor recusar.
func do(f func() error) tea.Cmd {
	return func() tea.Msg {
		if err := f(); err != nil {
			return errMsg{err}
		}
		return worldMsg(rooms.SnapshotWorld())
	}
}

func (m model) openConfig() (tea.Model, tea.Cmd) {
	m.cfg, m.cpage, m.ccur, m.notice, m.err = true, pageMenu, cfgName, "", ""
	return m, nil
}

func (m model) cfgKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k, id, code := msg.String(), m.id, m.room
	s := m.w.Rooms[code]
	if m.cname { // o nome é digitado na própria linha
		switch k {
		case "enter":
			name := strings.TrimSpace(m.in.Value())
			if name == "" {
				m.err = "a sala precisa de um nome"
				return m, nil
			}
			m.cname, m.err = false, ""
			return m, do(func() error { return rooms.Rename(id, code, name) })
		case "esc":
			m.cname, m.err = false, ""
			return m, nil
		}
		var cmd tea.Cmd
		m.in, cmd = m.in.Update(msg)
		return m, cmd
	}
	switch m.cpage {
	case pageGive, pageKick:
		others := m.others()
		switch k {
		case "up":
			m.mcur = max(0, m.mcur-1)
		case "down":
			m.mcur = min(len(others)-1, m.mcur+1)
		case "esc":
			m.cpage, m.err = pageMenu, ""
		case "enter":
			if len(others) == 0 {
				return m, nil
			}
			who := others[min(m.mcur, len(others)-1)]
			if m.cpage == pageKick {
				m.cpage = pageMenu
				return m, do(func() error { return rooms.Kick(id, code, who.ID) })
			}
			m.cpage, m.cwho, m.cyes = pageGiveYes, who.ID, false
		}
		return m, nil
	case pageGiveYes:
		who, ok := m.member(m.cwho)
		switch k {
		case "left", "right":
			m.cyes = !m.cyes
		case "esc":
			m.cpage, m.cyes = pageGive, false
		case "enter":
			if !m.cyes || !ok { // "não", ou a pessoa saiu da sala enquanto a pergunta estava na tela
				m.cpage, m.cyes = pageGive, false
				return m, nil
			}
			// passar a sala tira os seus poderes: o menu fecha junto
			m.cfg, m.cpage, m.cyes, m.notice = false, pageMenu, false, "a sala agora é de "+who.Nick
			return m, do(func() error { return rooms.Transfer(id, code, who.ID) })
		}
		return m, nil
	case pageEnd:
		switch k {
		case "left", "right":
			m.cyes = !m.cyes
		case "esc":
			m.cpage, m.cyes = pageMenu, false
		case "enter":
			if !m.cyes {
				m.cpage = pageMenu
				return m, nil
			}
			m.cfg, m.cpage, m.room, m.picking = false, pageMenu, "", true
			return m, do(func() error { return rooms.Dissolve(id, code) })
		}
		return m, nil
	}
	list := cfgList(s)
	m.ccur = min(m.ccur, len(list)-1) // a meta some quando a obra desliga: o cursor acompanha
	cur := list[m.ccur]
	switch k {
	case "up":
		m.ccur = max(0, m.ccur-1)
	case "down":
		m.ccur = min(len(list)-1, m.ccur+1)
	case "esc", "c":
		m.cfg, m.err = false, ""
	case "left", "right":
		d := 1
		if k == "left" {
			d = -1
		}
		switch cur {
		case cfgSize:
			n := s.Max + d
			return m, do(func() error { return rooms.Resize(id, code, n) })
		case cfgEntries:
			closed := !s.Closed
			return m, do(func() error { return rooms.SetClosed(id, code, closed) })
		case cfgObra:
			i := max(slices.Index(rooms.Obras, s.Obra), 0) + d
			if i < 0 || i >= len(rooms.Obras) {
				return m, nil
			}
			obra := rooms.Obras[i]
			return m, do(func() error { return rooms.SetObra(id, code, obra) })
		case cfgMeta:
			i := max(slices.Index(rooms.ObraMetas, s.Meta), 0) + d
			if i < 0 || i >= len(rooms.ObraMetas) {
				return m, nil
			}
			meta := rooms.ObraMetas[i]
			return m, do(func() error { return rooms.SetMeta(id, code, meta) })
		}
	case "enter":
		switch cur {
		case cfgName:
			m.cname, m.in, m.err = true, newInput("nome da sala", 30), ""
			m.in.Prompt = "" // o ▸ da linha já diz onde o texto entra
			m.in.SetValue(s.Name)
		case cfgEntries: // o enter também abre e fecha: ninguém adivinha que ali só vale ←/→
			closed := !s.Closed
			return m, do(func() error { return rooms.SetClosed(id, code, closed) })
		case cfgGive:
			m.cpage, m.mcur = pageGive, 0
		case cfgKick:
			m.cpage, m.mcur = pageKick, 0
		case cfgEnd:
			m.cpage, m.cyes = pageEnd, false
		}
	}
	return m, nil
}

func (m model) cfgHelp() string {
	switch {
	case m.cname:
		return "enter confirma · esc cancela"
	case m.cpage == pageGive, m.cpage == pageKick:
		if len(m.others()) == 0 { // não há o que escolher: a página só tem a saída
			return "esc volta"
		}
		if m.cpage == pageKick {
			return "↑/↓ escolhe · enter remove · esc volta"
		}
		return "↑/↓ escolhe · enter passa a sala · esc volta"
	case m.cpage == pageEnd, m.cpage == pageGiveYes:
		return "←/→ escolhe · enter confirma · esc volta"
	}
	return "↑/↓ escolhe · ←/→ muda · enter abre · esc volta"
}

func (m model) cfgView() string {
	s := m.w.Rooms[m.room]
	switch m.cpage {
	case pageGive, pageKick:
		titulo, vazio := "passar a sala para quem?", "você é a única pessoa na sala."
		if m.cpage == pageKick {
			titulo, vazio = "remover quem da sala?", "não há mais ninguém na sala."
		}
		others := m.others()
		if len(others) == 0 {
			return bold.Render(titulo) + "\n\n" + m.st.muted.Render(vazio)
		}
		var b strings.Builder
		b.WriteString(bold.Render(titulo) + "\n\n")
		for i, x := range others {
			b.WriteString(cfgRow(m, i == m.mcur, x.Nick, m.st.muted.Render(m.status(x))) + "\n")
		}
		return strings.TrimRight(b.String(), "\n")
	case pageGiveYes:
		who, _ := m.member(m.cwho)
		aviso := "o menu de configurar passa a ser de " + who.Nick + ". você continua na sala, como os outros."
		return bold.Render("passar "+s.Name+" para "+who.Nick+"?") + "\n\n" + m.st.muted.Render(aviso) +
			"\n\n" + m.yesNo()
	case pageEnd:
		aviso := "a sala some. o que você focou nela fica no seu histórico."
		if n := len(s.Members); n > 1 {
			aviso = fmt.Sprintf("a sala some para os %d membros. o histórico de cada um fica.", n)
		}
		return bold.Render("encerrar "+s.Name+"?") + "\n\n" + m.st.muted.Render(aviso) + "\n\n" + m.yesNo()
	}

	name := m.st.muted.Render(s.Name)
	if m.cname {
		name = m.in.View()
	}
	floor := max(len(s.Members), rooms.MinMembers)
	size := bold.Render(fmt.Sprintf("%d", len(s.Members))) + m.st.muted.Render(" / ") +
		m.arrows(m.ccur == cfgSize, fmt.Sprintf("%d", s.Max), s.Max > floor, s.Max < rooms.MaxMembers)
	list := cfgList(s)
	cur := list[min(m.ccur, len(list)-1)]
	entries := m.arrows(cur == cfgEntries, map[bool]string{true: "fechadas", false: "abertas"}[s.Closed], true, true)
	io := max(slices.Index(rooms.Obras, s.Obra), 0)
	im := max(slices.Index(rooms.ObraMetas, s.Meta), 0)

	label := map[int]string{cfgName: "nome", cfgSize: "tamanho", cfgEntries: "novas entradas",
		cfgObra: "obra", cfgMeta: "meta da obra", cfgGive: "passar a sala",
		cfgKick: "remover alguém", cfgEnd: "encerrar a sala"}
	value := map[int]string{
		cfgName:    name,
		cfgSize:    size,
		cfgEntries: entries,
		cfgObra:    m.arrows(cur == cfgObra, obraName(s.Obra), io > 0, io < len(rooms.Obras)-1),
		cfgMeta:    m.arrows(cur == cfgMeta, dur(float64(s.Meta)), im > 0, im < len(rooms.ObraMetas)-1),
	}
	var rows []string
	for _, r := range list {
		rows = append(rows, cfgRow(m, r == cur, label[r], value[r]))
	}
	return bold.Render("configurar") + m.st.muted.Render(" · "+s.Name) + "\n\n" + strings.Join(rows, "\n")
}

// yesNo: o "‹ não · sim ›" das confirmações, com a escolha em negrito. Começa no "não".
func (m model) yesNo() string {
	sim, nao := "sim", bold.Render("não")
	if m.cyes {
		sim, nao = bold.Render("sim"), "não"
	}
	return m.st.muted.Render("‹ ") + nao + m.st.muted.Render(" · ") + sim + m.st.muted.Render(" ›")
}

// member: o membro id da sala que você está vendo, se ele ainda estiver nela.
func (m model) member(id string) (rooms.Member, bool) {
	i := slices.IndexFunc(m.others(), func(x rooms.Member) bool { return x.ID == id })
	if i < 0 {
		return rooms.Member{}, false
	}
	return m.others()[i], true
}

// cfgRow: uma linha da lista, com o rótulo numa coluna fixa e o valor à direita. As linhas que
// abrem outra página não têm valor: o espaço vazio é que diz que ali se entra.
func cfgRow(m model, sel bool, label, value string) string {
	mark, st := "  ", lipgloss.NewStyle().Width(18)
	if sel {
		mark, st = m.st.sel.Render("▸ "), st.Bold(true)
	}
	return mark + st.Render(label) + value
}

// arrows: o valor com as setas, desenhadas só na linha escolhida. A seta que não anda mais some,
// para o fim do caminho aparecer sem precisar de explicação.
func (m model) arrows(sel bool, value string, left, right bool) string {
	if !sel {
		return m.st.muted.Render(value)
	}
	l, r := m.st.muted.Render("‹"), m.st.muted.Render("›")
	if !left {
		l = " "
	}
	if !right {
		r = " "
	}
	return l + " " + m.st.title.Render(value) + " " + r
}

// stillMine: a sala que você está vendo continua sua? Serve para o aviso de quem perdeu a vaga ou
// estava numa sala encerrada (ADR 0007).
func (m model) stillMine() bool {
	s, ok := m.w.Rooms[m.room]
	return ok && slices.ContainsFunc(s.Members, func(x rooms.Member) bool { return x.ID == m.id })
}
