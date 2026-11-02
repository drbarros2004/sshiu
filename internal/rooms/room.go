package rooms

// Foco individual e salas, em memória (ADR 0002 a 0004). Cada usuário tem no máximo uma tarefa em
// andamento, vista por todas as salas de que é membro; cada sala tem seus membros e seus eventos,
// com as reações. O banco guarda os membros, as tarefas encerradas e a em andamento (para sobreviver a
// um restart); eventos e reações vivem só aqui, e um restart os perde.
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
	DemoCode   = "demo"
	Reactions  = 4  // quantas reações existem (as teclas 1 a 4)
	MaxMembers = 10 // maior sala que dá para escolher (docs/regras.md)
	MinMembers = 1  // a sala solo, para focar sozinho (ADR 0006)
	ObraStages = 4  // estágios de cada peça, contando o vazio: um desenho cada (ADR 0008)
	maxEvents  = 20
)

var reconnectGrace = 2 * time.Minute // conexão instável (ou um restart) não interrompe a tarefa; var para o teste

var Demo bool // flag -demo: a sala demo, com bots

// Obras: os temas que a sala pode construir, na ordem em que o dono escolhe ("" = sem obra).
// O desenho de cada peça está na interface; aqui só mora o nome (ADR 0008).
var Obras = []string{"", "trem", "cidade", "horta", "estante", "feira", "acampamento", "orquestra",
	"passarinhos", "porto", "foguetes"}

// ObraMetas: as metas que dá para escolher, em minutos focados por pessoa (docs/regras.md).
var ObraMetas = []int{30, 60, 90, 120, 180}

// ponytail: um lock só para salas e usuários, que dá para uma turma; um lock por sala se houver muitas.
var (
	mu        sync.Mutex
	rooms     = map[string]*room{} // por código
	users     = map[string]*user{} // quem é membro de alguma sala ou está conectado
	lastEvent int
)

type Status int

const (
	Offline Status = iota
	Idle           // conectado, sem tarefa: "sem tarefa"
	Focusing
)

// Focus é a tarefa em andamento. Os minutos contam até o fim do tempo declarado: estar ausente
// nunca rende mais do que o tempo que o próprio usuário declarou.
type Focus struct {
	Task                  string
	Room                  string    // a sala em que a tarefa começou, para o histórico
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
	color    string // a cor escolhida para o pet ("" = a cor sai do humor)
	energy   float64
	today    string  // dia de todayMin (YYYY-MM-DD), como o pet guarda
	todayMin float64 // minutos focados hoje, sem a tarefa em andamento
	focus    *Focus
	online   int       // sessões abertas
	gone     time.Time // quando a última sessão fechou
	bot      *bot
}

// minutesToday: os minutos focados hoje, com a tarefa em andamento incluída. É o que faz a peça
// da obra crescer, e vale em todas as salas: os minutos são da pessoa, não da sala (ADR 0008).
func (u *user) minutesToday(now time.Time) float64 {
	m := 0.0
	if u.today == now.Format(time.DateOnly) {
		m = u.todayMin
	}
	if u.focus != nil {
		m += u.focus.Focused(now).Minutes()
	}
	return m
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
	owner      string // quem muda a sala (ADR 0007); a demo não tem dono
	max        int    // membros que a sala aceita, escolhido na criação e mudado pelo dono
	closed     bool   // fechada para novas entradas, mesmo com vaga
	obra       string // tema da obra ("" = sem obra, ADR 0008)
	meta       int    // minutos focados que completam uma peça
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
		rooms[r.Code] = &room{code: r.Code, name: r.Name, owner: r.Owner, max: ClampSize(r.Max),
			closed: r.Closed, obra: r.Obra, meta: r.ObraMeta, members: ids}
		for _, id := range ids {
			load(id)
		}
	}
	// o restart conta como desconexão de todo mundo, agora: quem volta a tempo continua a tarefa.
	open, err := store.OpenTasks()
	if err != nil {
		return err
	}
	now := time.Now()
	for _, t := range open {
		u := load(t.User)
		u.focus = &Focus{Task: t.Task, Room: t.Room, Start: t.Start, SegStart: t.SegStart, Ends: t.Ends,
			Declared: t.Declared, Counted: t.Counted}
		u.gone = now
		awaitReturn(u)
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
	u := &user{id: id, nick: nick, color: p.Color, energy: p.EnergyAt(time.Now()),
		today: p.Today, todayMin: p.TodayMin}
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
	awaitReturn(u)
}

// awaitReturn: se u não voltar dentro da tolerância, a tarefa é interrompida com os minutos até u.gone.
// Chame com mu travado.
func awaitReturn(u *user) {
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

// SetColor: a pessoa trocou a cor do pet. Quem guarda é o banco; aqui é só o retrato que as
// salas mostram, que ficaria com a cor velha até a próxima tarefa encerrada.
func SetColor(id, hex string) {
	mu.Lock()
	defer mu.Unlock()
	load(id).color = hex
}

// Start começa uma tarefa, se não há outra em andamento. A sala é a que estava na tela: a tarefa
// aparece em todas as suas salas (ADR 0002), mas o histórico guarda onde ela começou.
func Start(id, room, task string, minutes int) {
	mu.Lock()
	defer mu.Unlock()
	start(load(id), room, task, minutes, time.Duration(minutes)*time.Minute)
}

func start(u *user, room, task string, minutes int, d time.Duration) {
	if u.focus != nil {
		return
	}
	now := time.Now()
	u.focus = &Focus{Task: task, Room: room, Start: now, SegStart: now, Ends: now.Add(d), Declared: minutes}
	save(u)
	event(u, "começou", task)
}

// save grava a tarefa em andamento, para ela sobreviver a um restart. Os bots não gravam: a demo
// some com o servidor. Chame com mu travado.
func save(u *user) {
	if f := u.focus; f != nil && u.bot == nil {
		if err := store.SaveOpenTask(store.OpenTask{User: u.id, Task: f.Task, Room: f.Room, Start: f.Start,
			SegStart: f.SegStart, Ends: f.Ends, Declared: f.Declared, Counted: f.Counted}); err != nil {
			log.Error("saveOpenTask", "err", err)
		}
	}
}

// Continue dá mais tempo à tarefa cujo tempo declarado esgotou.
func Continue(id string, minutes int) {
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	u := load(id)
	if f := u.focus; f != nil && f.Expired(now) {
		f.Counted, f.SegStart = f.Focused(now), now
		f.Ends, f.Declared = now.Add(time.Duration(minutes)*time.Minute), f.Declared+minutes
		save(u)
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
	p, err := store.RecordTask(u.id, store.TaskRow{Start: f.Start, End: at, Task: f.Task,
		DeclaredMin: f.Declared, Focused: f.Focused(at), Result: r, Room: f.Room})
	if err != nil {
		log.Error("recordTask", "err", err)
		return
	}
	u.energy, u.today, u.todayMin = p.EnergyAt(at), p.Today, p.TodayMin
	u.color = p.Color
	if r == pet.Done {
		event(u, "cumpriu", f.Task)
	}
}

// event aparece em todas as salas de que u é membro (a tarefa é a mesma em todas).
// Chame com mu travado.
func event(u *user, verb, task string) {
	for _, r := range rooms {
		if slices.Contains(r.members, u.id) {
			roomEvent(r, u, verb, task)
		}
	}
}

// roomEvent põe um evento numa sala só: entrar e sair acontecem numa sala, não em todas.
// Chame com mu travado.
func roomEvent(r *room, u *user, verb, task string) {
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

// owned: a sala do código, se id for o dono dela. Chame com mu travado.
func owned(id, code string) (*room, error) {
	r := rooms[code]
	if r == nil {
		return nil, fmt.Errorf("nenhuma sala com o código %q", code)
	}
	if r.owner != id {
		return nil, errors.New("só o dono muda a sala")
	}
	return r, nil
}

// saveRoom grava a sala como ela está agora. A demo só existe em memória. Chame com mu travado.
func saveRoom(r *room) {
	if r.code == DemoCode {
		return
	}
	if err := store.SetRoom(store.RoomRow{Code: r.code, Name: r.name, Owner: r.owner,
		Max: r.max, Closed: r.closed, Obra: r.obra, ObraMeta: r.meta}); err != nil {
		log.Error("setRoom", "err", err)
	}
}

// Rename troca o nome da sala.
func Rename(id, code, name string) error {
	mu.Lock()
	defer mu.Unlock()
	r, err := owned(id, code)
	if err != nil {
		return err
	}
	if name = strings.TrimSpace(name); name == "" {
		return errors.New("a sala precisa de um nome")
	}
	r.name = name
	saveRoom(r)
	return nil
}

// Resize muda quantos membros a sala aceita, de quantos já estão lá até o teto de todas: ninguém
// é expulso por uma mudança de tamanho, então o chão é a sala como ela está (ADR 0007).
func Resize(id, code string, n int) error {
	mu.Lock()
	defer mu.Unlock()
	r, err := owned(id, code)
	if err != nil {
		return err
	}
	r.max = min(max(n, len(r.members), MinMembers), MaxMembers)
	saveRoom(r)
	return nil
}

// SetClosed fecha ou abre a sala para novas entradas. Quem já é membro não é afetado.
func SetClosed(id, code string, closed bool) error {
	mu.Lock()
	defer mu.Unlock()
	r, err := owned(id, code)
	if err != nil {
		return err
	}
	r.closed = closed
	saveRoom(r)
	return nil
}

// Kick tira um membro da sala e abre a vaga. Não é bloqueio: com o código a pessoa entra de novo.
func Kick(id, code, who string) error {
	mu.Lock()
	defer mu.Unlock()
	r, err := owned(id, code)
	if err != nil {
		return err
	}
	if who == id {
		return errors.New("para sair da sala, use sair")
	}
	if !slices.Contains(r.members, who) {
		return errors.New("essa pessoa não está na sala")
	}
	if code != DemoCode {
		if err := store.LeaveRoom(who, code); err != nil {
			return err
		}
	}
	r.members = slices.DeleteFunc(r.members, func(m string) bool { return m == who })
	roomEvent(r, load(who), "perdeu a vaga", "")
	return nil
}

// Transfer passa a sala para outro membro. Quem passa continua na sala, como os outros.
func Transfer(id, code, to string) error {
	mu.Lock()
	defer mu.Unlock()
	r, err := owned(id, code)
	if err != nil {
		return err
	}
	if !slices.Contains(r.members, to) {
		return errors.New("essa pessoa não está na sala")
	}
	r.owner = to
	saveRoom(r)
	roomEvent(r, load(to), "ficou com a sala", "")
	return nil
}

// SetObra escolhe o que a sala constrói hoje, ou "" para não construir nada. A sala solo não tem
// obra: uma peça sozinha não é um conjunto (ADR 0008).
func SetObra(id, code, obra string) error {
	mu.Lock()
	defer mu.Unlock()
	r, err := owned(id, code)
	if err != nil {
		return err
	}
	if !slices.Contains(Obras, obra) {
		return fmt.Errorf("não existe obra %q", obra)
	}
	if obra != "" && r.max == MinMembers {
		return errors.New("a sala solo não tem obra")
	}
	r.obra = obra
	saveRoom(r)
	return nil
}

// SetMeta escolhe quantos minutos focados completam uma peça.
func SetMeta(id, code string, minutes int) error {
	mu.Lock()
	defer mu.Unlock()
	r, err := owned(id, code)
	if err != nil {
		return err
	}
	if !slices.Contains(ObraMetas, minutes) {
		return fmt.Errorf("meta fora da lista: %d min", minutes)
	}
	r.meta = minutes
	saveRoom(r)
	return nil
}

// Dissolve encerra a sala: ela some para todo mundo. As tarefas ficam no histórico de cada um.
func Dissolve(id, code string) error {
	mu.Lock()
	defer mu.Unlock()
	if _, err := owned(id, code); err != nil { // a demo não tem dono: ninguém a encerra
		return err
	}
	if err := store.DeleteRoom(code); err != nil {
		return err
	}
	delete(rooms, code)
	return nil
}

// Stage: em que estágio está a peça de quem focou mins minutos numa sala de meta minutos, de 0
// (vazia) ao último desenho. Os estágios dividem a meta em partes iguais (ADR 0008).
func Stage(mins float64, meta int) int {
	last := ObraStages - 1
	if meta <= 0 {
		return 0
	}
	return min(int(mins*float64(last)/float64(meta)), last)
}

// ClampSize: o tamanho que a sala pode ter. Quem cria escolhe, dentro do teto de todas.
func ClampSize(n int) int { return max(MinMembers, min(n, MaxMembers)) }

// Create cria uma sala com id dentro, do tamanho escolhido, e devolve o código.
func Create(id, name string, size int) (string, error) {
	size = ClampSize(size)
	r, err := store.CreateRoom(id, name, size)
	if err != nil {
		return "", err
	}
	mu.Lock()
	defer mu.Unlock()
	// a sala solo nasce fechada: o código existe como em qualquer sala, e quem aumentar o tamanho
	// depois vê o código aparecer e decide quando abrir a porta (ADR 0006 e 0007)
	n := &room{code: r.Code, name: r.Name, owner: id, max: size, closed: size == MinMembers,
		meta: store.DefaultObraMeta, members: []string{id}}
	rooms[r.Code] = n
	if n.closed {
		saveRoom(n)
	}
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
	if len(r.members) >= r.max {
		return fmt.Errorf("%s já está cheia (%d membros)", r.name, r.max)
	}
	if r.closed {
		return fmt.Errorf("%s não está aceitando mais ninguém", r.name)
	}
	if code != DemoCode {
		if err := store.JoinRoom(id, code); err != nil {
			return err
		}
	}
	r.members = append(r.members, id)
	roomEvent(r, load(id), "entrou", "")
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
			delete(rooms, code) // o último saiu: a sala e os eventos somem com ela
		} else {
			roomEvent(r, load(id), "saiu", "")
			if r.owner == id { // a sala nunca fica órfã: vai para o membro mais antigo (ADR 0007)
				r.owner = r.members[0]
				saveRoom(r)
				roomEvent(r, load(r.owner), "ficou com a sala", "")
			}
		}
	}
	return nil
}

// --- o retrato ---

// Member: um usuário como as telas o veem. Focus só vale com Status == Focusing.
type Member struct {
	ID, Nick string
	Color    string // a cor escolhida para o pet
	Energy   float64
	TodayMin float64 // minutos focados hoje, com a tarefa em andamento: a peça da obra
	Status   Status
	Focus    Focus
}

type Snapshot struct {
	Code, Name string
	Owner      string // o dono da sala (ADR 0007); "" na demo
	Max        int    // membros que a sala aceita
	Closed     bool   // não aceita mais ninguém
	Obra       string // tema da obra ("" = sem obra, ADR 0008)
	Meta       int    // minutos focados que completam uma peça
	Members    []Member
	Events     []Event // cópias: seguras para ler fora do lock
}

// World: todas as salas e todos os usuários conhecidos, que cada sessão recebe no tick.
type World struct {
	Rooms map[string]Snapshot
	Users map[string]Member
}

func member(u *user) Member {
	m := Member{ID: u.id, Nick: u.nick, Color: u.color, Energy: u.energy, Status: u.status(),
		TodayMin: u.minutesToday(time.Now())}
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
		s := Snapshot{Code: code, Name: r.name, Owner: r.owner, Max: r.max, Closed: r.closed,
			Obra: r.obra, Meta: r.meta}
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
