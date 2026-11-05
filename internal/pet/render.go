package pet

// O pet: uma bolinha com um broto na cabeça. Floresce quando o usuário foca, murcha quando some.
// Os humores usam a mesma paleta, com dois quadros de respiração; dormindo tem quatro.
// A arte ASCII abaixo é o plano B para terminais com poucas cores.

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"charm.land/lipgloss/v2"
)

var art = [...][2]string{
	Radiant: {`
 *  \@/  *
     |
   .---.
  ( ^ ^ )
  ( \_/ )
   '---'`, `
 +  \@/  +
     |
  .-----.
 ( ^   ^ )
 (  \_/  )
  '-----'`},
	Good: {`
    \|/
     |
   .---.
  ( o o )
  (  u  )
   '---'`, `
    \|/
     |
  .-----.
 ( o   o )
 (   u   )
  '-----'`},
	Okay: {`
     ,
     |
   .---.
  ( . . )
  (  -  )
   '---'`, `
     ,
     |
  .-----.
 ( .   . )
 (   -   )
  '-----'`},
	Tired: {`
       z
   _,
   .-\-.
  ( - - )
  (  ~  )
   '---'`, `
        Z
   _,  z
  .--\--.
 ( -   - )
 (   ~   )
  '-----'`},
	Wilted: {`



  _,.---.
 ( T _ T )
  '-----'`, `



  _,.---.
 ( T _ T )
 .'-----'`},
	Sleeping: {`
     ,   z
     |
   .---.
  ( _ _ )
  (  o  )
   '---'`, `
     ,  z
     |
  .-----.
 ( _   _ )
 (   o   )
  '-----'`},
}

var (
	MoodName   = [...]string{Wilted: "murcho", Tired: "cansado", Okay: "normal", Good: "bem", Radiant: "radiante", Sleeping: "dormindo"}
	MoodColor  = [...]string{Wilted: "244", Tired: "180", Okay: "117", Good: "78", Radiant: "220", Sleeping: "67"}
	moodPeriod = [...]int{Wilted: 4, Tired: 3, Okay: 2, Good: 2, Radiant: 1, Sleeping: 3} // ticks por quadro: quanto mais cansado, mais lenta a respiração
	// quadros de cada humor na folha: respirar são dois, e o sono usa quatro para o z subir
	moodFrames = [...]int{Wilted: 2, Tired: 2, Okay: 2, Good: 2, Radiant: 2, Sleeping: 4}
)

// Render usa a mesma paleta em todos os humores. Só o desenho e a animação mudam.
func Render(m Mood, tick int, pixel bool, paletteID string) string {
	frame := tick / moodPeriod[m] % moodFrames[m]
	palette := ColorIndex(paletteID)
	if pixel {
		return renderedPalettes[palette].small[m][frame]
	}
	return lipgloss.NewStyle().Foreground(Colors[palette].Body).Width(11).
		Render(strings.TrimPrefix(art[m][frame%2], "\n"))
}

// pet.png é o template de marcadores: 64×96, quadros 16×16, uma linha por estado.
// As cores visíveis ficam em palettes.json, não nesta imagem.
//
//go:embed pet.png
var petPNG []byte

type spriteFrames struct{ small, big [6][4]string }

var renderedPalettes []spriteFrames // pronto ao iniciar, somente leitura durante as sessões

func init() {
	template, err := png.Decode(bytes.NewReader(petPNG))
	if err != nil {
		panic(fmt.Errorf("pet.png: %w", err))
	}
	if template.Bounds().Dx() != 64 || template.Bounds().Dy() != 96 {
		panic("pet.png: o template deve medir 64×96, com quadros 16×16")
	}
	Colors, err = readPalettes(palettesJSON, template)
	if err != nil {
		panic(err)
	}
	for _, palette := range Colors {
		renderedPalettes = append(renderedPalettes, renderPalette(template, palette))
	}
}

// renderPalette aplica as cores nomeadas e prepara os dois tamanhos do pet.
func renderPalette(template image.Image, palette Palette) spriteFrames {
	var frames spriteFrames
	for mood := range frames.small {
		for frame := 0; frame < moodFrames[mood]; frame++ {
			frames.small[mood][frame] = Pixels(template, frame*16, mood*16, 16, 16, 1, palette.parts)
			frames.big[mood][frame] = Pixels(template, frame*16, mood*16, 16, 16, 2, palette.parts)
		}
	}
	return frames
}

func RenderBig(m Mood, tick int, pixel bool, paletteID string) string {
	if pixel {
		return renderedPalettes[ColorIndex(paletteID)].big[m][tick/moodPeriod[m]%moodFrames[m]]
	}
	return Render(m, tick, false, paletteID)
}

// swapped aplica a cor final quando o pixel está no mapa de substituições.
func swapped(c color.Color, swap map[uint32]color.Color) color.Color {
	if swap == nil {
		return c
	}
	r, g, b, a := c.RGBA()
	if a == 0 {
		return c
	}
	if n, ok := swap[r>>8<<16|g>>8<<8|b>>8]; ok {
		return n
	}
	return c
}

// Pixels desenha um pedaço w×h de uma folha PNG juntando 2 pixels por caractere: "▀" com a cor de
// cima na frente e a de baixo no fundo. scale repete cada pixel scale×scale vezes, e swap, se
// vier, troca as cores que ele listar — é assim que o pet e a peça da obra saem no tom de alguém.
func Pixels(img image.Image, ox, oy, w, h, scale int, swap map[uint32]color.Color) string {
	var rows []string
	for y := 0; y < h*scale; y += 2 {
		var b strings.Builder
		for x := 0; x < w*scale; x++ {
			top := swapped(img.At(ox+x/scale, oy+y/scale), swap)
			bot := swapped(img.At(ox+x/scale, oy+(y+1)/scale), swap)
			st := lipgloss.NewStyle()
			switch {
			case opaque(top) && opaque(bot):
				b.WriteString(st.Foreground(top).Background(bot).Render("▀"))
			case opaque(top):
				b.WriteString(st.Foreground(top).Render("▀"))
			case opaque(bot):
				b.WriteString(st.Foreground(bot).Render("▄"))
			default:
				b.WriteByte(' ')
			}
		}
		rows = append(rows, b.String())
	}
	return strings.Join(rows, "\n")
}

func opaque(c color.Color) bool { _, _, _, a := c.RGBA(); return a > 0 }
