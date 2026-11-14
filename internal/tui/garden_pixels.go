package tui

// Os wallpapers em pixel art: café, árvore, estrada à noite, farol e o gato e o cachorro dormindo.
// Cada célula mostra dois pixels em meio bloco ("▄" invertido: a cor de cima no fundo e a de baixo
// no glifo), então a imagem tem o
// dobro de linhas e o pixel sai quase quadrado. A regra da banda (docs/research/wallpapers-ascii.md): o
// cenário fica parado e só o que é pequeno se mexe — o vapor, o café girando, as folhas, a faixa
// da pista. E o que se mexe fica junto: o renderer reescreve a linha da primeira à última célula
// que mudou, e no pontilhado cada célula leva o próprio código de cor. Duas coisas se mexendo nas
// pontas da mesma linha custam a linha inteira (medido no harness: a estrada com os prédios
// andando e três postes na tela mandava 140 KB/s).

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
)

// pix: a imagem de um quadro, em cores da paleta de 256. 0 é "nada": quem pinta com 0 deixa o que
// estava embaixo (o preto de verdade é o 16).
type pix struct {
	w, h int
	p    []uint8
}

func newPix(w, h int) *pix { return &pix{w, h, make([]uint8, w*h)} }

func (p *pix) set(x, y int, c uint8) {
	if c != 0 && x >= 0 && y >= 0 && x < p.w && y < p.h {
		p.p[y*p.w+x] = c
	}
}

// oval: f pinta cada pixel dentro da elipse de centro (cx, cy) e raios rx, ry. dx e dy vão de -1
// a 1 e d é a distância ao centro na mesma escala.
func (p *pix) oval(cx, cy, rx, ry float64, f func(x, y int, d, dx, dy float64) uint8) {
	for y := int(math.Floor(cy - ry)); y <= int(cy+ry); y++ {
		for x := int(math.Floor(cx - rx)); x <= int(cx+rx); x++ {
			dx, dy := (float64(x)+0.5-cx)/rx, (float64(y)+0.5-cy)/ry
			if d := math.Hypot(dx, dy); d <= 1 {
				p.set(x, y, f(x, y, d, dx, dy))
			}
		}
	}
}

// cloud: uma nuvem de três bolhas achatadas, clara em cima e escura embaixo.
func (p *pix) cloud(cx, cy, rx, ry float64, ramp []uint8) {
	top := cy - ry*1.9
	f := func(x, y int, d, dx, dy float64) uint8 { return shade(ramp, 1-(float64(y)-top)/(ry*2.9), x, y) }
	p.oval(cx, cy, rx, ry, f)
	p.oval(cx-rx*0.3, cy-ry*0.7, rx*0.45, ry, f)
	p.oval(cx+rx*0.15, cy-ry*0.9, rx*0.35, ry, f)
}

// board: o quadro do texto, em células (x, y, largura, altura), pintado em pixels: edge na borda
// de 2 pixels e face no miolo. O texto entra uma linha abaixo e três colunas à direita do canto.
func (p *pix) board(b [4]int, edge, face func(x, y int) uint8) {
	for y := 2 * b[1]; y < 2*(b[1]+b[3]); y++ {
		for x := b[0]; x < b[0]+b[2]; x++ {
			if y < 2*b[1]+2 || y >= 2*(b[1]+b[3])-2 || x < b[0]+2 || x >= b[0]+b[2]-2 {
				p.set(x, y, edge(x, y))
			} else {
				p.set(x, y, face(x, y))
			}
		}
	}
}

// plank: a placa de tábuas deitadas, na árvore e no píer do farol.
func (p *pix) plank(b [4]int) {
	p.board(b, func(x, y int) uint8 { return shade([]uint8{58, 94}, 0.3+0.4*hash(x/3, y, 22), x, y) },
		func(x, y int) uint8 {
			if (y-2*b[1])%5 == 1 {
				return 58 // a fresta entre duas tábuas
			}
			return 94
		})
}

// boardAt: o quadro em volta de um bloco de texto tw × th com o canto em (x, y): a borda e mais
// três colunas e uma linha de folga de cada lado, para o texto não encostar na moldura.
func boardAt(x, y, tw, th int) [4]int { return [4]int{x - 5, y - 2, tw + 10, th + 4} }

var colorName [256]string

func init() {
	for i := range colorName {
		colorName[i] = strconv.Itoa(i)
	}
}

// put: os pixels na tela, dois por célula.
func (c *canvas) put(p *pix) {
	for y := 0; y+1 < p.h && y/2 < c.h; y += 2 {
		for x := range min(p.w, c.w) {
			top, bot := p.p[y*p.w+x], p.p[(y+1)*p.w+x]
			cl := &c.cells[y/2*c.w+x]
			switch {
			case top == bot && top != 0:
				*cl = cell{r: ' ', b: colorName[top]}
			case top != 0: // invertido: o de cima no fundo, o de baixo no "▄" (ver solid)
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

var bayer = [16]float64{0, 8, 2, 10, 12, 4, 14, 6, 3, 11, 1, 9, 15, 7, 13, 5}

// shade: a cor da rampa para v (de 0 a 1). Entre um degrau e o outro entra o pontilhado fixo de
// Bayer, que é da posição: a transição fica suave e não treme quando o resto da tela anda. Um 0
// na rampa é transparente, o que dá meia-transparência ao vapor e aos halos.
func shade(ramp []uint8, v float64, x, y int) uint8 {
	v = min(max(v, 0), 1) * float64(len(ramp)-1)
	i := int(v)
	if i < len(ramp)-1 && v-float64(i) > (bayer[(y&3)*4+(x&3)]+0.5)/16 {
		i++
	}
	return ramp[i]
}

// hash: um número fixo de 0 a 1 para cada (a, b, semente): estrelas, janelas, textura.
func hash(a, b, seed int) float64 {
	h := uint32(a*73856093 ^ b*19349663 ^ seed*83492791)
	h ^= h >> 13
	h *= 0x5bd1e995
	h ^= h >> 15
	return float64(h&0xffff) / 0x10000
}

// --- Café: a xícara vista um pouco de cima, o café girando devagar e o vapor subindo ---

type coffee struct {
	board [4]int // a lousa do texto, em células: x, y, largura, altura (largura 0 = sem lousa)
}

// spot: o texto vai numa lousa de giz na parede, à esquerda da janela, escrito em giz.
func (s *coffee) spot(w, h, tw, th int) (int, int, [3]string) {
	x, y := fitText(w, h, tw, th, max(6, w*26/100-tw/2), max(3, (h*6/10-th)/2))
	s.board = boardAt(x, y, tw, th)
	return x, y, [3]string{"248", "229", "255"}
}

var (
	ceramic = []uint8{236, 101, 144, 187, 230, 231}
	brew    = []uint8{233, 236, 94, 137, 180, 223}
)

func (s *coffee) draw(c *canvas, t, dt float64) {
	W, H := c.w, c.h*2
	p := newPix(W, H)
	edge := H * 3 / 5                           // onde a mesa encosta na parede
	lx, ly := float64(W)*0.58, -float64(H)*0.15 // o abajur, fora da tela, em cima da xícara
	for y := range H {
		for x := range W {
			d := math.Hypot(float64(x)-lx, (float64(y)-ly)*1.4) / float64(W)
			if y < edge {
				p.set(x, y, shade([]uint8{232, 233, 234, 235, 236, 237}, 1.1-d*1.4, x, y))
				continue
			}
			k := int(math.Sqrt(float64(y-edge)) * 1.7) // as tábuas, mais largas perto de quem olha
			if k != int(math.Sqrt(float64(y-edge+1))*1.7) {
				p.set(x, y, 233)
				continue
			}
			grain := 0.06 * math.Sin(float64(x)*0.09+hash(k, 0, 5)*20+math.Sin(float64(x)*0.021+float64(k))*3)
			p.set(x, y, shade([]uint8{233, 234, 236, 94, 137}, 1.15-d*1.3+(hash(k, 1, 5)-0.5)*0.12+grain, x, y))
		}
	}

	// a janela atrás da xícara: a noite lá fora, as luzes da rua desfocadas e a chuva escorrendo
	// no vidro. Atrás da xícara, a chuva muda as mesmas linhas que o vapor já muda.
	x0, x1, y1, y2 := W*53/100, W*87/100, H*8/100, edge-H*8/100
	ww, wh := float64(x1-x0), float64(y2-y1)
	for y := y1; y < y2; y++ {
		for x := x0; x < x1; x++ {
			p.set(x, y, shade([]uint8{17, 17, 18, 54, 60}, float64(y-y1)/wh, 0, 0)) // em faixas lisas: barato de reescrever
		}
	}
	for i := range 12 {
		col := []uint8{137, 179, 67, 132}[i%4]
		r := 1 + 1.5*hash(i, 3, 13)
		p.oval(float64(x0)+hash(i, 1, 13)*ww, float64(y1)+(0.4+0.55*hash(i, 2, 13))*wh, r, r,
			func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{0, 0, col}, 1.5-d, x, y) })
	}
	for i := range int(ww * wh / 160) {
		x := x0 + int(hash(i, 4, 13)*ww)
		y := y1 + int(math.Mod(hash(i, 6, 13)*wh+t*(1.5+3*hash(i, 5, 13)), wh))
		p.set(x, y, 110)
		p.set(x, y-1, 67)
		p.set(x, y-2, 24)
	}
	for y := y1 - 2; y < y2+2; y++ { // o caixilho
		for x := x0 - 2; x < x1+2; x++ {
			if y < y1 || y >= y2 || x < x0 || x >= x1 || x == (x0+x1)/2 || y == (y1+y2)/2 {
				p.set(x, y, shade([]uint8{234, 236, 238}, 0.6-float64(y-y1)/wh*0.4, x, y))
			}
		}
	}
	for x := x0 - 4; x < x1+4; x++ { // o peitoril
		p.set(x, y2+2, 239)
		p.set(x, y2+3, 236)
	}

	if b := s.board; b[2] > 0 { // a lousa: moldura de madeira, ardósia com pó de giz e a canaleta
		s.board = [4]int{} // vale para este quadro: sem texto no próximo, sem lousa
		p.board(b, func(x, y int) uint8 {
			return shade([]uint8{94, 130, 137}, 0.55-float64(y-2*b[1])/float64(4*b[3]), x, y)
		}, func(x, y int) uint8 {
			return shade([]uint8{235, 236, 237}, 0.35+0.3*hash(x/5, y/3, 21), x, y) // o pó de giz
		})
		for x := b[0] - 1; x <= b[0]+b[2]; x++ {
			p.set(x, 2*(b[1]+b[3]), 137)
		}
		p.set(b[0]+4, 2*(b[1]+b[3])-1, 255)
		p.set(b[0]+5, 2*(b[1]+b[3])-1, 255)
	}

	R := min(float64(W)*0.085, float64(H)*0.15) // o raio da boca
	const k = 0.38                              // o quanto a gente vê de cima: a boca é uma elipse
	cx, hc := float64(W)*0.7, R*1.1
	yb := float64(edge) + float64(H-edge)*0.5 // o pé da xícara
	y0 := yb - hc                             // o centro da boca

	// o pires e a sombra dele na mesa
	sr := R * 1.8
	p.oval(cx+R*0.4, yb+3, sr, sr*k, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{0, 233}, 1.4-d, x, y) })
	p.oval(cx, yb+1.5, sr, sr*k, func(x, y int, d, dx, dy float64) uint8 { return shade(ceramic, 0.25-dx*0.15, x, y) })
	p.oval(cx, yb, sr, sr*k, func(x, y int, d, dx, dy float64) uint8 {
		if d > 0.92 {
			return shade(ceramic, 0.85-dx*0.2-dy*0.1, x, y) // a borda do pires pega luz
		}
		return shade(ceramic, 0.6-dx*0.2+dy*0.1-max(0, 0.6-d)*0.4, x, y)
	})
	p.oval(cx+R*0.3, yb+0.5, R*1.05, R*k*1.1, func(x, y int, d, dx, dy float64) uint8 { return shade(ceramic, 0.3, x, y) })

	// a asa, antes do corpo: o corpo cobre a parte de dentro
	ax, ay := cx+R*0.95, y0+hc*0.42
	p.oval(ax, ay, R*0.5, R*0.48, func(x, y int, d, dx, dy float64) uint8 {
		if math.Hypot(dx*0.5/0.26, dy*0.48/0.24) <= 1 {
			return 0
		}
		return shade(ceramic, 0.55-dy*0.25-dx*0.15, x, y)
	})
	// o corpo: elipses do pé até a boca, cada uma por cima da de baixo, com luz da esquerda
	for s := 1.0; s >= 0; s -= 0.25 / hc {
		r := R * (1 - 0.2*s*s)
		p.oval(cx, y0+s*hc, r, r*k, func(x, y int, d, dx, dy float64) uint8 {
			v := 0.62 - 0.45*dx + 0.12*math.Sqrt(max(0, 1-dx*dx)) - 0.12*s
			if dx > -0.62 && dx < -0.42 {
				v += 0.35 // o brilho da louça
			}
			if dx > 0.82 {
				v += 0.12 // a luz que volta da mesa
			}
			return shade(ceramic, v, x, y)
		})
	}
	// a boca: a borda, a parede de dentro e o café
	p.oval(cx, y0, R, R*k, func(x, y int, d, dx, dy float64) uint8 { return shade(ceramic, 0.9-dx*0.2, x, y) })
	ri := R - 1.6
	p.oval(cx, y0+0.4, ri, ri*k, func(x, y int, d, dx, dy float64) uint8 { return shade(ceramic, 0.35+dx*0.3-dy*0.1, x, y) })
	cr, depth := ri-0.6, R*0.13
	p.oval(cx, y0+depth, cr, cr*k, func(x, y int, d, dx, dy float64) uint8 {
		if math.Hypot((float64(x)+0.5-cx)/ri, (float64(y)+0.5-y0-0.4)/(ri*k)) > 1 {
			return 0 // atrás da borda
		}
		a := math.Atan2(dy, dx)
		v := 0.32 + 0.2*math.Sin(2*a+d*7-t*0.6)*(1-d) // o creme em espiral, girando devagar
		if d > 0.82 {
			v -= 0.15
		}
		if math.Hypot(dx+0.45, dy+0.35) < 0.16 {
			v = 1 // o reflexo do abajur
		}
		return shade(brew, v, x, y)
	})

	// o vapor: três fitas que sobem ondulando e somem
	top := y0 - R*k
	L := min(float64(H)*0.4, top)
	for i := range 3 {
		fi := float64(i)
		bx := cx + (fi-1)*R*0.42
		breath := 0.65 + 0.35*math.Sin(t*0.5+fi*2.1)
		for j := 0.0; j < L; j++ {
			x := bx + math.Sin(j*0.11-t*1.1+fi*2.3)*(1.5+j*0.14) + math.Sin(t*0.3+fi)*j*0.08
			fade := min(j/4, 1) * (1 - j/L) * breath
			for w := -1; w <= 1; w++ {
				px, py := int(x)+w, int(top-2-j)
				p.set(px, py, shade([]uint8{0, 244, 247, 250, 253}, fade*1.3*(1-0.45*float64(w*w)), px, py))
			}
		}
	}
	c.put(p)
}

// --- Árvore: no alto de um morro, no fim da tarde, com as folhas se soltando ---
// A copa tem volume (a luz vem do sol, à direita), mas não balança: balançar muda centenas de
// células por quadro.

type tree struct {
	board  [4]int // a placa do texto, em células (largura 0 = sem placa)
	leaves []struct{ x, y, vx, vy, ph, landed float64 }
	next   float64
}

// canopy: as bolas da copa, em unidades do raio. As de baixo vêm depois e ficam na frente.
var canopy = [][3]float64{
	{0, -0.8, 0.45}, {-0.38, -0.5, 0.55}, {0.38, -0.55, 0.55}, {0, -0.15, 0.7},
	{-0.68, 0.05, 0.52}, {0.7, 0.0, 0.55}, {-1.0, 0.38, 0.38}, {1.0, 0.35, 0.4},
	{-0.4, 0.35, 0.5}, {0.38, 0.38, 0.5},
}

// spot: o texto vai numa placa de madeira fincada no campo, à direita, abaixo do sol.
func (s *tree) spot(w, h, tw, th int) (int, int, [3]string) {
	x, y := fitText(w, h, tw, th, w*70/100-tw/2, h-th-6)
	s.board = boardAt(x, y, tw, th)
	return x, y, [3]string{"223", "230", "255"}
}

func (s *tree) draw(c *canvas, t, dt float64) {
	W, H := c.w, c.h*2
	p := newPix(W, H)
	fw, fh := float64(W), float64(H)
	hz := H * 62 / 100
	tx := fw * 0.24
	ground := func(x float64) float64 { // o morro da árvore e o campo
		u := (x - tx) / (fw * 0.24)
		return float64(hz) + fh*0.1 - fh*0.17*math.Exp(-u*u) + 1.5*math.Sin(x*0.05)
	}
	sky := []uint8{60, 61, 103, 139, 175, 181, 223}
	for y := range hz {
		for x := range W {
			p.set(x, y, shade(sky, float64(y)/float64(hz)*1.05, x, y))
		}
	}
	sx, sy, sr := fw*0.78, float64(hz)-fh*0.05, fh*0.06
	p.oval(sx, sy, sr*2.6, sr*2.6, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{0, 0, 223, 224}, 1-d, x, y) })
	p.oval(sx, sy, sr, sr, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{224, 230}, 1.2-d, x, y) })
	for i := range 4 { // nuvens compridas, passando bem devagar
		span := fw + 120
		cx := int(math.Mod(hash(i, 6, 9)*span-t*(0.4+0.2*float64(i%2))+span*100, span)) - 60
		cy := float64(hz) * (0.45 + 0.3*hash(i, 7, 9)) // abaixo do texto
		rx, ry := 16+24*hash(i, 8, 9), 1.5+1.5*hash(i, 9, 9)
		p.cloud(float64(cx), cy, rx, ry, []uint8{139, 175, 181, 224})
	}
	for x := range W { // morros ao longe, apagados pela distância
		fx := float64(x)
		far := float64(hz) - fh*0.07 - fh*0.03*math.Sin(fx*0.023+1) - fh*0.02*math.Sin(fx*0.061)
		mid := float64(hz) - fh*0.015 - fh*0.02*math.Sin(fx*0.037+4)
		for y := int(far); y < H; y++ {
			col := shade([]uint8{103, 139}, 0.4-(float64(y)-far)*0.08, x, y)
			if float64(y) >= mid {
				col = shade([]uint8{60, 66, 102}, 0.6-(float64(y)-mid)*0.06, x, y)
			}
			p.set(x, y, col)
		}
		g := ground(fx)
		slope := ground(fx+1) - g // positivo descendo para a direita: a encosta virada para o sol
		for y := int(g); y < H; y++ {
			v := 0.6 + slope*0.3 - (float64(y)-g)*0.014
			p.set(x, y, shade([]uint8{22, 23, 29, 65, 71, 107}, v, x, y))
		}
		if hash(x, 0, 4) < 0.35 { // tufos de grama na crista
			p.set(x, int(g)-1, shade([]uint8{29, 65, 71}, 0.4+slope*0.3, x, int(g)))
		}
	}

	gx := ground(tx)
	R := fh * 0.17
	ccx, ccy := tx, gx-fh*0.17-R*0.5
	p.oval(tx-R*0.3, gx+1, R*1.3, R*0.18, func(x, y int, d, dx, dy float64) uint8 {
		if float64(y) < ground(float64(x)) {
			return 0
		}
		return shade([]uint8{0, 22}, 1.5-d, x, y) // a sombra da copa, só no chão
	})
	for y := int(ccy); y < int(gx)+1; y++ { // o tronco, mais grosso no pé
		u := (float64(y) - ccy) / (gx - ccy)
		hw := 1.3 + 1.2*u + 2.5*math.Pow(u, 6)
		for x := int(tx - hw); x <= int(tx+hw); x++ {
			nx := (float64(x) + 0.5 - tx) / hw
			p.set(x, y, shade([]uint8{234, 236, 94, 137}, 0.45+nx*0.4+(hash(x, y/3, 8)-0.5)*0.15, x, y))
		}
	}
	for _, b := range canopy {
		bx, by, br := ccx+b[0]*R*1.15, ccy+b[1]*R, b[2]*R
		p.oval(bx, by, br, br, func(x, y int, d, dx, dy float64) uint8 {
			nz := math.Sqrt(max(0, 1-d*d))
			v := 0.2 + 0.45*(dx*0.6-dy*0.55+nz*0.55) - max(0, float64(y)-ccy)/R*0.25
			v += (hash(x/2, y/2, 3) - 0.5) * 0.3 // as folhas: textura fixa em blocos de 2
			if d > 0.9 && dx > 0.3 && dy < 0.3 {
				v += 0.25 // a borda que pega o sol
			}
			return shade([]uint8{22, 22, 23, 29, 65, 71, 107, 150}, v, x, y)
		})
	}

	if b := s.board; b[2] > 0 { // a placa: dois postes até o pé da tela e tábuas deitadas
		s.board = [4]int{}
		for y := 2 * (b[1] + b[3]); y < H; y++ {
			for _, x := range []int{b[0] + 3, b[0] + b[2] - 5} {
				p.set(x, y, 94)
				p.set(x+1, y, 58)
			}
		}
		p.plank(b)
	}

	if t >= s.next {
		s.next = t + 0.6 + rand.Float64()
		a := rand.Float64() * 2 * math.Pi
		s.leaves = append(s.leaves, struct{ x, y, vx, vy, ph, landed float64 }{
			ccx + R*math.Cos(a), ccy + R*0.7*math.Sin(a), 2 + 2*rand.Float64(), 2 + rand.Float64(), rand.Float64() * 6.3, 0})
	}
	live := s.leaves[:0]
	for _, l := range s.leaves {
		x := l.x
		if l.landed == 0 {
			l.x += l.vx * dt
			l.y += l.vy * dt
			if l.y >= ground(l.x) {
				l.y, l.landed = ground(l.x), t
			}
			x += 2 * math.Sin(t*1.5+l.ph)
		} else if t-l.landed > 5 { // no chão por 5 s, depois some: nada se acumula
			continue
		}
		if l.x > fw+4 {
			continue
		}
		live = append(live, l)
		col := []uint8{179, 172, 136, 143, 107}[int(l.ph)%5]
		p.set(int(x), int(l.y), col)
		if math.Cos(t*3+l.ph) > 0 { // vira de lado e de frente enquanto cai
			p.set(int(x)+1, int(l.y), col)
		}
	}
	s.leaves = live
	c.put(p)
}

// --- Estrada à noite: o carro parado na tela e a cidade passando atrás ---
// Cada camada anda no seu ritmo, em pixels por segundo: quanto mais perto, mais rápido.

type road struct {
	dash, tw int // as linhas do painel do carro (0 = sem painel) e a largura do texto no meio dele
}

// spot: o texto vai no painel do carro, embaixo e de ponta a ponta, nas luzes âmbar dos painéis
// antigos: o relógio e a tarefa numa tela no meio, os mostradores dos lados. O nome do wallpaper
// cai na linha de baixo da tela.
func (s *road) spot(w, h, tw, th int) (int, int, [3]string) {
	if h-th-8 < roadMinScene { // terminal baixo: o painel cortaria o carro, então o texto vai para o céu
		s.dash = 0
		return (w - tw) / 2, 1, [3]string{"252", "215", "223"}
	}
	// de cima para baixo: a borda, uma linha, a moldura e a folga da tela, o texto, a folga e a
	// moldura, o nome do wallpaper e a última linha
	s.dash, s.tw = th+8, tw
	return (w - tw) / 2, h - th - 4, [3]string{"137", "215", "223"}
}

// roadMinScene: as linhas de cena que o carro pede acima do painel; com menos, o painel sai.
const roadMinScene = 12

const roadCity, roadLamps, roadLane = 0.25, 10.0, 26.0

func (s *road) draw(c *canvas, t, dt float64) {
	dash, tw := s.dash, s.tw
	s.dash = 0                // vale para este quadro: sem texto no próximo, sem painel
	W, H := c.w, (c.h-dash)*2 // a cena fica acima do painel
	p := newPix(W, c.h*2)
	fw, fh := float64(W), float64(H)
	hz := H * 58 / 100 // o pé dos prédios, atrás da mureta
	base := hz + 4     // onde começa a pista

	sky := []uint8{233, 17, 17, 54, 60, 97, 139}
	for y := range hz {
		for x := range W {
			p.set(x, y, shade(sky, float64(y)/float64(hz), x, y))
		}
	}
	for i := range W * hz / 70 { // estrelas, mais no alto; uma em cada quatro pisca, devagar
		x, y := int(hash(i, 1, 7)*fw), int(hash(i, 2, 7)*hash(i, 3, 7)*float64(hz)*0.8)
		b := hash(i, 5, 7)
		if i%4 == 0 {
			b = (math.Sin(t*(0.2+hash(i, 4, 7)*0.4)+b*6.3) + 1) / 2
		}
		p.set(x, y, []uint8{0, 60, 103, 146, 189}[int(b*4.99)])
	}
	mx, my, mr := fw*0.8, float64(hz)*0.25, max(3, fh/15)
	p.oval(mx, my, mr*2.5, mr*2.5, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{0, 0, 60, 61}, 1-d, x, y) })
	p.oval(mx, my, mr, mr, func(x, y int, d, dx, dy float64) uint8 {
		v := 1 - 0.6*max(0, d-0.6) - 0.15*(dx+dy)
		if hash(x/2, y/2, 3) < 0.18 {
			v -= 0.35 // as crateras
		}
		return shade([]uint8{144, 187, 230}, v, x, y)
	})
	for i := range 3 { // nuvens compridas, iluminadas por cima pela lua
		span := fw + 120
		cx := int(math.Mod(hash(i, 6, 9)*span-t*(0.3+0.2*float64(i%2))+span*100, span)) - 60
		cy := float64(hz) * (0.48 + 0.25*hash(i, 7, 9)) // abaixo do texto
		rx, ry := 18+26*hash(i, 8, 9), 1.5+1.5*hash(i, 9, 9)
		p.cloud(float64(cx), cy, rx, ry, []uint8{54, 60, 96, 139})
	}

	city := func(o, lot, maxH int, body uint8, lit []uint8, chance float64, seed int) {
		for x := range W {
			wx := x + o
			n, bx := wx/lot, wx%lot
			bw := lot - 1 - int(hash(n, 1, seed)*float64(lot)/3)
			if bx >= bw {
				continue
			}
			top := hz - int(float64(maxH)*(0.3+0.7*hash(n, 2, seed)))
			if hash(n, 3, seed) > 0.6 && (bx == 0 || bx == bw-1) {
				top += 2 // o topo em degrau
			}
			if hash(n, 4, seed) > 0.8 && bx == bw/2 { // antena com a luz piscando
				p.set(x, top-1, body)
				p.set(x, top-2, body)
				if math.Sin(t*1.3+float64(n)) > 0.3 {
					p.set(x, top-3, 167)
				}
			}
			for y := max(top, 0); y < hz; y++ {
				col := body
				if bx%2 == 1 && bx < bw-1 && (y-top)%3 == 2 && hash(n*97+bx, y-top, seed) < chance {
					col = lit[int(hash(n, bx*7+y, seed+1)*float64(len(lit)))]
				}
				p.set(x, y, col)
			}
		}
	}
	// os prédios do fundo não andam e os da frente andam um pixel a cada 4 s: cada passo reescreve
	// todas as linhas da cidade
	city(0, 9, hz*48/100, 60, []uint8{97, 103, 139}, 0.3, 11)
	city(int(t*roadCity), 13, hz*32/100, 235, []uint8{179, 222, 216, 222, 116, 180}, 0.35, 12)

	o := int(t * roadLamps)
	lampTop := hz - hz*36/100
	// os postes, atrás da mureta, com o halo da lâmpada. Um por vez: dois na mesma linha
	// reescrevem tudo o que fica entre eles.
	gap := W + 48
	for lx := -(o % gap); lx < W+16; lx += gap {
		hx, hy := float64(lx)+5, float64(lampTop)+1
		p.oval(hx, hy, 6, 5, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{0, 0, 0, 94, 137, 179}, 1-d, x, y) })
		for y := lampTop + 1; y < base; y++ {
			p.set(lx, y, 238)
		}
		p.set(lx+1, lampTop, 238)
		p.set(lx+2, lampTop, 238)
		p.set(lx+3, lampTop, 238)
		p.set(lx+4, lampTop, 230)
		p.set(lx+5, lampTop, 230)
		p.set(lx+4, lampTop+1, 223)
		p.set(lx+5, lampTop+1, 223)
	}
	for x := range W { // a mureta, com os pilares passando
		p.set(x, hz, 103)
		p.set(x, hz+1, 60)
		col := uint8(237)
		if (x+o)%18 < 2 {
			col = 235
		}
		p.set(x, hz+2, col)
		p.set(x, hz+3, col)
	}

	lane := int(t * roadLane)
	for y := base; y < H; y++ { // a pista: em faixas lisas, porque a faixa que anda reescreve a linha
		for x := range W {
			p.set(x, y, shade([]uint8{237, 236, 235}, float64(y-base)/float64(H-base)*1.2, 0, 0))
		}
	}
	for x := range W {
		p.set(x, base, 234)
		if (x+lane)%24 < 11 {
			p.set(x, base+(H-base)*30/100, 103)
		}
		p.set(x, H-1, 60) // a última linha de pixels: a de texto acima dela fica limpa
	}

	cx := max(2, W*28/100-20)
	wy := base + (H-base)*70/100 // onde as rodas encostam
	p.oval(float64(cx)+21, float64(wy), 23, 1.6, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{0, 233}, 1.6-d, x, y) })
	bob := 0
	if math.Sin(t*1.7)+math.Sin(t*2.9) > 1.6 { // um buraco na pista, de vez em quando
		bob = 1
	}
	fx, fy := float64(cx)+40, float64(wy-15-bob)+6.5
	p.oval(fx+16, float64(wy)-0.5, 18, 2.5, func(x, y int, d, dx, dy float64) uint8 {
		return shade([]uint8{0, 0, 237, 238, 239}, (1-d)*(0.9-0.4*dx), x, y) // a luz do farol na pista
	})
	p.oval(fx, fy, 3.5, 3, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{0, 0, 101, 144, 187}, 1-d, x, y) })
	drawCar(p, cx, wy-15, t*roadLane/4.3, bob)
	if dash == 0 {
		c.put(p)
		return
	}

	// o painel: a borda de cima pega a luz da rua. No meio, a tela; dos lados, o que couber: os
	// mostradores e, com mais lugar, a gasolina, a temperatura e o hodômetro da viagem
	for y := H; y < p.h; y++ {
		col := uint8(234) // liso atrás do texto
		switch y - H {
		case 0:
			col = 239
		case 1:
			col = 237
		}
		for x := range W {
			p.set(x, y, col)
		}
	}
	// a tela, funda no painel, em volta do texto (o lugar que o spot devolveu)
	tx, ty, th := (W-tw)/2, c.h-dash+4, dash-8
	p.board([4]int{tx - 4, ty - 2, tw + 8, th + 4}, func(x, y int) uint8 {
		if y < 2*ty-2 {
			return 238 // a moldura pega a luz em cima
		}
		return 236
	}, func(x, y int) uint8 { return 233 })

	kmh := 62 + 5*math.Sin(t*0.21) + 2*math.Sin(t*0.53)
	r := float64(min(11, dash-3))
	cy := float64(H) + float64(p.h-H)/2 + 1
	row := int(cy) / 2
	sx, rx := 3+r, float64(W)-4-r // o velocímetro e o conta-giros
	dials := W >= tw+8+2*int(2*r+6)
	extras := W >= tw+8+2*int(2*r+30)
	if dials {
		dial(p, sx, cy, r, kmh/180, 0)
		dial(p, rx, cy, r, (kmh/26)/8, 5) // na quarta marcha, o giro anda junto com a velocidade
	}
	bx := int(sx+r) + 4
	if extras {
		fuel := max(2, 6-int(t/900))       // a gasolina baixa um traço a cada 15 min
		temp := min(5, int(t/12))          // o motor esquenta no primeiro minuto e fica no meio
		icon(p, bx, 2*row-5, gasPump, 137) // a bomba de gasolina
		icon(p, bx+1, 2*row+1, thermometer, 137)
		for i := range 6 {
			x := bx + 7 + 3*i
			gas, heat := uint8(236), uint8(236)
			if i < fuel {
				gas = 179
			}
			if i == min(temp, 3) {
				heat = 215
			}
			for k := range 2 {
				p.set(x+k, 2*row-4, gas)
				p.set(x+k, 2*row-3, gas)
				p.set(x+k, 2*row+2, heat)
				p.set(x+k, 2*row+3, heat)
			}
		}
		for x := int(rx-r) - 15; x < int(rx-r)-4; x++ { // o hodômetro, fundo e preto
			p.set(x, 2*row, 16)
			p.set(x, 2*row+1, 16)
		}
	}
	c.put(p)
	if dials {
		c.ink(int(sx)-1, row+2, fmt.Sprint(int(kmh)), "223")
		c.ink(int(sx)-1, row+3, "km/h", "137")
		c.ink(int(rx)-1, row+2, "rpm", "137")
		c.ink(int(rx)-2, row+3, "x1000", "137")
	}
	if extras {
		c.ink(int(rx-r)-15, row-1, c.lang.t("viagem"), "137")
		c.ink(int(rx-r)-15, row, strings.Replace(fmt.Sprintf("%06.1f km", t*62/3600), ".", ",", 1), "252")
	}
}

// dial: um mostrador redondo com aro, sete marcas e o ponteiro em frac (de 0 a 1). Com red, as
// marcas a partir dela ficam vermelhas.
func dial(p *pix, cx, cy, r, frac float64, red int) {
	p.oval(cx, cy, r, r, func(x, y int, d, dx, dy float64) uint8 {
		if d > 0.86 {
			return shade([]uint8{236, 239, 243}, 0.5-dy*0.5, x, y) // o aro, claro em cima
		}
		return 233
	})
	for k := range 7 {
		a := math.Pi * (1.25 - 1.5*float64(k)/6)
		col := uint8(137)
		if red > 0 && k >= red {
			col = 167
		}
		p.set(int(cx+0.7*r*math.Cos(a)), int(cy-0.7*r*math.Sin(a)), col)
	}
	a := math.Pi * (1.25 - 1.5*min(frac, 1))
	for j := 0.0; j < 0.75*r; j += 0.5 { // o ponteiro
		p.set(int(cx+j*math.Cos(a)), int(cy-j*math.Sin(a)), 209)
	}
	p.set(int(cx), int(cy), 245)
}

// os ícones do painel, de 4 pixels de altura
var (
	gasPump     = []string{"###..", "#.#.#", "###.#", "####."}
	thermometer = []string{".#.#", ".#..", ".#.#", "###."}
)

// icon: um desenho pequeno de "#" com o canto em (x, y).
func icon(p *pix, x, y int, rows []string, col uint8) {
	for dy, l := range rows {
		for dx, ch := range l {
			if ch == '#' {
				p.set(x+dx, y+dy, col)
			}
		}
	}
}

// drawCar: um cupê de lado, virado para a direita: 40 pixels de largura por 15 de altura, com o
// canto de cima à esquerda em (x0, y0). A carroceria sobe bob pixels; as rodas giram com spin.
func drawCar(p *pix, x0, y0 int, spin float64, bob int) {
	const light, body, dark, deep, glass, glare = 210, 174, 131, 95, 60, 103
	wheels := [2]float64{9.5, 31.5}
	for y := range 11 {
		for x := range 40 {
			var col uint8
			switch {
			case y < 5: // a cabine
				l, r := 13-y, 25+2*y
				switch {
				case x < l || x > r:
				case y == 0:
					col = body
				case x == l || x == r || x == 20 || x == 21:
					col = deep
				case (x+y)%8 == 0 || (x+y)%8 == 1:
					col = glare
				default:
					col = glass
				}
			case y == 5 && (x < 2 || x > 37), y == 10 && (x < 1 || x > 38):
			case x >= 38 && y >= 6 && y <= 7:
				col = 229 // o farol
			case x == 0 && y >= 6 && y <= 7:
				col = 167 // a lanterna
			case y == 9 && (x < 3 || x > 36):
				col = 250 // o para-choque
			case y == 5:
				col = light
			case x == 21 || y == 10:
				col = deep
			case y == 6 && x >= 24 && x <= 25:
				col = light // a maçaneta
			case y <= 7:
				col = body
			default:
				col = dark
			}
			for _, wx := range wheels { // as caixas de roda
				if y >= 6 && math.Hypot(float64(x)+0.5-wx, float64(y)+0.5-10.5) < 5.6 {
					col = 0
				}
			}
			p.set(x0+x, y0+y-bob, col)
		}
	}
	for _, wx := range wheels {
		p.oval(float64(x0)+wx, float64(y0)+10.5, 4.4, 4.4, func(x, y int, d, dx, dy float64) uint8 {
			switch {
			case d > 0.6:
				return 234
			case d < 0.25:
				return 250
			case math.Abs(math.Sin(math.Atan2(dy, dx)-spin)) < 0.4:
				return 238 // os raios
			}
			return 244
		})
	}
}

// --- Farol: numa ilha de pedra ao longe, à noite, com o facho girando na neblina ---
// O desenho é medido numa tela de 100×60 pixels e cresce por k; o farol, longe, cresce por kl. u e
// v são as medidas dele, a partir do pé da torre (lx) e do horizonte (hz). O céu e o mar ficam em
// faixas lisas; de um quadro para o outro mudam só o facho, as cristas e o reflexo da lanterna.

type lighthouse struct {
	board [4]int // a placa do píer, em células (largura 0 = sem placa)
}

// lighthouseTurn: segundos por volta do facho. Ele anda 4 vezes por segundo, não a cada quadro:
// cada passo reescreve as linhas do céu por onde ele passa (docs/regras.md).
const lighthouseTurn, lighthouseSteps = 20.0, 4.0

// spot: o texto vai numa placa de madeira no píer, à esquerda do farol.
func (s *lighthouse) spot(w, h, tw, th int) (int, int, [3]string) {
	pier := lighthousePier(w, h*2) / 2 // a linha do píer, em células
	x, y := fitText(w, h, tw, th, max(6, w*17/100-tw/2), max(2, pier-th-3))
	s.board = boardAt(x, y, tw, th)
	return x, y, [3]string{"223", "230", "255"}
}

// fitText: o canto do texto dentro da tela. O render e as cenas com quadro (a lousa, as placas, o
// calendário) usam a mesma conta, para o quadro ficar em volta do texto em qualquer tela.
func fitText(w, h, tw, th, x, y int) (int, int) {
	return min(max(x, 1), w-tw-1), min(max(y, 0), h-th)
}

// lighthousePier: a linha do píer, em pixels: logo abaixo do horizonte, na escala da cena.
func lighthousePier(W, H int) int {
	return H*63/100 + int(2*min(float64(W)/100, float64(H)/60))
}

func (s *lighthouse) draw(c *canvas, t, dt float64) {
	W, H := c.w, c.h*2
	p := newPix(W, H)
	fw := float64(W)
	k := min(fw/100, float64(H)/60)
	kl := k * 0.6 // o farol, longe
	hz := H * 63 / 100
	b := s.board
	s.board = [4]int{} // vale para este quadro: sem texto no próximo, sem placa
	lx := fw * 0.7     // o pé da torre: à direita da placa, se ela for larga, sem sair da tela
	if b[2] > 0 {
		lx = min(max(lx, float64(b[0]+b[2])+23*kl), fw-18*kl)
	}
	ly := float64(hz) - 30*kl // a lanterna

	a := math.Floor(t*lighthouseSteps) / lighthouseSteps * 2 * math.Pi / lighthouseTurn
	ca, sa := math.Cos(a), math.Sin(a)
	reach, dir := 20+160*math.Abs(ca), math.Copysign(1, ca) // o facho encurta quando aponta para quem olha
	beam := func(x, y int) float64 {
		d, dy := (float64(x)-lx)/kl, (float64(y)-ly)/kl
		v := 0.0
		if d*dir > 0 && math.Abs(d) < reach { // a luz cai suave dos lados e some aos poucos na neblina
			spread := 2.7 + math.Abs(d)*0.2
			v = math.Pow(1-math.Abs(d)/reach, 1.5) * math.Exp(-2*(dy/spread)*(dy/spread)) * 0.85
			v *= 0.8 + 0.2*math.Sin(float64(x)*0.05/k+float64(y)*0.21/k) // a neblina, mais grossa aqui e ali
		}
		dist := math.Hypot(d, dy*1.3) // e a lanterna brilha mais
		return min(1, v+max(0, sa)*max(0, 1-dist/20)*0.7+0.25*max(0, 1-dist/10))
	}
	sky := []uint8{232, 233, 234, 235, 60, 103, 146}
	for y := range hz {
		g := 0.14 + math.Pow(float64(y)/float64(hz), 2)*0.24 // a neblina clareia o horizonte
		for x := range W {
			if v := beam(x, y); v > 0.04 {
				p.set(x, y, shade(sky, g+v*0.6, x, y))
			} else {
				p.set(x, y, shade(sky, g, 0, 0))
			}
		}
	}
	for i := range W * hz / 120 { // as estrelas, que o facho apaga quando passa
		x, y := int(hash(i, 1, 71)*fw), int(hash(i, 2, 71)*float64(hz)*0.6)
		if beam(x, y) < 0.05 {
			p.set(x, y, 243)
		}
	}

	sea, glow := []uint8{232, 233, 234, 235, 23, 24}, []uint8{94, 137, 179, 222}
	tide := math.Floor(t)     // as cristas andam uma vez por segundo: andar a cada quadro reescrevia o mar todo (83 KB/s)
	for y := hz; y < H; y++ { // devagar, uma linha para cada lado
		r := float64(y - hz)
		step, drift := 2+r*0.25, tide*(0.4+r*0.05)
		if (y-hz)%2 == 0 {
			drift = -drift
		}
		for x := range W {
			crest := hash(int(math.Floor((float64(x)+drift)/step)), y, 3) < 0.22
			if crest && math.Abs(float64(x)-lx) < 2*kl+r*0.15 { // o reflexo da lanterna
				p.set(x, y, shade(glow, 0.3+max(0, sa)*0.6-r/k*0.012, x, y))
				continue
			}
			v := 0.3 - r/k*0.006
			if crest {
				v += 0.26
			}
			p.set(x, y, shade(sea, v, 0, 0))
		}
	}

	uv := func(x, y int) (float64, float64) {
		return (float64(x) + 0.5 - lx) / kl, (float64(y) + 0.5 - float64(hz)) / kl
	}
	foam := hz + int(2*kl)
	for x := int(lx - 22*kl); x <= int(lx+28*kl); x++ { // a ilha
		u, _ := uv(x, 0)
		top := -3 - 1.5*(1+math.Sin(u*0.5+1)) + max(0, math.Abs(u-4)-20)*0.6
		if u > -13 && u < 22 {
			top -= 4 + math.Sin(u*0.9)
		}
		for y := hz + int(top*kl); y <= foam; y++ {
			_, v := uv(x, y)
			col := shade([]uint8{233, 234, 236, 238, 240}, 0.6-(v-top)*0.07+(hash(x/2, y/2, 64)-0.5)*0.4-(u+22)/50*0.25, x, y)
			if y == foam && hash(x, y, 63) < 0.5 {
				col = 244 // a espuma
			}
			p.set(x, y, col)
		}
	}
	for y := hz - int(38*kl); y < hz; y++ { // o farol e a cabana do faroleiro
		for x := int(lx - 7*kl); x <= int(lx+20*kl); x++ {
			u, v := uv(x, y)
			p.set(x, y, lighthouseAt(u, v, x, y))
		}
	}

	pier, end := lighthousePier(W, H), int(fw*0.34)
	if b[2] > 0 {
		end = max(end, b[0]+b[2]+4)
	}
	for x := range end { // o píer, com os pilares dentro da água
		col := shade([]uint8{94, 137}, 0.7-float64(x)/fw, x, pier)
		if int(float64(x)/k)%5 == 4 {
			col = 233
		}
		p.set(x, pier, col)
		p.set(x, pier+1, 94)
	}
	for _, u := range []float64{3, 14, 25, 33} {
		for y := pier + 2; y < pier+int(8*k); y++ {
			col := uint8(235)
			if y > pier+int(5*k) {
				col = 233
			}
			p.set(int(u*k), y, col)
		}
	}
	if b[2] > 0 { // a placa, em dois postes fincados no píer
		for y := 2 * (b[1] + b[3]); y < pier; y++ {
			for _, x := range []int{b[0] + 3, b[0] + b[2] - 5} {
				p.set(x, y, 94)
				p.set(x+1, y, 58)
			}
		}
		p.plank(b)
	}
	c.put(p)
}

// lighthouseAt: a cor do farol e da casa em (u, v), nas medidas da tela de 100×60; 0 é nada.
func lighthouseAt(u, v float64, x, y int) uint8 {
	au := math.Abs(u)
	switch {
	case au < 0.6 && v >= -37 && v < -35.4: // o para-raios
		return 240
	case math.Hypot(u/4.2, (v+33.4)/2.6) <= 1 && v < -33: // a cúpula
		return shade([]uint8{52, 88, 131}, 0.6-u/4.2*0.4, x, y)
	case v >= -33 && v < -28 && au < 3.5: // a lanterna, entre os montantes
		switch r := v + 33; {
		case au > 2.9 || au < 0.5:
			return 238
		case r < 1 || r >= 4:
			return 179
		case r >= 2 && r < 3:
			return 230
		}
		return 222
	case v >= -28 && v < -27 && au <= 6.5: // a varanda: o corrimão, os balaústres e o piso
		return 239
	case v >= -27 && v < -26 && au <= 6.5:
		if int(math.Floor(u+7))%2 == 0 {
			return 240
		}
		return 0
	case v >= -26 && v < -25 && au <= 6.5:
		if u < 0 {
			return 239
		}
		return 237
	case v >= -25 && v < -4 && au <= 3.8+(v+25)*0.11: // a torre listrada, com a lua pela esquerda
		switch {
		case au < 1.5 && v >= -9: // a porta
			if v < -8 {
				return 236
			}
			return 235
		case au < 0.6 && v >= -20 && v < -18: // as janelinhas
			if v < -19 {
				return 179
			}
			return 137
		case math.Abs(u+1) < 0.6 && v >= -14 && v < -12:
			return 236
		}
		lit := 0.75 - u/(3.8+(v+25)*0.11)*0.45
		if int(math.Floor((v+25)/5))%2 == 1 {
			return shade([]uint8{52, 95, 131}, lit, x, y)
		}
		return shade([]uint8{240, 244, 248, 251}, lit, x, y)
	case u >= 11.5 && u < 13 && v >= -5 && v < -3.5: // a cabana: a janela acesa, a parede e o telhado
		return 179
	case u >= 10 && u < 16 && v >= -6 && v < -2:
		return shade([]uint8{239, 243, 246}, 0.65-(u-10)/6*0.5, x, y)
	case v >= -10 && v < -6 && math.Abs(u-13) <= v+11.5:
		return shade([]uint8{52, 95, 131}, 0.55-(u-13)/5*0.4, x, y)
	}
	return 0
}

// --- Gato e cachorro dormindo: no fim da tarde, com o sol entrando pela janela ---
// O bicho respira (o corpo cresce e volta a cada 4,2 s); o gato mexe a ponta do rabo e o
// cachorro, sonhando, a pata. Medidas de uma tela de 100×60 pixels,
// multiplicadas por k, como no farol.

type sleeper struct {
	dog   bool
	board [4]int // o calendário, em células (largura 0 = sem calendário)
}

// sleeperWindow: o tamanho k do desenho e as bordas da janela, em pixels.
func sleeperWindow(W, H int) (k float64, x0, x1 int) {
	k = min(float64(W)/100, float64(H)/60)
	x0 = int(float64(W) * 0.06)
	return k, x0, x0 + int(24*k)
}

// spot: o texto vai num calendário pendurado na parede, à direita da janela, escrito a caneta.
func (s *sleeper) spot(w, h, tw, th int) (int, int, [3]string) {
	k, _, x1 := sleeperWindow(w, h*2)
	x, y := fitText(w, h, tw, th, max(x1+int(7*k)+7, min(w*72/100-tw/2, w-tw-7)), 4)
	s.board = boardAt(x, y, tw, th)
	return x, y, [3]string{"234", "234", "234"} // no papel liso, tudo escuro; o nome do wallpaper, no chão
}

func (s *sleeper) draw(c *canvas, t, dt float64) {
	W, H := c.w, c.h*2
	p := newPix(W, H)
	fw := float64(W)
	k, x0, x1 := sleeperWindow(W, H)
	edge := H * 2 / 3 // onde o chão encosta na parede
	y1 := int(5 * k)
	y2 := max(y1+4, min(y1+int(25*k), edge-int(6*k)))
	sun := func(x, y int) bool { // a mancha de sol no chão
		r := float64(y - edge)
		return y > edge && float64(x) >= float64(x1)-6*k+r*1.3 && float64(x) <= float64(x1)+22*k+r*1.7
	}
	for y := range edge {
		for x := range W {
			p.set(x, y, 138) // a parede rosada
		}
	}
	for x := range W {
		p.set(x, edge-2, 187) // o rodapé
		p.set(x, edge-1, 144)
	}
	plank := max(8, int(24*k))
	for y := edge; y < H; y++ { // o assoalho, em cor chapada: as tábuas do café e as emendas
		n := int(math.Sqrt(float64(y-edge)) * 1.7)
		seam := n != int(math.Sqrt(float64(y-edge+1))*1.7)
		for x := range W {
			joint, in, col := seam || (x+int(hash(n, 2, 5)*float64(plank)))%plank == 0, sun(x, y), uint8(137)
			if in && !joint {
				col = 180 // o sol no chão
			}
			if joint && !in {
				col = 94
			}
			p.set(x, y, col)
		}
	}

	wh := float64(y2 - y1) // a janela, com o pôr do sol, os morros longe e as cortinas
	for y := y1; y < y2; y++ {
		for x := x0; x < x1; x++ {
			p.set(x, y, shade([]uint8{175, 181, 216, 223}, float64(y-y1)/wh, 0, 0))
		}
	}
	p.oval(float64(x0)+16*k, float64(y2)-5*k, 4*k, 4*k, func(x, y int, d, dx, dy float64) uint8 {
		if d < 0.7 {
			return 230
		}
		return 229
	})
	cloud := func(x, y int, d, dx, dy float64) uint8 { return 224 }
	p.oval(float64(x0)+6*k, float64(y1)+6*k, 5*k, 1.4*k, cloud)
	p.oval(float64(x0)+19*k, float64(y1)+3*k, 4*k, 1.1*k, cloud)
	for x := x0; x < x1; x++ {
		for y := y2 - int((4+2*math.Sin(float64(x)/k*0.2))*k); y < y2; y++ {
			p.set(x, y, 138)
		}
	}
	mx, my := (x0+x1)/2, (y1+y2)/2
	for y := y1 - 2; y < y2+2; y++ {
		for x := x0 - 2; x < x1+2; x++ {
			if y < y1 || y >= y2 || x < x0 || x >= x1 || x == mx || y == my {
				col := uint8(144)
				if x < mx {
					col = 187
				}
				p.set(x, y, col)
			}
		}
	}
	for x := x0 - 4; x < x1+4; x++ {
		p.set(x, y2+2, 230) // o peitoril
		p.set(x, y2+3, 187)
	}
	for _, cx := range []int{x0 - 3 - int(4*k), x1 + 2} {
		for y := int(2 * k); y < y2+int(6*k); y++ {
			for x := cx; x <= cx+int(4*k); x++ {
				p.set(x, y, shade([]uint8{65, 108, 151}, 0.55+0.3*math.Sin(float64(x)*1.6/k), x, y))
			}
		}
	}
	for x := x0 - 3 - int(4*k); x <= x1+2+int(4*k); x++ {
		p.set(x, int(k), 137) // o varão
	}

	px, pb := fw-6*k, float64(edge)+1 // a costela-de-adão, no canto
	for y := int(pb - 9*k); y <= int(pb); y++ {
		r := (float64(y) - (pb - 9*k)) / k
		for x := int(px - (4-r*0.2)*k); x <= int(px+(4-r*0.2)*k); x++ {
			col := shade([]uint8{95, 131, 173}, 0.6-(float64(x)-px+4*k)/(8*k)*0.4, x, y)
			if r < 1 {
				col = 174
			}
			p.set(x, y, col)
		}
	}
	for i := range 6 {
		a, l := -math.Pi/2+(float64(i)-2.5)*0.5, (7+hash(i, 1, 99)*6)*k
		sx, sy := px, pb-9*k
		ex, ey := sx+math.Cos(a)*l, sy+math.Sin(a)*l
		for j := 0.0; j <= 1; j += 0.5 / l {
			p.set(int(sx+(ex-sx)*j), int(sy+(ey-sy)*j), 22)
		}
		p.oval(ex, ey, 3*k, 2.2*k, func(x, y int, d, dx, dy float64) uint8 {
			if d > 0.8 {
				return 22
			}
			v := 0.55 - dx*0.3
			if x%3 == 0 && d > 0.4 {
				v -= 0.3 // os recortes da folha
			}
			return shade([]uint8{28, 65, 71, 108}, v, x, y)
		})
	}

	ax, ay := fw*0.44, float64(edge)+5*k // onde o bicho dorme
	if s.dog {                           // o tapetinho branco
		for y := int(ay + 2*k); y <= int(ay+12*k); y++ {
			r := (float64(y) - ay - 2*k) / k
			for x := int(ax + (r*0.4-24)*k); x <= int(ax+(32-r*0.7)*k); x++ {
				col := uint8(252)
				if hash(x/4, y/2, 7) < 0.18 {
					col = 248 // as dobras
				}
				p.set(x, y, col)
			}
		}
		p.oval(ax+2*k, ay+3.5*k, 18*k, 3.15*k, func(x, y int, d, dx, dy float64) uint8 { return 245 })
	} else { // o tapete e a almofada
		p.oval(ax+2*k, ay+7*k, 36*k, 7.5*k, func(x, y int, d, dx, dy float64) uint8 {
			v := 0.4
			if math.Sin(d*20) > 0.3 {
				v = 0.75
			}
			return shade([]uint8{95, 131, 174}, v-dx*0.15-d*0.2, x, y)
		})
		p.oval(ax+2*k, ay+4.5*k, 21*k, 6.2*k, func(x, y int, d, dx, dy float64) uint8 { return shade([]uint8{23, 66}, 0.4-dx*0.3, x, y) })
		p.oval(ax+2*k, ay+3*k, 21*k, 6*k, func(x, y int, d, dx, dy float64) uint8 {
			v := 0.75 + dx*0.15 - dy*0.25
			if d > 0.9 {
				v -= 0.2
			}
			return shade([]uint8{23, 66, 109, 152}, v, x, y)
		})
		p.oval(ax+2*k, ay+2.2*k, 15.3*k, 2.88*k, func(x, y int, d, dx, dy float64) uint8 { return 66 }) // a marca do gato
	}

	breath := 1 + 0.035*math.Sin(t*2*math.Pi/4.2) // a respiração
	if s.dog {
		drawDog(p, ax, ay, k, t, breath)
	} else {
		drawCat(p, ax, ay, k, t, breath)
	}
	if b := s.board; b[2] > 0 { // o calendário, num prego: por cima do bicho, para o fundo do texto não mexer
		s.board = [4]int{}
		bx, by, bw, bh := b[0], 2*b[1], b[2], 2*b[3]
		nx := bx + bw/2
		p.set(nx, by-3, 247)
		for i := range 13 {
			f := float64(i) / 12
			p.set(nx-int(f*float64(bw/2-2)), by-3+int(f*3), 95)
			p.set(nx+int(f*float64(bw/2-2)), by-3+int(f*3), 95)
		}
		for y := by; y < by+bh; y++ {
			for x := bx; x < bx+bw; x++ {
				col := uint8(230) // o papel liso: a tarefa não some no pontilhado
				if y < by+2 {
					col = 131 // a espiral
					if y == by && (x-bx-3)%5 == 0 {
						col = 95
					}
				}
				p.set(x, y, col)
			}
		}
		for x := bx + 1; x <= bx+bw; x++ {
			p.set(x, by+bh, 95) // a sombra
		}
		for y := by + 1; y <= by+bh; y++ {
			p.set(bx+bw, y, 95)
		}
	}

	c.put(p)
}

// drawCat: o gato tigrado de máscara, peito e patas brancos, enrolado e de cara para quem olha,
// em tons chapados.
func drawCat(p *pix, ax, ay, k, t, breath float64) {
	tabby := func(v float64) uint8 {
		switch {
		case v > 0.72:
			return 137
		case v > 0.38:
			return 95
		}
		return 237
	}
	white := func(v float64) uint8 {
		if v > 0.55 {
			return 252
		}
		return 247
	}
	p.oval(ax+4*k, ay-(breath-1)*7*k, 14.5*k*breath, 7.2*k*breath, func(x, y int, d, dx, dy float64) uint8 {
		switch {
		case dy > 0.5 && dx > -0.3 && dx < 0.55:
			return white(0.9 - dy*0.4) // a barriga
		case dy < 0.45 && d < 0.9 && math.Sin((float64(x)*0.8+float64(y)*0.55)/k) > 0.55:
			return 236 // as listras
		}
		v := 0.62 - dx*0.25 - dy*0.35
		if d > 0.9 {
			v -= 0.3
		}
		return tabby(v)
	})
	p.oval(ax-4*k, ay+2.5*k, 5*k, 4.5*k, func(x, y int, d, dx, dy float64) uint8 { return white(0.9 - dx*0.2 - dy*0.3) }) // o peito
	p.oval(ax-2.5*k, ay+5.8*k, 3*k, 1.5*k, func(x, y int, d, dx, dy float64) uint8 { return white(0.95 - dy*0.5) })       // a pata
	flick := math.Mod(t, 9) < 0.7                                                                                         // a ponta do rabo mexe
	for u := 0.0; u <= 1; u += 0.025 / k {
		x, y := ax+(17-21*u)*k, ay+(5.4+1.1*math.Sin(u*math.Pi))*k
		if flick && u > 0.88 {
			y -= k
		}
		p.oval(x, y, 1.9*k, 1.5*k, func(_, _ int, d, dx, dy float64) uint8 {
			switch {
			case u > 0.93:
				return 234 // a ponta escura
			case int(u*10)%2 == 1:
				return 236
			case dy < -0.2:
				return 137
			}
			return 95
		})
	}
	hx, hy := ax-10*k, ay+0.5*k
	catEar(p, hx-3.5*k, hy-9*k, k, 95)
	catEar(p, hx+3.5*k, hy-9*k, k, 237)
	p.oval(hx, hy, 6.6*k, 5.4*k, func(x, _ int, d, dx, dy float64) uint8 {
		v := 0.85 - dx*0.25 - dy*0.2
		if d > 0.92 {
			v -= 0.3
		}
		switch {
		case dy > 0.1 || math.Abs(dx) < 0.18 && dy > -0.8:
			return white(v) // a máscara, subindo pela testa
		case math.Abs(dx) > 0.3 && dy < -0.3 && x%2 == 0:
			return 236
		}
		return tabby(v - 0.15)
	})
	for _, side := range []float64{-1, 1} { // os olhos fechados
		for u := -1.0; u <= 1; u += 0.5 / k {
			p.set(int(hx+(side*3+u)*k), int(hy+(1-math.Abs(u))*k), 235)
		}
	}
	p.set(int(hx), int(hy+2*k), 174) // o nariz
}

// catEar: a orelha, um triângulo de 4k pixels de altura com o rosa por dentro.
func catEar(p *pix, ax, ay, k float64, col uint8) {
	for j := 0.0; j <= 4*k; j++ {
		hw := j * 0.6
		for i := math.Ceil(-hw); i <= math.Floor(hw); i++ {
			c := col
			if j >= 2*k && math.Abs(i) < hw-0.9 {
				c = 138
			}
			p.set(int(ax+i), int(ay+j), c)
		}
	}
}

// drawDog: o cachorro amarelo de focinho branco, deitado com a cabeça nas patas, virado para a
// esquerda.
func drawDog(p *pix, ax, ay, k, t, breath float64) {
	coat := func(v float64) uint8 { // em tons chapados, como o gato
		switch {
		case v > 0.78:
			return 223
		case v > 0.45:
			return 180
		case v > 0.2:
			return 137
		}
		return 94
	}
	for u := 0.0; u <= 1; u += 0.05 / k { // o rabo
		p.oval(ax+(19+5*u)*k, ay+(3+3*u)*k, 1.5*k, 1.5*k, func(x, y int, d, dx, dy float64) uint8 {
			return coat(0.6 - dy*0.2 + u*0.2)
		})
	}
	p.oval(ax+4*k, ay-((breath-1)*6.5-0.5)*k, 16*k*breath, 6.6*k*breath, func(x, y int, d, dx, dy float64) uint8 {
		v := 0.68 - dx*0.2 - dy*0.35
		if d > 0.9 {
			v -= 0.2
		}
		if dy > 0.55 {
			v += 0.15 // a barriga, mais clara
		}
		return coat(v)
	})
	p.oval(ax+12*k, ay+2*k, 5.5*k, 4*k, func(x, y int, d, dx, dy float64) uint8 { // a coxa
		v := 0.62 - dx*0.25 - dy*0.3
		if d > 0.85 {
			v -= 0.25
		}
		return coat(v)
	})
	twitch, droop := 0.0, 0.0
	if math.Mod(t, 11) < 0.5 {
		twitch = math.Round(math.Sin(t * 25)) // sonhando: a pata mexe
	}
	if math.Mod(t, 8) < 0.3 {
		droop = -1 // e a orelha
	}
	p.oval(ax-8*k, ay+5.9*k, 3.4*k, 1.4*k, func(x, y int, d, dx, dy float64) uint8 { return coat(0.8 - dy*0.2) })
	p.oval(ax-14*k+twitch, ay+5.6*k, 3.6*k, 1.4*k, func(x, y int, d, dx, dy float64) uint8 { return coat(0.85 - dy*0.2) })
	p.oval(ax-11*k, ay, 6*k, 4.8*k, func(x, y int, d, dx, dy float64) uint8 { // a cabeça
		v := 0.72 - dx*0.25 - dy*0.3
		if d > 0.92 {
			v -= 0.2
		}
		return coat(v)
	})
	p.oval(ax-17*k, ay+2.2*k, 4.2*k, 2.6*k, func(x, y int, d, dx, dy float64) uint8 { // o focinho branco
		v := 0.85 - dy*0.4
		if dx < -0.85 {
			v -= 0.3
		}
		if v > 0.6 {
			return 253
		}
		return 247
	})
	for y := int(ay + k); y <= int(ay+2*k); y++ {
		for x := int(ax - 21*k); x <= int(ax-20*k); x++ {
			p.set(x, y, 233) // o nariz
		}
	}
	for x := int(ax - 14*k); x <= int(ax-13*k); x++ {
		p.set(x, int(ay-k), 236) // o olho fechado
	}
	p.oval(ax-8*k, ay+1.5*k+droop, 2.8*k, 4.8*k, func(x, y int, d, dx, dy float64) uint8 { // a orelha caída
		v := 0.15 - dy*0.1
		if dx < -0.5 {
			v += 0.1
		}
		return coat(v)
	})
}
