// Harness do trailer: entra no sshiu por SSH, joga a saída num emulador vt e salva a tela em PNG.
// Uso: harness -user ana -cols 160 -rows 45 passo passo ...
// Passos: sleep:2.5 | wait:texto | type:texto | enter | esc | left | right | up | down | tab | key:x | shot:arq.png | dump
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/vt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	defFg = color.RGBA{0xd8, 0xd8, 0xd8, 0xff}
	defBg = color.RGBA{0x12, 0x12, 0x14, 0xff}
)

func main() {
	user := flag.String("user", "ana", "")
	addr := flag.String("addr", "localhost:23235", "")
	cols := flag.Int("cols", 160, "")
	rows := flag.Int("rows", 45, "")
	cw := flag.Int("cw", 12, "largura da célula em px")
	ch := flag.Int("ch", 24, "altura da célula em px")
	flag.Parse()

	cfg := &ssh.ClientConfig{User: *user, Auth: []ssh.AuthMethod{ssh.PublicKeys(userKey(*user))},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), Timeout: 5 * time.Second}
	client, err := ssh.Dial("tcp", *addr, cfg)
	if err != nil {
		log.Fatal(err)
	}
	sess, _ := client.NewSession()
	if err := sess.RequestPty("xterm-256color", *rows, *cols, ssh.TerminalModes{}); err != nil {
		log.Fatal(err)
	}
	stdin, _ := sess.StdinPipe()
	stdout, _ := sess.StdoutPipe()
	if err := sess.Shell(); err != nil {
		log.Fatal(err)
	}

	em := vt.NewEmulator(*cols, *rows)
	var mu sync.Mutex
	go func() { // respostas do terminal (consultas do Bubble Tea) voltam para o servidor
		buf := make([]byte, 4096)
		for {
			n, err := em.Read(buf)
			if err != nil {
				return
			}
			stdin.Write(buf[:n])
		}
	}()
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				mu.Lock()
				em.Write(buf[:n])
				mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()

	keys := map[string]string{"enter": "\r", "esc": "\x1b", "tab": "\t", "up": "\x1b[A", "down": "\x1b[B", "right": "\x1b[C", "left": "\x1b[D", "space": " ", "bs": "\x7f"}
	for _, step := range flag.Args() {
		name, arg, _ := strings.Cut(step, ":")
		switch {
		case name == "sleep":
			d, _ := strconv.ParseFloat(arg, 64)
			time.Sleep(time.Duration(d * float64(time.Second)))
		case name == "wait":
			deadline := time.Now().Add(15 * time.Second)
			for {
				mu.Lock()
				s := em.String()
				mu.Unlock()
				if strings.Contains(s, arg) {
					break
				}
				if time.Now().After(deadline) {
					fmt.Println(s)
					log.Fatalf("não apareceu: %q", arg)
				}
				time.Sleep(100 * time.Millisecond)
			}
		case name == "until": // until:tecla:texto aperta a tecla até o texto aparecer (no máximo 20 vezes)
			key, text, _ := strings.Cut(arg, ":")
			for i := 0; ; i++ {
				mu.Lock()
				found := strings.Contains(em.String(), text)
				mu.Unlock()
				if found {
					break
				}
				if i == 20 {
					log.Fatalf("until: %q não apareceu", text)
				}
				stdin.Write([]byte(keys[key]))
				time.Sleep(300 * time.Millisecond)
			}
		case name == "type":
			for _, r := range arg {
				stdin.Write([]byte(string(r)))
				time.Sleep(90 * time.Millisecond)
			}
		case name == "key":
			stdin.Write([]byte(arg))
		case keys[name] != "":
			stdin.Write([]byte(keys[name]))
		case name == "rec": // rec:pasta grava 30 quadros por segundo até o stop
			recDir, recStop = arg, make(chan bool)
			recDone = make(chan [][]uv.Cell)
			go record(em, &mu, *cols, *rows, recStop, recDone)
		case name == "stop":
			close(recStop)
			frames := <-recDone
			os.MkdirAll(recDir, 0o755)
			var wg sync.WaitGroup
			sem := make(chan bool, 8)
			for i, f := range frames {
				wg.Add(1)
				sem <- true
				go func() {
					defer func() { <-sem; wg.Done() }()
					img := raster(func(x, y int) *uv.Cell { return &f[y*(*cols)+x] }, *cols, *rows, *cw, *ch)
					out, _ := os.Create(fmt.Sprintf("%s/%05d.png", recDir, i))
					png.Encode(out, img)
					out.Close()
				}()
			}
			wg.Wait()
			fmt.Printf("%s: %d quadros\n", recDir, len(frames))
		case name == "dump":
			mu.Lock()
			fmt.Println(em.String())
			mu.Unlock()
		case name == "shot":
			mu.Lock()
			img := raster(em.CellAt, *cols, *rows, *cw, *ch)
			mu.Unlock()
			f, _ := os.Create(arg)
			png.Encode(f, img)
			f.Close()
		default:
			log.Fatalf("passo desconhecido: %s", step)
		}
		time.Sleep(150 * time.Millisecond)
	}
	sess.Close()
}

// userKey guarda uma chave ed25519 por pessoa em keys/<user>.seed: a mesma chave é a mesma identidade no sshiu.
func userKey(user string) ssh.Signer {
	path := "keys/" + user + ".seed"
	seed, err := os.ReadFile(path)
	if err != nil {
		seed = make([]byte, ed25519.SeedSize)
		rand.Read(seed)
		os.MkdirAll("keys", 0o700)
		os.WriteFile(path, seed, 0o600)
	}
	signer, err := ssh.NewSignerFromKey(ed25519.NewKeyFromSeed(seed))
	if err != nil {
		log.Fatal(err)
	}
	return signer
}

var (
	recDir  string
	recStop chan bool
	recDone chan [][]uv.Cell
)

// record copia a grade de células a 30 fps; rasterizar fica para depois, senão atrasa os quadros.
func record(em *vt.Emulator, mu *sync.Mutex, cols, rows int, stop chan bool, done chan [][]uv.Cell) {
	var frames [][]uv.Cell
	t := time.NewTicker(time.Second / 30)
	defer t.Stop()
	for {
		select {
		case <-stop:
			done <- frames
			return
		case <-t.C:
			f := make([]uv.Cell, cols*rows)
			mu.Lock()
			for y := 0; y < rows; y++ {
				for x := 0; x < cols; x++ {
					if c := em.CellAt(x, y); c != nil {
						f[y*cols+x] = *c
					}
				}
			}
			mu.Unlock()
			frames = append(frames, f)
		}
	}
}

var fonts [2]*opentype.Font

func parseFont(path string) *opentype.Font {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	f, err := opentype.Parse(b)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func init() {
	home, _ := os.UserHomeDir()
	fonts[0] = parseFont(home + "/Library/Fonts/JetBrainsMono-Regular.ttf")
	fonts[1] = parseFont(home + "/Library/Fonts/JetBrainsMono-Bold.ttf")
}

// raster desenha a grade; cria as próprias faces porque uma font.Face não é segura entre goroutines.
func raster(cellAt func(x, y int) *uv.Cell, cols, rows, cw, ch int) *image.RGBA {
	var faces [2]font.Face
	for i, f := range fonts {
		faces[i], _ = opentype.NewFace(f, &opentype.FaceOptions{Size: float64(ch) * 0.78, DPI: 72, Hinting: font.HintingFull})
	}
	img := image.NewRGBA(image.Rect(0, 0, cols*cw, rows*ch))
	draw.Draw(img, img.Bounds(), &image.Uniform{defBg}, image.Point{}, draw.Src)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			c := cellAt(x, y)
			if c == nil || c.Width == 0 {
				continue
			}
			fg, bg := toRGBA(c.Style.Fg, defFg), toRGBA(c.Style.Bg, defBg)
			if c.Style.Attrs&uv.AttrReverse != 0 {
				fg, bg = bg, fg
			}
			if c.Style.Attrs&uv.AttrFaint != 0 {
				fg = mix(fg, bg, 0.5)
			}
			w := max(c.Width, 1)
			cell := image.Rect(x*cw, y*ch, (x+w)*cw, (y+1)*ch)
			fill(img, cell, bg)
			r := []rune(c.Content)
			if len(r) == 0 || r[0] == ' ' {
				continue
			}
			if blocks(img, r[0], cell, fg, bg) || pasteEmoji(img, r, cell) {
				continue
			}
			face := faces[0]
			if c.Style.Attrs&uv.AttrBold != 0 {
				face = faces[1]
			}
			d := font.Drawer{Dst: img, Src: &image.Uniform{fg}, Face: face}
			m := face.Metrics()
			base := cell.Min.Y + (ch-(m.Ascent+m.Descent).Ceil())/2 + m.Ascent.Ceil()
			d.Dot = fixed.P(cell.Min.X, base)
			d.DrawString(c.Content)
		}
	}
	return img
}

// blocks desenha à mão o que precisa encostar na borda da célula: meio bloco, quadrantes e linhas de caixa.
func blocks(img *image.RGBA, r rune, c image.Rectangle, fg, bg color.RGBA) bool {
	x0, y0, x1, y1 := c.Min.X, c.Min.Y, c.Max.X, c.Max.Y
	mx, my := (x0+x1)/2, (y0+y1)/2
	quad := map[rune][4]bool{ // sup-esq, sup-dir, inf-esq, inf-dir
		'▀': {true, true, false, false}, '▄': {false, false, true, true}, '█': {true, true, true, true},
		'▌': {true, false, true, false}, '▐': {false, true, false, true},
		'▘': {true, false, false, false}, '▝': {false, true, false, false}, '▖': {false, false, true, false}, '▗': {false, false, false, true},
		'▚': {true, false, false, true}, '▞': {false, true, true, false},
		'▙': {true, false, true, true}, '▛': {true, true, true, false}, '▜': {true, true, false, true}, '▟': {false, true, true, true},
	}
	if q, ok := quad[r]; ok {
		rects := []image.Rectangle{image.Rect(x0, y0, mx, my), image.Rect(mx, y0, x1, my), image.Rect(x0, my, mx, y1), image.Rect(mx, my, x1, y1)}
		for i, on := range q {
			if on {
				fill(img, rects[i], fg)
			}
		}
		return true
	}
	if s, ok := map[rune]float64{'░': 0.25, '▒': 0.5, '▓': 0.75}[r]; ok {
		fill(img, c, mix(fg, bg, s))
		return true
	}
	// linhas de caixa finas: cima, baixo, esquerda, direita
	lines := map[rune][4]bool{
		'─': {false, false, true, true}, '│': {true, true, false, false},
		'┌': {false, true, false, true}, '╭': {false, true, false, true}, '┐': {false, true, true, false}, '╮': {false, true, true, false},
		'└': {true, false, false, true}, '╰': {true, false, false, true}, '┘': {true, false, true, false}, '╯': {true, false, true, false},
		'├': {true, true, false, true}, '┤': {true, true, true, false}, '┬': {false, true, true, true}, '┴': {true, false, true, true}, '┼': {true, true, true, true},
	}
	if corner, ok := map[rune][2]int{'╭': {1, 1}, '╮': {-1, 1}, '╰': {1, -1}, '╯': {-1, -1}}[r]; ok {
		t := max(1, (x1-x0)/8)
		R := float64(min(x1-x0, y1-y0)) / 2
		dx, dy := corner[0], corner[1]
		cx, cy := float64(mx)+float64(dx)*R, float64(my)+float64(dy)*R
		for a := 0.0; a <= math.Pi/2; a += 0.02 { // quarto de círculo até o meio da célula
			px := cx - float64(dx)*R*math.Cos(a)
			py := cy - float64(dy)*R*math.Sin(a)
			fill(img, image.Rect(int(px)-t/2, int(py)-t/2, int(px)-t/2+t, int(py)-t/2+t), fg)
		}
		if dy > 0 {
			fill(img, image.Rect(mx-t/2, int(cy), mx-t/2+t, y1), fg)
		} else {
			fill(img, image.Rect(mx-t/2, y0, mx-t/2+t, int(cy)+1), fg)
		}
		if dx > 0 {
			fill(img, image.Rect(int(cx), my-t/2, x1, my-t/2+t), fg)
		} else {
			fill(img, image.Rect(x0, my-t/2, int(cx)+1, my-t/2+t), fg)
		}
		return true
	}
	if l, ok := lines[r]; ok {
		t := max(1, (x1-x0)/8)
		if l[0] {
			fill(img, image.Rect(mx-t/2, y0, mx-t/2+t, my+t/2+1), fg)
		}
		if l[1] {
			fill(img, image.Rect(mx-t/2, my-t/2, mx-t/2+t, y1), fg)
		}
		if l[2] {
			fill(img, image.Rect(x0, my-t/2, mx+t/2+1, my-t/2+t), fg)
		}
		if l[3] {
			fill(img, image.Rect(mx-t/2, my-t/2, x1, my-t/2+t), fg)
		}
		return true
	}
	return false
}

func fill(img *image.RGBA, r image.Rectangle, c color.RGBA) {
	draw.Draw(img, r, &image.Uniform{c}, image.Point{}, draw.Src)
}

func toRGBA(c color.Color, def color.RGBA) color.RGBA {
	if c == nil {
		return def
	}
	r, g, b, _ := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 0xff}
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	l := func(x, y uint8) uint8 { return uint8(float64(x)*t + float64(y)*(1-t)) }
	return color.RGBA{l(a.R, b.R), l(a.G, b.G), l(a.B, b.B), 0xff}
}

var (
	emojiCache = map[string]image.Image{}
	emojiMu    sync.Mutex
)

// pasteEmoji cola o PNG da Apple Color Emoji (gerado pelo emoji.swift) no lugar do glifo.
func pasteEmoji(img *image.RGBA, r []rune, cell image.Rectangle) bool {
	var parts []string
	for _, x := range r {
		if x != 0xfe0f {
			parts = append(parts, strconv.FormatInt(int64(x), 16))
		}
	}
	name := strings.Join(parts, "-")
	emojiMu.Lock()
	defer emojiMu.Unlock()
	e, ok := emojiCache[name]
	if !ok {
		if f, err := os.Open("emoji/" + name + ".png"); err == nil {
			e, _ = png.Decode(f)
			f.Close()
		}
		emojiCache[name] = e
	}
	if e == nil {
		return false
	}
	side := min(cell.Dx(), cell.Dy())
	x0 := cell.Min.X + (cell.Dx()-side)/2
	draw.CatmullRom.Scale(img, image.Rect(x0, cell.Min.Y, x0+side, cell.Min.Y+side), e, e.Bounds(), draw.Over, nil)
	return true
}
