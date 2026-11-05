package pet

import (
	"math"
	"testing"
)

// A obra mantém sua recoloração por matiz, independente das paletas fixas do pet.
func TestArtworkHue(t *testing.T) {
	h := colorHue(0x8EC9F0) // azul
	for _, c := range []uint32{0xFFD166, 0xB07818, 0x9A9A9A} {
		r, g, b, _ := withHue(c, h).RGBA()
		got := [3]float64{float64(r >> 8), float64(g >> 8), float64(b >> 8)}
		want := [3]float64{float64(c >> 16 & 255), float64(c >> 8 & 255), float64(c & 255)}
		lo, hi := faixa(got)
		wlo, whi := faixa(want)
		if math.Abs(lo-wlo) > 1 || math.Abs(hi-whi) > 1 {
			t.Errorf("withHue(%06X): brilho %v..%v, queria %v..%v", c, lo, hi, wlo, whi)
		}
		if lo == hi { // cinza não tem matiz para conferir
			continue
		}
		novo := uint32(got[0])<<16 | uint32(got[1])<<8 | uint32(got[2])
		if d := math.Abs(colorHue(novo) - h); d > 0.02 {
			t.Errorf("withHue(%06X): matiz saiu %v do pedido", c, d)
		}
	}
}

func faixa(c [3]float64) (lo, hi float64) {
	return math.Min(c[0], math.Min(c[1], c[2])), math.Max(c[0], math.Max(c[1], c[2]))
}

// A cor some da lista (ou nunca esteve): o pet volta para o padrão em vez de sumir.
func TestColorIndex(t *testing.T) {
	if got := ColorIndex("#8EC9F0"); Colors[got].Name != "azul" {
		t.Errorf("achou %q", Colors[got].Name)
	}
	if got := ColorIndex("#123456"); got != 0 {
		t.Errorf("cor que não existe virou %d", got)
	}
	if got := ColorIndex(""); got != 0 {
		t.Errorf("padrão virou %d", got)
	}
}
