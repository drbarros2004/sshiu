package tui

// Edição na própria tela: base é o retrato salvo; changes é o rascunho, nunca o banco.
import (
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type inlineDev struct {
	base        rooms.DevScene
	changes     []rooms.DevChange
	obra        *string
	selected    string
	locked      bool
	field       string
	value       string
	number      float64
	picker      int
	target      string
	busy        bool
	freshNumber bool
	rawNumber   string
	help        bool
}
type inlineMsg struct {
	actor      string
	tab        tab
	scene      rooms.DevScene
	err        error
	save, exit bool
	target     string
}
type assumeMsg struct {
	err           error
	actor, target string
	next          model
}

var devAccent = lipgloss.NewStyle().Foreground(lipgloss.Color("#5FD7D7"))

func (m model) inline() *inlineDev {
	if !Dev {
		return nil
	}
	if m.tab == tabRoom {
		return m.roomDev
	}
	if m.tab == tabProfile {
		return m.profileDev
	}
	return nil
}
func (d *inlineDev) dirty() bool {
	return len(d.changes) > 0 || d.obra != nil || d.target != "" || d.field != ""
}
func (d *inlineDev) change(id string) *rooms.DevChange {
	for i := range d.changes {
		if d.changes[i].ID == id {
			return &d.changes[i]
		}
	}
	d.changes = append(d.changes, rooms.DevChange{ID: id})
	return &d.changes[len(d.changes)-1]
}
func (m model) readInline() tea.Cmd {
	actor, code, t := m.id, m.room, m.tab
	if t == tabProfile {
		code = ""
	}
	return func() tea.Msg { s, e := rooms.DevRead(code); return inlineMsg{actor: actor, tab: t, scene: s, err: e} }
}
func (m model) saveInline(exit bool) tea.Cmd {
	d := m.inline()
	base, changes, obra, target := d.base, slices.Clone(d.changes), d.obra, d.target
	actor, t, original := m.id, m.tab, m.original
	d.busy = true
	return func() tea.Msg {
		s, e := rooms.DevSaveAs(base, changes, obra, time.Now(), original, target)
		return inlineMsg{actor: actor, tab: t, scene: s, err: e, save: true, exit: exit, target: target}
	}
}
func (m model) assume(target string) tea.Cmd {
	actor, invite, origin := m.id, m.invite, m.original
	return func() tea.Msg {
		if err := rooms.DevAssume(origin, target); err != nil {
			return assumeMsg{actor: actor, err: err}
		}
		n := New(origin, invite, "")
		n.original = origin
		return assumeMsg{actor: actor, target: target, next: n}
	}
}
func (m model) inlineScene() (rooms.DevScene, error) {
	d := m.inline()
	return rooms.DevPreview(d.base, d.changes, d.obra, m.now)
}
func (m model) inlinePerson() rooms.DevPerson {
	s, _ := m.inlineScene()
	for _, p := range s.People {
		if p.ID == m.inline().selected {
			return p
		}
	}
	return rooms.DevPerson{}
}
func (m model) inlineRender() model {
	d := m.inline()
	if d == nil {
		return m
	}
	s, err := m.inlineScene()
	if err != nil {
		return m
	}
	m.w = rooms.DevWorld(m.w, s, d.changes, m.now)
	if m.tab == tabProfile {
		for _, p := range s.People {
			if p.ID == m.id {
				m.pet = p.Pet
				m.nick = p.Nick
				if !m.custom {
					m.pcolor = pet.ColorIndex(p.Pet.Color)
				}
			}
		}
	}
	return m
}

// inlineKey trata só o contexto dev. O perfil delega as demais teclas ao editor normal.
func (m model) inlineKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	d := m.inline()
	k := msg.String()
	if d.busy {
		return m, nil
	}
	m.err = ""
	p := m.inlinePerson()
	if d.field == "" && k == "?" {
		d.help = !d.help
		return m, nil
	}
	d.help = false
	if d.field != "" {
		if k == "esc" {
			d.field = ""
			return m, nil
		}
		if d.field == "people" {
			people := m.availablePeople()
			switch k {
			case "up":
				d.picker = max(0, d.picker-1)
			case "down":
				d.picker = min(max(0, len(people)-1), d.picker+1)
			case "ctrl+n":
				d.field = "new"
				m.in = newInput(m.t("apelido da pessoa nova"), nickMax)
			case "enter":
				if len(people) > 0 {
					id := people[d.picker].ID
					on := true
					d.change(id).Member = &on
					d.selected = id
					d.field = ""
				}
			}
			return m, nil
		}
		if d.field == "task" {
			pending := ""
			for _, c := range d.changes {
				if c.ID == p.ID {
					pending = c.Action
				}
			}
			if pending != "" {
				if pending == "start" && k == "p" {
					c := d.change(p.ID)
					c.Action, c.Task, c.Minutes = "", "", 0
					d.field = ""
					return m, nil
				}
				if k == "n" || k == "f" || k == "p" {
					m.err = m.t("salve a alteração de tarefa antes de outra ação")
					return m, nil
				}
			}
			switch k {
			case "n":
				if p.Focus == nil {
					d.field = "taskText"
					m.in = newInput(m.t("ex.: fazer o exercício 1 da lista 3"), 80)
				}
			case "f", "p":
				if p.Focus != nil {
					c := d.change(p.ID)
					c.Action = "done"
					if k == "p" {
						c.Action = "stop"
					}
					d.field = ""
				}
			}
			return m, nil
		}
		switch d.field {
		case "name", "new", "taskText":
			if k != "enter" {
				var cmd tea.Cmd
				m.in, cmd = m.in.Update(msg)
				return m, cmd
			}
			value := strings.TrimSpace(m.in.Value())
			if d.field == "taskText" {
				if value == "" {
					return m, nil
				}
				d.value = value
				d.field = "taskMinutes"
				d.number = 25
				d.freshNumber = true
				d.rawNumber = strconv.FormatFloat(d.number, 'f', -1, 64)
				return m, nil
			}
			if m.err = nickProblem(m.lang, value); m.err != "" {
				return m, nil
			}
			scene, _ := m.inlineScene()
			for _, other := range scene.People {
				if other.Nick == value && (d.field == "new" || other.ID != p.ID) {
					m.err = m.t("esse apelido já está em uso")
					return m, nil
				}
			}
			if d.field == "new" {
				id := "dev:" + value
				c := d.change(id)
				c.Create = true
				c.Nick = &value
				on := true
				c.Member = &on
				d.selected = id
			} else {
				d.change(p.ID).Nick = &value
			}
			d.field = ""
			d.locked = false
			return m, nil
		case "energy", "taskMinutes":
			if k == "left" || k == "right" || k == "up" || k == "down" || k == "backspace" {
				d.freshNumber = false
			}
			step := 1.0
			if d.field == "taskMinutes" {
				step = 5
			}
			switch k {
			case "left", "down":
				d.number = max(0, d.number-step)
			case "right", "up":
				d.number = min(120, d.number+step)
			case "backspace":
				if d.rawNumber == "" {
					d.rawNumber = strconv.FormatFloat(d.number, 'f', -1, 64)
				}
				d.rawNumber = d.rawNumber[:max(0, len(d.rawNumber)-1)]
				d.number, _ = strconv.ParseFloat(strings.TrimSuffix(d.rawNumber, "."), 64)

			case "enter":
				if d.field == "energy" {
					if d.number > 100 {
						m.err = m.t("energia de 0 a 100")
						return m, nil
					}
					v := d.number
					d.change(p.ID).Energy = &v
				} else {
					if d.number != float64(int(d.number)) {
						m.err = m.t("a duração precisa ser um número inteiro")
						return m, nil
					}
					if err := rooms.ValidateTask(d.value, int(d.number)); err != nil {
						m.err = m.lang.errText(err)
						return m, nil
					}
					c := d.change(p.ID)
					c.Action, c.Task, c.Minutes = "start", d.value, int(d.number)
				}
				d.field = ""
				d.locked = false
				return m, nil
			default:
				if k != "" && strings.Trim(k, "0123456789.") == "" {
					if d.freshNumber {
						d.rawNumber = ""
						d.freshNumber = false
					}
					next := d.rawNumber + k
					if len(next) <= 8 && strings.Count(next, ".") <= 1 {
						d.rawNumber = next
						d.number, _ = strconv.ParseFloat(strings.TrimSuffix(next, "."), 64)
					}
				}
			}
			if k == "left" || k == "right" || k == "up" || k == "down" {
				if d.field == "energy" {
					d.number = min(d.number, 100)
				}
				d.rawNumber = strconv.FormatFloat(d.number, 'f', -1, 64)
			}

			return m, nil
		case "color", "stage", "obra":
			delta := 0
			if k == "left" {
				delta = -1
			}
			if k == "right" {
				delta = 1
			}
			if delta != 0 {
				n := len(pet.Colors)
				if d.field == "stage" {
					n = rooms.ObraStages + 1
				}
				if d.field == "obra" {
					n = len(rooms.Obras)
				}
				d.picker = (d.picker + delta + n) % n
			}
			if k == "enter" {
				switch d.field {
				case "color":
					v := pet.Colors[d.picker].ID
					d.change(p.ID).Color = &v
				case "stage":
					v := d.picker - 1
					d.change(p.ID).Stage = &v
				case "obra":
					if d.base.Room != nil && d.base.Room.Max == rooms.MinMembers && rooms.Obras[d.picker] != "" {
						m.err = m.t("a sala solo não tem obra")
						return m, nil
					}
					v := rooms.Obras[d.picker]
					d.obra = &v
				}
				d.field = ""
				d.locked = false
			}
			return m, nil
		}
	}
	switch k {
	case "ctrl+s", "x":
		if m.tab == tabProfile && m.custom {
			v := m.myColor()
			d.change(m.id).Color = &v
			m.custom = false
		}
		return m, m.saveInline(k == "x")
	case "esc":
		if d.locked {
			d.locked = false
			return m, nil
		}
		if m.tab == tabRoom {
			m.roomDev = nil
		} else {
			m.profileDev = nil
			m.custom = false
			m.pcolor = m.pcSaved
		}
		return m, m.loadProfile()
	case "ctrl+r":
		d.target = m.original
		return m, nil
	case "ctrl+n":
		if m.tab == tabProfile || d.locked {
			d.field = "name"
			m.in = newInput(m.t("seu apelido"), nickMax)
			m.in.SetValue(p.Nick)
		}
		return m, nil
	case "ctrl+e":
		d.field = "energy"
		d.number = p.Pet.EnergyAt(m.now)
		d.freshNumber = true
		d.rawNumber = strconv.FormatFloat(d.number, 'f', -1, 64)
		return m, nil
	}
	if m.tab == tabProfile {
		if m.custom && (k == "x" || k == "ctrl+s") {
			v := m.myColor()
			d.change(m.id).Color = &v
			m.custom = false
			return m, m.saveInline(k == "x")
		}
		if m.custom && k == "enter" {
			v := m.myColor()
			d.change(m.id).Color = &v
			m.custom = false
			return m, nil
		}
		return m.profileKey(msg)
	}
	if !d.locked {
		switch k {
		case "left", "right":
			scene, _ := m.inlineScene()
			ids := scene.Members
			if len(ids) == 0 {
				return m, nil
			}
			i := slices.Index(ids, d.selected)
			delta := 1
			if k == "left" {
				delta = -1
			}
			d.selected = ids[(max(i, 0)+delta+len(ids))%len(ids)]
		case "enter":
			if p.ID != "" {
				d.locked = true
			}
		case "ctrl+b":
			scene, _ := m.inlineScene()
			if scene.Room != nil && len(scene.Members) >= scene.Room.Max {
				m.err = m.t("a sala já está cheia")
				return m, nil
			}
			d.field = "people"
			d.picker = 0
		case "ctrl+o":
			d.field = "obra"
			if d.base.Room != nil {
				scene, _ := m.inlineScene()
				d.picker = max(0, slices.Index(rooms.Obras, scene.Room.Obra))
			}
		case "ctrl+u":
			if p.ID != "" {
				d.target = p.ID
			}
		}
		return m, nil
	}
	switch k {
	case "ctrl+u":
		d.target = p.ID
	case "ctrl+l":
		d.field = "color"
		d.picker = pet.ColorIndex(p.Pet.Color)
	case "ctrl+o":
		d.field = "stage"
		d.picker = 1
		scene, _ := m.inlineScene()
		world := rooms.DevWorld(m.w, scene, d.changes, m.now)
		if scene.Room != nil {
			for _, x := range world.Rooms[scene.Room.Code].Members {
				if x.ID == p.ID {
					d.picker = rooms.Stage(x.TodayMin, scene.Room.ObraMeta) + 1
				}
			}
		}
	case "ctrl+p":
		v := !p.Online
		d.change(p.ID).Online = &v
	case "ctrl+d":
		v := false
		d.change(p.ID).Member = &v
		d.locked = false
	case "ctrl+t":
		d.field = "task"
	}
	return m, nil
}
func (m model) availablePeople() []rooms.DevPerson {
	s, _ := m.inlineScene()
	return slices.DeleteFunc(s.People, func(p rooms.DevPerson) bool { return slices.Contains(s.Members, p.ID) })
}

// previewField incorpora apenas a edição em curso à cópia de desenho; Esc nunca deixa essa edição no rascunho.
func (m model) previewField() model {
	d := m.inline()
	if d == nil || d.field == "" {
		return m
	}
	copy := *d
	copy.changes = slices.Clone(d.changes)
	c := copy.change(d.selected)
	switch d.field {
	case "energy":
		v := d.number
		c.Energy = &v
	case "color":
		v := pet.Colors[d.picker].ID
		c.Color = &v
	case "stage":
		v := d.picker - 1
		c.Stage = &v
	case "obra":
		v := rooms.Obras[d.picker]
		copy.obra = &v
	case "name":
		v := m.in.Value()
		c.Nick = &v
	default:
		return m
	}
	if m.tab == tabRoom {
		m.roomDev = &copy
	} else {
		m.profileDev = &copy
	}
	return m
}

func (m model) inlinePanel(w int) string {
	d := m.inline()
	if d == nil {
		return ""
	}
	title := m.t("modo dev · prévia local")
	if d.dirty() {
		title += m.t(" · alterações por salvar")
	}
	if d.target != "" {
		title += m.tf(" · assumir: %s", d.target)
	}
	body := m.tf("seleção: %s", m.inlinePerson().Nick)
	switch d.field {
	case "name", "new", "taskText":
		body = m.in.View()
	case "energy":
		body = m.tf("energia: %g · ←/→ ou ↑/↓ ajusta · digite o valor", d.number)
	case "taskMinutes":
		body = m.tf("duração: %.0f min · ←/→ ajusta · digite o valor", d.number)
	case "people":
		people := m.availablePeople()
		body = m.t("↑/↓ escolhe · enter adiciona · ctrl+n cria usuário")
		if len(people) > 0 {
			body += "\n‹ " + people[min(d.picker, len(people)-1)].Nick + " ›"
		}
	case "color":
		body = m.t("cor") + ": ‹ " + m.t(pet.Colors[d.picker].Name) + " ›"
	case "stage":
		body = m.tf("estágio da obra: %d · -1 usa o progresso real", d.picker-1)
	case "obra":
		body = m.t("obra") + ": ‹ " + m.t(rooms.Obras[d.picker]) + " ›"
	case "task":
		body = m.t("n inicia tarefa · f marca feito · p para")
	}
	if d.busy {
		body = m.t("salvando…")
	}
	if d.help {
		full := m
		full.width, full.height = 120, 44
		body = full.inlineHelp()
	}
	if m.height < 30 {
		return devAccent.Render(title) + "\n" + lipgloss.NewStyle().Width(max(1, w)).Render(body)
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#5FD7D7")).Width(max(1, w-2)).Render(devAccent.Render(title) + "\n" + body)
}
func (m model) inlineHelp() string {
	d := m.inline()
	if d == nil {
		return ""
	}
	if d.field != "" {
		return m.t("enter confirma prévia · esc cancela campo")
	}
	if m.tab == tabRoom && (m.width < 100 || m.height < 30) {
		return m.t("enter edita · ←/→ membro · ? atalhos · ctrl+s salva · x salva e sai · esc volta")
	}
	base := m.t("ctrl+s salva · x salva e sai · esc volta · ctrl+r volta a você")
	if m.tab == tabProfile {
		actions := m.t("ctrl+n nome") + " · " + m.t("ctrl+e energia")
		if m.custom {
			return m.t("←/→ cor · ↑/↓ prévia do humor · enter confirma prévia · esc volta") + " · " + actions + " · " + base
		}
		return m.profileHelp() + " · " + actions + " · " + base
	}
	if d.locked {
		return m.t("ctrl+n nome · ctrl+e energia · ctrl+l cor · ctrl+o estágio da obra · ctrl+p presença · ctrl+d remove · ctrl+t tarefa · ctrl+u assume") + " · " + base
	}
	return m.t("←/→ membro · enter edita · ctrl+b adiciona · ctrl+o obra · ctrl+u assume") + " · " + base
}

// As respostas carregam a identidade de origem: um perfil atrasado não troca a identidade nova.
func (m model) updateInline(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch v := msg.(type) {
	case inlineMsg:
		if v.actor != m.id {
			return m, nil, true
		}
		d := m.roomDev
		if v.tab == tabProfile {
			d = m.profileDev
		}
		if d == nil {
			return m, nil, true
		}
		d.busy = false
		if v.err != nil {
			m.err = m.lang.errText(v.err)
			return m, nil, true
		}
		selected := d.selected
		*d = inlineDev{base: v.scene, selected: selected}
		if d.selected == "" {
			d.selected = m.id
		}
		for _, p := range v.scene.People {
			if p.ID == m.id {
				m.nick = p.Nick
				m.pcolor = pet.ColorIndex(p.Pet.Color)
				m.pcSaved = m.pcolor
			}
		}
		if !slices.Contains(v.scene.Members, d.selected) && v.tab == tabRoom && len(v.scene.Members) > 0 {
			d.selected = v.scene.Members[0]
		}
		if v.save {
			m.notice = m.t("salvo")
			if v.tab == tabRoom && v.scene.Room == nil {
				m.roomDev = nil
			}
			if v.target != "" && v.target != m.id {
				return m, m.assume(v.target), true
			}
			if v.exit {
				if v.tab == tabRoom {
					m.roomDev = nil
				} else {
					m.profileDev = nil
					m.custom = false
				}
			}
			return m, tea.Batch(act(func() {}), m.loadProfile()), true
		}
		return m, nil, true
	case assumeMsg:
		if v.err != nil {
			m.err = m.lang.errText(v.err)
			return m, nil, true
		}
		if v.actor != m.id {
			return m, nil, true
		}
		n := v.next
		n.width, n.height, n.pixel, n.tick = m.width, m.height, m.pixel, m.tick
		n.tab = m.tab
		n.original = m.original
		// Preferências visuais são da sessão; permissões, perfil e salas são da identidade assumida.
		n.theme, n.saved, n.st, n.lang, n.langSaved, n.vis, n.visSaved, n.statsList, n.listSaved = m.theme, m.saved, m.st, m.lang, m.langSaved, m.vis, m.visSaved, m.statsList, m.listSaved
		n.notice = n.t("identidade alterada; rascunhos anteriores descartados")
		return n, n.loadProfile(), true
	}
	return m, nil, false
}
