package rooms

// Salas em memória: cada uma tem fila, sprint atual, chat e o relógio que troca as fases.
// A pública existe sempre (com bots na demo); as privadas vêm do SQLite e se entra por código.
// As telas não recebem eventos: cada sessão lê um retrato (World) a cada tick.

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"plantao/internal/pet"
	"plantao/internal/store"
	"slices"
	"sync"
	"time"

	"charm.land/log/v2"
)

const PublicCode = "publica"

var (
	Demo     bool // tempo comprimido (flag -demo)
	roomsMu  sync.Mutex
	rooms    = map[string]*Room{} // por código
	onlineMu sync.Mutex
	online   = map[string]int{} // sessões abertas por usuário
)

type Phase int

const (
	Lobby Phase = iota // esperando o próximo sprint
	Focus
	Break // check-out + chat
)

type timing struct {
	focus, pause time.Duration
	next         func(after time.Time) time.Time // início do próximo sprint
}

func newTiming(slots []int) timing {
	if Demo {
		return timing{time.Minute, 40 * time.Second, func(t time.Time) time.Time { return t.Add(20 * time.Second) }}
	}
	return timing{25 * time.Minute, 5 * time.Minute, func(t time.Time) time.Time { return nextSlot(slots, t) }}
}

// nextSlot: o primeiro horário cheio (slots em ordem crescente) a partir de after.
func nextSlot(slots []int, after time.Time) time.Time {
	y, mo, d := after.Date()
	for _, h := range slots {
		if t := time.Date(y, mo, d, h, 0, 0, 0, after.Location()); !t.Before(after) {
			return t
		}
	}
	return time.Date(y, mo, d+1, slots[0], 0, 0, 0, after.Location())
}

type Member struct {
	ID, Nick, Lock string
	Energy         float64
	bot            *bot
	Result         pet.Result // "" enquanto foca ou aguarda o check-out
}

type Room struct {
	code, name string
	mu         sync.Mutex
	t          timing
	phase      Phase
	ends       time.Time // fim da fase atual (no lobby: início do sprint)
	start      time.Time // início do sprint atual
	queue      []*Member // fila do próximo sprint
	room       []*Member // sprint atual
	chat       []string
	events     []string // atividade da sala
}

// Open liga a sala em memória (uma vez só) e devolve ela.
func Open(code, name string, slots []int) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	if h := rooms[code]; h != nil {
		return h
	}
	h := &Room{code: code, name: name, t: newTiming(slots)}
	h.ends = h.t.next(time.Now())
	rooms[code] = h
	h.botsJoin()
	go h.run()
	return h
}

func Get(code string) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	return rooms[code]
}

func allRooms() []*Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	return slices.Collect(maps.Values(rooms))
}

// JoinQueue entra na fila de uma sala e sai de qualquer outra: uma fila por pessoa.
func JoinQueue(code, id, nick, lock string) {
	for _, h := range allRooms() {
		if h.code != code {
			h.Cancel(id)
		}
	}
	if h := Get(code); h != nil {
		h.join(id, nick, lock, nil)
	}
}

func Connect(id string) {
	onlineMu.Lock()
	defer onlineMu.Unlock()
	online[id]++
}

// Disconnect: quando a última sessão da pessoa fecha, ela sai de todas as salas.
func Disconnect(id string) {
	onlineMu.Lock()
	online[id]--
	last := online[id] <= 0
	if last {
		delete(online, id)
	}
	onlineMu.Unlock()
	if last {
		for _, h := range allRooms() {
			h.leave(id)
		}
	}
}

func (h *Room) run() {
	for {
		h.mu.Lock()
		wait := time.Until(h.ends)
		h.mu.Unlock()
		time.Sleep(wait)
		h.advance()
	}
}

func (h *Room) advance() {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	switch h.phase {
	case Lobby:
		h.phase, h.start, h.ends = Focus, now, now.Add(h.t.focus)
		h.room, h.queue, h.events = h.queue, nil, nil
		h.event(fmt.Sprintf("sprint começou com %d pessoas", len(h.room)))
		for _, m := range h.room {
			if m.bot != nil && rand.Float64() < m.bot.pQuit {
				time.AfterFunc(rand.N(h.t.focus), func() { h.Abandon(m.ID) })
			}
		}
	case Focus:
		h.phase, h.ends = Break, now.Add(h.t.pause)
		h.event("fim do foco: check-out e chat abertos")
		for _, m := range h.room {
			if m.bot != nil && m.Result == "" {
				h.botBreak(m)
			}
		}
	case Break:
		for _, m := range h.room {
			if m.Result == "" { // não respondeu o check-out
				h.finish(m, pet.NotDone)
			}
		}
		h.phase, h.ends = Lobby, h.t.next(now)
		h.room, h.chat = nil, nil
		h.botsJoin()
	}
}

// nextStart é o início do próximo sprint, em qualquer fase.
func (h *Room) nextStart() time.Time {
	switch h.phase {
	case Focus:
		return h.t.next(h.ends.Add(h.t.pause))
	case Break:
		return h.t.next(h.ends)
	}
	return h.ends
}

func (h *Room) event(s string) {
	h.events = append(h.events, time.Now().Format("15:04")+"  "+s)
	if len(h.events) > 5 {
		h.events = h.events[1:]
	}
}

// finish fecha a participação de m, grava no SQLite e atualiza o pet na hora.
func (h *Room) finish(m *Member, r pet.Result) {
	m.Result = r
	p, err := store.Record(m.ID, h.start, m.Lock, r)
	if err != nil {
		log.Error("record", "err", err)
		return
	}
	m.Energy = p.EnergyAt(time.Now())
}

func Find(ms []*Member, id string) *Member {
	if i := slices.IndexFunc(ms, func(m *Member) bool { return m.ID == id }); i >= 0 {
		return ms[i]
	}
	return nil
}

func (h *Room) join(id, nick, lock string, b *bot) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if Find(h.queue, id) != nil {
		return
	}
	_, p, _, err := store.GetUser(id)
	if err != nil {
		log.Error("getUser", "err", err)
	}
	h.queue = append(h.queue, &Member{ID: id, Nick: nick, Lock: lock, bot: b, Energy: p.EnergyAt(time.Now())})
	h.event(nick + " entrou na fila")
}

func (h *Room) Cancel(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.queue = slices.DeleteFunc(h.queue, func(m *Member) bool { return m.ID == id })
}

func (h *Room) Abandon(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := Find(h.room, id); h.phase == Focus && m != nil && m.Result == "" {
		h.finish(m, pet.Abandoned)
		h.event(m.Nick + " abandonou o sprint")
	}
}

func (h *Room) Checkout(id string, r pet.Result) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := Find(h.room, id); h.phase == Break && m != nil && m.Result == "" {
		h.finish(m, r)
	}
}

func (h *Room) Say(id, text string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := Find(h.room, id); h.phase == Break && m != nil && text != "" {
		h.chat = append(h.chat, m.Nick+": "+text)
	}
}

// leave: quem estava na fila faltou; quem estava focando abandonou.
func (h *Room) leave(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := Find(h.queue, id); m != nil {
		h.queue = slices.DeleteFunc(h.queue, func(x *Member) bool { return x == m })
		if _, err := store.Record(id, h.nextStart(), m.Lock, pet.Missed); err != nil {
			log.Error("record", "err", err)
		}
		h.event(m.Nick + " saiu da fila (falta)")
	}
	if m := Find(h.room, id); h.phase == Focus && m != nil && m.Result == "" {
		h.finish(m, pet.Abandoned)
		h.event(m.Nick + " saiu no meio do foco")
	}
}

type Snapshot struct {
	Code, Name   string
	Phase        Phase
	Ends, Start  time.Time
	Next         time.Time
	Queue, Room  []*Member // cópias: seguras para ler fora do lock
	Chat, Events []string
}

func (h *Room) snapshot() Snapshot {
	h.mu.Lock()
	defer h.mu.Unlock()
	cp := func(ms []*Member) (out []*Member) {
		for _, m := range ms {
			c := *m
			out = append(out, &c)
		}
		return out
	}
	return Snapshot{h.code, h.name, h.phase, h.ends, h.start, h.nextStart(), cp(h.queue), cp(h.room),
		slices.Clone(h.chat), slices.Clone(h.events)}
}

// World: o retrato de todas as salas e de quem está online, que cada sessão recebe no tick.
type World struct {
	Rooms  map[string]Snapshot // por código
	Online map[string]bool
}

func SnapshotWorld() World {
	w := World{Rooms: map[string]Snapshot{}, Online: map[string]bool{}}
	for _, h := range allRooms() {
		w.Rooms[h.code] = h.snapshot()
	}
	onlineMu.Lock()
	for id := range online {
		w.Online[id] = true
	}
	onlineMu.Unlock()
	return w
}

// --- bots da demo (só na sala pública) ---

type bot struct {
	nick         string
	energy       float64 // humor inicial, para a sala ter todos os estados
	pDone, pQuit float64
	locks        []string
}

var bots = []*bot{
	{"ana", 92, 0.9, 0, []string{"terminar a lista 3 de Cálculo", "revisar a prova de Álgebra"}},
	{"bruno", 68, 0.7, 0, []string{"revisar o PR do login", "escrever os testes do parser"}},
	{"caio", 48, 0.5, 0.05, []string{"ler o cap. 4 de Sistemas Operacionais", "resumir o artigo de IHC"}},
	{"duda", 27, 0.35, 0.2, []string{"começar o relatório de Física", "configurar o Docker do projeto"}},
	{"edu", 8, 0.15, 0.4, []string{"instalar o ambiente do TCC", "abrir o projeto do TCC"}},
}

var botLines = []string{"alguém mais travou no exercício 4?", "café e volta", "fechei o PR, finalmente",
	"bora pro próximo", "esse rendeu", "minha cabeça fritou", "alguém tem o link da lista?"}

func SeedBots() {
	for _, b := range bots {
		if err := store.PutUser("bot:"+b.nick, b.nick, b.energy); err != nil {
			log.Fatal("seed bots", "err", err)
		}
	}
}

// botsJoin espalha a entrada dos bots pela contagem regressiva, para a fila ir enchendo.
func (h *Room) botsJoin() {
	if !Demo || h.code != PublicCode {
		return
	}
	window := time.Until(h.ends) * 8 / 10
	for _, b := range bots {
		time.AfterFunc(rand.N(window+1), func() { h.join("bot:"+b.nick, b.nick, b.locks[rand.N(len(b.locks))], b) })
	}
}

func (h *Room) botBreak(m *Member) {
	r := pet.NotDone
	if rand.Float64() < m.bot.pDone {
		r = pet.Done
	}
	time.AfterFunc(time.Second+rand.N(5*time.Second), func() { h.Checkout(m.ID, r) })
	if rand.N(2) == 0 {
		time.AfterFunc(3*time.Second+rand.N(h.t.pause/2), func() { h.Say(m.ID, botLines[rand.N(len(botLines))]) })
	}
}
