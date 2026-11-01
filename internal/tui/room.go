package tui

// Aba "sala": o seu foco em cima, um card por membro e os eventos com as reações embaixo
// (o layout A do protótipo, ver IDEIAS.md).

import (
	"fmt"
	"plantao/internal/pet"
	"plantao/internal/rooms"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var reactionGlyphs = [rooms.Reactions]string{"👏", "🔥", "💪", "☕"}

const (
	shownEvents = 6
	panelCols   = 54 // largura do painel da sala solo
	minGarden   = 36 // menos que isso não dá jardim: a tela fica só com o painel
	soloEvents  = 3  // linhas do mini log da sala solo
	clockCols   = 20 // o relógio grande: 5 glifos de 4 colunas
)

// solo: a sala é de uma pessoa (ADR 0006), então a tela não tem cards, reações nem código.
func (m model) solo() bool { return m.w.Rooms[m.room].Max == rooms.MinMembers }

// soloGardenWidth: o que sobra para o jardim ao lado do painel (0 = não cabe).
func soloGardenWidth(inner int) int {
	if w := inner - panelCols - 2; w >= minGarden {
		return w
	}
	return 0
}

// soloRows: a altura das colunas da sala solo, fora o cabeçalho, o título e a ajuda.
func soloRows(height int) int { return max(12, height-9) }

// minuteChoices: os tempos declarados que dá para escolher (na demo, 1 min para a plateia ver o fluxo).
func minuteChoices() []int {
	if rooms.Demo {
		return []int{1, 10, 15, 25, 45, 60}
	}
	return []int{10, 15, 25, 45, 60}
}

// events: os eventos navegáveis. A sala solo não tem nenhum (nem cursor, nem reações); o que
// você fez nela aparece no mini log.
func (m model) events() []rooms.Event {
	if m.solo() {
		return nil
	}
	ev := m.w.Rooms[m.room].Events
	return ev[:min(len(ev), shownEvents)]
}

func (m model) selIndex() int {
	return max(0, slices.IndexFunc(m.events(), func(e rooms.Event) bool { return e.ID == m.sel }))
}

// inPicker: o menu de salas está na tela, seja porque você abriu com s, seja porque você ainda
// não está em nenhuma sala (o primeiro acesso cai nele).
func (m model) inPicker() bool { return m.picking || m.room == "" }

func (m model) roomKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.inPicker() {
		return m.pickerKey(msg)
	}
	k, id, code := msg.String(), m.id, m.room
	m.notice = "" // o aviso de "entrou na sala" some na primeira tecla
	me := m.me()
	focusing := me.Status == rooms.Focusing
	expired := focusing && me.Focus.Expired(m.now)
	var cmd tea.Cmd
	switch {
	case m.entering:
		switch k {
		case "enter":
			if strings.TrimSpace(m.in.Value()) == "" {
				m.err = "escreva a próxima coisa concreta que você vai fazer"
				return m, nil
			}
			m.entering, m.choosing, m.cont, m.err = false, true, false, ""
		case "esc":
			m.entering, m.err = false, ""
		default:
			m.in, cmd = m.in.Update(msg)
		}
		return m, cmd
	case m.choosing:
		choices := minuteChoices()
		m.mins = min(m.mins, len(choices)-1)
		switch k {
		case "left":
			m.mins = max(0, m.mins-1)
		case "right":
			m.mins = min(len(choices)-1, m.mins+1)
		case "esc":
			m.choosing = false
		case "enter":
			m.choosing = false
			task, mins := strings.TrimSpace(m.in.Value()), choices[m.mins]
			if m.cont {
				return m, act(func() { rooms.Continue(id, mins) })
			}
			return m, act(func() { rooms.Start(id, code, task, mins) })
		}
		return m, nil
	}
	switch k {
	case "up", "down":
		ev := m.events()
		if len(ev) == 0 {
			return m, nil
		}
		i := m.selIndex()
		if k == "up" {
			i = max(0, i-1)
		} else {
			i = min(len(ev)-1, i+1)
		}
		m.sel = ev[i].ID
	case "1", "2", "3", "4":
		if ev := m.events(); len(ev) > 0 {
			eid, r := ev[m.selIndex()].ID, int(k[0]-'1')
			return m, act(func() { rooms.React(id, code, eid, r) })
		}
	case "s":
		return m.openPicker()
	case "enter":
		if !focusing {
			m.entering, m.in = true, newInput("ex.: fazer o exercício 1 da lista 3", 80)
		}
	case "f":
		if focusing { // em sequência: o loadProfile só vale depois de a tarefa estar gravada
			return m, tea.Sequence(act(func() { rooms.Finish(id, pet.Done) }), m.loadProfile())
		}
	case "t": // encerra como não cumprida, sem evento: ninguém é exposto por trocar
		if focusing {
			m.entering, m.in = true, newInput("ex.: fazer o exercício 1 da lista 3", 80)
			return m, tea.Sequence(act(func() { rooms.Finish(id, pet.NotDone) }), m.loadProfile())
		}
	case "p":
		if focusing {
			return m, tea.Sequence(act(func() { rooms.Finish(id, pet.NotDone) }), m.loadProfile())
		}
	case "c":
		if expired {
			m.choosing, m.cont = true, true
		}
	case "y": // o servidor não sabe se o terminal aceitou o OSC 52: o aviso diz só que mandou
		if !m.solo() && code != rooms.DemoCode { // as duas salas sem código para passar adiante
			m.notice = "código enviado para a área de transferência"
			return m, tea.SetClipboard(code)
		}
	case "j": // na sala solo o jardim é a outra versão da tela, e abre a qualquer momento
		if m.solo() || focusing && !expired {
			return m.openGarden(false)
		}
	}
	return m, nil
}

func (m model) roomHelp() string {
	me := m.me()
	focusing := me.Status == rooms.Focusing
	switch {
	case m.inPicker() && m.form == formSize:
		return "←/→ escolhe · enter cria · esc volta"
	case m.inPicker() && m.form == formJoin:
		return "enter entra · esc volta"
	case m.inPicker() && m.form != formNone:
		return "enter confirma · esc volta"
	case m.inPicker():
		return "↑/↓ escolhe · enter abre" + map[bool]string{true: " · esc volta", false: ""}[m.room != ""]
	case m.entering:
		return "enter confirma · esc volta"
	case m.choosing:
		return "←/→ escolhe · enter começa · esc volta"
	case focusing && me.Focus.Expired(m.now): // as escolhas já estão em destaque no painel
		return "s salas" + m.reactHelp()
	case focusing:
		return "f feito · t trocar · p parar · j modo jardim" + m.reactHelp() + " · s salas"
	}
	if m.solo() { // na sala solo o jardim abre sem tarefa: é a outra versão da tela
		return "enter começa uma tarefa · j modo jardim · s salas"
	}
	return "enter começa uma tarefa" + m.reactHelp() + " · s salas"
}

// reactHelp: as teclas de evento e reação, que a sala solo não tem.
func (m model) reactHelp() string {
	if m.solo() {
		return ""
	}
	return " · ↑/↓ evento · 1-4 reage"
}

// roomView: a sala em w×h células. Quando falta altura, a lista de eventos encurta e depois os
// pets viram um ponto da cor do humor.
// ponytail: com muita gente num terminal baixo os cards ainda passam da altura e são cortados;
// rolagem nos cards se isso aparecer de verdade.
func (m model) roomView(w, h int) string {
	if m.inPicker() {
		return m.pickerView()
	}
	s := m.w.Rooms[m.room]
	title := bold.Render(s.Name)
	if m.room != rooms.DemoCode && !m.solo() { // a sala solo não tem código para passar
		title += m.st.muted.Render(" · código " + m.room)
		if !m.entering && !m.choosing { // digitando a tarefa, o y é texto
			title += m.st.muted.Render(" · y copia")
		}
	}
	if !m.solo() { // quanto a sala comporta, embaixo do nome; a solo é sempre uma pessoa
		title += "\n" + m.st.muted.Render(fmt.Sprintf("%d/%d pessoas", len(s.Members), s.Max))
	}
	if m.notice != "" {
		title += "\n" + m.st.ok.Render(m.notice)
	}
	if m.solo() { // a sala solo tem uma pessoa e nenhum card: ela já cabe
		return title + "\n\n" + m.soloView()
	}
	alone := len(s.Members) == 1 && m.room != rooms.DemoCode // sozinho numa sala de mais
	var out string
	for _, compact := range []bool{false, true} {
		for n := shownEvents; n >= 0 && (compact || n > 0); n -= 2 { // sem eventos só com os pets já encolhidos
			out = title + "\n\n" + m.focusPanel() + "\n\n" + m.cards(s, w/18, compact)
			switch {
			case alone: // o convite vale mais que o feed vazio, e não encolhe
				out += "\n\n" + m.invitePanel(s)
			case n > 0:
				out += "\n\n" + m.feed(n) + "\n" + m.legend()
			}
			if lipgloss.Height(out) <= h {
				return out
			}
		}
	}
	return out
}

// invitePanel: a sala recém-criada é a tela mais vazia do app, e o que falta nela é gente.
// Enquanto você é o único membro, o convite ocupa o lugar do feed, com os dois caminhos:
// quem já usa o app entra pelo menu de salas, e quem está chegando cola o comando no terminal.
func (m model) invitePanel(s rooms.Snapshot) string {
	row := func(label, text string) string {
		return m.st.muted.Render(fmt.Sprintf("  %-13s", label)) + text
	}
	return bold.Render("você é o primeiro aqui.") + m.st.muted.Render(" para a turma entrar na sala:") + "\n\n" +
		row("no Plantão", m.st.muted.Render("aperta ")+"s"+m.st.muted.Render(", escolhe ")+"entrar com código"+
			m.st.muted.Render(" e digita o código da sala")) + "\n" +
		row("no terminal", m.st.title.Render(m.invite+" "+s.Code))
}

// soloView: o painel do foco à esquerda e o modo jardim à direita (ADR 0006). O jardim anda só
// durante o foco; parado, a coluna fica quieta. Em terminal estreito ele não cabe e sobra o painel,
// com o pet ao lado das informações.
func (m model) soloView() string {
	// o pet com quem você é ao lado, na altura dele: embaixo o nome ficava solto, longe do desenho
	// (o sprite é centralizado numa caixa de 16, então ele nunca encosta na margem).
	who := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(20).Render(pet.Render(pet.MoodOf(m.me().Energy), m.tick, m.pixel)),
		m.soloWho())
	panel := strings.Join([]string{m.focusPanel(), "", who, "", m.soloLog()}, "\n")
	if soloGardenWidth(max(m.width-4, 40)) == 0 {
		return panel // terminal estreito: não cabe coluna nenhuma
	}
	// a direita muda com o estado: focando é o jardim, parado são os seus números.
	right := m.statsBlock(true)
	if m.side != nil {
		right = m.side.frame
	}
	// a barra separa as colunas pela altura da mais alta: o jardim e os números têm alturas diferentes
	bar := m.st.muted.Render(strings.TrimRight(strings.Repeat("│\n", max(lipgloss.Height(panel), lipgloss.Height(right))), "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(panelCols).Render(panel), bar, " ", right)
}

// soloWho: as informações ao lado do pet, em colunas, como a saída de um comando: o rótulo
// apagado e o valor na cor do texto. O streak e os minutos vêm do banco, então só aparecem
// depois da primeira tarefa encerrada no dia.
func (m model) soloWho() string {
	rows := [][2]string{{"humor", pet.MoodName[pet.MoodOf(m.me().Energy)]}}
	if p := m.pet; p.Today == m.now.Format("2006-01-02") {
		if n := p.StreakAt(m.now); n > 0 {
			rows = append(rows, [2]string{"streak", fmt.Sprintf("%d %s", n, plural(n, "dia seguido", "dias seguidos"))})
		}
		rows = append(rows, [2]string{"hoje", fmt.Sprintf("%d min focados", int(p.TodayMin))})
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(m.st.muted.Render(fmt.Sprintf("%-8s", r[0])) + r[1] + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func plural(n int, um, varios string) string {
	if n == 1 {
		return um
	}
	return varios
}

// soloLog: as últimas coisas que você fez nesta sala, sem reações e sem cursor.
func (m model) soloLog() string {
	ev := m.w.Rooms[m.room].Events
	if len(ev) == 0 {
		return m.st.muted.Render("a primeira tarefa que você começar aparece aqui")
	}
	var b strings.Builder
	for _, e := range ev[:min(len(ev), soloEvents)] {
		verb := e.Verb
		if e.Verb == "cumpriu" {
			verb = m.st.ok.Render(verb)
		}
		b.WriteString(m.st.muted.Render(e.At.Format("15:04")) + "  " + verb + " \"" + e.Task + "\"\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// focusPanel: o seu foco, conforme o momento.
func (m model) focusPanel() string {
	me := m.me()
	f := me.Focus
	choice := func(label string) string {
		return label + "  " + m.st.muted.Render("‹ ") + m.st.title.Render(fmt.Sprintf("%d min", minuteChoices()[min(m.mins, len(minuteChoices())-1)])) + m.st.muted.Render(" ›")
	}
	switch {
	case m.entering:
		return bold.Render("qual a próxima coisa concreta que você vai fazer?") + "\n" + m.in.View()
	case m.choosing && m.cont:
		return bold.Render(f.Task) + "\n" + choice("mais quanto tempo?")
	case m.choosing:
		return bold.Render(strings.TrimSpace(m.in.Value())) + "\n" + choice("quanto tempo?")
	case me.Status != rooms.Focusing:
		return bold.Render("nenhuma tarefa em andamento.") + "\n" + m.st.muted.Render("enter começa uma tarefa")
	}
	// A tarefa e os números ficam no topo do relógio, nunca ao lado dele por inteiro: o painel da
	// sala solo tem 54 colunas, e texto comprido ao lado do relógio quebrava a tela em pedaços.
	side := max(20, m.panelWidth()-clockCols-3)
	lines := []string{lipgloss.NewStyle().Width(side).MaxHeight(2).Render(bold.Render(f.Task))}
	c := m.st.title
	if f.Expired(m.now) {
		c = m.st.bad
		lines = append(lines, m.st.bad.Render("tempo esgotado"))
	}
	lines = append(lines, m.st.muted.Render(fmt.Sprintf("%d de %d min focados",
		int(f.Focused(m.now).Minutes()), f.Declared)))
	block := lipgloss.JoinHorizontal(lipgloss.Top, c.Render(bigText(mmss(f.Ends.Sub(m.now)))), "   ",
		strings.Join(lines, "\n"))
	if f.Expired(m.now) { // as escolhas embaixo do relógio, onde cabem em qualquer largura
		block += "\n\n" + m.st.bad.Render("c continuar") + m.st.muted.Render(" · ") +
			m.st.bad.Render("t trocar") + m.st.muted.Render(" · ") + m.st.bad.Render("p parar") +
			m.st.muted.Render(" · ") + m.st.ok.Render("f feito")
	}
	return block
}

// panelWidth: a largura útil de quem mostra o foco — o painel da sala solo ou a tela inteira.
func (m model) panelWidth() int {
	if m.solo() {
		return panelCols
	}
	return max(m.width-4, 40)
}

// received: quantas reações cada membro recebeu nos eventos da sala (o contador do card, ADR 0004).
func received(s rooms.Snapshot) map[string]int {
	n := map[string]int{}
	for _, e := range s.Events {
		for _, who := range e.Reacts {
			n[e.Who] += len(who)
			if slices.Contains(who, e.Who) { // reagir ao próprio evento não enche o próprio contador
				n[e.Who]--
			}
		}
	}
	return n
}

// orderedMembers: você primeiro, depois quem está focando, quem está sem tarefa e quem está offline.
// Numa sala cheia, quem está focando agora fica sempre nas primeiras posições.
func orderedMembers(id string, s rooms.Snapshot) []rooms.Member {
	rank := func(x rooms.Member) int {
		switch {
		case x.ID == id:
			return 0
		case x.Status == rooms.Focusing:
			return 1
		case x.Status == rooms.Idle:
			return 2
		}
		return 3
	}
	ms := slices.Clone(s.Members)
	slices.SortStableFunc(ms, func(a, b rooms.Member) int { return rank(a) - rank(b) })
	return ms
}

// cards: per cards por fileira; compact troca o desenho do pet por um ponto da cor do humor.
func (m model) cards(s rooms.Snapshot, per int, compact bool) string {
	ms := orderedMembers(m.id, s)
	got := received(s)
	per = max(1, per)
	var rows []string
	for i := 0; i < len(ms); i += per {
		var row []string
		for _, x := range ms[i:min(i+per, len(ms))] {
			row = append(row, m.card(x, got[x.ID], compact))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	}
	return strings.Join(rows, "\n")
}

func (m model) card(x rooms.Member, got int, compact bool) string {
	mood := pet.MoodOf(x.Energy)
	st, name, task := m.st.card, bold.Render(x.Nick), ""
	if x.ID == m.id {
		st, name = st.BorderForeground(m.st.accent), bold.Render("você")
	}
	if x.Status == rooms.Focusing {
		task = x.Focus.Task
	}
	reacts := " "
	switch {
	case got == 1:
		reacts = m.st.muted.Render("1 reação")
	case got > 1:
		reacts = m.st.muted.Render(fmt.Sprintf("%d reações", got))
	}
	art := pet.Render(mood, m.tick, m.pixel)
	if compact {
		art = lipgloss.NewStyle().Foreground(lipgloss.Color(pet.MoodColor[mood])).Render("●")
	}
	return st.Render(lipgloss.JoinVertical(lipgloss.Center,
		art,
		name+" "+m.st.muted.Render(pet.MoodName[mood]),
		m.st.task.Render(task),
		m.status(x),
		reacts))
}

// feed: os n eventos mais novos, com as reações numa coluna só, em ordem fixa.
func (m model) feed(n int) string {
	ev := m.events()
	ev = ev[:min(max(n, 0), len(ev))]
	if len(ev) == 0 {
		return m.st.muted.Render("nada por aqui ainda: quando alguém começar ou cumprir uma tarefa, aparece aqui")
	}
	text := func(e rooms.Event) string {
		nick := e.Nick
		if e.Who == m.id {
			nick = "você"
		}
		line := m.st.muted.Render(e.At.Format("15:04")) + " " + bold.Render(nick) + " " + e.Verb
		if e.Task != "" { // entrar e sair não têm tarefa
			line += " \"" + e.Task + "\""
		}
		return line
	}
	col := 0
	for _, e := range ev {
		col = max(col, lipgloss.Width(text(e)))
	}
	var b strings.Builder
	sel := m.selIndex()
	for i, e := range ev {
		mark := "  "
		if i == sel {
			mark = m.st.title.Render("> ")
		}
		line := lipgloss.NewStyle().Width(col).Render(text(e))
		for k, who := range e.Reacts { // ordem fixa 1-4: quem tem zero sai e os seguintes andam para a esquerda
			if len(who) == 0 {
				continue
			}
			r := fmt.Sprintf("%s %d", reactionGlyphs[k], len(who)) // o espaço protege o número de emoji mais largos que 2 colunas
			if slices.Contains(who, m.id) {
				r = m.st.title.Render(r)
			}
			line += "  " + r
		}
		b.WriteString(mark + line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m model) legend() string {
	var ks []string
	for k, g := range reactionGlyphs {
		ks = append(ks, fmt.Sprintf("%d %s", k+1, g))
	}
	return m.st.muted.Render("reagir ao evento selecionado: ") + strings.Join(ks, "  ")
}
