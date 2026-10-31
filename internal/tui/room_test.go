package tui

import (
	"plantao/internal/rooms"
	"strings"
	"testing"
	"time"
)

func TestSoloGardenWidth(t *testing.T) {
	if got := soloGardenWidth(116); got != 60 { // terminal de 120 colunas
		t.Fatalf("120 colunas: %d", got)
	}
	if got := soloGardenWidth(76); got != 0 { // 80 colunas: o jardim não cabe
		t.Fatalf("80 colunas: %d", got)
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
