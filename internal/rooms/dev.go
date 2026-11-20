package rooms

// O modo dev (flag -dev): criar pessoas e mexer nos números delas com o servidor rodando. Tudo
// passa por aqui para o banco e a memória das salas mudarem juntos (a sala lê da memória).

import (
	"errors"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/store"
	"strings"
	"sync"
	"time"

	"charm.land/log/v2"
)

// Os passos de cada número no painel dev (docs/regras.md).
const (
	DevEnergyStep = 5  // energia guardada
	DevTodayStep  = 15 // minutos de uma tarefa dev de hoje
	DevBeforeStep = 60 // minutos de uma tarefa dev num dia anterior
	DevFocusMin   = 25 // o tempo declarado da tarefa que o painel põe em andamento
	DevFocusStep  = 5  // minutos a mais (+) ou já passados (-) na tarefa em andamento
)

// devMu: um ajuste por vez. O Bubble Tea roda os Cmd em paralelo, e dois "+" seguidos liam o
// mesmo pet e um sobrescrevia o outro. Trave antes de mu, nunca depois.
var devMu sync.Mutex

// devHere: quem está online por uma sessão de mentira do painel, além das de verdade.
var devHere = map[string]bool{}

type DevField int

const (
	DevEnergy DevField = iota
	DevToday           // minutos focados hoje
	DevBefore          // minutos focados antes de hoje: a semana, o mês e o total da tabela de foco
	DevStreak
	DevColor
	DevIdle   // dias sem focar: a energia decai 8 por dia
	DevOnline // uma sessão de mentira, para a sala ver a pessoa
	DevFocus  // a tarefa em andamento
)

// DevCreate cria uma pessoa com o apelido nick. O id vem do apelido, então quem entra com
// "ssh nick@localhost" no modo dev é ela.
func DevCreate(nick string) (string, error) {
	if err := store.ValidateNick(nick); err != nil {
		return "", err
	}
	id := "dev:" + nick
	if _, _, ok, _ := store.GetUser(id); ok { // o PutUser só redefiniria a energia
		return "", errors.New("esse apelido já está em uso")
	}
	if err := store.PutUser(id, nick, pet.EnergyStart); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") { // o apelido é de outro id, como um bot da demo
			return "", errors.New("esse apelido já está em uso")
		}
		return "", err
	}
	Created(id)
	return id, nil
}

// DevUser: a identidade de quem entra no modo dev, pelo nome do SSH. Quem ainda não existe nasce
// com esse nome de apelido; se o apelido já tem dono, o primeiro acesso pede outro, como sempre.
func DevUser(name string) string {
	id := "dev:" + name
	if _, _, ok, _ := store.GetUser(id); !ok {
		if _, err := DevCreate(name); err != nil {
			log.Warn("dev", "nome", name, "err", err)
		}
	}
	return id
}

// DevAdjust muda uma coluna do painel dev em d passos: d negativo diminui, na cor troca de paleta,
// em online liga (d > 0) ou desliga. Os minutos viram tarefas dev de verdade, para a tabela de foco
// somar o mesmo que o pet mostra.
func DevAdjust(id string, f DevField, d int, now time.Time) error {
	devMu.Lock()
	defer devMu.Unlock()
	switch f { // os dois que só existem na memória das salas
	case DevOnline:
		setHere(id, d > 0)
		return nil
	case DevFocus:
		shiftFocus(id, d, now)
		return nil
	}
	// mu do começo ao fim, como no finish: uma tarefa encerrada no meio leria o pet antes e o
	// gravaria depois, desfazendo o ajuste (ou o ajuste desfaria a tarefa)
	mu.Lock()
	defer mu.Unlock()
	_, p, ok, err := store.GetUser(id)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("ninguém com esse id")
	}
	today := now.Format(time.DateOnly)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if p.Today != today {
		p.Today, p.TodayMin = today, 0
	}
	switch f {
	case DevEnergy:
		p.Energy = max(0, min(100, p.Energy+float64(d*DevEnergyStep))) // os dias parados continuam decaindo
	case DevIdle: // a última atividade anda um dia por passo, nunca para depois de agora
		if p.LastActive = p.LastActive.AddDate(0, 0, -d); p.LastActive.After(now) {
			p.LastActive = now
		}
	case DevStreak:
		p.Streak, p.StreakDay = max(0, p.Streak+d), today
	case DevColor:
		a := p.Look()
		variants := pet.Pets.Character(a.Character).Variants
		i := slices.IndexFunc(variants, func(v pet.CharacterVariant) bool { return v.ID == a.Variant })
		a.Variant = variants[((i+d)%len(variants)+len(variants))%len(variants)].ID
		p.Appearance, p.Color = a, pet.Pets.ArtworkColor(a)
	case DevToday:
		if d > 0 {
			start := now.Add(-DevTodayStep * time.Minute)
			if start.Before(midnight) {
				start = midnight
			}
			err = store.AddDevTask(id, start, DevTodayStep)
			p.TodayMin += DevTodayStep
		} else {
			var m float64
			m, err = store.DeleteDevTask(id, midnight, midnight.AddDate(0, 0, 1))
			if m == 0 { // sem tarefa dev hoje: tira dos minutos assim mesmo
				m = DevTodayStep
			}
			p.TodayMin = max(0, p.TodayMin-m)
		}
	case DevBefore:
		if d > 0 { // cada uma num dia mais para trás: ontem, anteontem...
			var n int
			n, err = store.DevTasks(id, time.Unix(0, 0), midnight)
			if err == nil {
				err = store.AddDevTask(id, midnight.AddDate(0, 0, -(n+1)).Add(12*time.Hour), DevBeforeStep)
			}
		} else {
			_, err = store.DeleteDevTask(id, time.Unix(0, 0), midnight)
		}
	}
	if err != nil {
		return err
	}
	if err := store.SetPet(id, p); err != nil {
		return err
	}
	if u := users[id]; u != nil {
		u.appearance = p.Look()
		u.energy, u.today, u.todayMin, u.color = p.EnergyAt(now), p.Today, p.TodayMin, p.Color
	}
	return nil
}

// setHere abre ou fecha a sessão de mentira de id, pelo mesmo caminho das de verdade: sem sessão
// nenhuma, a tarefa em andamento espera a tolerância de reconexão e é interrompida. Chame com devMu.
func setHere(id string, on bool) {
	if Dev {
		mu.Lock()
		defer mu.Unlock()
		u := load(id)
		v := on
		u.devOnline = &v
		devSaved.Presence[id] = on
		if err := store.SaveDevScene(nil, nil, nil, devSaved); err != nil {
			log.Error("presença dev", "err", err)
		}
		return
	}

	if on == devHere[id] {
		return
	}
	if on {
		devHere[id] = true
		Connect(id)
	} else {
		delete(devHere, id)
		Disconnect(id)
	}
}

// DevStart põe id para focar em task, na sala room, com uma sessão de mentira para a sala vê-la
// focando. Se já há uma tarefa em andamento, só troca o texto dela.
func DevStart(id, room, task string) {
	devMu.Lock()
	defer devMu.Unlock()
	devStart(id, room, task)
}

func devStart(id, room, task string) {
	setHere(id, true)
	mu.Lock()
	defer mu.Unlock()
	u := load(id)
	if u.focus != nil {
		u.focus.Task = task
		save(u)
		return
	}
	start(u, room, task, DevFocusMin, DevFocusMin*time.Minute)
}

// shiftFocus: d > 0 dá mais tempo à tarefa; d < 0 a faz ter começado antes, como se o tempo
// tivesse passado (até esgotar). Sem tarefa, o + começa uma. Chame com devMu.
func shiftFocus(id string, d int, now time.Time) {
	mu.Lock()
	u := load(id)
	f := u.focus
	if f == nil {
		mu.Unlock()
		if d > 0 {
			devStart(id, "", store.DevTask)
		}
		return
	}
	defer mu.Unlock()
	step := time.Duration(d*DevFocusStep) * time.Minute
	if d > 0 {
		if f.Expired(now) { // como o continuar: o tempo esgotado sem resposta não conta
			f.Counted, f.SegStart, f.Ends = f.Focused(now), now, now
		}
		f.Ends, f.Declared = f.Ends.Add(step), f.Declared+d*DevFocusStep
	} else {
		f.Start, f.SegStart, f.Ends = f.Start.Add(step), f.SegStart.Add(step), f.Ends.Add(step)
	}
	save(u)
}
