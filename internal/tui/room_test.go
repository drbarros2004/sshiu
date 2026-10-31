package tui

import "testing"

func TestSoloGardenWidth(t *testing.T) {
	if got := soloGardenWidth(116); got != 60 { // terminal de 120 colunas
		t.Fatalf("120 colunas: %d", got)
	}
	if got := soloGardenWidth(76); got != 0 { // 80 colunas: o jardim não cabe
		t.Fatalf("80 colunas: %d", got)
	}
}
