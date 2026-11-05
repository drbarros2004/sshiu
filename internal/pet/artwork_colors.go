package pet

import (
	"fmt"
	"image/color"
	"math"
)

// ArtworkColorMap mantém a recoloração das peças da obra; não é usado pelos pets.
func ArtworkColorMap(base, shade uint32, hex string) map[uint32]color.Color {
	if hex == "" {
		hex = Colors[0].ID
	}
	var c uint32
	if _, err := fmt.Sscanf(hex, "#%06x", &c); err != nil {
		return nil
	}
	h := colorHue(c)
	return map[uint32]color.Color{base: withHue(base, h), shade: withHue(shade, h)}
}

// colorHue: o matiz de uma cor, de 0 a 1. Um cinza não tem matiz e devolve 0.
func colorHue(c uint32) float64 {
	r, g, b := float64(c>>16&255)/255, float64(c>>8&255)/255, float64(c&255)/255
	mx := math.Max(r, math.Max(g, b))
	d := mx - math.Min(r, math.Min(g, b))
	if d == 0 {
		return 0
	}
	var h float64
	switch mx {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return math.Mod(h/6+1, 1)
}

// withHue troca o matiz de uma cor e mantém o resto: a mesma distância entre o claro e o escuro,
// e o mesmo brilho. A obra conserva seus detalhes ao usar o tom escolhido pelo usuário.
func withHue(c uint32, h float64) color.Color {
	r, g, b := float64(c>>16&255)/255, float64(c>>8&255)/255, float64(c&255)/255
	mn := math.Min(r, math.Min(g, b))
	d := math.Max(r, math.Max(g, b)) - mn
	x := d * (1 - math.Abs(math.Mod(h*6, 2)-1))
	var rr, gg, bb float64
	switch int(h * 6) {
	case 0:
		rr, gg, bb = d, x, 0
	case 1:
		rr, gg, bb = x, d, 0
	case 2:
		rr, gg, bb = 0, d, x
	case 3:
		rr, gg, bb = 0, x, d
	case 4:
		rr, gg, bb = x, 0, d
	default:
		rr, gg, bb = d, 0, x
	}
	u := func(v float64) uint8 { return uint8(math.Round((v + mn) * 255)) }
	return color.RGBA{u(rr), u(gg), u(bb), 255}
}
