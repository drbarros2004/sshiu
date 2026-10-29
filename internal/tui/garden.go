package tui

// Modo jardim: a tela de descanso do foco. Um wallpaper animado em tela cheia, com o relógio, a
// tarefa e quem mais está focando por cima, conforme a aba visual (IDEIAS.md, docs/regras.md).
// Liga com j ou depois de 30 s sem tecla, só durante o foco; qualquer tecla volta sem executar nada.
//
// Wallpapers escritos do zero (ADR 0005, créditos em CREDITS.md): céu noturno, chuva e lâmpada de
// lava inspirados em projetos MIT/Apache; café, árvore e estrada são desenhos nossos.

import (
	"fmt"
	"math"
	"math/rand/v2"
	"plantao/internal/rooms"
	"plantao/internal/store"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	gardenFPS  = 8                // docs/regras.md
	gardenIdle = 30 * time.Second // sem tecla durante o foco: liga sozinho
)

// wallpapers: o id fica no banco; o nome aparece na aba visual.
type wallpaper struct {
	id, name string
	new      func() scene
}

var wallpapers = []wallpaper{
	{"sky", "céu noturno", func() scene { return &sky{} }},
	{"rain", "chuva lenta", func() scene { return &rain{} }},
	{"lava", "lâmpada de lava", func() scene { return &lava{} }},
	{"coffee", "café", func() scene { return &coffee{} }},
	{"tree", "árvore", func() scene { return &tree{} }},
	{"road", "estrada à noite", func() scene { return &road{} }},
}

var clockModes = []string{"mm:ss", "só os minutos", "nenhum"}

// garden: o modo jardim ligado. É ponteiro no modelo porque as cenas guardam estado (as partículas).
type garden struct {
	scene       scene
	start, last time.Time
	gen         int  // descarta ticks de um jardim anterior
	preview     bool // aberto pela aba visual: fica ligado sem tarefa
	frame       string
}

type gardenTickMsg struct {
	t   time.Time
	gen int
}

func gardenTick(gen int) tea.Cmd {
	return tea.Tick(time.Second/gardenFPS, func(t time.Time) tea.Msg { return gardenTickMsg{t, gen} })
}

// wallpaperFor: o escolhido, ou um sorteado por tarefa (o mesmo durante toda a tarefa).
func wallpaperFor(g store.Garden, f rooms.Focus) int {
	if i := slices.IndexFunc(wallpapers, func(w wallpaper) bool { return w.id == g.Wallpaper }); i >= 0 {
		return i
	}
	return int(f.Start.UnixNano()/int64(time.Millisecond)) % len(wallpapers)
}

// openGarden liga o modo jardim (preview: aberto pela aba visual, mesmo sem tarefa).
func (m model) openGarden(preview bool) (model, tea.Cmd) {
	gen := 1
	if m.garden != nil {
		gen = m.garden.gen + 1
	}
	i := wallpaperFor(m.vis, m.me().Focus)
	m.garden = &garden{scene: wallpapers[i].new(), start: m.now, last: m.now, gen: gen, preview: preview}
	m.garden.render(m)
	return m, gardenTick(gen)
}

// othersFocusing: quem mais está focando agora, nas suas salas, sem repetir.
func (m model) othersFocusing() []string {
	var nicks []string
	for _, code := range m.w.MyRooms(m.id) {
		for _, x := range m.w.Rooms[code].Members {
			if x.ID != m.id && x.Status == rooms.Focusing && !slices.Contains(nicks, x.Nick) {
				nicks = append(nicks, x.Nick)
			}
		}
	}
	return nicks
}

// focusingLine: "ana, bruno e duda também estão focando".
func focusingLine(nicks []string) string {
	switch {
	case len(nicks) == 0:
		return "ninguém mais está focando agora"
	case len(nicks) > 4:
		nicks = append(nicks[:3:3], fmt.Sprintf("mais %d", len(nicks)-3))
	}
	s := nicks[len(nicks)-1]
	if len(nicks) > 1 {
		s = strings.Join(nicks[:len(nicks)-1], ", ") + " e " + s
	}
	if len(nicks) == 1 {
		return s + " também está focando"
	}
	return s + " também estão focando"
}

// render avança a cena até agora e monta o quadro, com o texto por cima numa caixa.
func (g *garden) render(m model) {
	t, dt := m.now.Sub(g.start).Seconds(), min(m.now.Sub(g.last).Seconds(), 0.5)
	g.last = m.now
	w, h := max(m.width, 20), max(m.height, 10)
	c := newCanvas(w, h)
	g.scene.draw(c, t, dt)

	me := m.me()
	var lines, cols []string
	add := func(text, col string) { // um bloco, separado do anterior por uma linha vazia
		if len(lines) > 0 {
			lines, cols = append(lines, ""), append(cols, "")
		}
		for _, l := range strings.Split(text, "\n") {
			lines, cols = append(lines, l), append(cols, col)
		}
	}
	if g.preview {
		add("prévia de "+wallpapers[wallpaperFor(m.vis, me.Focus)].name+" · qualquer tecla volta", "244")
	}
	if me.Status == rooms.Focusing {
		left := me.Focus.Ends.Sub(m.now)
		switch m.vis.Clock {
		case 0:
			add(bigText(mmss(left)), "208")
		case 1: // só os minutos: o número muda uma vez por minuto
			add(bigText(fmt.Sprint(int(math.Ceil(max(left, 0).Minutes())))), "208")
			lines, cols = append(lines, "min"), append(cols, "208")
		}
		if m.vis.Task {
			add(me.Focus.Task, "252")
		}
		if m.vis.Room {
			add(focusingLine(m.othersFocusing()), "244")
		}
	}
	wd := 0
	for _, l := range lines {
		wd = max(wd, len([]rune(l)))
	}
	y0 := (h - len(lines)) / 2
	if len(lines) > 0 { // a caixa: o texto não briga com o wallpaper
		for y := y0 - 1; y <= y0+len(lines); y++ {
			for x := (w-wd)/2 - 3; x < (w+wd)/2+3; x++ {
				c.set(x, y, ' ', "")
			}
		}
	}
	for i, l := range lines {
		c.text((w-len([]rune(l)))/2, y0+i, l, cols[i])
	}
	g.frame = c.String()
}

// --- a tela é uma grade de células: as cenas e o relógio desenham nela ---

type cell struct {
	r rune
	c string // cor 256 ("" = cor do terminal)
}

type canvas struct {
	w, h  int
	cells []cell
}

func newCanvas(w, h int) *canvas {
	c := &canvas{w, h, make([]cell, w*h)}
	for i := range c.cells {
		c.cells[i] = cell{' ', ""}
	}
	return c
}

func (c *canvas) set(x, y int, r rune, col string) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	if r == ' ' {
		col = ""
	}
	c.cells[y*c.w+x] = cell{r, col}
}

func (c *canvas) text(x, y int, s, col string) {
	for _, r := range s {
		c.set(x, y, r, col)
		x++
	}
}

var styleOf = map[string]lipgloss.Style{}

func (c *canvas) String() string {
	var b strings.Builder
	for y := range c.h {
		var run strings.Builder
		col := ""
		flush := func() {
			if col == "" {
				b.WriteString(run.String())
			} else {
				st, ok := styleOf[col]
				if !ok {
					st = lipgloss.NewStyle().Foreground(lipgloss.Color(col))
					styleOf[col] = st
				}
				b.WriteString(st.Render(run.String()))
			}
			run.Reset()
		}
		for x := range c.w {
			cl := c.cells[y*c.w+x]
			if cl.c != col {
				flush()
				col = cl.c
			}
			run.WriteRune(cl.r)
		}
		flush()
		if y < c.h-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// --- os wallpapers: t é o tempo em segundos, dt o tempo desde o último quadro ---
// O movimento depende do tempo, não dos quadros: trocar o fps muda a suavidade, não a velocidade.

type scene interface {
	draw(c *canvas, t, dt float64)
}

// Céu noturno: estrelas paradas piscando devagar e, de vez em quando, uma estrela cadente.
type sky struct {
	w, h  int
	stars []struct {
		x, y  int
		f, ph float64
	}
	shoot struct{ x, y, at float64 } // at: quando a próxima começa
}

func (s *sky) draw(c *canvas, t, dt float64) {
	if s.w != c.w || s.h != c.h {
		s.w, s.h, s.stars = c.w, c.h, nil
		for range c.w * c.h / 30 {
			s.stars = append(s.stars, struct {
				x, y  int
				f, ph float64
			}{rand.N(c.w), rand.N(c.h), 0.2 + rand.Float64()*0.5, rand.Float64() * 6.3})
		}
		s.shoot.at = t + 8
	}
	for _, st := range s.stars {
		b := (math.Sin(t*st.f+st.ph) + 1) / 2
		switch {
		case b < 0.4:
			c.set(st.x, st.y, '.', "238")
		case b < 0.75:
			c.set(st.x, st.y, '.', "245")
		case b < 0.93:
			c.set(st.x, st.y, '+', "250")
		default:
			c.set(st.x, st.y, '*', "255")
		}
	}
	if t < s.shoot.at {
		return
	}
	if t-dt < s.shoot.at { // começou agora
		s.shoot.x, s.shoot.y = float64(rand.N(max(1, c.w/2))), float64(rand.N(max(1, c.h/3)))
	}
	e := t - s.shoot.at // segundos desde o começo: atravessa em ~1,2 s
	if e > 1.2 {
		s.shoot.at = t + 15 + rand.Float64()*20
		return
	}
	for i, col := range []string{"255", "250", "245", "240", "237"} {
		k := e - float64(i)*0.04
		c.set(int(s.shoot.x+k*30), int(s.shoot.y+k*8), map[bool]rune{true: '*', false: '·'}[i == 0], col)
	}
}

// Chuva lenta: duas camadas, as gotas do fundo menores, mais lentas e mais apagadas.
type rain struct {
	w, h int
	dr   []struct {
		x, y, v float64
		near    bool
	}
}

func (s *rain) draw(c *canvas, t, dt float64) {
	if s.w != c.w || s.h != c.h {
		s.w, s.h, s.dr = c.w, c.h, nil
		for range c.w * c.h / 45 {
			near := rand.N(3) == 0
			v := 3 + rand.Float64()*2
			if near {
				v = 7 + rand.Float64()*3
			}
			s.dr = append(s.dr, struct {
				x, y, v float64
				near    bool
			}{rand.Float64() * float64(c.w), rand.Float64() * float64(c.h), v, near})
		}
	}
	for i := range s.dr {
		d := &s.dr[i]
		d.y += d.v * dt
		if d.y >= float64(c.h) {
			d.y, d.x = 0, rand.Float64()*float64(c.w)
		}
		if d.near {
			c.set(int(d.x), int(d.y), '|', "67")
			c.set(int(d.x), int(d.y)-1, '¦', "24")
		} else {
			c.set(int(d.x), int(d.y), '.', "24")
		}
	}
}

// Lâmpada de lava: bolhas (metaballs) subindo e descendo bem devagar, cada uma no seu ritmo.
type lava struct {
	blobs []struct{ x, r, period, ph float64 }
}

func (s *lava) draw(c *canvas, t, dt float64) {
	if s.blobs == nil {
		for i := range 6 {
			s.blobs = append(s.blobs, struct{ x, r, period, ph float64 }{(float64(i) + 0.5) / 6, 3 + rand.Float64()*3, 40 + rand.Float64()*50, rand.Float64() * 6.3})
		}
	}
	type p struct{ x, y, r float64 }
	var ps []p
	for _, b := range s.blobs {
		x := b.x*float64(c.w) + 4*math.Sin(t*2*math.Pi/(b.period*1.7)+b.ph)
		y := (0.5 + 0.42*math.Sin(t*2*math.Pi/b.period+b.ph)) * float64(c.h)
		ps = append(ps, p{x, y, b.r})
	}
	for y := range c.h {
		for x := range c.w {
			v := 0.0
			for _, b := range ps {
				dx, dy := float64(x)-b.x, (float64(y)-b.y)*2 // a célula é duas vezes mais alta que larga
				v += b.r * b.r / (dx*dx + dy*dy + 0.01)
			}
			switch {
			case v > 1.6:
				c.set(x, y, '█', "130")
			case v > 1.1:
				c.set(x, y, '▓', "130")
			case v > 0.8:
				c.set(x, y, '▒', "94")
			case v > 0.6:
				c.set(x, y, '░', "52")
			}
		}
	}
}

// art desenha um desenho opaco: só os espaços antes do primeiro caractere de cada linha são transparentes.
func (c *canvas) art(x, y int, lines []string, col string) {
	for i, l := range lines {
		t := strings.TrimLeft(l, " ")
		c.text(x+len(l)-len(t), y+i, t, col)
	}
}

// Café: uma xícara parada e o vapor subindo, balançando e sumindo.
type coffee struct {
	steam []struct{ x, y, born, ph float64 }
	next  float64
}

var cup = []string{
	"  _________",
	" |         |__",
	" |         |  )",
	" |         |_/",
	"  \\_______/",
	"═════════════",
}

func (s *coffee) draw(c *canvas, t, dt float64) {
	x0, y0 := c.w*2/3, c.h-2-len(cup)
	c.art(x0, y0, cup, "180")
	if t >= s.next {
		s.next = t + 0.3
		s.steam = append(s.steam, struct{ x, y, born, ph float64 }{float64(x0 + 3 + rand.N(7)), float64(y0), t, rand.Float64() * 6.3})
	}
	live := s.steam[:0]
	for _, p := range s.steam {
		age := t - p.born // sobe ~1 linha/s e some em 8 s
		if age > 8 {
			continue
		}
		live = append(live, p)
		ch := ')'
		if math.Cos(t*1.2+p.ph) < 0 {
			ch = '('
		}
		col := []string{"252", "250", "248", "246", "244", "242", "240", "238"}[int(age)]
		c.set(int(p.x+1.5*math.Sin(t*1.2+p.ph)+age*0.3), int(p.y-age), ch, col)
	}
	s.steam = live
}

// Árvore: copa e tronco parados, folhas se soltando de vez em quando, caindo com o vento.
type tree struct {
	leaves []struct{ x, y, vx, vy, ph, landed float64 }
	next   float64
}

func (s *tree) draw(c *canvas, t, dt float64) {
	ground := c.h - 2
	tx, rx, ry := c.w/5, 13.0, 6.0 // centro do tronco e raios da copa
	cy := float64(ground - 4 - 6)
	for x := range c.w {
		c.set(x, ground, '_', "58")
	}
	for y := ground - 4; y < ground; y++ {
		c.text(tx-1, y, "| |", "94")
	}
	c.text(tx-3, ground, "/   \\", "94")
	for y := int(cy - ry); y <= int(cy+ry); y++ {
		for x := tx - int(rx); x <= tx+int(rx); x++ {
			dx, dy := (float64(x)-float64(tx))/rx, (float64(y)-cy)/ry
			if dx*dx+dy*dy <= 1 {
				h := int(uint32(x*73856093^y*19349663) % 5) // folhagem fixa: mesmo caractere sempre na mesma célula
				c.set(x, y, []rune("&%@#&")[h], []string{"22", "28", "64", "70", "28"}[h])
			}
		}
	}
	if t >= s.next {
		s.next = t + 0.8 + rand.Float64()
		a := rand.Float64() * 2 * math.Pi
		s.leaves = append(s.leaves, struct{ x, y, vx, vy, ph, landed float64 }{
			float64(tx) + rx*0.9*math.Cos(a), cy + ry*0.9*math.Sin(a), 0.5 + rand.Float64(), 0.7 + rand.Float64()*0.5, rand.Float64() * 6.3, 0})
	}
	live := s.leaves[:0]
	for _, l := range s.leaves {
		if l.landed == 0 {
			l.x += l.vx * dt
			l.y += l.vy * dt
			if l.y >= float64(ground) {
				l.y, l.landed = float64(ground), t
			}
		} else if t-l.landed > 5 { // no chão por 5 s, depois some: nada se acumula
			continue
		}
		live = append(live, l)
		x := l.x
		if l.landed == 0 {
			x += 1.2 * math.Sin(t*1.5+l.ph)
		}
		c.set(int(x), int(l.y), []rune(",`'*")[int(l.ph)%4], []string{"130", "136", "166", "178"}[int(l.ph)%4])
	}
	s.leaves = live
}

// Estrada à noite: o carro parado na tela e o mundo passando em camadas (paralaxe).
type road struct {
	stars [][2]int
	w, h  int
}

var car = []string{
	"     ______",
	" ___/__|___\\____",
	"'--(_)----(_)--'",
}

func (s *road) draw(c *canvas, t, dt float64) {
	if s.w != c.w || s.h != c.h {
		s.w, s.h, s.stars = c.w, c.h, nil
		for range c.w * c.h / 60 {
			s.stars = append(s.stars, [2]int{rand.N(c.w), rand.N(max(1, c.h-12))})
		}
	}
	for _, st := range s.stars {
		c.set(st[0], st[1], '.', "238")
	}
	base := c.h - 4      // linha das rodas
	for x := range c.w { // morros ao fundo: devagar
		u := float64(x) + t*1.5
		top := base - 3 - int(3+2*math.Sin(u*0.05)+1.5*math.Sin(u*0.13))
		for y := top; y < base; y++ {
			c.set(x, y, '░', "238")
		}
	}
	for x := range c.w { // postes: velocidade média
		if (x+int(t*6))%28 == 0 {
			for y := base - 5; y < base; y++ {
				c.set(x, y, '│', "240")
			}
			c.text(x-1, base-5, "─┼─", "240")
		}
	}
	c.art(c.w/4, base-2, car, "174")
	for x := range c.w {
		c.set(x, base+1, '─', "240")
		if (x+int(t*14))%8 < 3 { // faixa: rápida
			c.set(x, base+2, '━', "178")
		}
	}
}
