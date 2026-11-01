package pet

// O pet: uma bolinha com um broto na cabeça. Floresce quando o usuário foca, murcha quando some.
// Dois quadros por estado; o quadro 2 é a "inspiração" da respiração.
// A arte ASCII abaixo é o plano B para terminais com poucas cores.

import (
	"bytes"
	_ "embed"
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

// Render desenha o pet: pixel art se o terminal tem 256 cores ou mais, senão ASCII.
func Render(m Mood, tick int, pixel bool) string {
	f := tick / moodPeriod[m] % moodFrames[m]
	if pixel {
		return sprites[m][f]
	}
	// a arte ASCII tem dois quadros em todos os humores: o z do sono anda só na pixel art
	return lipgloss.NewStyle().Foreground(lipgloss.Color(MoodColor[m])).Width(11).Render(strings.TrimPrefix(art[m][f%2], "\n"))
}

// pet.png: folha 64x96, linha = humor (murcho..radiante, e dormindo por último, 16px cada),
// coluna = quadro. Os humores usam as duas primeiras colunas (a respiração); o sono usa as quatro,
// com o z saindo da cabeça. Quantos quadros cada linha tem está em moodFrames.
// Edite no Piskel/Aseprite/Lospec à vontade; o servidor relê no próximo build.
//
//go:embed pet.png
var petPNG []byte

var sprites, spritesBig [6][4]string // 16x8 nos cards, 32x16 no perfil

func init() {
	img, err := png.Decode(bytes.NewReader(petPNG))
	if err != nil {
		panic(err)
	}
	for m := range sprites {
		for f := range sprites[m] {
			sprites[m][f] = halfBlocks(img, f*16, m*16, 1)
			spritesBig[m][f] = halfBlocks(img, f*16, m*16, 2)
		}
	}
}

func RenderBig(m Mood, tick int, pixel bool) string {
	if pixel {
		return spritesBig[m][tick/moodPeriod[m]%moodFrames[m]]
	}
	return Render(m, tick, false)
}

// halfBlocks junta 2 pixels por caractere: "▀" com a cor de cima na frente e a de baixo no fundo.
// scale repete cada pixel scale×scale vezes.
func halfBlocks(img image.Image, ox, oy, scale int) string {
	var rows []string
	for y := 0; y < 16*scale; y += 2 {
		var b strings.Builder
		for x := 0; x < 16*scale; x++ {
			top, bot := img.At(ox+x/scale, oy+y/scale), img.At(ox+x/scale, oy+(y+1)/scale)
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
