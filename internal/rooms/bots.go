package rooms

// Bots da sala demo (flag -demo): os mesmos 5 de sempre, com os humores iniciais de sempre,
// focando em tarefas de cerca de 1 min e reagindo a quem está na sala.

import (
	"math/rand/v2"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/store"
	"strings"
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
	for i, b := range bots {
		if err := putBot(b); err != nil {
			log.Fatal("seed bots", "err", err)
		}
		// cada bot numa cor, a partir da segunda: quem entra fica com a primeira, e a obra mostra
		// um tom (e, na orquestra, um instrumento) por pessoa
		if err := store.SetPetColor("bot:"+b.nick, pet.Colors[(i+1)%len(pet.Colors)].ID); err != nil {
			log.Fatal("seed bots", "err", err)
		}
		// os minutos de hoje saem da energia: é a escada que a demo precisa para a obra mostrar
		// os quatro estágios ao mesmo tempo. Vai no banco, senão a primeira tarefa que o bot
		// encerra lê de lá o zero e a obra esvazia enquanto a plateia olha.
		if err := store.SetToday("bot:"+b.nick, time.Now().Format(time.DateOnly), b.energy); err != nil {
			log.Fatal("seed bots", "err", err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	// a demo é a vitrine: ela já chega com uma obra, senão a plateia não vê que existe
	r := &room{code: DemoCode, name: "sala demo", max: MaxMembers,
		obra: "trem", meta: store.DefaultObraMeta}
	rooms[DemoCode] = r
	for i, b := range bots {
		u := load("bot:" + b.nick)
		u.bot = b
		r.members = append(r.members, u.id)
		if !b.offline {
			u.online = 1
		}
		if i != 2 && !b.offline { // caio começa sem tarefa
			start(u, DemoCode, b.tasks[0], 1, 20*time.Second+rand.N(60*time.Second))
		}
	}
	go runBots()
}

// putBot grava o bot. Se o apelido já é de alguém (no modo dev, a ana do exemplo; em qualquer
// banco, quem escolheu o nome antes da primeira demo), o bot fica com "<apelido>-bot".
func putBot(b *bot) error {
	err := store.PutUser("bot:"+b.nick, b.nick, b.energy)
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		err = store.PutUser("bot:"+b.nick, b.nick+"-bot", b.energy)
	}
	return err
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
				start(u, DemoCode, b.tasks[rand.N(len(b.tasks))], 1, 40*time.Second+rand.N(50*time.Second))
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
