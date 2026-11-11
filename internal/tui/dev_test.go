package tui

import (
	"fmt"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// Quem nasce no meio da lista não muda a pessoa escolhida, e uma leitura velha não volta a tela.
func TestDevSetRows(t *testing.T) {
	t0 := time.Now()
	d := devPanel{}
	d.setRows([]store.DevRow{{ID: "dev:ana"}, {ID: "dev:caio"}}, t0, "")
	d.row = 1
	d.setRows([]store.DevRow{{ID: "dev:ana"}, {ID: "dev:bia"}, {ID: "dev:caio"}}, t0.Add(time.Second), "")
	if d.rows[d.row].ID != "dev:caio" {
		t.Fatalf("o cursor foi para %s", d.rows[d.row].ID)
	}
	d.setRows([]store.DevRow{{ID: "dev:ana"}}, t0, "") // mais velha que a da tela
	if len(d.rows) != 3 {
		t.Fatal("a leitura velha substituiu a nova")
	}
	d.setRows(d.rows, t0.Add(2*time.Second), "dev:ana") // quem acabou de nascer
	if d.row != 0 {
		t.Fatalf("não escolheu a pessoa nova: %d", d.row)
	}
}

// A aba cabe num terminal de 80 colunas (76 de miolo) e rola: o cursor nunca sai da tela.
func TestDevViewFits(t *testing.T) {
	m := model{st: newStyles(themes[0]), now: time.Now(), w: rooms.World{Users: map[string]rooms.Member{}}}
	for i := range 30 {
		p := pet.Pet{Color: "turquesa", LastActive: m.now, Energy: 100}
		m.dev.rows = append(m.dev.rows, store.DevRow{ID: fmt.Sprint(i), Nick: fmt.Sprintf("p%02d-apelidolongo", i), Pet: p, Before: 9999})
	}
	m.dev.row, m.dev.col = 20, 5
	out := m.devView(12)
	for _, l := range strings.Split(out, "\n") {
		if w := ansi.StringWidth(l); w > 76 {
			t.Fatalf("linha com %d colunas: %q", w, ansi.Strip(l))
		}
	}
	if lines := strings.Count(out, "\n") + 1; lines > 12 {
		t.Fatalf("%d linhas numa altura de 12", lines)
	}
	if !strings.Contains(ansi.Strip(out), "› p20-apeli…") {
		t.Fatalf("o cursor sumiu:\n%s", ansi.Strip(out))
	}
}
