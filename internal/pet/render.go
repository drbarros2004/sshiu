package pet

// O pet: uma bolinha com um broto na cabeça. Floresce quando o usuário foca, murcha quando some.
// Dois quadros por estado; o quadro 2 é a "inspiração" da respiração.
// A arte ASCII abaixo é o plano B para terminais com poucas cores.

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"sync"

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

// Colors: as cores que dá para escolher para o pet. O matiz é o escolhido e o humor continua
// aparecendo no quanto a cor é viva — murcho fica apagado, radiante fica vivo, os dois no seu
// tom. Quem nunca escolheu fica com a primeira da lista.
var Colors = []struct{ Name, Hex string }{
	{"azul", "#8EC9F0"},
	{"verde", "#8EE08A"},
	{"amarelo", "#FFD166"},
	{"coral", "#F0908E"},
	{"roxo", "#C7A0E8"},
	{"turquesa", "#7FD8D0"},
	{"laranja", "#F2B07A"},
	{"rosa", "#E890C0"},
}

// ColorIndex: onde está a cor hex em Colors (a primeira, para quem nunca escolheu ou para uma
// cor que saiu da lista).
func ColorIndex(hex string) int {
	for i, c := range Colors {
		if c.Hex == hex {
			return i
		}
	}
	return 0
}

// body: a cor do corpo do pet em cada humor, base e sombra. São as únicas que a cor escolhida
// troca: contorno, broto, flor e olhos ficam como estão.
var body = [...][2]uint32{
	Wilted:   {0x9A9A9A, 0x5C5C5C},
	Tired:    {0xD9B98C, 0x8A6A3E},
	Okay:     {0x8EC9F0, 0x3B78A8},
	Good:     {0x8EE08A, 0x3D8B47},
	Radiant:  {0xFFD166, 0xB07818},
	Sleeping: {0x8EC9F0, 0x3B78A8}, // o sono foi desenhado com as cores do normal
}

// Render desenha o pet: pixel art se o terminal tem 256 cores ou mais, senão ASCII.
func Render(m Mood, tick int, pixel bool, hex string) string {
	f := tick / moodPeriod[m] % moodFrames[m]
	if pixel {
		return skinOf(hex).small[m][f]
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

// skin: o pet inteiro numa cor. small vai nos cards, big no perfil.
type skin struct{ small, big [6][4]string }

var (
	sheet  image.Image
	skins  = map[string]*skin{}
	skinMu sync.Mutex
)

func init() {
	img, err := png.Decode(bytes.NewReader(petPNG))
	if err != nil {
		panic(err)
	}
	sheet = img
}

// Tint: o que trocar para um desenho sair no tom escolhido — a cor base e a sombra dela viram
// o mesmo matiz, com o brilho onde estava. Devolve nil para "" (cor padrão), e aí nada muda.
// O pet usa isso por humor; a obra usa nas cores marcadoras do corpo da peça.
func Tint(base, shade uint32, hex string) map[uint32]color.Color {
	if hex == "" {
		hex = Colors[0].Hex
	}
	var c uint32
	if _, err := fmt.Sscanf(hex, "#%06x", &c); err != nil {
		return nil
	}
	h := hueOf(c)
	return map[uint32]color.Color{base: rehue(base, h), shade: rehue(shade, h)}
}

// skinOf: o pet na cor escolhida, montado uma vez por cor e guardado. São poucas cores e o
// servidor é um só, então um mapa com cadeado dá conta.
func skinOf(hex string) *skin {
	if hex == "" {
		hex = Colors[0].Hex // quem nunca escolheu fica com a primeira cor da lista
	}
	skinMu.Lock()
	defer skinMu.Unlock()
	if s, ok := skins[hex]; ok {
		return s
	}
	// hex que não dá para ler devolve swap nil, e aí o pet sai com a cor que está no PNG
	s := build(func(m Mood) map[uint32]color.Color {
		return Tint(body[m][0], body[m][1], hex)
	})
	skins[hex] = &s
	return &s
}

// build monta os dois tamanhos do pet; swap, se vier, diz o que fazer com a cor do corpo.
func build(swap func(Mood) map[uint32]color.Color) skin {
	var s skin
	for m := range s.small {
		var sw map[uint32]color.Color
		if swap != nil {
			sw = swap(Mood(m))
		}
		for f := range s.small[m] {
			s.small[m][f] = halfBlocks(sheet, f*16, m*16, 1, sw)
			s.big[m][f] = halfBlocks(sheet, f*16, m*16, 2, sw)
		}
	}
	return s
}

// hueOf: o matiz de uma cor, de 0 a 1. Um cinza não tem matiz e devolve 0.
func hueOf(c uint32) float64 {
	r, g, b := float64(c>>16&255)/255, float64(c>>8&255)/255, float64(c&255)/255
	mx := math.Max(r, math.Max(g, b))
	d := mx - math.Min(r, math.Min(g, b))
	if d == 0 {
		return 0
	}
	var h float64
	switch mx {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return math.Mod(h/6+1, 1)
}

// rehue troca o matiz de uma cor e mantém o resto: a mesma distância entre o claro e o escuro,
// e o mesmo brilho. É por isso que o humor continua legível depois de trocar a cor do pet.
func rehue(c uint32, h float64) color.Color {
	r, g, b := float64(c>>16&255)/255, float64(c>>8&255)/255, float64(c&255)/255
	mn := math.Min(r, math.Min(g, b))
	d := math.Max(r, math.Max(g, b)) - mn
	x := d * (1 - math.Abs(math.Mod(h*6, 2)-1))
	var rr, gg, bb float64
	switch int(h * 6) {
	case 0:
		rr, gg, bb = d, x, 0
	case 1:
		rr, gg, bb = x, d, 0
	case 2:
		rr, gg, bb = 0, d, x
	case 3:
		rr, gg, bb = 0, x, d
	case 4:
		rr, gg, bb = x, 0, d
	default:
		rr, gg, bb = d, 0, x
	}
	u := func(v float64) uint8 { return uint8(math.Round((v + mn) * 255)) }
	return color.RGBA{u(rr), u(gg), u(bb), 255}
}

func RenderBig(m Mood, tick int, pixel bool, hex string) string {
	if pixel {
		return skinOf(hex).big[m][tick/moodPeriod[m]%moodFrames[m]]
	}
	return Render(m, tick, false, hex)
}

// swapped: a cor que vai no pixel, trocada se ela for uma das cores do corpo do pet.
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

// halfBlocks: o pet, que é sempre 16x16 na folha.
func halfBlocks(img image.Image, ox, oy, scale int, swap map[uint32]color.Color) string {
	return Pixels(img, ox, oy, 16, 16, scale, swap)
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
