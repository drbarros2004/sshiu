package rooms

import (
	"fmt"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/store"
	"strings"
	"testing"
	"time"
)

func TestFocused(t *testing.T) {
	t0 := time.Date(2026, 10, 29, 14, 0, 0, 0, time.Local)
	f := Focus{Start: t0, SegStart: t0, Ends: t0.Add(25 * time.Minute), Declared: 25}
	if got := f.Focused(t0.Add(10 * time.Minute)); got != 10*time.Minute || f.Expired(t0.Add(10*time.Minute)) {
		t.Fatalf("no meio: %v", got)
	}
	late := t0.Add(40 * time.Minute) // esgotou e ninguém respondeu: os minutos param no tempo declarado
	if got := f.Focused(late); got != 25*time.Minute || !f.Expired(late) {
		t.Fatalf("esgotado: %v", got)
	}
	// continuar aos 40 min, por mais 10: os 15 sem resposta não contam
	f.Counted, f.SegStart, f.Ends = f.Focused(late), late, late.Add(10*time.Minute)
	if got := f.Focused(late.Add(5 * time.Minute)); got != 30*time.Minute {
		t.Fatalf("continuou: %v", got)
	}
}

// A virada descarta minutos encerrados ontem e preserva a tarefa em andamento (ADR 0008).
func TestObraMidnight(t *testing.T) {
	before := time.Date(2026, 11, 4, 23, 59, 59, 0, time.FixedZone("servidor", -3*60*60))
	after := before.Add(time.Second)
	u := user{today: before.Format(time.DateOnly), todayMin: 60}
	if Stage(u.minutesToday(before), 60) != ObraStages-1 || u.minutesToday(after) != 0 {
		t.Fatal("os minutos do dia anterior não foram descartados na virada")
	}
	start := after.Add(-10 * time.Minute)
	u.focus = &Focus{Start: start, SegStart: start, Ends: after.Add(15 * time.Minute), Declared: 25}
	if got := u.minutesToday(after); got != 10 || Stage(got, 60) != 0 {
		t.Fatalf("a tarefa atravessando a meia-noite não foi preservada: %v", got)
	}
	if got := u.minutesToday(after.Add(10 * time.Minute)); got != 20 || Stage(got, 60) != 1 {
		t.Fatalf("a obra não continuou crescendo no novo dia: %v", got)
	}
	if u.focus.Start != start || u.focus.Declared != 25 {
		t.Fatal("a virada alterou a tarefa")
	}
}

func TestReact(t *testing.T) {
	rooms["t"] = &room{code: "t", members: []string{"ana", "bia"}, events: []*Event{{ID: 1, Who: "ana"}}}
	defer delete(rooms, "t")
	react("bia", "t", 1, 0)
	react("ana", "t", 1, 0) // a mesma reação de outra pessoa soma, inclusive a de quem fez o evento
	react("eve", "t", 1, 0) // quem não é da sala, não
	if got := rooms["t"].events[0].Reacts[0]; len(got) != 2 || got[0] != "bia" || got[1] != "ana" {
		t.Fatalf("reagiu: %v", got)
	}
	react("bia", "t", 1, 0) // de novo: desfaz só a sua
	if got := rooms["t"].events[0].Reacts[0]; len(got) != 1 || got[0] != "ana" {
		t.Fatalf("desfez: %v", got)
	}
}

// Só o dono muda a sala, e remover abre a vaga sem bloquear quem saiu (ADR 0007).
func TestDono(t *testing.T) {
	rooms[DemoCode] = &room{code: DemoCode, name: "sala", owner: "ana", max: MaxMembers,
		members: []string{"ana", "bia", "cris"}}
	defer delete(rooms, DemoCode)
	for _, id := range []string{"ana", "bia", "cris", "dan"} {
		users[id] = &user{id: id, nick: id}
		defer delete(users, id)
	}
	r := rooms[DemoCode]
	if err := Rename("bia", DemoCode, "outro nome"); err == nil || r.name != "sala" {
		t.Fatalf("quem não é dono mudou o nome: %v", err)
	}
	if err := Rename("ana", DemoCode, "turma de cálculo"); err != nil || r.name != "turma de cálculo" {
		t.Fatalf("o dono não mudou o nome: %v", err)
	}
	if err := Resize("ana", DemoCode, 99); err != nil || r.max != MaxMembers {
		t.Fatalf("tamanho fora do teto: %v %d", err, r.max)
	}
	if err := Resize("ana", DemoCode, 1); err != nil || r.max != 3 { // 3 membros é o chão
		t.Fatalf("tamanho abaixo de quem já está na sala: %v %d", err, r.max)
	}
	if err := Kick("ana", DemoCode, "ana"); err == nil {
		t.Fatal("o dono se removeu da própria sala")
	}
	if err := Kick("ana", DemoCode, "cris"); err != nil || slices.Contains(r.members, "cris") {
		t.Fatalf("remover: %v %v", err, r.members)
	}
	if err := Join("cris", DemoCode); err != nil { // sem bloqueio: com o código, volta
		t.Fatalf("voltar depois de removido: %v", err)
	}
	if err := SetClosed("ana", DemoCode, true); err != nil {
		t.Fatal(err)
	}
	if err := Join("dan", DemoCode); err == nil { // tem vaga, mas a sala está fechada
		t.Fatal("entrou numa sala fechada")
	}
	if err := Transfer("ana", DemoCode, "bia"); err != nil || r.owner != "bia" {
		t.Fatalf("passar a sala: %v %q", err, r.owner)
	}
	if err := Rename("ana", DemoCode, "sem poderes"); err == nil {
		t.Fatal("quem passou a sala continuou mandando")
	}
}

// Sair sendo dono passa a sala para o membro mais antigo que ficou: ela nunca fica órfã.
func TestSucessao(t *testing.T) {
	if err := store.Open(":memory:"); err != nil { // sem arquivo: nada para limpar depois
		t.Fatal(err)
	}
	for _, id := range []string{"a", "b", "c"} {
		if err := store.PutUser(id, id, pet.EnergyStart); err != nil {
			t.Fatal(err)
		}
	}
	code, err := Create("a", "turma", MaxMembers)
	if err != nil {
		t.Fatal(err)
	}
	defer delete(rooms, code)
	for _, id := range []string{"b", "c"} {
		if err := Join(id, code); err != nil {
			t.Fatal(err)
		}
	}
	if err := Leave("a", code); err != nil {
		t.Fatal(err)
	}
	if got := rooms[code].owner; got != "b" {
		t.Fatalf("dono depois da saída: %q", got)
	}
	if err := Leave("b", code); err != nil { // de novo: a sala segue com quem sobrou
		t.Fatal(err)
	}
	if got := rooms[code].owner; got != "c" {
		t.Fatalf("dono depois da segunda saída: %q", got)
	}
}

// A sala solo nasce fechada, e crescer não abre a porta sozinho (ADR 0006 e 0007).
func TestSoloNasceFechada(t *testing.T) {
	if err := store.Open(":memory:"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"s1", "s2"} {
		if err := store.PutUser(id, id, pet.EnergyStart); err != nil {
			t.Fatal(err)
		}
	}
	code, err := Create("s1", "só eu", MinMembers)
	if err != nil {
		t.Fatal(err)
	}
	defer delete(rooms, code)
	if r := rooms[code]; !r.closed || r.max != MinMembers {
		t.Fatalf("sala solo: fechada=%v tamanho=%d", r.closed, r.max)
	}
	if err := Resize("s1", code, 5); err != nil {
		t.Fatal(err)
	}
	if err := Join("s2", code); err == nil { // tem vaga, mas a porta continua fechada
		t.Fatal("a sala abriu sozinha ao crescer")
	}
	if err := SetClosed("s1", code, false); err != nil {
		t.Fatal(err)
	}
	if err := Join("s2", code); err != nil {
		t.Fatalf("entrar depois de aberta: %v", err)
	}
	// o que o dono mudou tem que voltar igual depois de um restart
	if err := Rename("s1", code, "a sala"); err != nil {
		t.Fatal(err)
	}
	clear(rooms)
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	if r := rooms[code]; r == nil || r.name != "a sala" || r.owner != "s1" || r.max != 5 || r.closed {
		t.Fatalf("depois do restart: %+v", r)
	}
}

// A peça anda nos terços da meta, e a mesma pessoa está em etapas diferentes em salas de metas
// diferentes: os minutos são dela, a meta é da sala (ADR 0008).
func TestStage(t *testing.T) {
	for _, c := range []struct {
		mins float64
		meta int
		want int
	}{
		{0, 60, 0}, {19, 60, 0}, {20, 60, 1}, {39, 60, 1}, {40, 60, 2}, {60, 60, 3},
		{90, 60, 3},  // passar da meta não cria etapa nova
		{30, 60, 1},  // meia hora numa sala de 1 h
		{30, 120, 0}, // a mesma meia hora numa sala de 2 h
		{30, 0, 0},   // sala sem meta não constrói nada
	} {
		if got := Stage(c.mins, c.meta); got != c.want {
			t.Errorf("Stage(%v, %d) = %d, queria %d", c.mins, c.meta, got, c.want)
		}
	}
}

// Os minutos do dia somam o que já foi gravado e a tarefa em andamento, e viram do dia.
func TestMinutesToday(t *testing.T) {
	now := time.Date(2026, 11, 1, 14, 0, 0, 0, time.Local)
	u := &user{id: "a", today: now.Format(time.DateOnly), todayMin: 25}
	if got := u.minutesToday(now); got != 25 {
		t.Fatalf("só o gravado: %v", got)
	}
	u.focus = &Focus{Start: now.Add(-10 * time.Minute), SegStart: now.Add(-10 * time.Minute),
		Ends: now.Add(15 * time.Minute), Declared: 25}
	if got := u.minutesToday(now); got != 35 {
		t.Fatalf("com a tarefa em andamento: %v", got)
	}
	u.focus = nil
	u.today = now.AddDate(0, 0, -1).Format(time.DateOnly) // o que ficou é de ontem
	if got := u.minutesToday(now); got != 0 {
		t.Fatalf("virou o dia: %v", got)
	}
}

// Só o dono escolhe a obra, e a sala solo não tem obra (ADR 0008).
func TestObra(t *testing.T) {
	rooms[DemoCode] = &room{code: DemoCode, name: "sala", owner: "ana", max: MaxMembers,
		meta: 60, members: []string{"ana"}}
	defer delete(rooms, DemoCode)
	users["ana"] = &user{id: "ana", nick: "ana"}
	defer delete(users, "ana")
	r := rooms[DemoCode]
	if err := SetObra("bia", DemoCode, "trem"); err == nil || r.obra != "" {
		t.Fatalf("quem não é dono escolheu a obra: %v", err)
	}
	if err := SetObra("ana", DemoCode, "foguete"); err == nil {
		t.Fatal("aceitou um tema que não existe")
	}
	if err := SetObra("ana", DemoCode, "cidade"); err != nil || r.obra != "cidade" {
		t.Fatalf("escolher: %v %q", err, r.obra)
	}
	if err := SetMeta("ana", DemoCode, 17); err == nil || r.meta != 60 {
		t.Fatalf("meta fora da lista: %v %d", err, r.meta)
	}
	if err := SetMeta("ana", DemoCode, 120); err != nil || r.meta != 120 {
		t.Fatalf("meta: %v %d", err, r.meta)
	}
	if err := SetObra("ana", DemoCode, ""); err != nil || r.obra != "" { // desligar
		t.Fatalf("desligar: %v %q", err, r.obra)
	}
	r.max = MinMembers
	if err := SetObra("ana", DemoCode, "trem"); err == nil {
		t.Fatal("a sala solo ganhou obra")
	}
}

// Fechar o terminal com uma tarefa de pé: a sala vê offline na hora, antes de a tolerância de
// reconexão interromper a tarefa.
func TestStatusOffline(t *testing.T) {
	u := &user{id: "ana", online: 1, focus: &Focus{}}
	if u.status() != Focusing {
		t.Fatalf("focando: %v", u.status())
	}
	u.online = 0
	if u.status() != Offline {
		t.Fatalf("saiu: %v", u.status())
	}
	u.focus = nil
	if u.online = 1; u.status() != Idle {
		t.Fatalf("voltou sem tarefa: %v", u.status())
	}
}

// A sala demo é a única que não passa pelo banco, e os usuários já em memória não são procurados
// lá: dá para testar o teto sem abrir o SQLite.
func TestJoinFull(t *testing.T) {
	rooms[DemoCode] = &room{code: DemoCode, name: "demo", max: MaxMembers}
	defer delete(rooms, DemoCode)
	for i := range MaxMembers {
		id := fmt.Sprintf("u%d", i)
		users[id] = &user{id: id}
		defer delete(users, id)
		if err := Join(id, DemoCode); err != nil {
			t.Fatalf("membro %d: %v", i, err)
		}
	}
	users["tarde"] = &user{id: "tarde"}
	defer delete(users, "tarde")
	if err := Join("tarde", DemoCode); err == nil || !strings.Contains(err.Error(), "demo já está cheia") {
		t.Fatalf("entrou numa sala cheia: %v", err)
	}
	if got := len(rooms[DemoCode].members); got != MaxMembers {
		t.Fatalf("membros: %d", got)
	}
	if err := Join("u0", DemoCode); err != nil { // quem já é membro não ocupa outra vaga
		t.Fatalf("já membro: %v", err)
	}
}

// O restart é uma desconexão de todo mundo: quem volta dentro da tolerância continua a tarefa de onde
// estava; quem não volta fica com ela interrompida, com os minutos até o restart.
func TestRestart(t *testing.T) {
	// uma conexão só (SetMaxOpenConns): o banco em memória dura o teste todo
	if err := store.Open(":memory:"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"ana", "bia"} {
		if err := store.PutUser(id, id, pet.EnergyStart); err != nil {
			t.Fatal(err)
		}
	}
	defer func(g time.Duration) { reconnectGrace = g }(reconnectGrace)
	reconnectGrace = 200 * time.Millisecond

	Start("ana", "abc", "lista 3", 25) // grava a tarefa aberta
	ana := *users["ana"].focus
	ten := time.Now().Add(-10 * time.Minute) // bia começou há 10 min
	if err := store.SaveOpenTask(store.OpenTask{User: "bia", Task: "relatório", Room: "abc",
		Start: ten, SegStart: ten, Ends: ten.Add(25 * time.Minute), Declared: 25}); err != nil {
		t.Fatal(err)
	}

	clear(users) // o servidor caiu: só o banco sobrou
	clear(rooms)
	defer clear(users)
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	Connect("ana") // dentro da tolerância
	time.Sleep(3 * reconnectGrace)

	mu.Lock()
	f, bia := users["ana"].focus, users["bia"].focus
	mu.Unlock()
	if f == nil || f.Task != "lista 3" || f.Start.Unix() != ana.Start.Unix() || f.Ends.Unix() != ana.Ends.Unix() {
		t.Fatalf("ana não continuou: %+v", f)
	}
	if bia != nil {
		t.Fatalf("bia não foi interrompida: %+v", bia)
	}
	ts, err := store.Tasks("bia")
	if err != nil || len(ts) != 1 || ts[0].Result != pet.Interrupted ||
		ts[0].Focused < 10*time.Minute || ts[0].Focused > 10*time.Minute+2*time.Second {
		t.Fatalf("tarefa de bia: %+v %v", ts, err)
	}
	if open, _ := store.OpenTasks(); len(open) != 1 || open[0].User != "ana" { // a de bia saiu das abertas
		t.Fatalf("abertas: %+v", open)
	}
}
