package tui

// Os brinquedos do brainrot (#79): a prensa hidráulica esmagando coisas e o sabão sendo fatiado,
// desenhados aqui em pixel art (nada de vídeo de terceiros, ADR 0005). Eles ocupam o vão entre os
// dois vídeos, em cima e embaixo do relógio, que era espaço vazio. Cada um anda num ciclo fixo
// pelo tempo, como os outros wallpapers, e só pinta a própria caixa.

import "math"

// toyBox: a caixa de um brinquedo, em células.
type toyBox struct{ x, y, w, h int }

// pressObjects: o que a prensa esmaga, um por ciclo.
var pressObjects = []string{"duck", "can", "melon", "bear"}

// pressCycle: um ciclo da prensa, em segundos: desce, esmaga, segura, sobe.
const pressCycle = 4.0

// drawPress: a prensa numa caixa de b.w × 2·b.h pixels. A placa desce até encostar no objeto,
// achata ele até virar uma panqueca (que espirra para os lados) e sobe; o próximo objeto aparece.
func drawPress(c *canvas, b toyBox, t float64) {
	p := newPix(b.w, 2*b.h)
	w, h := p.w, p.h
	n := int(t / pressCycle)
	ph := math.Mod(t, pressCycle) / pressCycle // de 0 a 1 no ciclo
	baseY := h - 3                             // a mesa de baixo ocupa as últimas 3 linhas
	topY := 3                                  // a viga de cima
	objH0, objW0 := min(h/2, 14), min(w/3, 14) // o objeto inteiro
	flat := 2                                  // a panqueca
	plateH := 3
	// a altura do objeto agora: inteira até a placa encostar, depois esmagada
	free := float64(topY + plateH)           // a placa lá em cima
	touch := float64(baseY - objH0 - plateH) // a placa encostando no objeto
	bottom := float64(baseY - flat - plateH) // a placa no fim do esmagamento
	var plateY float64
	switch {
	case ph < 0.3: // desce rápido
		plateY = free + (touch-free)*ease(ph/0.3)
	case ph < 0.55: // esmaga devagar: é a parte que prende o olho
		plateY = touch + (bottom-touch)*(ph-0.3)/0.25
	case ph < 0.75: // segura
		plateY = bottom
	default: // sobe
		plateY = bottom + (free-bottom)*ease((ph-0.75)/0.25)
	}
	objH := objH0
	switch {
	case ph >= 0.75: // depois de esmagado, fica achatado até o próximo ciclo
		objH = flat
	case ph >= 0.3:
		objH = max(flat, min(objH0, baseY-int(plateY)-plateH))
	}
	// a coluna e a viga
	steel := func(x, y int) uint8 { return shade([]uint8{240, 243, 246}, 0.2+0.6*float64(x%4)/4, x, y) }
	for y := 0; y < baseY; y++ {
		for _, x := range []int{1, 2, w - 3, w - 2} {
			p.set(x, y, steel(x, y))
		}
	}
	for y := 0; y < topY; y++ {
		for x := 1; x < w-1; x++ {
			p.set(x, y, 242)
		}
	}
	// a haste do pistão e a placa, com a faixa de perigo
	py := int(plateY)
	for y := topY; y < py; y++ {
		for x := w/2 - 2; x <= w/2+1; x++ {
			p.set(x, y, []uint8{250, 252, 252, 248}[x-(w/2-2)])
		}
	}
	for y := py; y < py+plateH; y++ {
		for x := 3; x < w-3; x++ {
			col := uint8(244)
			if y == py+plateH-1 {
				col = 240
			} else if y == py && (x/2)%2 == 0 {
				col = 220
			} else if y == py {
				col = 16
			}
			p.set(x, y, col)
		}
	}
	// a mesa
	for y := baseY; y < h; y++ {
		for x := 0; x < w; x++ {
			p.set(x, y, []uint8{245, 241, 238}[y-baseY])
		}
	}
	// o objeto, esticado para os lados conforme achata
	objW := min(w-8, objW0*objH0/max(objH, 1))
	drawObject(p, pressObjects[n%len(pressObjects)], w/2-objW/2, baseY-objH, objW, objH, objH < objH0)
	// os respingos, logo depois do esmagamento
	if ph > 0.5 && ph < 0.8 {
		k := (ph - 0.5) / 0.3
		for i := range 10 {
			dir := float64(1 - 2*(i%2))
			sx := float64(w/2) + dir*(float64(objW)/2+k*float64(w)/3*(0.4+hash(i, n, 3)))
			sy := float64(baseY) - 2 - k*6*hash(i, n, 5) + k*k*8
			if sy < float64(baseY) {
				p.set(int(sx), int(sy), objectColor(pressObjects[n%len(pressObjects)]))
			}
		}
	}
	c.putAt(p, b.x, b.y)
}

// ease: começa e termina devagar.
func ease(v float64) float64 { return v * v * (3 - 2*v) }

// objectColor: a cor dos respingos de cada objeto.
func objectColor(kind string) uint8 {
	switch kind {
	case "duck":
		return 226
	case "can":
		return 160
	case "melon":
		return 203
	}
	return 213
}

// drawObject: o objeto em ow × oh pixels com o canto em (x, y). Esmagado, ele perde os detalhes.
func drawObject(p *pix, kind string, x, y, ow, oh int, squashed bool) {
	cx, cy := float64(x)+float64(ow)/2, float64(y)+float64(oh)/2
	rx, ry := float64(ow)/2, float64(oh)/2
	switch kind {
	case "duck": // o pato de borracha: corpo, cabeça e bico
		p.oval(cx, cy+ry*0.25, rx, ry*0.75, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{226, 220, 214}, 0.3+dy*0.4+d*0.3, x, y) })
		if !squashed {
			p.oval(cx+rx*0.35, cy-ry*0.45, rx*0.42, ry*0.45, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{226, 220}, 0.3+dy*0.5, x, y) })
			p.set(int(cx+rx*0.5), int(cy-ry*0.55), 16)
			for i := range 3 {
				p.set(int(cx+rx*0.8)+i, int(cy-ry*0.35), 208)
			}
		}
	case "can": // a latinha: vermelha com a tampa prateada
		for yy := y; yy < y+oh; yy++ {
			for xx := x; xx < x+ow; xx++ {
				col := shade([]uint8{124, 160, 196, 160}, float64(xx-x)/float64(max(ow, 1)), xx, yy)
				if yy == y || yy == y+oh-1 {
					col = 250
				} else if !squashed && yy == y+oh/2 {
					col = 255
				}
				p.set(xx, yy, col)
			}
		}
	case "melon": // a melancia: casca verde listrada e o miolo vermelho que aparece ao esmagar
		p.oval(cx, cy, rx, ry, func(xx, yy int, d, dx, dy float64) uint8 {
			if squashed && d < 0.8 {
				if hash(xx, yy, 9) < 0.08 {
					return 16
				}
				return 203
			}
			if int((dx+1)*4)%2 == 0 {
				return 28
			}
			return 34
		})
	default: // o ursinho de goma, translúcido
		p.oval(cx, cy+ry*0.3, rx*0.7, ry*0.7, func(xx, yy int, d, dx, dy float64) uint8 { return shade([]uint8{219, 213, 207}, d, xx, yy) })
		if !squashed {
			p.oval(cx, cy-ry*0.5, rx*0.5, ry*0.4, func(xx, yy int, d, dx, dy float64) uint8 { return shade([]uint8{219, 213}, d, xx, yy) })
			p.oval(cx-rx*0.4, cy-ry*0.8, rx*0.2, ry*0.2, func(int, int, float64, float64, float64) uint8 { return 213 })
			p.oval(cx+rx*0.4, cy-ry*0.8, rx*0.2, ry*0.2, func(int, int, float64, float64, float64) uint8 { return 213 })
		}
	}
}

// soapColors: o sabão de cada ciclo, do claro (o topo) para o escuro (a lateral).
var soapColors = [][3]uint8{{225, 218, 175}, {189, 183, 141}, {194, 158, 115}, {195, 153, 110}}

// soapCut: quanto tempo leva uma fatia, e quantas fatias tem um sabão.
const (
	soapCut    = 1.2
	soapSlices = 8
)

// drawSoap: o sabão numa caixa de b.w × 2·b.h pixels. A faca desce na ponta do sabão, a fatia
// tomba para a direita e cai na pilha; quando o sabão acaba, vem outro de outra cor.
func drawSoap(c *canvas, b toyBox, t float64) {
	p := newPix(b.w, 2*b.h)
	w, h := p.w, p.h
	cut := int(t / soapCut)
	bar, done := cut/soapSlices, cut%soapSlices // qual sabão, quantas fatias já saíram
	ph := math.Mod(t, soapCut) / soapCut
	col := soapColors[bar%len(soapColors)]
	tableY := h - 2
	for y := tableY; y < h; y++ {
		for x := range w {
			p.set(x, y, []uint8{137, 94}[y-tableY]) // a tábua de madeira
		}
	}
	sliceW := max(2, w/24)
	fullW := soapSlices * sliceW
	bx := w/2 - fullW/2 - w/8
	bh := min(h-6, 12)
	top := tableY - bh
	left := fullW - done*sliceW // o que sobra do sabão
	soap := func(x0, y0, sw, sh int) {
		for y := y0; y < y0+sh; y++ {
			for x := x0; x < x0+sw; x++ {
				v := col[1]
				switch {
				case y < y0+2:
					v = col[0] // o topo, mais claro
				case x == x0+sw-1:
					v = col[2] // a face cortada
				}
				p.set(x, y, v)
			}
		}
	}
	soap(bx, top, left, bh)
	// a fatia que está saindo: em pé até a faca passar, depois tomba deitada para a direita
	sx := bx + left - sliceW
	if ph > 0.55 {
		k := min((ph-0.55)/0.35, 1)
		lean := int(k * float64(bh-sliceW))
		soap(sx+sliceW+lean/2, top+lean, sliceW+lean, max(sliceW, bh-lean))
	}
	// a pilha de fatias, deitadas ao lado
	pileX := bx + fullW + 3
	for i := range done {
		y := tableY - (i+1)*sliceW/2 - 1
		soap(pileX+int(hash(i, bar, 4)*3), y, bh, max(1, sliceW/2))
	}
	// a faca: lâmina prateada que desce até a tábua e volta
	kx := bx + left - sliceW - 1
	depth := 0.0
	switch {
	case ph < 0.45:
		depth = ph / 0.45
	case ph < 0.55:
		depth = 1
	default:
		depth = 1 - (ph-0.55)/0.45
	}
	tip := top - 4 + int(depth*float64(bh+4))
	for y := max(0, tip-12); y <= tip; y++ {
		p.set(kx, y, 252)
		p.set(kx+1, y, 248)
	}
	for y := max(0, tip-18); y < max(0, tip-12); y++ { // o cabo
		p.set(kx, y, 94)
		p.set(kx+1, y, 130)
	}
	// as lascas que voam no corte
	if ph > 0.2 && ph < 0.6 {
		for i := range 4 {
			k := (ph - 0.2) / 0.4
			p.set(kx+2+int(k*6*hash(i, cut, 7)), top+int(k*float64(bh)*hash(i, cut, 8)), col[0])
		}
	}
	c.putAt(p, b.x, b.y)
}

// putAt: os pixels na tela com o canto em (x0, y0), dois por célula; o 0 deixa o que estava.
func (c *canvas) putAt(p *pix, x0, y0 int) {
	for y := 0; y+1 < p.h; y += 2 {
		cy := y0 + y/2
		if cy < 0 || cy >= c.h {
			continue
		}
		for x := range p.w {
			cx := x0 + x
			if cx < 0 || cx >= c.w {
				continue
			}
			top, bot := p.p[y*p.w+x], p.p[(y+1)*p.w+x]
			cl := &c.cells[cy*c.w+cx]
			switch {
			case top == bot && top != 0:
				*cl = cell{r: ' ', b: colorName[top]}
			case top != 0:
				*cl = cell{r: '▄', c: colorName[top], rev: true}
				if bot != 0 {
					cl.b = colorName[bot]
				}
			case bot != 0:
				*cl = cell{r: '▄', c: colorName[bot]}
			}
		}
	}
}

// auraWords: o que sobe nos vãos do brainrot. Emoji, W, L e sigma são iguais nos dois idiomas;
// as frases de auraText passam pela tradução.
var (
	auraWords = []string{"💀", "🗿", "😭😭", "🔥", "sigma", "💀💀💀", "W", "L"}
	auraText  = []string{"+1000 de aura", "-500 de aura"}
)

// auraColors: uma cor por palavra que sobe, das mais berrantes da paleta.
var auraColors = []string{"201", "51", "226", "46", "208", "196", "123"}

// drawAura: três palavras subindo pela caixa, cada uma no próprio ritmo; só entra a palavra que
// cabe na largura.
func drawAura(c *canvas, b toyBox, t float64, seed int) {
	if b.w < 2 || b.h < 6 {
		return
	}
	for i := range 3 {
		period := 3 + 2*hash(i, seed, 11)
		k := t/period + hash(i, seed, 12) // a fase de cada uma
		n, ph := int(k), k-math.Floor(k)
		word := ""
		if j := int(hash(n, i+seed*7, 13) * float64(len(auraWords)+len(auraText))); j < len(auraWords) {
			word = auraWords[j]
		} else {
			word = c.lang.t(auraText[j-len(auraWords)])
		}
		gs, ws := glyphs(word)
		width := 0
		for _, w := range ws {
			width += w
		}
		if len(gs) == 0 || width > b.w {
			continue
		}
		x := b.x + int(hash(n, i+seed*7, 14)*float64(b.w-width+1))
		y := b.y + b.h - 1 - int(ph*float64(b.h-1))
		c.text(x, y, word, auraColors[(n+i)%len(auraColors)])
	}
}
