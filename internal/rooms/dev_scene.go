package rooms

// O editor guarda intenções; só DevSave toca no banco e nas salas. DevPreview usa o mesmo
// planejamento, sem efeitos, para desenhar os cards e a obra com os renderizadores normais.
import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/store"
	"time"
)

var Dev bool // ligado no main junto de -dev; não é uma permissão de uma sessão SSH.

type devOffset struct {
	Day     string
	Minutes float64
}
type devState struct {
	ViewAs   map[string]string
	Presence map[string]bool
	Progress map[string]map[string]devOffset
}

var devSaved = devState{Presence: map[string]bool{}, Progress: map[string]map[string]devOffset{}}

type DevPerson struct {
	Coins    int
	ID, Nick string
	Pet      pet.Pet
	Online   bool
	Focus    *Focus
}
type DevScene struct {
	People  []DevPerson
	Room    *store.RoomRow
	Members []string
	Token   string
}
type DevChange struct {
	Coins          *int
	History        *store.TaskRow
	ID             string
	Create         bool
	Nick           *string
	Energy         *float64
	Color          *string
	Appearance     *pet.Appearance
	Stage          *int // -1 volta aos minutos reais, sem apagar tarefas.
	Online, Member *bool
	Action, Task   string // start, done ou stop; o tempo só é escolhido ao iniciar.
	Minutes        int
}

func LoadDev() error {
	if !Dev {
		return nil
	}
	if err := store.DevState("inline", &devSaved); err != nil {
		return err
	}
	if devSaved.Presence == nil {
		devSaved.Presence = map[string]bool{}
	}
	if devSaved.Progress == nil {
		devSaved.Progress = map[string]map[string]devOffset{}
	}
	if devSaved.ViewAs == nil {
		devSaved.ViewAs = map[string]string{}
	}
	mu.Lock()
	defer mu.Unlock()
	for id, on := range devSaved.Presence {
		u := load(id)
		v := on
		u.devOnline = &v
	}
	return nil
}

// Chame com mu: o token não inclui o relógio, mas inclui tudo que alguém poderia sobrescrever.
func devScene(code string) (DevScene, error) {
	rows, err := store.DevUsers(time.Now())
	if err != nil {
		return DevScene{}, err
	}
	s := DevScene{}
	for _, r := range rows {
		u := load(r.ID)
		p := DevPerson{ID: r.ID, Nick: r.Nick, Pet: r.Pet, Online: u.status() != Offline}
		if u.focus != nil {
			f := *u.focus
			p.Focus = &f
		}
		economy, e := store.ReadEconomy(r.ID)
		if e != nil {
			return s, e
		}
		p.Coins = economy.Balance
		s.People = append(s.People, p)
	}
	if r := rooms[code]; r != nil {
		s.Room = &store.RoomRow{Code: r.code, Name: r.name, Owner: r.owner, Max: r.max, Closed: r.closed, Obra: r.obra, ObraMeta: r.meta, BoardOpen: r.boardOpen, BoardRank: r.boardRank}
		s.Members = slices.Clone(r.members)
	}
	data, _ := json.Marshal(struct {
		Scene DevScene
		State devState
	}{s, devSaved})
	s.Token = fmt.Sprintf("%x", sha256.Sum256(data))
	return s, nil
}

func DevRead(code string) (DevScene, error) {
	if !Dev {
		return DevScene{}, errors.New("modo dev desativado")
	}
	mu.Lock()
	defer mu.Unlock()
	return devScene(code)
}

// DevPreview é puro: inclusive o fim de uma tarefa calcula só uma cópia do pet.
func DevPreview(base DevScene, changes []DevChange, obra *string, now time.Time) (DevScene, error) {
	s := base
	s.People = slices.Clone(base.People)
	s.Members = slices.Clone(base.Members)
	if base.Room != nil {
		r := *base.Room
		s.Room = &r
	}
	if obra != nil {
		if s.Room == nil || !slices.Contains(Obras, *obra) {
			return s, errors.New("obra inválida")
		}
		if s.Room.Max == MinMembers && *obra != "" {
			return s, errors.New("a sala solo não tem obra")
		}
		s.Room.Obra = *obra
	}
	for _, c := range changes {
		i := slices.IndexFunc(s.People, func(p DevPerson) bool { return p.ID == c.ID })
		if c.Create {
			if i >= 0 || c.Nick == nil {
				return s, errors.New("esse apelido já está em uso")
			}
			s.People = append(s.People, DevPerson{ID: c.ID, Nick: *c.Nick, Pet: pet.Pet{Energy: pet.EnergyStart, LastActive: now}})
			i = len(s.People) - 1
		}
		if i < 0 {
			return s, errors.New("ninguém com esse id")
		}
		p := &s.People[i]
		if c.Nick != nil {
			if err := store.ValidateNick(*c.Nick); err != nil {
				return s, err
			}
			p.Nick = *c.Nick
		}
		if c.Coins != nil {
			if *c.Coins < 0 || *c.Coins > pet.Coins.Maximum {
				return s, errors.New("saldo inválido")
			}
			p.Coins = *c.Coins
		}
		if c.History != nil {
			h := c.History
			if h.Task == "" || len([]rune(h.Task)) > MaxTaskLength || !h.End.After(h.Start) || h.End.After(now) {
				return s, errors.New("tarefa histórica inválida")
			}
		}
		if c.Energy != nil {
			if math.IsNaN(*c.Energy) || math.IsInf(*c.Energy, 0) || *c.Energy < 0 || *c.Energy > 100 {
				return s, errors.New("energia de 0 a 100")
			}
			// Materializa o valor exibido sem apagar o streak ou somar bônus oculto.
			p.Pet.Energy, p.Pet.LastActive, p.Pet.DirectEnergy = *c.Energy, now, true
		}
		if c.Color != nil {
			a := p.Pet.Look()
			found := false
			for _, v := range pet.Pets.Character(a.Character).Variants {
				candidate := a
				candidate.Variant = v.ID
				if pet.Pets.ArtworkColor(candidate) == *c.Color {
					a = candidate
					found = true
					break
				}
			}
			if !found {
				return s, errors.New("cor inválida")
			}
			p.Pet.Appearance, p.Pet.Color = a, *c.Color
		}
		if c.Appearance != nil {
			if err := pet.Pets.ValidateAppearance(*c.Appearance); err != nil {
				return s, err
			}
			p.Pet.Appearance = *c.Appearance
			p.Pet.Color = pet.Pets.ArtworkColor(*c.Appearance)
		}
		if c.Online != nil {
			p.Online = *c.Online
		}
		if c.Member != nil {
			if s.Room == nil {
				return s, errors.New("abra uma sala na aba sala para pôr gente nela")
			}
			s.Members = slices.DeleteFunc(s.Members, func(id string) bool { return id == c.ID })
			if *c.Member {
				s.Members = append(s.Members, c.ID)
			}
		}
		if c.Stage != nil && (s.Room == nil || *c.Stage < -1 || *c.Stage >= ObraStages) {
			return s, errors.New("estágio inválido")
		}
		switch c.Action {
		case "":
		case "start":
			if p.Focus != nil {
				return s, errors.New("já existe uma tarefa em andamento")
			}
			if err := ValidateTask(c.Task, c.Minutes); err != nil {
				return s, err
			}
			code := ""
			if s.Room != nil {
				code = s.Room.Code
			}
			p.Focus = &Focus{Task: c.Task, Room: code, Start: now, SegStart: now, Ends: now.Add(time.Duration(c.Minutes) * time.Minute), Declared: c.Minutes}
			p.Online = true
		case "done", "stop":
			if p.Focus == nil {
				return s, errors.New("nenhuma tarefa em andamento")
			}
			p.Pet.Focus(p.Focus.Focused(now).Minutes(), c.Action == "done", now)
			p.Focus = nil
		default:
			return s, errors.New("ação inválida")
		}
	}
	names := map[string]bool{}
	for _, p := range s.People {
		if names[p.Nick] {
			return s, errors.New("esse apelido já está em uso")
		}
		names[p.Nick] = true
	}
	if s.Room != nil {
		if len(s.Members) > s.Room.Max {
			return s, errors.New("a sala já está cheia")
		}
		if !slices.Contains(s.Members, s.Room.Owner) {
			s.Room.Owner = ""
			if len(s.Members) > 0 && s.Room.Code != DemoCode {
				s.Room.Owner = s.Members[0]
			}
		}
	}
	return s, nil
}

func ValidateTask(task string, minutes int) error {
	if task == "" || len([]rune(task)) > MaxTaskLength {
		return errors.New("tarefa de 1 a 80 caracteres")
	}
	if minutes < MinTaskMinutes || minutes > MaxTaskMinutes {
		if !(Demo && minutes == 1) {
			return errors.New("tempo de 5 a 120 min")
		}
	}
	return nil
}

func personMinutes(p DevPerson, now time.Time) float64 {
	m := 0.0
	if p.Pet.Today == now.Format(time.DateOnly) {
		m = p.Pet.TodayMin
	}
	if p.Focus != nil {
		m += p.Focus.Focused(now).Minutes()
	}
	return m
}

// DevWorld projeta um cenário sobre o retrato normal, sem modificar os mapas compartilhados.
func DevWorld(w World, s DevScene, changes []DevChange, now time.Time) World {
	out := World{Rooms: map[string]Snapshot{}, Users: map[string]Member{}}
	for k, v := range w.Rooms {
		out.Rooms[k] = v
	}
	for k, v := range w.Users {
		out.Users[k] = v
	}
	for _, p := range s.People {
		x := Member{ID: p.ID, Nick: p.Nick, Color: p.Pet.Color, Appearance: p.Pet.Look(), Energy: p.Pet.EnergyAt(now), TodayMin: personMinutes(p, now), Status: Offline}
		if p.Online {
			x.Status = Idle
		}
		if p.Focus != nil {
			x.Focus = *p.Focus
			if p.Online {
				x.Status = Focusing
			}
		}
		out.Users[p.ID] = x
	}
	if s.Room != nil {
		r := out.Rooms[s.Room.Code]
		r.Code, r.Name, r.Owner, r.Max, r.Closed, r.Obra, r.Meta = s.Room.Code, s.Room.Name, s.Room.Owner, s.Room.Max, s.Room.Closed, s.Room.Obra, s.Room.ObraMeta
		r.Members = nil
		for _, id := range s.Members {
			x := out.Users[id]
			// A projeção conserva o progresso já salvo, a menos que este rascunho o substitua.
			if old, ok := w.Rooms[r.Code]; ok {
				for _, oldX := range old.Members {
					if oldX.ID == id {
						x.TodayMin += oldX.TodayMin - w.Users[id].TodayMin
						break
					}
				}
			}
			for _, c := range changes {
				if c.ID == id && c.Stage != nil {
					if *c.Stage < 0 {
						x.TodayMin = personMinutes(s.People[slices.IndexFunc(s.People, func(p DevPerson) bool { return p.ID == id })], now)
					} else {
						x.TodayMin = float64(*c.Stage*r.Meta) / float64(ObraStages-1)
					}
				}
			}
			r.Members = append(r.Members, x)
		}
		out.Rooms[r.Code] = r
	}
	return out
}

func DevSave(base DevScene, changes []DevChange, obra *string, now time.Time) (DevScene, error) {
	return DevSaveAs(base, changes, obra, now, "", "")
}

func DevSaveAs(base DevScene, changes []DevChange, obra *string, now time.Time, original, target string) (DevScene, error) {
	if !Dev {
		return DevScene{}, errors.New("modo dev desativado")
	}
	devMu.Lock()
	defer devMu.Unlock()
	mu.Lock()
	defer mu.Unlock()
	code := ""
	if base.Room != nil {
		code = base.Room.Code
	}
	current, err := devScene(code)
	if err != nil {
		return current, err
	}
	if current.Token != base.Token {
		return current, errors.New("o cenário mudou; cancele e abra o editor de novo")
	}
	next, err := DevPreview(current, changes, obra, now)
	if err != nil {
		return current, err
	}
	stateData, _ := json.Marshal(devSaved)
	var state devState
	_ = json.Unmarshal(stateData, &state)
	if state.ViewAs == nil {
		state.ViewAs = map[string]string{}
	}
	if target != "" {
		if !slices.ContainsFunc(next.People, func(p DevPerson) bool { return p.ID == target }) {
			return current, errors.New("ninguém com esse id")
		}
		state.ViewAs[original] = target
	}
	var writes []store.DevWrite
	for _, c := range changes {
		p := next.People[slices.IndexFunc(next.People, func(p DevPerson) bool { return p.ID == c.ID })]
		w := store.DevWrite{ID: p.ID, Nick: p.Nick, Pet: p.Pet, New: c.Create, Coins: c.Coins, History: c.History}
		if c.Action == "done" || c.Action == "stop" {
			old := current.People[slices.IndexFunc(current.People, func(p DevPerson) bool { return p.ID == c.ID })]
			f := old.Focus
			// recordTask soma o foco uma vez, usando o pet anterior ao encerramento.
			w.Pet = old.Pet
			if c.Energy != nil {
				w.Pet.Energy, w.Pet.LastActive, w.Pet.DirectEnergy = *c.Energy, now, true
			}
			w.Pet.Color = p.Pet.Color
			w.Pet.Appearance = p.Pet.Appearance
			result := pet.NotDone
			if c.Action == "done" {
				result = pet.Done
			}
			w.Finish = &store.TaskRow{Start: f.Start, End: now, Task: f.Task, DeclaredMin: f.Declared, Focused: f.Focused(now), Result: result, Room: f.Room}
		}
		if f := p.Focus; f != nil {
			w.Open = &store.OpenTask{User: p.ID, Task: f.Task, Room: f.Room, Start: f.Start, SegStart: f.SegStart, Ends: f.Ends, Declared: f.Declared, Counted: f.Counted}
		}
		writes = append(writes, w)
		if c.Online != nil || c.Action == "start" {
			state.Presence[p.ID] = p.Online
		}
		if c.Stage != nil {
			if state.Progress[code] == nil {
				state.Progress[code] = map[string]devOffset{}
			}
			if *c.Stage < 0 {
				delete(state.Progress[code], p.ID)
			} else {
				state.Progress[code][p.ID] = devOffset{now.Format(time.DateOnly), float64(*c.Stage*next.Room.ObraMeta)/float64(ObraStages-1) - personMinutes(p, now)}
			}
		}
	}
	if err = store.SaveDevScene(writes, next.Room, next.Members, state); err != nil {
		return current, err
	}
	devSaved = state
	for _, c := range changes {
		p := next.People[slices.IndexFunc(next.People, func(p DevPerson) bool { return p.ID == c.ID })]
		u := load(p.ID)
		u.appearance = p.Pet.Look()
		u.nick, u.color, u.energy, u.today, u.todayMin, u.focus = p.Nick, p.Pet.Color, p.Pet.EnergyAt(now), p.Pet.Today, p.Pet.TodayMin, p.Focus
		if on, ok := state.Presence[p.ID]; ok {
			v := on
			u.devOnline = &v
		}
	}
	if r := rooms[code]; r != nil {
		before := slices.Clone(r.members)
		r.members = slices.Clone(next.Members)
		r.owner, r.obra = next.Room.Owner, next.Room.Obra
		if len(r.members) == 0 {
			delete(rooms, code)
		}
		for _, id := range before {
			if !slices.Contains(r.members, id) {
				roomEvent(r, load(id), "saiu", "")
			}
		}
		for _, id := range r.members {
			if !slices.Contains(before, id) {
				roomEvent(r, load(id), "entrou", "")
			}
		}
	}
	for _, c := range changes {
		if c.History != nil {
			eventReceipt(load(c.ID), *c.History)
		}
		if c.Action == "start" {
			event(load(c.ID), "começou", c.Task)
		}
		if c.Action == "done" {
			old := current.People[slices.IndexFunc(current.People, func(p DevPerson) bool { return p.ID == c.ID })]
			eventReceipt(load(c.ID), store.TaskRow{Start: old.Focus.Start, End: now, Task: old.Focus.Task})
		}
	}
	return devScene(code)
}

// DevIdentity conserva a identidade SSH e restaura só a escolha de simulação daquele login.
func DevIdentity(original string) string {
	if !Dev {
		return original
	}
	mu.Lock()
	defer mu.Unlock()
	if target := devSaved.ViewAs[original]; target != "" {
		return target
	}
	return original
}

func DevAssume(original, target string) error {
	if !Dev {
		return errors.New("modo dev desativado")
	}
	devMu.Lock()
	defer devMu.Unlock()
	mu.Lock()
	defer mu.Unlock()
	if _, _, ok, err := store.GetUser(target); err != nil {
		return err
	} else if !ok {
		return errors.New("ninguém com esse id")
	}
	data, _ := json.Marshal(devSaved)
	var state devState
	_ = json.Unmarshal(data, &state)
	if state.ViewAs == nil {
		state.ViewAs = map[string]string{}
	}
	state.ViewAs[original] = target
	if err := store.SaveDevScene(nil, nil, nil, state); err != nil {
		return err
	}
	devSaved = state
	return nil
}
