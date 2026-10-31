package rooms

// Bots da sala demo (flag -demo): os mesmos 5 de sempre, com os humores iniciais de sempre,
// focando em tarefas de cerca de 1 min e reagindo a quem está na sala.

import (
	"math/rand/v2"
	"plantao/internal/pet"
	"plantao/internal/store"
	"slices"
	"time"

	"charm.land/log/v2"
)

type bot struct {
	nick          string
	energy, pDone float64
	tasks         []string
	offline       bool // começa offline: a sala mostra os três status
}

var bots = []*bot{
	{"ana", 92, 0.9, []string{"terminar a lista 3 de Cálculo", "revisar a prova de Álgebra"}, false},
	{"bruno", 68, 0.7, []string{"revisar o PR do login", "escrever os testes do parser"}, false},
	{"caio", 48, 0.5, []string{"ler o cap. 4 de Sistemas Operacionais", "resumir o artigo de IHC"}, false},
	{"duda", 27, 0.35, []string{"começar o relatório de Física", "configurar o Docker do projeto"}, false},
	{"edu", 8, 0.15, []string{"instalar o ambiente do TCC", "abrir o projeto do TCC"}, true},
}

type botReaction struct {
	at         time.Time
	who, code  string
	eventID, k int
}

var pending []botReaction // reações agendadas pelos bots

// SeedBots cria a sala demo com os bots, alguns já focando, e põe os bots para andar.
func SeedBots() {
	for _, b := range bots {
		if err := store.PutUser("bot:"+b.nick, b.nick, b.energy); err != nil {
			log.Fatal("seed bots", "err", err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	r := &room{code: DemoCode, name: "sala demo"}
	rooms[DemoCode] = r
	for i, b := range bots {
		u := load("bot:" + b.nick)
		u.bot = b
		r.members = append(r.members, u.id)
		if !b.offline {
			u.online = 1
		}
		if i != 2 && !b.offline { // caio começa sem tarefa
			start(u, b.tasks[0], 1, 20*time.Second+rand.N(60*time.Second))
		}
	}
	go runBots()
}

func runBots() {
	for range time.Tick(time.Second) {
		mu.Lock()
		now := time.Now()
		for _, u := range users {
			b := u.bot
			switch {
			case b == nil:
			case u.focus != nil && u.focus.Expired(now):
				r := pet.NotDone
				if rand.Float64() < b.pDone {
					r = pet.Done
				}
				finish(u, r, now)
			case u.focus == nil && u.online > 0 && rand.N(90) == 0: // uns 90 s parado entre uma tarefa e outra: a sala não fica barulhenta
				start(u, b.tasks[rand.N(len(b.tasks))], 1, 40*time.Second+rand.N(50*time.Second))
			case u.focus == nil && rand.N(200) == 0:
				u.online = 1 - u.online // sem tarefa <-> offline
			}
		}
		rest := pending[:0]
		for _, p := range pending {
			if now.Before(p.at) {
				rest = append(rest, p)
			} else if r := rooms[p.code]; r != nil {
				if i := slices.IndexFunc(r.events, func(e *Event) bool { return e.ID == p.eventID }); i >= 0 && !slices.Contains(r.events[i].Reacts[p.k], p.who) {
					react(p.who, p.code, p.eventID, p.k)
				}
			}
		}
		pending = rest
		mu.Unlock()
	}
}

// botsReactTo agenda reações dos bots a um evento: sempre a quem não é bot, às vezes entre eles.
func botsReactTo(r *room, e *Event) {
	n := 1 + rand.N(3)
	if users[e.Who] != nil && users[e.Who].bot != nil {
		if rand.N(10) >= 3 {
			return
		}
		n = 1
	}
	for range n {
		b := bots[rand.N(len(bots))]
		if "bot:"+b.nick == e.Who {
			continue
		}
		pending = append(pending, botReaction{time.Now().Add(2*time.Second + rand.N(6*time.Second)), "bot:" + b.nick, r.code, e.ID, rand.N(Reactions)})
	}
}
