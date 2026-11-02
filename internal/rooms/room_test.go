package rooms

import (
	"fmt"
	"plantao/internal/pet"
	"plantao/internal/store"
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
