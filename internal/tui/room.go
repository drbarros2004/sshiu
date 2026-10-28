package tui

// Aba "sala": lobby, Lock, foco, check-out e chat do intervalo da sala que você está vendo.

import (
	"fmt"
	"plantao/internal/pet"
	"plantao/internal/rooms"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) chatOpen() bool {
	me := rooms.Find(m.snap.Room, m.id)
	return m.tab == tabRoom && me != nil && m.snap.Phase == rooms.Break && me.Result != ""
}

func (m model) roomKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.picking {
		return m.pickerKey(msg)
	}
	k, s, id, code := msg.String(), m.snap, m.id, m.room
	me := rooms.Find(s.Room, id)
	h := func(f func(*rooms.Room)) tea.Cmd { // efeito na sala atual
		return act(func() {
			if r := rooms.Get(code); r != nil {
				f(r)
			}
		})
	}
	var cmd tea.Cmd
	switch {
	case m.entering:
		switch k {
		case "enter":
			lock := strings.TrimSpace(m.in.Value())
			if lock == "" {
				m.err = "escreva a primeira ação concreta"
				return m, nil
			}
			nick := m.nick
			m.entering, m.err = false, ""
			return m, act(func() { rooms.JoinQueue(code, id, nick, lock) })
		case "esc":
			m.entering, m.err = false, ""
			return m, nil
		}
		m.in, cmd = m.in.Update(msg)
		return m, cmd
	case me != nil && s.Phase == rooms.Break && me.Result == "": // check-out
		switch k {
		case "s":
			return m, h(func(r *rooms.Room) { r.Checkout(id, pet.Done) })
		case "n":
			return m, h(func(r *rooms.Room) { r.Checkout(id, pet.NotDone) })
		}
	case m.chatOpen():
		if k == "enter" {
			text := strings.TrimSpace(m.chat.Value())
			m.chat.Reset()
			return m, h(func(r *rooms.Room) { r.Say(id, text) })
		}
		m.chat, cmd = m.chat.Update(msg)
		return m, cmd
	case me != nil && s.Phase == rooms.Focus && me.Result == "": // focando
		if m.confirm && k == "s" {
			m.confirm = false
			return m, h(func(r *rooms.Room) { r.Abandon(id) })
		}
		m.confirm = k == "x"
	default: // lobby
		switch k {
		case "enter":
			if rooms.Find(s.Queue, id) == nil {
				m.entering, m.in = true, newInput("ex.: terminar a lista 3 de Cálculo", 80)
			}
		case "c":
			return m, h(func(r *rooms.Room) { r.Cancel(id) })
		case "s":
			return m.openPicker()
		}
	}
	return m, nil
}

func (m model) roomHelp() string {
	s := m.snap
	me := rooms.Find(s.Room, m.id)
	switch {
	case m.picking && m.form != formNone:
		return "enter confirma · esc volta"
	case m.picking:
		return "↑/↓ escolhe · enter abre · esc volta"
	case m.entering:
		return "enter entra na fila · esc volta"
	case me != nil && s.Phase == rooms.Break && me.Result == "":
		return "s cumpri · n não cumpri"
	case m.chatOpen():
		return "enter envia · o chat fecha quando o intervalo acaba"
	case me != nil && s.Phase == rooms.Focus && me.Result == "":
		return "x abandona (sair do ssh também conta)"
	case rooms.Find(s.Queue, m.id) != nil:
		return "c sai da fila (desconectar conta como falta) · s salas"
	}
	return "enter entra no próximo sprint · s troca ou cria sala"
}

func (m model) roomView() string {
	if m.picking {
		return m.pickerView()
	}
	title := bold.Render(m.snap.Name)
	if m.room != rooms.PublicCode {
		title += m.st.muted.Render(" · código " + m.room)
	}
	if m.notice != "" {
		title += "\n" + m.st.ok.Render(m.notice)
	}
	title += "\n\n"
	if m.entering {
		return title + bold.Render("Lock:") + " qual a primeira ação que você vai fazer neste sprint?\n\n" + m.in.View()
	}
	if me := rooms.Find(m.snap.Room, m.id); me != nil && (m.snap.Phase == rooms.Break || me.Result == "") {
		return title + m.sprintView(me)
	}
	return title + m.lobbyView()
}

func (m model) sprintView(me *rooms.Member) string {
	s := m.snap
	var b strings.Builder
	label, color := "foco", m.st.accent
	if s.Phase == rooms.Break {
		label, color = "intervalo", m.st.good
	}
	clock := lipgloss.NewStyle().Foreground(color).Render(bigClock(s.Ends.Sub(m.now)))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, clock, "  ",
		bold.Render(label)+"\n"+m.st.muted.Render("sprint das "+s.Start.Format("15:04"))))
	b.WriteString("\n\n" + m.cards(s.Room, false) + "\n\n")

	if s.Phase == rooms.Focus {
		b.WriteString(m.st.muted.Render("🔒 chat travado durante o foco"))
		for _, e := range s.Events {
			b.WriteString("\n" + m.st.muted.Render(e))
		}
		if m.confirm {
			b.WriteString("\n\n" + m.st.bad.Render("abandonar o sprint? [s] sim · qualquer tecla cancela"))
		}
		return b.String()
	}

	done := 0
	for _, x := range s.Room {
		if x.Result == pet.Done {
			done++
		}
	}
	b.WriteString(bold.Render(fmt.Sprintf("%d/%d cumpriram", done, len(s.Room))) + "\n\n")
	if me.Result == "" {
		b.WriteString(m.st.title.Render("check-out:") + " você cumpriu o Lock \"" + me.Lock + "\"?")
		return b.String()
	}
	chat := s.Chat
	if len(chat) > 8 {
		chat = chat[len(chat)-8:]
	}
	for _, c := range chat {
		b.WriteString(c + "\n")
	}
	b.WriteString("\n" + m.chat.View())
	return b.String()
}

func (m model) lobbyView() string {
	s := m.snap
	var b strings.Builder
	b.WriteString("próximo sprint às " + bold.Render(s.Next.Format("15:04")) + "\n\n")
	b.WriteString(m.st.title.Render(bigClock(s.Next.Sub(m.now))) + "\n\n")
	if s.Phase != rooms.Lobby {
		b.WriteString(m.st.muted.Render(fmt.Sprintf("sprint em andamento com %d pessoas", len(s.Room))) + "\n\n")
	}
	if len(s.Queue) == 0 {
		b.WriteString(m.st.muted.Render("ninguém na fila ainda. seja o primeiro."))
	} else {
		b.WriteString(bold.Render(fmt.Sprintf("na fila (%d)", len(s.Queue))) + "\n" + m.cards(s.Queue, true))
	}
	for _, e := range s.Events {
		b.WriteString("\n" + m.st.muted.Render(e))
	}
	return b.String()
}

func (m model) cards(ms []*rooms.Member, queued bool) string {
	per := max(1, (m.width-4)/22)
	var rows []string
	for i := 0; i < len(ms); i += per {
		var row []string
		for _, x := range ms[i:min(i+per, len(ms))] {
			row = append(row, m.card(x, queued))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	}
	return strings.Join(rows, "\n")
}

func (m model) card(x *rooms.Member, queued bool) string {
	mood := pet.MoodOf(x.Energy)
	status := resultLabel[x.Result]
	if x.Result == "" {
		status = map[rooms.Phase]string{rooms.Focus: "focando…", rooms.Break: "check-out…"}[m.snap.Phase]
		if queued {
			status = "na fila"
		}
	}
	switch x.Result {
	case pet.Done:
		status = m.st.ok.Render(status)
	case pet.NotDone, pet.Abandoned:
		status = m.st.bad.Render(status)
	}
	st, name := m.st.card, bold.Render(x.Nick)
	if x.ID == m.id {
		st = st.BorderForeground(m.st.accent)
	}
	if m.isFriend(x.ID) {
		name = m.st.sel.Render("★ ") + name
	}
	return st.Render(lipgloss.JoinVertical(lipgloss.Center,
		pet.Render(mood, m.tick, m.pixel),
		name+" "+m.st.muted.Render(pet.MoodName[mood]),
		m.st.lock.Render(x.Lock),
		status))
}
