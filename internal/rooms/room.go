package rooms

// Foco individual e salas, em memória (ADR 0002 a 0004). Cada usuário tem no máximo uma tarefa em
// andamento, vista por todas as salas de que é membro; cada sala tem seus membros e seus eventos,
// com as reações. O banco guarda os membros e as tarefas encerradas; eventos e reações vivem só aqui.
// As telas não recebem eventos: cada sessão lê um retrato (World) a cada tick.

import (
	"errors"
	"fmt"
	"plantao/internal/pet"
	"plantao/internal/store"
	"slices"
	"strings"
	"sync"
	"time"

	"charm.land/log/v2"
)

const (
	DemoCode       = "demo"
	Reactions      = 4               // quantas reações existem (as teclas 1 a 4)
	MaxMembers     = 10              // teto de membros por sala (docs/regras.md)
	reconnectGrace = 2 * time.Minute // conexão instável não interrompe a tarefa
	maxEvents      = 20
)

var Demo bool // flag -demo: a sala demo, com bots

// ponytail: um lock só para salas e usuários, que dá para uma turma; um lock por sala se houver muitas.
// ponytail: foco e eventos só em memória, e um restart perde as tarefas em andamento; gravar a tarefa aberta no banco quando houver deploy.
var (
	mu        sync.Mutex
	rooms     = map[string]*room{} // por código
	users     = map[string]*user{} // quem é membro de alguma sala ou está conectado
	lastEvent int
)

type Status int

const (
	Offline Status = iota
	Idle           // conectado, sem tarefa: "parado"
	Focusing
)

// Focus é a tarefa em andamento. Os minutos contam até o fim do tempo declarado: estar ausente
// nunca rende mais do que o tempo que o próprio usuário declarou.
type Focus struct {
	Task                  string
	Start, SegStart, Ends time.Time // SegStart: início do trecho que está contando (muda ao continuar)
	Declared              int       // minutos declarados, somando os de "continuar"
	Counted               time.Duration
}

func (f Focus) Expired(now time.Time) bool { return !now.Before(f.Ends) }

// Focused: os minutos que contaram até now (param no fim do tempo declarado).
func (f Focus) Focused(now time.Time) time.Duration {
	if now.After(f.Ends) {
		now = f.Ends
	}
	return f.Counted + max(0, now.Sub(f.SegStart))
}

type user struct {
	id, nick string
	energy   float64
	focus    *Focus
	online   int       // sessões abertas
	gone     time.Time // quando a última sessão fechou
	bot      *bot
}

func (u *user) status() Status {
	switch {
	case u.focus != nil:
		return Focusing
	case u.online > 0:
		return Idle
	}
	return Offline
}

type Event struct {
	ID                    int
	Who, Nick, Verb, Task string
	At                    time.Time
	Reacts                [Reactions][]string // quem reagiu, por tipo
}

type room struct {
	code, name string
	members    []string
	events     []*Event // mais novo primeiro
}

// Load sobe as salas do banco, com os membros. Chamado uma vez, ao ligar o servidor.
func Load() error {
	rs, err := store.Rooms()
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	for _, r := range rs {
		ids, err := store.RoomMembers(r.Code)
		if err != nil {
			return err
		}
		rooms[r.Code] = &room{code: r.Code, name: r.Name, members: ids}
		for _, id := range ids {
			load(id)
		}
	}
	return nil
}

// load devolve o usuário em memória, lendo do banco na primeira vez. Chame com mu travado.
func load(id string) *user {
	if u := users[id]; u != nil {
		return u
	}
	nick, p, _, err := store.GetUser(id)
	if err != nil {
		log.Error("getUser", "err", err)
	}
	u := &user{id: id, nick: nick, energy: p.EnergyAt(time.Now())}
	users[id] = u
	return u
}

func Connect(id string) {
	mu.Lock()
	defer mu.Unlock()
	load(id).online++
}

// Disconnect: quando a última sessão fecha, a tarefa espera a tolerância de reconexão e,
// se o usuário não voltar, é interrompida com os minutos até a desconexão.
func Disconnect(id string) {
	mu.Lock()
	defer mu.Unlock()
	u := load(id)
	if u.online--; u.online > 0 {
		return
	}
	u.gone = time.Now()
	if f := u.focus; f != nil {
		time.AfterFunc(reconnectGrace, func() {
			mu.Lock()
			defer mu.Unlock()
			if u.online == 0 && u.focus == f {
				finish(u, pet.Interrupted, u.gone)
			}
		})
	}
}

// Created: o usuário acabou de ser criado no primeiro acesso (antes disso, load não achou ninguém).
func Created(id string) {
	mu.Lock()
	defer mu.Unlock()
	nick, p, _, err := store.GetUser(id)
	if err != nil {
		log.Error("getUser", "err", err)
	}
	u := load(id)
	u.nick, u.energy = nick, p.EnergyAt(time.Now())
}

// Start começa uma tarefa, se não há outra em andamento.
func Start(id, task string, minutes int) {
	mu.Lock()
	defer mu.Unlock()
	start(load(id), task, minutes, time.Duration(minutes)*time.Minute)
}

func start(u *user, task string, minutes int, d time.Duration) {
	if u.focus != nil {
		return
	}
	now := time.Now()
	u.focus = &Focus{Task: task, Start: now, SegStart: now, Ends: now.Add(d), Declared: minutes}
	event(u, "começou", task)
}

// Continue dá mais tempo à tarefa cujo tempo declarado esgotou.
func Continue(id string, minutes int) {
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	if f := load(id).focus; f != nil && f.Expired(now) {
		f.Counted, f.SegStart = f.Focused(now), now
		f.Ends, f.Declared = now.Add(time.Duration(minutes)*time.Minute), f.Declared+minutes
	}
}

// Finish encerra a tarefa em andamento: feito (Done), trocar ou parar (NotDone).
func Finish(id string, r pet.Result) {
	mu.Lock()
	defer mu.Unlock()
	if u := load(id); u.focus != nil {
		finish(u, r, time.Now())
	}
}

// finish grava a tarefa e atualiza o pet na hora. Chame com mu travado.
func finish(u *user, r pet.Result, at time.Time) {
	f := u.focus
	u.focus = nil
	p, err := store.RecordTask(u.id, store.TaskRow{Start: f.Start, End: at, Task: f.Task, DeclaredMin: f.Declared, Focused: f.Focused(at), Result: r})
	if err != nil {
		log.Error("recordTask", "err", err)
		return
	}
	u.energy = p.EnergyAt(at)
	if r == pet.Done {
		event(u, "cumpriu", f.Task)
	}
}

// event aparece em todas as salas de que u é membro. Chame com mu travado.
func event(u *user, verb, task string) {
	for _, r := range rooms {
		if !slices.Contains(r.members, u.id) {
			continue
		}
		lastEvent++
		e := &Event{ID: lastEvent, Who: u.id, Nick: u.nick, Verb: verb, Task: task, At: time.Now()}
		r.events = append([]*Event{e}, r.events...)
		if len(r.events) > maxEvents {
			r.events = r.events[:maxEvents]
		}
		if r.code == DemoCode {
			botsReactTo(r, e)
		}
	}
}

// React liga ou desliga a reação k de id a um evento. Ninguém reage ao próprio evento.
func React(id, code string, eventID, k int) {
	mu.Lock()
	defer mu.Unlock()
	react(id, code, eventID, k)
}

func react(id, code string, eventID, k int) {
	r := rooms[code]
	if r == nil || k < 0 || k >= Reactions || !slices.Contains(r.members, id) {
		return
	}
	for _, e := range r.events {
		if e.ID != eventID || e.Who == id {
			continue
		}
		if i := slices.Index(e.Reacts[k], id); i >= 0 {
			e.Reacts[k] = slices.Delete(e.Reacts[k], i, i+1)
		} else {
			e.Reacts[k] = append(e.Reacts[k], id)
		}
	}
}

// Create cria uma sala com id dentro e devolve o código.
func Create(id, name string) (string, error) {
	r, err := store.CreateRoom(id, name)
	if err != nil {
		return "", err
	}
	mu.Lock()
	defer mu.Unlock()
	rooms[r.Code] = &room{code: r.Code, name: r.Name, members: []string{id}}
	load(id)
	return r.Code, nil
}

// Join põe id na sala do código, até o teto de membros (a demo só existe em memória).
// O teto é conferido antes de gravar: o banco nunca ganha um membro que a sala recusa.
func Join(id, code string) error {
	mu.Lock()
	defer mu.Unlock()
	r := rooms[code]
	if r == nil {
		return fmt.Errorf("nenhuma sala com o código %q", code)
	}
	if slices.Contains(r.members, id) {
		return nil
	}
	if len(r.members) >= MaxMembers {
		return fmt.Errorf("%s já está cheia (%d membros)", r.name, MaxMembers)
	}
	if code != DemoCode {
		if err := store.JoinRoom(id, code); err != nil {
			return err
		}
	}
	r.members = append(r.members, id)
	load(id)
	return nil
}

// Leave tira id da sala; a sala some quando o último membro sai (ADR 0003).
func Leave(id, code string) error {
	if code == DemoCode {
		return errors.New("a sala demo não tem saída: ela some quando o servidor desliga")
	}
	if err := store.LeaveRoom(id, code); err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	if r := rooms[code]; r != nil {
		r.members = slices.DeleteFunc(r.members, func(m string) bool { return m == id })
		if len(r.members) == 0 {
			delete(rooms, code)
		}
	}
	return nil
}

// --- o retrato ---

// Member: um usuário como as telas o veem. Focus só vale com Status == Focusing.
type Member struct {
	ID, Nick string
	Energy   float64
	Status   Status
	Focus    Focus
}

type Snapshot struct {
	Code, Name string
	Members    []Member
	Events     []Event // cópias: seguras para ler fora do lock
}

// World: todas as salas e todos os usuários conhecidos, que cada sessão recebe no tick.
type World struct {
	Rooms map[string]Snapshot
	Users map[string]Member
}

func member(u *user) Member {
	m := Member{ID: u.id, Nick: u.nick, Energy: u.energy, Status: u.status()}
	if u.focus != nil {
		m.Focus = *u.focus
	}
	return m
}

func SnapshotWorld() World {
	mu.Lock()
	defer mu.Unlock()
	w := World{Rooms: map[string]Snapshot{}, Users: map[string]Member{}}
	for id, u := range users {
		w.Users[id] = member(u)
	}
	for code, r := range rooms {
		s := Snapshot{Code: code, Name: r.name}
		for _, id := range r.members {
			s.Members = append(s.Members, w.Users[id])
		}
		for _, e := range r.events {
			c := *e
			for k := range c.Reacts {
				c.Reacts[k] = slices.Clone(e.Reacts[k])
			}
			s.Events = append(s.Events, c)
		}
		w.Rooms[code] = s
	}
	return w
}

// MyRooms: os códigos das salas de que id é membro, em ordem de nome.
func (w World) MyRooms(id string) []string {
	var out []string
	for code, s := range w.Rooms {
		if slices.ContainsFunc(s.Members, func(m Member) bool { return m.ID == id }) {
			out = append(out, code)
		}
	}
	slices.SortFunc(out, func(a, b string) int { return strings.Compare(w.Rooms[a].Name+a, w.Rooms[b].Name+b) })
	return out
}
