package tui

// PROTÓTIPO: o wallpaper de piada "para desenvolvedores de verdade" (IDEIAS.md). Um vídeo num lado da
// tela, o relógio e a tarefa do outro, como nos cortes com a tela dividida.
// ponytail: os quadros vêm de PLANTAO_VIDEOS (uma pasta por vídeo, PNGs feitos com o ffmpeg), fora do
// git, até a licença estar decidida; embutir com go:embed se o wallpaper ficar.
//
//	ffmpeg -i subway.mp4 -vf "fps=8,scale=96:54:flags=area" <PLANTAO_VIDEOS>/subway/%04d.png

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
)

func init() {
	dir := os.Getenv("PLANTAO_VIDEOS")
	if dir == "" {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Error("vídeos", "err", err)
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		c, err := loadClip(filepath.Join(dir, e.Name()))
		if err != nil {
			log.Error("vídeo", "nome", e.Name(), "err", err)
			continue
		}
		wallpapers = append(wallpapers, wallpaper{"video:" + e.Name(), e.Name() + " (para devs de verdade)", func() scene { return c }})
		log.Info("vídeo carregado", "nome", e.Name(), "quadros", len(c.sizes[0].frames))
	}
}

// clip: o vídeo em alguns tamanhos, do maior ao menor; na tela entra o maior que couber.
// Cada tamanho guarda os quadros já em ANSI de 256 cores, linha por linha (cada "▀" mostra 2 pixels).
type clip struct{ sizes []clipSize }

type clipSize struct {
	frames [][]string
	w, h   int // em células
}

var clipScales = []float64{1, 0.75, 0.5, 0.35}

func loadClip(dir string) (*clip, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.png"))
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("nenhum PNG em %s", dir)
	}
	c := &clip{sizes: make([]clipSize, len(clipScales))}
	for _, f := range files { // o Glob devolve em ordem: 0001.png, 0002.png...
		r, err := os.Open(f)
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(r)
		r.Close()
		if err != nil {
			return nil, err
		}
		for i, k := range clipScales {
			w, h := int(float64(img.Bounds().Dx())*k), int(float64(img.Bounds().Dy())*k)/2*2
			rows := halfBlockRows(shrink(img, w, h))
			c.sizes[i] = clipSize{append(c.sizes[i].frames, rows), w, h / 2}
		}
	}
	return c, nil
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

// halfBlockRows: pixel de cima na cor da frente, o de baixo no fundo; só repete o código de cor quando muda.
func halfBlockRows(img image.Image) []string {
	b := img.Bounds()
	var rows []string
	for y := b.Min.Y; y+1 < b.Max.Y; y += 2 {
		var s strings.Builder
		fg, bg := -1, -1
		for x := b.Min.X; x < b.Max.X; x++ {
			f, k := xterm256(img.At(x, y)), xterm256(img.At(x, y+1))
			switch {
			case f != fg && k != bg:
				fmt.Fprintf(&s, "\x1b[38;5;%d;48;5;%dm", f, k)
			case f != fg:
				fmt.Fprintf(&s, "\x1b[38;5;%dm", f)
			case k != bg:
				fmt.Fprintf(&s, "\x1b[48;5;%dm", k)
			}
			fg, bg = f, k
			s.WriteString("▀")
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

// size: o maior tamanho que cabe em 60% da largura (o resto é do texto) e na altura da tela.
func (s *clip) size(w, h int) clipSize {
	for _, z := range s.sizes {
		if z.w <= w*6/10 && z.h <= h-2 {
			return z
		}
	}
	return s.sizes[len(s.sizes)-1]
}

// x0: o vídeo fica no quarto esquerdo da tela, encostando na margem se não couber.
func x0(z clipSize, w int) int { return max(2, w/4-z.w/2) }

// textCenter: o meio do espaço à direita do vídeo.
func (s *clip) textCenter(w, h int) int {
	z := s.size(w, h)
	return (x0(z, w) + z.w + w) / 2
}

func (s *clip) draw(c *canvas, t, dt float64) {
	z := s.size(c.w, c.h)
	f := z.frames[int(t*gardenFPS)%len(z.frames)]
	x, y0 := x0(z, c.w), (c.h-len(f))/2
	if c.raw == nil {
		c.raw = map[int]rawSeg{}
	}
	for i, row := range f {
		if y := y0 + i; y >= 0 && y < c.h && x+z.w <= c.w {
			c.raw[y] = rawSeg{x, z.w, row}
		}
	}
}
