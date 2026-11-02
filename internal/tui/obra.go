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
	sheet  image.Image
	chao   [2]string               // as duas cores do chão: a de cima e a de baixo do meio-bloco
	engate string                  // a cor do traço que liga uma peça na outra ("" = não liga)
	linkY  int                     // em que linha do desenho o engate cai
	lead   bool                    // tem uma peça fixa na frente, que não é de ninguém
	wide   bool                    // ocupa a largura da tela (o trem); senão cabe no canto
	rows   func(n, cols int) []int // quantas peças em cada fileira
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
	// o engate cai na linha do estrado, que é igual nos cinco quadros do trem
	obras["trem"] = &obra{sheet: read(tremPNG), chao: [2]string{"#8A8A94", "#44444E"},
		engate: "#8A8A94", linkY: 4, lead: true, wide: true, rows: fitRows}
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
	col := pieceW
	for _, x := range ms {
		col = max(col, lipgloss.Width(m.obraName(x)))
	}
	col += pieceGap
	vao := col - pieceW      // o que sobra da coluna é o vão, metade de cada lado da peça
	avail := cols - pieceGap // margem para o bloco não encostar na borda da tela
	if o.lead {
		avail -= col // a peça da frente também ocupa uma coluna
	}

	var out []string
	i := 0
	for r, n := range rows(len(ms), avail) {
		// cada célula é uma peça com o seu rótulo; a da frente não é de ninguém e não leva nome
		arts, names := []string{gapBlock(vao/2, nil)}, []string{strings.Repeat(" ", vao/2)}
		cells := 0
		add := func(art, nome string) {
			if cells > 0 {
				arts = append(arts, gapBlock(vao, o))
			}
			arts, names = append(arts, art), append(names, center(nome, col))
			cells++
		}
		if o.lead && r == 0 {
			add(m.frame(o, rooms.ObraStages, ""), "")
		}
		for _, x := range ms[i : i+n] {
			add(m.piece(o, s, x), m.obraName(x))
		}
		arts = append(arts, gapBlock(vao-vao/2, nil))
		out = append(out,
			lipgloss.JoinHorizontal(lipgloss.Top, arts...),
			lipgloss.NewStyle().Foreground(lipgloss.Color(o.chao[0])).Background(lipgloss.Color(o.chao[1])).
				Render(strings.Repeat("▀", max(col*cells, 1))),
			strings.Join(names, ""))
		i += n
	}
	return m.obraHead(s) + "\n\n" + strings.Join(out, "\n")
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
	return m.frame(o, rooms.Stage(x.TodayMin, s.Meta), x.Color)
}

// frame: um quadro da folha do tema. hex "" não tinge nada — é o caso da locomotiva, que é
// desenhada com cores próprias e não com o marcador do corpo.
func (m model) frame(o *obra, i int, hex string) string {
	return pet.Pixels(o.sheet, i*pieceW, 0, pieceW, pieceH, 1,
		pet.Tint(pieceBody[0], pieceBody[1], hex))
}

// center: o desenho ou o apelido no meio de uma coluna de col células.
func center(s string, col int) string {
	return lipgloss.NewStyle().Width(col).Align(lipgloss.Center).Render(s)
}
