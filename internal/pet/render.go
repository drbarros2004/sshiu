package pet

import (
	"charm.land/lipgloss/v2"
	"image"
	"image/color"
	"strings"
)

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

// Pixels desenha um pedaço w×h de uma folha PNG juntando 2 pixels por caractere: "▄" invertido,
// com a cor de cima no fundo e a de baixo no meio bloco. scale repete cada pixel scale×scale
// vezes, e swap, se vier, troca as cores que ele listar — a obra usa essa troca; os personagens já têm as cores no PNG.
//
// Por que invertido (issue #56): o Terminal.app desenha "█ ▀ ▄" com a fonte, que não ocupa a
// célula toda (medido: de 4 a 26 px numa célula de 28), e sobravam frestas entre as linhas. O
// fundo cobre a célula inteira, e o "▄" encosta embaixo; com a cor de cima no fundo, a fresta
// vira 1 px na última linha, que o sublinhado (na cor do glifo) cobre: o Terminal.app o desenha
// bem nessa linha. Nos terminais que desenham os blocos por conta própria (Windows Terminal), o
// sublinhado cai dentro do "▄", da mesma cor, e não aparece.
func Pixels(img image.Image, ox, oy, w, h, scale int, swap map[uint32]color.Color) string {
	var rows []string
	for y := 0; y < h*scale; y += 2 {
		var b strings.Builder
		for x := 0; x < w*scale; x++ {
			top := swapped(img.At(ox+x/scale, oy+y/scale), swap)
			bot := swapped(img.At(ox+x/scale, oy+(y+1)/scale), swap)
			st := lipgloss.NewStyle().Underline(true)
			switch {
			case opaque(top): // invertido: a frente (top) vira o fundo, e o fundo (bot ou o do terminal) o "▄"
				if opaque(bot) {
					st = st.Background(bot)
				}
				b.WriteString(st.Reverse(true).Foreground(top).Render("▄"))
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
