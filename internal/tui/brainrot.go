package tui

// PROTÓTIPO: o wallpaper de piada "para desenvolvedores de verdade" (IDEIAS.md). Dois vídeos, um em cada
// borda, e o relógio e a tarefa no meio, como nos cortes com a tela dividida, só que em dobro.
// ponytail: os quadros vêm de SSHIU_VIDEOS (uma pasta por vídeo, PNGs feitos com o ffmpeg), fora do
// git, até a licença estar decidida; embutir com go:embed se o wallpaper ficar. Vertical, como no
// celular: o centro do vídeo 16:9 cortado em 9:16, 54 colunas × 48 linhas no tamanho cheio.
//
//	ffmpeg -i subway.mp4 -vf "fps=8,crop=ih*9/16:ih,scale=54:96:flags=area" <SSHIU_VIDEOS>/subway/%04d.png

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
	dir := os.Getenv("SSHIU_VIDEOS")
	if dir == "" {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Error("vídeos", "err", err)
		return
	}
	var clips []*clip
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		c, err := loadClip(filepath.Join(dir, e.Name()))
		if err != nil {
			log.Error("vídeo", "nome", e.Name(), "err", err)
			continue
		}
		clips = append(clips, c)
		log.Info("vídeo carregado", "nome", e.Name(), "quadros", len(c.sizes[0].frames))
	}
	if len(clips) < 2 {
		log.Error("brainrot precisa de dois vídeos", "achados", len(clips))
		return
	}
	b := brainrot{clips[0], clips[1]} // os dois primeiros em ordem alfabética: minecraft à esquerda, subway à direita
	wallpapers = append(wallpapers, wallpaper{"brainrot", "brainrot (para devs de verdade)", func() scene { return b }})
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

// size: o maior tamanho que cabe em maxW × maxH células.
func (s *clip) size(maxW, maxH int) clipSize {
	for _, z := range s.sizes {
		if z.w <= maxW && z.h <= maxH {
			return z
		}
	}
	return s.sizes[len(s.sizes)-1]
}

// put: o quadro do instante t, com a borda esquerda em x e centrado na altura.
func (z clipSize) put(c *canvas, t float64, x int) {
	f := z.frames[int(t*gardenFPS)%len(z.frames)]
	y0 := (c.h - len(f)) / 2
	for i, row := range f {
		if y := y0 + i; y >= 0 && y < c.h && x >= 0 && x+z.w <= c.w {
			c.raw[y] = append(c.raw[y], rawSeg{x, z.w, row})
		}
	}
}

// brainrot: um vídeo em cada borda e o texto no meio. Estímulo demais, de propósito: é a piada.
type brainrot struct{ l, r *clip }

func (d brainrot) draw(c *canvas, t, dt float64) {
	maxW := (c.w-48)/2 - 2 // 48 colunas no meio para o relógio, a tarefa e quem mais está focando
	l, r := d.l.size(maxW, c.h-2), d.r.size(maxW, c.h-2)
	l.put(c, t, 2)
	r.put(c, t, c.w-2-r.w)
}
