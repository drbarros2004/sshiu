package tui

import (
	"plantao/internal/pet"
	"plantao/internal/rooms"
	"testing"
)

func TestPresence(t *testing.T) {
	w := rooms.World{
		Rooms: map[string]rooms.Snapshot{
			rooms.PublicCode: {Name: "sala pública", Phase: rooms.Focus,
				Room:  []*rooms.Member{{ID: "ana"}, {ID: "edu", Result: pet.Abandoned}},
				Queue: []*rooms.Member{{ID: "bia"}}},
			"k7m2qp": {Name: "Turma de Cálculo", Phase: rooms.Break, Room: []*rooms.Member{{ID: "lia", Result: pet.Done}}},
		},
		Online: map[string]bool{"edu": true, "caio": true},
	}
	cases := []struct{ id, label, room string }{
		{"ana", "focando", "sala pública"},
		{"lia", "no intervalo", "Turma de Cálculo"},
		{"bia", "na fila", "sala pública"},
		{"edu", "online", ""}, // abandonou: está conectado, mas não foca mais
		{"caio", "online", ""},
		{"dani", "offline", ""},
	}
	for _, c := range cases {
		if label, room, _ := presence(c.id, w); label != c.label || room != c.room {
			t.Errorf("%s: got %q em %q, want %q em %q", c.id, label, room, c.label, c.room)
		}
	}
}
