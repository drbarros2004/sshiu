package rooms

import (
	"fmt"
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
	react("ana", "t", 1, 0) // ninguém reage ao próprio evento
	react("eve", "t", 1, 0) // nem quem não é da sala
	if got := rooms["t"].events[0].Reacts[0]; len(got) != 1 || got[0] != "bia" {
		t.Fatalf("reagiu: %v", got)
	}
	react("bia", "t", 1, 0) // de novo: desfaz
	if got := rooms["t"].events[0].Reacts[0]; len(got) != 0 {
		t.Fatalf("desfez: %v", got)
	}
}

// A sala demo é a única que não passa pelo banco, e os usuários já em memória não são procurados
// lá: dá para testar o teto sem abrir o SQLite.
func TestJoinFull(t *testing.T) {
	rooms[DemoCode] = &room{code: DemoCode, name: "demo"}
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
