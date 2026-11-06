//go:build ignore

// Gera as tiras do brainrot a partir de assets/videos/ (CREDITS.md): go generate ./internal/tui.
// Precisa do ffmpeg. Cada tira é um PNG com os quadros um embaixo do outro, já nas 216 cores do
// cubo do xterm, que é o que a tela mostra: assim o PNG fica 7× menor e a tela não muda nada.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	fps  = 8  // gardenFPS
	secs = 60 // os trechos têm 1 min
)

func main() {
	tmp, err := os.MkdirTemp("", "brainrot")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)
	for _, v := range []string{"minecraft", "subway"} {
		// vertical, como no celular: o centro do vídeo 16:9 cortado em 9:16, 54 × 96 pixels
		strip := filepath.Join(tmp, v+".png")
		vf := fmt.Sprintf("fps=%d,crop=ih*9/16:ih,scale=54:96:flags=area,tile=1x%d", fps, fps*secs)
		cmd := exec.Command("ffmpeg", "-v", "error", "-y", "-i", "../../assets/videos/"+v+".mp4", "-vf", vf, "-frames:v", "1", strip)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		if err := quantize(strip, "brainrot-"+v+".png"); err != nil {
			panic(err)
		}
	}
}

// quantize: cada pixel na cor do cubo 6×6×6 que o xterm256 do brainrot.go escolheria.
func quantize(in, out string) error {
	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return err
	}
	levels := []uint8{0, 95, 135, 175, 215, 255}
	var pal color.Palette
	for _, r := range levels {
		for _, g := range levels {
			for _, b := range levels {
				pal = append(pal, color.RGBA{r, g, b, 255})
			}
		}
	}
	level := func(v uint32) int { // o mesmo do xterm256
		v >>= 8
		switch {
		case v < 48:
			return 0
		case v < 115:
			return 1
		}
		return int((v - 35) / 40)
	}
	b := img.Bounds()
	p := image.NewPaletted(b, pal)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			p.SetColorIndex(x, y, uint8(36*level(r)+6*level(g)+level(bl)))
		}
	}
	w, err := os.Create(out)
	if err != nil {
		return err
	}
	defer w.Close()
	return (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(w, p)
}
