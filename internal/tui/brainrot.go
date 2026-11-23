package tui

// O wallpaper de piada "para desenvolvedores de verdade" (IDEIAS.md): dois vídeos, um em cada
// borda, e o relógio e a tarefa no meio, como nos cortes com a tela dividida, só que em dobro. Só
// aparece para quem pede na aba visual (#36). Os vídeos vêm de assets/videos/ (ADR 0005,
// CREDITS.md); o CC BY do Minecraft pede crédito para quem vê, e ele fica embaixo do vídeo.
// Vertical, como no celular: 54 colunas × 48 linhas no tamanho cheio, 8 quadros por segundo.

//go:generate go run brainrot_gen.go

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"sync"
)

var (
	//go:embed brainrot-minecraft.png
	minecraftPNG []byte
	//go:embed brainrot-subway.png
	subwayPNG []byte
)

// brainrotScene: os quadros em ANSI ocupam ~60 MB, então só são montados na primeira vez que
// alguém abre o brainrot, e daí servem para todo mundo.
var brainrotScene = sync.OnceValue(func() scene {
	return brainrot{mustClip(minecraftPNG), mustClip(subwayPNG)}
})

func mustClip(b []byte) *clip {
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		panic(err) // embutido no binário: se não abre, o build está quebrado
	}
	return newClip(img)
}

// clip: o vídeo em alguns tamanhos, do maior ao menor; na tela entra o maior que couber.
// Cada tamanho guarda os quadros já em ANSI de 256 cores, linha por linha (cada "▄" mostra 2 pixels).
type clip struct{ sizes []clipSize }

type clipSize struct {
	frames [][]string
	w, h   int // em células
}

var clipScales = []float64{1, 0.75, 0.5, 0.35}

// newClip: a tira tem os quadros um embaixo do outro, cada um com 16:9 de altura para cada 9 de
// largura (brainrot_gen.go).
func newClip(strip image.Image) *clip {
	b := strip.Bounds()
	fw := b.Dx()
	fh := fw * 16 / 9
	c := &clip{sizes: make([]clipSize, len(clipScales))}
	for y := b.Min.Y; y+fh <= b.Max.Y; y += fh {
		img := strip.(interface {
			SubImage(image.Rectangle) image.Image
		}).SubImage(image.Rect(b.Min.X, y, b.Max.X, y+fh))
		for i, k := range clipScales {
			w, h := int(float64(fw)*k), int(float64(fh)*k)/2*2
			rows := halfBlockRows(shrink(img, w, h))
			c.sizes[i] = clipSize{append(c.sizes[i].frames, rows), w, h / 2}
		}
	}
	return c
}

// shrink: redimensiona pegando o pixel mais próximo (pixelado de propósito).
func shrink(img image.Image, w, h int) image.Image {
	b := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			out.Set(x, y, img.At(b.Min.X+x*b.Dx()/w, b.Min.Y+y*b.Dy()/h))
		}
	}
	return out
}

// halfBlockRows: pixel de baixo na cor da frente (o "▄"), o de cima no fundo; só repete o código de cor quando muda.
func halfBlockRows(img image.Image) []string {
	b := img.Bounds()
	var rows []string
	for y := b.Min.Y; y+1 < b.Max.Y; y += 2 {
		var s strings.Builder
		s.WriteString("\x1b[4m") // sublinhado: cobre a última linha de pixels do "▄" (ver pet.Pixels)
		fg, bg := -1, -1
		for x := b.Min.X; x < b.Max.X; x++ {
			f, k := xterm256(img.At(x, y+1)), xterm256(img.At(x, y)) // "▄": o de cima no fundo (ver pet.Pixels)
			switch {
			case f != fg && k != bg:
				fmt.Fprintf(&s, "\x1b[38;5;%d;48;5;%dm", f, k)
			case f != fg:
				fmt.Fprintf(&s, "\x1b[38;5;%dm", f)
			case k != bg:
				fmt.Fprintf(&s, "\x1b[48;5;%dm", k)
			}
			fg, bg = f, k
			s.WriteString("▄")
		}
		s.WriteString("\x1b[0m")
		rows = append(rows, s.String())
	}
	return rows
}

// xterm256: a cor mais próxima no cubo 6×6×6 da paleta de 256 cores.
func xterm256(c color.Color) int {
	r, g, b, _ := c.RGBA()
	level := func(v uint32) int {
		v >>= 8
		switch {
		case v < 48:
			return 0
		case v < 115:
			return 1
		}
		return int((v - 35) / 40)
	}
	return 16 + 36*level(r) + 6*level(g) + level(b)
}

// size: o maior tamanho que cabe em maxW × maxH células.
func (s *clip) size(maxW, maxH int) clipSize {
	for _, z := range s.sizes {
		if z.w <= maxW && z.h <= maxH {
			return z
		}
	}
	return s.sizes[len(s.sizes)-1]
}

// put: o quadro do instante t, com a borda esquerda em x e centrado na altura. Devolve a linha
// logo abaixo do vídeo.
func (z clipSize) put(c *canvas, t float64, x int) int {
	f := z.frames[int(t*gardenFPS)%len(z.frames)]
	y0 := (c.h - len(f)) / 2
	for i, row := range f {
		if y := y0 + i; y >= 0 && y < c.h && x >= 0 && x+z.w <= c.w {
			c.raw[y] = append(c.raw[y], rawSeg{x, z.w, row})
		}
	}
	return y0 + len(f)
}

// brainrot: um vídeo em cada borda e o texto no meio. Estímulo demais, de propósito: é a piada.
type brainrot struct{ l, r *clip }

func (d brainrot) draw(c *canvas, t, dt float64) {
	maxW := (c.w-48)/2 - 2 // 48 colunas no meio para o relógio, a tarefa e quem mais está focando
	l, r := d.l.size(maxW, c.h-2), d.r.size(maxW, c.h-2)
	credit := l.put(c, t, 2)
	r.put(c, t, c.w-2-r.w)
	// o crédito que o CC BY pede vai por último: nada do que se mexe pode passar por cima dele
	defer c.text(2, credit, c.lang.t("vídeo de orbital (CC BY)"), "240")
	// o vão entre os vídeos, em cima e embaixo do relógio, que fica no meio com ~7 linhas para
	// cada lado: a prensa em cima e o sabão embaixo (#79)
	x0, x1 := 2+l.w+2, c.w-2-r.w-2
	gw := min(x1-x0, 32)
	if gw < 16 {
		return
	}
	gx := x0 + (x1-x0-gw)/2
	if rows := c.h/2 - 7 - 1; rows >= 5 {
		rows = min(rows, 14)
		drawPress(c, toyBox{gx, c.h/2 - 7 - rows, gw, rows}, t)
	}
	if rows := c.h - 5 - (c.h/2 + 7); rows >= 4 { // 5: o rodapé da prévia e a linha em volta dele
		rows = min(rows, 10)
		drawSoap(c, toyBox{gx, c.h/2 + 7, gw, rows}, t)
	}
	// e, nos vãos dos lados, a aura subindo
	drawAura(c, toyBox{x0, 1, gx - 1 - x0, c.h - 5}, t, 0)
	drawAura(c, toyBox{gx + gw + 1, 1, x1 - gx - gw - 1, c.h - 5}, t, 1)
}
