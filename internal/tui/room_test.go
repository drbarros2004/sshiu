package tui

import (
	"plantao/internal/rooms"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestRoomCode(t *testing.T) {
	for in, want := range map[string]string{
		"fg7dhm":                         "fg7dhm",
		"  FG7DHM\n":                     "fg7dhm", // colado com espaço e quebra de linha
		"ssh -t -p 23234 plantao fg7dhm": "fg7dhm", // o convite inteiro
		"":                               "",
	} {
		if got := roomCode(in); got != want {
			t.Errorf("roomCode(%q) = %q, queria %q", in, got, want)
		}
	}
}

// O card conta as reações que o membro recebeu, menos as que ele mesmo deu no próprio evento.
func TestReceived(t *testing.T) {
	s := rooms.Snapshot{Events: []rooms.Event{{Who: "ana", Reacts: [rooms.Reactions][]string{
		0: {"bia", "ana"}, 1: {"bia"}}}}}
	if got := received(s)["ana"]; got != 2 {
		t.Fatalf("recebidas: %d", got)
	}
}

func TestSoloGardenWidth(t *testing.T) {
	if got := soloGardenWidth(116); got != 60 { // terminal de 120 colunas
		t.Fatalf("120 colunas: %d", got)
	}
	if got := soloGardenWidth(76); got != 0 { // 80 colunas: o jardim não cabe
		t.Fatalf("80 colunas: %d", got)
	}
}

func TestOrderedMembers(t *testing.T) {
	s := rooms.Snapshot{Members: []rooms.Member{
		{ID: "sumiu", Status: rooms.Offline},
		{ID: "foca1", Status: rooms.Focusing},
		{ID: "eu", Status: rooms.Idle},
		{ID: "parado", Status: rooms.Idle},
		{ID: "foca2", Status: rooms.Focusing},
	}}
	var got []string
	for _, x := range orderedMembers("eu", s) {
		got = append(got, x.ID)
	}
	if want := []string{"eu", "foca1", "foca2", "parado", "sumiu"}; !slices.Equal(got, want) {
		t.Fatalf("ordem %v, queria %v", got, want)
	}
}

// Sem tarefa a coluna da direita são os números, e o jardim (m.side) é nil: a tela não pode
// depender dele para se desenhar.
func TestSoloViewWithoutGarden(t *testing.T) {
	m := model{id: "a", room: "s", width: 120, height: 36, st: newStyles(themes[0]), now: time.Now(),
		w: rooms.World{Users: map[string]rooms.Member{"a": {ID: "a"}},
			Rooms: map[string]rooms.Snapshot{"s": {Code: "s", Name: "solo", Max: rooms.MinMembers}}}}
	if got := m.soloView(); !strings.Contains(got, "│") {
		t.Fatalf("sem as duas colunas: %q", got)
	}
}
