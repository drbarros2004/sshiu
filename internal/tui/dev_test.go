package tui

import (
	"sshiu/internal/store"
	"testing"
	"time"
)

// Quem nasce no meio da lista não muda a pessoa escolhida, e uma leitura velha não volta a tela.
func TestDevSetRows(t *testing.T) {
	t0 := time.Now()
	d := devPanel{}
	d.setRows([]store.DevRow{{ID: "dev:ana"}, {ID: "dev:caio"}}, t0)
	d.row = 1
	d.setRows([]store.DevRow{{ID: "dev:ana"}, {ID: "dev:bia"}, {ID: "dev:caio"}}, t0.Add(time.Second))
	if d.rows[d.row].ID != "dev:caio" {
		t.Fatalf("o cursor foi para %s", d.rows[d.row].ID)
	}
	d.setRows([]store.DevRow{{ID: "dev:ana"}}, t0) // mais velha que a da tela
	if len(d.rows) != 3 {
		t.Fatal("a leitura velha substituiu a nova")
	}
}
