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

const shownEvents = 6

// minuteChoices: os tempos declarados que dá para escolher (na demo, 1 min para a plateia ver o fluxo).
func minuteChoices() []int {
	if rooms.Demo {
		return []int{1, 10, 15, 25, 45, 60}
	}
	return []int{10, 15, 25, 45, 60}
}

func (m model) events() []rooms.Event {
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
			return m, act(func() { rooms.Start(id, task, mins) })
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
		if focusing {
			return m, act(func() { rooms.Finish(id, pet.Done) })
		}
	case "t": // encerra como não cumprida, sem evento: ninguém é exposto por trocar
		if focusing {
			m.entering, m.in = true, newInput("ex.: fazer o exercício 1 da lista 3", 80)
			return m, act(func() { rooms.Finish(id, pet.NotDone) })
		}
	case "p":
		if focusing {
			return m, act(func() { rooms.Finish(id, pet.NotDone) })
		}
	case "c":
		if expired {
			m.choosing, m.cont = true, true
		}
	case "j":
		if focusing && !expired {
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
	case m.inPicker() && m.form != formNone:
		return "enter confirma · esc volta"
	case m.inPicker():
		return "↑/↓ escolhe · enter abre" + map[bool]string{true: " · esc volta", false: ""}[m.room != ""]
	case m.entering:
		return "enter confirma · esc volta"
	case m.choosing:
		return "←/→ escolhe · enter começa · esc volta"
	case focusing && me.Focus.Expired(m.now):
		return "c continuar · t trocar · p parar · f feito · ↑/↓ evento · 1-4 reage"
	case focusing:
		return "f feito · t trocar · p parar · j modo jardim · ↑/↓ evento · 1-4 reage · s salas"
	}
	return "enter começa uma tarefa · ↑/↓ evento · 1-4 reage · s salas"
}

func (m model) roomView() string {
	if m.inPicker() {
		return m.pickerView()
	}
	s := m.w.Rooms[m.room]
	title := bold.Render(s.Name)
	if m.room != rooms.DemoCode {
		title += m.st.muted.Render(" · código " + m.room)
	}
	if m.notice != "" {
		title += "\n" + m.st.ok.Render(m.notice)
	}
	return title + "\n\n" + m.focusPanel() + "\n\n" + m.cards(s) + "\n\n" + m.feed() + "\n" + m.legend()
}

// focusPanel: o seu foco, conforme o momento.
func (m model) focusPanel() string {
	me := m.me()
	f := me.Focus
	choice := func(label string) string {
		return label + "  " + m.st.muted.Render("‹ ") + m.st.title.Render(fmt.Sprintf("%d min", minuteChoices()[min(m.mins, len(minuteChoices())-1)])) + m.st.muted.Render(" ›")
	}
	var text string
	switch {
	case m.entering:
		return bold.Render("Qual a próxima coisa concreta que você vai fazer?") + "\n" + m.in.View()
	case m.choosing && m.cont:
		return bold.Render(f.Task) + "\n" + choice("Mais quanto tempo?")
	case m.choosing:
		return bold.Render(strings.TrimSpace(m.in.Value())) + "\n" + choice("Quanto tempo?")
	case me.Status != rooms.Focusing:
		return bold.Render("Você está parado.") + "\n" + m.st.muted.Render("enter começa uma tarefa")
	case f.Expired(m.now):
		text = m.st.bad.Render("Tempo esgotado: ") + bold.Render(f.Task) + "\n" +
			"c continuar · t trocar · p parar · f feito\n" +
			m.st.muted.Render(fmt.Sprintf("os minutos pararam de contar · %d min focados", int(f.Focused(m.now).Minutes())))
	default:
		text = bold.Render(f.Task) + "\n" +
			m.st.muted.Render(fmt.Sprintf("de %d min · %d min focados", f.Declared, int(f.Focused(m.now).Minutes())))
	}
	c := m.st.title
	if f.Expired(m.now) {
		c = m.st.bad
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, c.Render(bigText(mmss(f.Ends.Sub(m.now)))), "   ", text)
}

// received: quantas reações cada membro recebeu nos eventos da sala (o contador do card, ADR 0004).
func received(s rooms.Snapshot) map[string]int {
	n := map[string]int{}
	for _, e := range s.Events {
		for _, r := range e.Reacts {
			n[e.Who] += len(r)
		}
	}
	return n
}

func (m model) cards(s rooms.Snapshot) string {
	var ms []rooms.Member // você primeiro
	for _, x := range s.Members {
		if x.ID == m.id {
			ms = append([]rooms.Member{x}, ms...)
		} else {
			ms = append(ms, x)
		}
	}
	got := received(s)
	per := max(1, (m.width-4)/18)
	var rows []string
	for i := 0; i < len(ms); i += per {
		var row []string
		for _, x := range ms[i:min(i+per, len(ms))] {
			row = append(row, m.card(x, got[x.ID]))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	}
	return strings.Join(rows, "\n")
}

func (m model) card(x rooms.Member, got int) string {
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
	return st.Render(lipgloss.JoinVertical(lipgloss.Center,
		pet.Render(mood, m.tick, m.pixel),
		name+" "+m.st.muted.Render(pet.MoodName[mood]),
		m.st.task.Render(task),
		m.status(x),
		reacts))
}

// feed: os eventos mais novos, com as reações numa coluna só, em ordem fixa.
func (m model) feed() string {
	ev := m.events()
	if len(ev) == 0 {
		return m.st.muted.Render("nada por aqui ainda: quando alguém começar ou cumprir uma tarefa, aparece aqui")
	}
	text := func(e rooms.Event) string {
		nick := e.Nick
		if e.Who == m.id {
			nick = "você"
		}
		return m.st.muted.Render(e.At.Format("15:04")) + " " + bold.Render(nick) + " " + e.Verb + " \"" + e.Task + "\""
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
