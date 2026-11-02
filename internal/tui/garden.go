package tui

// Modo jardim: a tela de descanso do foco. Um wallpaper animado em tela cheia, com o relógio, a
// tarefa e quem mais está focando por cima, conforme a aba visual (IDEIAS.md, docs/regras.md).
// Liga com j ou depois de 30 s sem tecla, só durante o foco. ←/→ trocam o wallpaper (a troca vale
// para a tarefa); qualquer outra tecla volta sem executar nada.
//
// Wallpapers escritos do zero (ADR 0005, créditos em CREDITS.md): céu noturno, chuva e lâmpada de
// lava inspirados em projetos MIT/Apache; café, árvore e estrada são desenhos nossos, em pixels
// (garden_pixels.go).

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
)

const (
	gardenFPS  = 8                // docs/regras.md
	gardenIdle = 30 * time.Second // sem tecla durante o foco: liga sozinho
	gardenHint = 5 * time.Second  // o nome do wallpaper e "← → troca", ao abrir e a cada troca
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
// Serve para a tela cheia e para a coluna da direita da sala solo (w/h e bare, ADR 0006).
type garden struct {
	scene       scene
	start, last time.Time
	gen         int       // descarta ticks de um jardim anterior
	preview     bool      // aberto pela aba visual
	keep        bool      // fica ligado sem tarefa (a prévia e a sala solo)
	bare        bool      // sem texto por cima: na sala solo quem mostra relógio e tarefa é o painel
	w, h        int       // 0 = a tela toda
	ticking     bool      // já tem um tick de 8 fps a caminho
	wall        int       // o wallpaper na tela
	hint        time.Time // até quando mostra o nome do wallpaper e "← → troca"
	frame       string
}

type gardenTickMsg struct {
	t   time.Time
	gen int
}

func gardenTick(gen int) tea.Cmd {
	return tea.Tick(time.Second/gardenFPS, func(t time.Time) tea.Msg { return gardenTickMsg{t, gen} })
}

// sideTickMsg: o jardim da coluna da sala solo, no mesmo ritmo da tela cheia.
type sideTickMsg struct {
	t   time.Time
	gen int
}

func sideTick(gen int) tea.Cmd {
	return tea.Tick(time.Second/gardenFPS, func(t time.Time) tea.Msg { return sideTickMsg{t, gen} })
}

// wallpaperIndex: a posição do wallpaper na lista (-1 = sorteado).
func wallpaperIndex(id string) int {
	return slices.IndexFunc(wallpapers, func(w wallpaper) bool { return w.id == id })
}

// cycleWallpaper: o próximo (step 1) ou o anterior (-1) da aba visual, passando pelo sorteado ("").
func cycleWallpaper(id string, step int) string {
	n := len(wallpapers) + 1
	if i := (wallpaperIndex(id)+1+step+n)%n - 1; i >= 0 {
		return wallpapers[i].id
	}
	return ""
}

// wallpaperFor: o escolhido, ou um sorteado por tarefa (o mesmo durante toda a tarefa).
// Sem tarefa (a prévia), sorteia na hora.
func wallpaperFor(g store.Garden, f rooms.Focus) int {
	if i := wallpaperIndex(g.Wallpaper); i >= 0 {
		return i
	}
	if f.Start.IsZero() { // sem tarefa em andamento não há instante de onde sortear
		return rand.IntN(len(wallpapers))
	}
	return int(f.Start.UnixMilli() % int64(len(wallpapers)))
}

// gardenWall: o wallpaper ao abrir o jardim. Fora da prévia, vale a troca feita nesta tarefa.
func (m model) gardenWall(preview bool) int {
	f := m.me().Focus
	if !preview && !f.Start.IsZero() && f.Start.Equal(m.wallFor) {
		return m.wall
	}
	return wallpaperFor(m.vis, f)
}

// openGarden liga o modo jardim (preview: aberto pela aba visual, mesmo sem tarefa).
func (m model) openGarden(preview bool) (model, tea.Cmd) {
	gen := 1
	if m.garden != nil {
		gen = m.garden.gen + 1
	}
	i := m.gardenWall(preview)
	// na sala solo o jardim aberto sem tarefa é a outra versão da tela e fica ligado; aberto
	// durante o foco, ele fecha quando o tempo esgota, senão você não vê que precisa responder.
	m.garden = &garden{scene: wallpapers[i].new(), start: m.now, last: m.now, gen: gen,
		preview: preview, keep: preview || m.solo() && !m.active(), wall: i}
	if !preview {
		m.garden.hint = m.now.Add(gardenHint)
	}
	m.garden.render(m)
	return m, gardenTick(gen)
}

// swapWallpaper: ←/→ no jardim g, o da tela cheia ou o da coluna da sala solo. Na prévia, muda a
// escolha da aba visual (passando pelo sorteado); durante a tarefa, troca só para esta tarefa.
func (m model) swapWallpaper(g *garden, step int) (model, tea.Cmd) {
	if g.preview {
		m.vis.Wallpaper, m.notice = cycleWallpaper(m.vis.Wallpaper, step), ""
		g.wall = wallpaperFor(m.vis, m.me().Focus)
	} else {
		n := len(wallpapers)
		g.wall = (g.wall + step + n) % n
		m.wall, m.wallFor = g.wall, m.me().Focus.Start
		g.hint = m.now.Add(gardenHint)
	}
	g.scene, g.start, g.last = wallpapers[g.wall].new(), m.now, m.now
	g.render(m)
	return m, nil
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

// spot: o wallpaper que escolhe onde o texto fica e em que cores, em vez da caixa no meio. Ele
// recebe o tamanho do bloco de texto (tw × th células) antes de desenhar, para abrir lugar para
// ele no desenho (a lousa do café). Devolve o canto de cima à esquerda e as cores do resto, do
// relógio e da tarefa.
type spot interface {
	spot(w, h, tw, th int) (x, y int, ink [3]string)
}

// splitTask: a tarefa comprida em duas linhas, quebrando no espaço mais perto do meio. Sem espaço
// por perto, corta a palavra no meio com hífen.
func splitTask(s string) []string {
	const fits, slack = 40, 12 // até 40 cabe numa linha; o espaço vale até 12 do meio
	r := []rune(s)
	if len(r) <= fits {
		return []string{s}
	}
	mid := len(r) / 2
	for d := 0; d <= slack; d++ {
		for _, i := range []int{mid - d, mid + d} {
			if r[i] == ' ' {
				return []string{strings.TrimSpace(string(r[:i])), strings.TrimSpace(string(r[i:]))}
			}
		}
	}
	return []string{string(r[:mid]) + "-", string(r[mid:])}
}

// render avança a cena até agora e monta o quadro, com o texto por cima: numa caixa no meio, ou
// no lugar que a cena escolheu.
func (g *garden) render(m model) {
	t, dt := m.now.Sub(g.start).Seconds(), min(m.now.Sub(g.last).Seconds(), 0.5)
	g.last = m.now
	w, h := max(m.width, 20), max(m.height, 10)
	if g.w > 0 { // a coluna da sala solo, não a tela toda
		w, h = g.w, g.h
	}
	c := newCanvas(w, h)
	if g.bare {
		g.scene.draw(c, t, dt)
		g.frame = c.String()
		return
	}

	const dim, clock, task = 0, 1, 2 // o papel de cada linha, que vira a cor
	me := m.me()
	var lines []string
	var roles []int
	add := func(text string, role int) { // um bloco, separado do anterior por uma linha vazia
		if len(lines) > 0 {
			lines, roles = append(lines, ""), append(roles, dim)
		}
		for _, l := range strings.Split(text, "\n") {
			lines, roles = append(lines, l), append(roles, role)
		}
	}
	var hint []string // o nome do wallpaper fica no rodapé, fora do bloco: some sem mexer no resto
	if g.preview {
		name := wallpapers[g.wall].name
		if wallpaperIndex(m.vis.Wallpaper) < 0 {
			name = "sorteado a cada tarefa · agora, " + name
		}
		hint = []string{"prévia: " + name, "← → troca · qualquer tecla volta"}
	} else if m.now.Before(g.hint) {
		hint = []string{wallpapers[g.wall].name + " · ← → troca"}
	}
	if me.Status == rooms.Focusing {
		left := me.Focus.Ends.Sub(m.now)
		switch m.vis.Clock {
		case 0:
			add(bigText(mmss(left)), clock)
		case 1: // só os minutos: o número muda uma vez por minuto
			add(bigText(fmt.Sprint(int(math.Ceil(max(left, 0).Minutes())))), clock)
			lines, roles = append(lines, "min"), append(roles, clock)
		}
		if m.vis.Task {
			add(strings.Join(splitTask(me.Focus.Task), "\n"), task)
		}
		if m.vis.Room {
			add(focusingLine(m.othersFocusing()), dim)
		}
	}
	wd := 0
	for _, l := range lines {
		wd = max(wd, len([]rune(l)))
	}
	y0 := (h - len(lines)) / 2
	cx := max(wd/2+3, min(w/2, w-wd/2-4)) // o texto nunca passa da borda
	ink := [3]string{"244", "208", "252"}
	sp, placed := g.scene.(spot)
	if placed && len(lines) > 0 {
		var x int
		x, y0, ink = sp.spot(w, h, wd, len(lines))
		cx = min(max(x, 1), w-wd-1) + wd/2
		y0 = min(max(y0, 0), h-len(lines))
	}
	g.scene.draw(c, t, dt)
	put := func(ls []string, cx, y0, wd int, col func(i int) string) { // linhas centradas em cx
		if len(ls) == 0 {
			return
		}
		if !placed { // a caixa: o texto não briga com o wallpaper
			for y := y0 - 1; y <= y0+len(ls); y++ {
				for x := cx - wd/2 - 3; x < cx+wd/2+3; x++ {
					c.set(x, y, ' ', "")
				}
			}
		}
		for i, l := range ls {
			x := cx - len([]rune(l))/2
			if placed {
				c.ink(x, y0+i, l, col(i))
			} else {
				c.text(x, y0+i, l, col(i))
			}
		}
	}
	put(lines, cx, y0, wd, func(i int) string { return ink[roles[i]] })
	hw := 0
	for _, l := range hint {
		hw = max(hw, len([]rune(l)))
	}
	put(hint, w/2, h-len(hint)-1, hw, func(int) string { return ink[dim] })
	g.frame = c.String()
}

// --- a tela é uma grade de células: as cenas e o relógio desenham nela ---

type cell struct {
	r    rune
	c, b string // cor 256 da frente e do fundo ("" = a do terminal)
}

type canvas struct {
	w, h  int
	cells []cell
	raw   map[int][]rawSeg // por linha: trechos já em ANSI (os vídeos), que cobrem as células de baixo
}

type rawSeg struct {
	x, w int
	s    string
}

func newCanvas(w, h int) *canvas {
	c := &canvas{w: w, h: h, cells: make([]cell, w*h), raw: map[int][]rawSeg{}}
	for i := range c.cells {
		c.cells[i] = cell{r: ' '}
	}
	return c
}

// set escreve um caractere sem fundo: o texto por cima de um wallpaper em pixels apaga o pixel.
func (c *canvas) set(x, y int, r rune, col string) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	if r == ' ' {
		col = ""
	}
	c.cells[y*c.w+x] = cell{r: r, c: col}
}

func (c *canvas) text(x, y int, s, col string) {
	for _, r := range s {
		c.set(x, y, r, col)
		x++
	}
}

// ink escreve por cima do desenho sem apagar o fundo: a letra pega a cor de fundo da célula, e
// o espaço deixa o desenho aparecer.
func (c *canvas) ink(x, y int, s, col string) {
	for _, r := range s {
		if r != ' ' && x >= 0 && y >= 0 && x < c.w && y < c.h {
			cl := &c.cells[y*c.w+x]
			bg := cl.b
			if bg == "" && cl.r != '▀' { // "▄" e o meio bloco sem fundo: o pixel aceso é o de baixo
				bg = cl.c
			}
			*cl = cell{r, col, bg}
		}
		x++
	}
}

// String: a grade em ANSI, com o código de cor só quando a cor muda.
func (c *canvas) String() string {
	var b strings.Builder
	for y := range c.h {
		fg, bg := "", ""
		sgr := func(f, k string) {
			if f == fg && k == bg {
				return
			}
			b.WriteString("\x1b[0")
			if f != "" {
				b.WriteString(";38;5;" + f)
			}
			if k != "" {
				b.WriteString(";48;5;" + k)
			}
			b.WriteByte('m')
			fg, bg = f, k
		}
		segs := c.raw[y]
		for x := 0; x < c.w; x++ {
			if i := slices.IndexFunc(segs, func(s rawSeg) bool { return s.x == x }); i >= 0 {
				sgr("", "")
				b.WriteString(segs[i].s) // termina com o reset
				x += segs[i].w - 1
				continue
			}
			cl := c.cells[y*c.w+x]
			sgr(cl.c, cl.b)
			b.WriteRune(cl.r)
		}
		sgr("", "")
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
