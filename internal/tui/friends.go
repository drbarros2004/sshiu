package tui

// Aba "amigos": quem você segue e o que cada um está fazendo agora.

import (
	"fmt"
	"plantao/internal/pet"
	"plantao/internal/rooms"
	"plantao/internal/store"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// presence diz o que alguém está fazendo e em que sala, só olhando o retrato (sem tocar nas salas).
func presence(id string, w rooms.World) (label, roomName string, m *rooms.Member) {
	for _, s := range w.Rooms {
		if m := rooms.Find(s.Room, id); m != nil && (s.Phase == rooms.Break || m.Result == "") {
			if s.Phase == rooms.Focus {
				return "focando", s.Name, m
			}
			return "no intervalo", s.Name, m
		}
	}
	for _, s := range w.Rooms {
		if m := rooms.Find(s.Queue, id); m != nil {
			return "na fila", s.Name, m
		}
	}
	if w.Online[id] {
		return "online", "", nil
	}
	return "offline", "", nil
}

func (m model) isFriend(id string) bool {
	for _, f := range m.friends {
		if f.ID == id {
			return true
		}
	}
	return false
}

// friendsActive conta amigos que não estão offline (vai no rótulo da aba).
func (m model) friendsActive() (n int) {
	for _, f := range m.friends {
		if p, _, _ := presence(f.ID, m.w); p != "offline" {
			n++
		}
	}
	return n
}

func (m model) friendsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k, id := msg.String(), m.id
	if m.adding {
		switch k {
		case "enter":
			nick := strings.TrimSpace(m.in.Value())
			m.adding, m.err = false, ""
			return m, func() tea.Msg {
				if err := store.AddFriend(id, nick); err != nil {
					return errMsg{err}
				}
				return loadFriends(id)()
			}
		case "esc":
			m.adding = false
			return m, nil
		}
		var cmd tea.Cmd
		m.in, cmd = m.in.Update(msg)
		return m, cmd
	}
	switch k {
	case "up", "k":
		m.cur = max(m.cur-1, 0)
	case "down", "j":
		m.cur = max(0, min(m.cur+1, len(m.friends)-1))
	case "a":
		m.adding, m.err, m.in = true, "", newInput("apelido de quem você quer seguir", 16)
	case "x":
		if len(m.friends) > 0 {
			fid := m.friends[m.cur].ID
			return m, func() tea.Msg {
				if err := store.RemoveFriend(id, fid); err != nil {
					return errMsg{err}
				}
				return loadFriends(id)()
			}
		}
	}
	return m, nil
}

func (m model) friendsView() string {
	var b strings.Builder
	if m.adding {
		b.WriteString("Seguir quem? " + m.st.muted.Render("(enter confirma · esc cancela)") + "\n\n" + m.in.View() + "\n\n")
	}
	if len(m.friends) == 0 {
		b.WriteString(m.st.muted.Render("Você ainda não segue ninguém. Aperte a e digite um apelido.\n" +
			"Na demo, os bots ana, bruno, caio, duda e edu também valem."))
		return b.String()
	}
	b.WriteString(bold.Render(fmt.Sprintf("seguindo (%d) · %d ativos agora", len(m.friends), m.friendsActive())) + "\n\n")
	for i, f := range m.friends {
		p, where, mem := presence(f.ID, m.w)
		dot, pst := "●", m.st.ok
		switch p {
		case "focando":
			pst = m.st.sel
		case "offline":
			dot, pst = "○", m.st.muted
		}
		mood := pet.MoodOf(f.Pet.EnergyAt(m.now))
		if mem != nil {
			mood = pet.MoodOf(mem.Energy) // humor ao vivo, se está na sala
		}
		cursor, nick := "  ", fmt.Sprintf("%-16s", f.Nick)
		if i == m.cur {
			cursor, nick = m.st.sel.Render("▸ "), m.st.sel.Render(nick)
		}
		line := cursor + pst.Render(dot) + " " + nick + pst.Render(fmt.Sprintf("%-14s", p)) +
			lipgloss.NewStyle().Foreground(lipgloss.Color(pet.MoodColor[mood])).Render(fmt.Sprintf("%-10s", pet.MoodName[mood]))
		if mem != nil {
			line += m.st.muted.Render(where + " · “" + mem.Lock + "”")
		}
		b.WriteString(line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
