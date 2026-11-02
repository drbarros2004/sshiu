package tui

// A obra da sala: cada membro tem uma peça que cresce com os minutos que ele focou hoje, e as
// peças juntas formam um desenho só (ADR 0008). Duas coisas valem a pena saber antes de mexer:
//
//   - O chão não é da peça. O bloco desenha ele de ponta a ponta, e é ele que emenda um vagão no
//     outro e uma casa na vizinha. Peça com o próprio chão deixava buraco entre elas.
//   - O arranjo é do tema. O trem é uma fileira que quebra quando não cabe; a cidade tem uma
//     tabela escrita na mão para cada quantidade de gente. Em tela apertada a tabela dá lugar à
//     fileira: perder o formato do conjunto é melhor que perder a obra inteira.

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"plantao/internal/pet"
	"plantao/internal/rooms"
	"strings"

	"charm.land/lipgloss/v2"
)

//go:embed obra-trem.png
var tremPNG []byte

//go:embed obra-cidade.png
var cidadePNG []byte

//go:embed obra-horta.png
var hortaPNG []byte

//go:embed obra-estante.png
var estantePNG []byte

//go:embed obra-feira.png
var feiraPNG []byte

//go:embed obra-acampamento.png
var acampamentoPNG []byte

//go:embed obra-orquestra.png
var orquestraPNG []byte

//go:embed obra-passarinhos.png
var passarinhosPNG []byte

//go:embed obra-porto.png
var portoPNG []byte

//go:embed obra-foguetes.png
var foguetesPNG []byte

const (
	pieceW, pieceH = 12, 12 // pixels de cada peça: 12 colunas por 6 linhas no terminal
	pieceGap       = 2      // colunas entre uma peça e a outra
	obraGap        = 4      // colunas entre o log e a obra, quando ela fica ao lado dele
)

// pieceBody: as duas cores marcadoras do corpo da peça na folha, base e sombra. São elas que
// viram o tom de quem construiu; telhado, chão e janela acesa são iguais para todo mundo, senão
// a rua deixa de parecer uma rua só. São saturadas de propósito: num marcador apagado, a sombra
// tingida perde o matiz no terminal de 256 cores e todo mundo vira marrom.
var pieceBody = [2]uint32{0xE0A050, 0xA86A28}

// obra: o que um tema tem além dos quatro desenhos.
type obra struct {
	sheet  image.Image
	pw     int                    // largura da peça em pixels (0 = pieceW)
	chao   [2]string              // as duas cores do chão: a de cima e a de baixo do meio-bloco
	engate string                 // a cor do traço que liga uma peça na outra ("" = não liga)
	linkY  int                    // em que linha do desenho o engate cai
	fixo   int                    // largura em pixels da peça que não é de ninguém, no fim da folha (0 = sem)
	centro bool                   // a peça fixa fica no meio da fileira do meio; senão, na frente
	inst   bool                   // do terceiro estágio em diante, um par de quadros por cor do pet
	wide   bool                   // ocupa a largura da tela (o trem); senão cabe no canto, centrada
	rows   func(n, per int) []int // quantas peças em cada fileira, cabendo per peças numa
}

// w: a largura da peça do tema, em pixels (e em colunas, que é um pixel por coluna).
func (o *obra) w() int {
	if o.pw > 0 {
		return o.pw
	}
	return pieceW
}

var obras = map[string]*obra{}

func init() {
	read := func(b []byte) image.Image {
		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			panic(err)
		}
		return img
	}
	// o engate cai na linha do estrado, que é igual nos vagões e na locomotiva. O vagão é mais
	// comprido que as outras peças, como um carro de passageiros, e a maria-fumaça mais ainda,
	// porque é o que cabe as rodas dela
	obras["trem"] = &obra{sheet: read(tremPNG), pw: 16, chao: [2]string{"#8A8A94", "#44444E"},
		engate: "#8A8A94", linkY: 4, fixo: 24, wide: true, rows: fitRows}
	obras["cidade"] = &obra{sheet: read(cidadePNG), chao: [2]string{"#6E6E78", "#3A3A44"},
		rows: cityRows}
	obras["horta"] = &obra{sheet: read(hortaPNG), chao: [2]string{"#6B4A2E", "#3E2A1A"},
		wide: true, rows: fitRows}
	obras["estante"] = &obra{sheet: read(estantePNG), chao: [2]string{"#4A4250", "#2E2A36"},
		wide: true, rows: fitRows}
	obras["feira"] = &obra{sheet: read(feiraPNG), chao: [2]string{"#9A8A78", "#5E5246"},
		wide: true, rows: fitRows}
	obras["acampamento"] = &obra{sheet: read(acampamentoPNG), chao: [2]string{"#3D8B47", "#24602C"},
		fixo: pieceW, centro: true, rows: campRows}
	// o instrumento sai da cor do pet, na ordem de pet.Colors: quem é azul toca violino
	obras["orquestra"] = &obra{sheet: read(orquestraPNG), chao: [2]string{"#8A6A3E", "#5A4529"},
		inst: true, wide: true, rows: fitRows}
	// o fio é o engate: ele sai do poste, que não é de ninguém, e passa por todo mundo
	obras["passarinhos"] = &obra{sheet: read(passarinhosPNG), chao: [2]string{"#6E6E78", "#3A3A44"},
		engate: "#8A8A94", linkY: 4, fixo: pieceW, wide: true, rows: fitRows}
	obras["porto"] = &obra{sheet: read(portoPNG), chao: [2]string{"#3E7CB1", "#2A5580"},
		wide: true, rows: fitRows}
	obras["foguetes"] = &obra{sheet: read(foguetesPNG), chao: [2]string{"#8A8A94", "#44444E"},
		wide: true, rows: fitRows}
}

// fitRows: a fileira que quebra quando não cabe na largura, como os cards já fazem.
func fitRows(n, per int) []int {
	per = max(1, per)
	var out []int
	for i := 0; i < n; i += per {
		out = append(out, min(per, n-i))
	}
	return out
}

// cityRows: o arranjo da cidade para cada quantidade de gente. É uma tabela de propósito: com
// fórmula, sete pessoas viravam uma grade torta em vez da praça que 2-3-2 desenha (ADR 0008).
func cityRows(n, _ int) []int {
	table := [][]int{{}, {1}, {2}, {2, 1}, {2, 2}, {3, 2}, {3, 3}, {2, 3, 2}, {4, 4}, {3, 3, 3}, {5, 5}}
	return table[min(max(n, 0), rooms.MaxMembers)]
}

// campRows: as barracas em volta da fogueira, que entra no meio da fileira do meio. Tabela pelo
// mesmo motivo da cidade: com quatro, 1-2-1 faz a roda que uma grade não faz.
func campRows(n, _ int) []int {
	table := [][]int{{}, {1}, {2}, {1, 2}, {1, 2, 1}, {2, 2, 1}, {2, 2, 2}, {3, 2, 2}, {3, 2, 3}, {2, 4, 3}, {3, 4, 3}}
	return table[min(max(n, 0), rooms.MaxMembers)]
}

// obraOf: o tema que a sala está construindo, ou nil quando não há obra para desenhar — sala solo,
// obra desligada, ou um tema que ficou guardado no banco e não tem mais desenho.
func (m model) obraOf(s rooms.Snapshot) *obra {
	if s.Obra == "" || s.Max == rooms.MinMembers {
		return nil
	}
	return obras[s.Obra]
}

// obraBlock: a obra inteira em cols colunas — o cabeçalho com a meta e as fileiras de peças, cada
// uma com o seu chão e os apelidos embaixo. Com flat, o arranjo do tema dá lugar a uma fileira só,
// que é mais baixa: é o que a tela usa quando a tabela da cidade não cabe na altura.
func (m model) obraBlock(s rooms.Snapshot, cols int, flat bool) string {
	o := m.obraOf(s)
	if o == nil {
		return ""
	}
	rows := o.rows
	if flat {
		rows = fitRows
	}
	ms := orderedMembers(m.id, s)
	col := o.w()
	for _, x := range ms {
		col = max(col, lipgloss.Width(m.obraName(x)))
	}
	col += pieceGap
	vao := col - o.w()       // o que sobra da coluna é o vão, metade de cada lado da peça
	avail := cols - pieceGap // margem para o bloco não encostar na borda da tela
	if o.fixo > 0 {
		avail -= o.fixo + vao // a peça fixa também ocupa espaço na fileira dela
	}

	var out [][3]string // cada fileira: o desenho, o chão e os apelidos
	i := 0
	rs := rows(len(ms), avail/col)
	fixo := -1 // a fileira da peça fixa, e onde ela entra: na frente ou no meio
	at := func(int) int { return 0 }
	switch {
	case o.fixo > 0 && o.centro:
		fixo, at = len(rs)/2, func(n int) int { return n / 2 }
	case o.fixo > 0:
		fixo = 0
	}
	for r, n := range rs {
		// cada célula é uma peça com o seu rótulo; a fixa não é de ninguém e não leva nome
		arts, names := []string{gapBlock(vao/2, nil)}, []string{strings.Repeat(" ", vao/2)}
		floor := 0 // a largura do chão: cada célula é a peça mais o vão dela
		add := func(art, nome string) {
			if floor > 0 {
				arts = append(arts, gapBlock(vao, o))
			}
			w := lipgloss.Width(art) + vao
			arts, names = append(arts, art), append(names, center(nome, w))
			floor += w
		}
		for k, x := range ms[i : i+n] {
			if r == fixo && k == at(n) {
				add(m.fixed(o), "")
			}
			add(m.piece(o, s, x), m.obraName(x))
		}
		arts = append(arts, gapBlock(vao-vao/2, nil))
		out = append(out, [3]string{
			lipgloss.JoinHorizontal(lipgloss.Top, arts...),
			lipgloss.NewStyle().Foreground(lipgloss.Color(o.chao[0])).Background(lipgloss.Color(o.chao[1])).
				Render(strings.Repeat("▀", max(floor, 1))),
			strings.Join(names, "")})
		i += n
	}
	// o arranjo de tabela é um lugar, não uma lista: as fileiras ficam centradas umas nas outras,
	// e é isso que põe a fogueira no meio de todo mundo e faz o 2-3-2 da cidade virar uma praça
	widest := 0
	for _, f := range out {
		widest = max(widest, lipgloss.Width(f[0]))
	}
	var lines []string
	for _, f := range out {
		pad := lipgloss.NewStyle()
		if !o.wide && !flat {
			pad = pad.PaddingLeft((widest - lipgloss.Width(f[0])) / 2)
		}
		for _, l := range f {
			lines = append(lines, pad.Render(l))
		}
	}
	return m.obraHead(s) + "\n\n" + strings.Join(lines, "\n")
}

// gapBlock: o vão entre duas peças, da altura do desenho. Com o tema que engata — o trem — ele
// leva um traço na altura do estrado, que é o que liga um vagão no outro sem colar os dois.
func gapBlock(w int, o *obra) string {
	if w <= 0 {
		return ""
	}
	lines := make([]string, pieceH/2)
	for i := range lines {
		lines[i] = strings.Repeat(" ", w)
	}
	if o != nil && o.engate != "" && o.linkY >= 0 && o.linkY < len(lines) {
		lines[o.linkY] = lipgloss.NewStyle().Foreground(lipgloss.Color(o.engate)).
			Render(strings.Repeat("▄", w))
	}
	return strings.Join(lines, "\n")
}

// obraView: a obra sozinha na tela, na tecla o. A tela da sala encolhe até a obra sair, e num
// terminal pequeno ela sai sempre; aqui ela não divide espaço com nada, então existe em qualquer
// tamanho. É o mesmo caminho do modo jardim, que também é uma tela inteira atrás de uma tecla.
func (m model) obraView(s rooms.Snapshot, w, h int) string {
	head := bold.Render(s.Name) + m.st.muted.Render(" · a obra de hoje") + "\n\n"
	ob := m.obraBlock(s, w, false)
	if lipgloss.Height(head+ob) > h { // a tabela do tema não cabe: vira uma fileira só
		ob = m.obraBlock(s, w, true)
	}
	return head + ob
}

// obraHead: o tema e o que cada peça custa. Sem o passo, o desenho não diz o que falta para
// crescer; com ele, quem está a dez minutos da próxima mudança sabe disso.
func (m model) obraHead(s rooms.Snapshot) string {
	return m.st.muted.Render(fmt.Sprintf("%s · meta %s · a peça muda a cada %s",
		s.Obra, dur(float64(s.Meta)), dur(float64(s.Meta/(rooms.ObraStages-1)))))
}

// obraName: o apelido embaixo da peça, que é o que deixa a sala ver quem focou e quem não focou.
func (m model) obraName(x rooms.Member) string {
	if x.ID == m.id {
		return "você"
	}
	return x.Nick
}

// piece: a peça de alguém, no estágio dos minutos que a pessoa focou hoje e no tom dela.
func (m model) piece(o *obra, s rooms.Snapshot, x rooms.Member) string {
	return m.frame(o, stageFrame(o, rooms.Stage(x.TodayMin, s.Meta), x.Color), x.Color)
}

// stageFrame: o quadro da folha para um estágio. Na orquestra, cadeira e estante são de todo
// mundo, e do instrumento em diante cada cor tem o seu par (parado, tocando).
func stageFrame(o *obra, stage int, hex string) int {
	if !o.inst || stage < 2 {
		return stage
	}
	return 2 + 2*pet.ColorIndex(hex) + stage - 2
}

// frame: um quadro da folha do tema, com o corpo no tom de quem construiu.
func (m model) frame(o *obra, i int, hex string) string {
	return pet.Pixels(o.sheet, i*o.w(), 0, o.w(), pieceH, 1,
		pet.Tint(pieceBody[0], pieceBody[1], hex))
}

// fixed: a peça que não é de ninguém (a locomotiva, a fogueira), no fim da folha. Ela tem cores
// próprias e não é tingida.
func (m model) fixed(o *obra) string {
	return pet.Pixels(o.sheet, o.sheet.Bounds().Dx()-o.fixo, 0, o.fixo, pieceH, 1, nil)
}

// center: o desenho ou o apelido no meio de uma coluna de col células.
func center(s string, col int) string {
	return lipgloss.NewStyle().Width(col).Align(lipgloss.Center).Render(s)
}
