package tui

// A obra da sala: cada membro tem uma peça que cresce com os minutos que ele focou hoje, e as
// peças juntas formam um desenho só (ADR 0008). Duas coisas valem a pena saber antes de mexer:
//
//   - O chão não é da peça. O bloco desenha ele de ponta a ponta, e é ele que emenda um vagão no
//     outro e uma casa na vizinha. Peça com o próprio chão deixava buraco entre elas.
//   - O arranjo é do tema. O trem é uma fileira que quebra quando não cabe; a cidade tem uma
//     tabela escrita na mão para cada quantidade de gente.

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

const (
	pieceW, pieceH = 12, 12 // pixels de cada peça: 12 colunas por 6 linhas no terminal
	pieceGap       = 2      // colunas entre uma peça e a outra
	obraGap        = 4      // colunas entre o log e a obra, quando ela fica ao lado dele
)

// pieceBody: as duas cores marcadoras do corpo da peça na folha, base e sombra. São elas que
// viram o tom de quem construiu; telhado, chão e janela acesa são iguais para todo mundo, senão
// a rua deixa de parecer uma rua só.
var pieceBody = [2]uint32{0xC09A6E, 0x8A6A45}

// obra: o que um tema tem além dos quatro desenhos.
type obra struct {
	sheet image.Image
	chao  [2]string               // as duas cores do chão: a de cima e a de baixo do meio-bloco
	wide  bool                    // ocupa a largura da tela (o trem); senão cabe no canto, ao lado do log
	rows  func(n, cols int) []int // quantas peças em cada fileira
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
	obras["trem"] = &obra{sheet: read(tremPNG), chao: [2]string{"#8A8A94", "#44444E"},
		wide: true, rows: fitRows}
	obras["cidade"] = &obra{sheet: read(cidadePNG), chao: [2]string{"#6E6E78", "#3A3A44"},
		rows: cityRows}
}

// fitRows: a fileira que quebra quando não cabe na largura, como os cards já fazem.
func fitRows(n, cols int) []int {
	per := max(1, (cols+pieceGap)/(pieceW+pieceGap))
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

// obraOf: o tema que a sala está construindo, ou nil quando não há obra para desenhar — sala solo,
// obra desligada, ou um tema que ficou guardado no banco e não tem mais desenho.
func (m model) obraOf(s rooms.Snapshot) *obra {
	if s.Obra == "" || s.Max == rooms.MinMembers {
		return nil
	}
	return obras[s.Obra]
}

// obraBlock: a obra inteira em cols colunas — o cabeçalho com a meta e as fileiras de peças, cada
// uma com o seu chão e os apelidos embaixo.
func (m model) obraBlock(s rooms.Snapshot, cols int) string {
	o := m.obraOf(s)
	if o == nil {
		return ""
	}
	ms := orderedMembers(m.id, s)
	col := pieceW
	for _, x := range ms {
		col = max(col, lipgloss.Width(m.obraName(x)))
	}
	col += pieceGap

	var out []string
	i := 0
	for _, n := range o.rows(len(ms), cols) {
		group := ms[i : i+n]
		var arts, names []string
		for _, x := range group {
			arts = append(arts, m.piece(o, s, x, col))
			names = append(names, center(m.obraName(x), col))
		}
		w := col*n - pieceGap // o chão para antes do vão que separa a última peça da margem
		out = append(out,
			lipgloss.JoinHorizontal(lipgloss.Top, arts...),
			lipgloss.NewStyle().Foreground(lipgloss.Color(o.chao[0])).Background(lipgloss.Color(o.chao[1])).
				Render(strings.Repeat("▀", max(w, 1))),
			strings.Join(names, ""))
		i += n
	}
	return m.obraHead(s) + "\n" + strings.Join(out, "\n")
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
func (m model) piece(o *obra, s rooms.Snapshot, x rooms.Member, col int) string {
	art := pet.Pixels(o.sheet, rooms.Stage(x.TodayMin, s.Meta)*pieceW, 0, pieceW, pieceH, 1,
		pet.Tint(pieceBody[0], pieceBody[1], x.Color))
	return center(art, col)
}

// center: o desenho ou o apelido no meio de uma coluna de col células.
func center(s string, col int) string {
	return lipgloss.NewStyle().Width(col).Align(lipgloss.Center).Render(s)
}
