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
	"image/color"
	"image/png"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
	pieceW, pieceH     = 12, 12 // pixels de cada peça: 12 colunas por 6 linhas no terminal
	pieceGap           = 2      // colunas entre uma peça e a outra
	obraGap            = 4      // colunas entre o log e a obra, quando ela fica ao lado dele
	maxArtworkRowRatio = 4      // a última fileira tem ao menos um quarto da anterior
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
	anim   int                    // ticks por quadro do movimento (0 = parado); o quadro B fica embaixo na folha
	nota   bool                   // a peça que toca canta: um ♫ onde a folha marca com notaCor
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
	// Cinco temas se mexem, como o pet respira: o trem anda (rodas, biela e fumaça), o porto
	// balança na água, a orquestra toca, o passarinho se abaixa no fio e canta, e a chama da
	// fogueira baixa e sobe. O movimento é pequeno de propósito: é um app de foco, e a obra fica
	// no canto do olho (docs/regras.md).
	obras["trem"] = &obra{sheet: read(tremPNG), pw: 16, chao: [2]string{"#8A8A94", "#44444E"},
		engate: "#8A8A94", linkY: 4, fixo: 24, anim: 1, wide: true, rows: fitRows}
	obras["cidade"] = &obra{sheet: read(cidadePNG), chao: [2]string{"#6E6E78", "#3A3A44"},
		rows: fitRows}
	obras["horta"] = &obra{sheet: read(hortaPNG), chao: [2]string{"#6B4A2E", "#3E2A1A"},
		wide: true, rows: fitRows}
	obras["estante"] = &obra{sheet: read(estantePNG), chao: [2]string{"#4A4250", "#2E2A36"},
		wide: true, rows: fitRows}
	obras["feira"] = &obra{sheet: read(feiraPNG), chao: [2]string{"#9A8A78", "#5E5246"},
		wide: true, rows: fitRows}
	// só a fogueira se mexe: barraca e cabana são iguais nas duas faixas da folha
	obras["acampamento"] = &obra{sheet: read(acampamentoPNG), chao: [2]string{"#3D8B47", "#24602C"},
		fixo: pieceW, centro: true, anim: 1, rows: fitRows}
	// o instrumento sai da cor do pet, na ordem de pet.Colors: quem é azul toca violino
	obras["orquestra"] = &obra{sheet: read(orquestraPNG), chao: [2]string{"#8A6A3E", "#5A4529"},
		inst: true, anim: 2, nota: true, wide: true, rows: fitRows}
	// o fio é o engate: ele sai do poste, que não é de ninguém, e passa por todo mundo
	obras["passarinhos"] = &obra{sheet: read(passarinhosPNG), chao: [2]string{"#6E6E78", "#3A3A44"},
		engate: "#8A8A94", linkY: 4, fixo: pieceW, anim: 2, nota: true, wide: true, rows: fitRows}
	obras["porto"] = &obra{sheet: read(portoPNG), chao: [2]string{"#3E7CB1", "#2A5580"},
		anim: 3, wide: true, rows: fitRows}
	obras["foguetes"] = &obra{sheet: read(foguetesPNG), chao: [2]string{"#8A8A94", "#44444E"},
		wide: true, rows: fitRows}
}

// fitRows: preenche cada fileira antes de abrir a próxima, conforme a largura disponível.
func fitRows(n, per int) []int {
	per = max(1, per)
	var rows []int
	for n > 0 {
		count := min(n, per)
		rows = append(rows, count)
		n -= count
	}
	if last := len(rows) - 1; last > 0 {
		for rows[last]*maxArtworkRowRatio < rows[last-1] {
			rows[last-1]--
			rows[last]++
		}
	}
	return rows
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
	per := avail / col
	sozinha := false // a peça fixa não cabe ao lado de ninguém: ganha uma fileira só dela
	if o.fixo > 0 {
		if lado := (avail - o.fixo - vao) / col; lado >= 1 {
			per = lado // a peça fixa também ocupa espaço na fileira dela
		} else {
			sozinha = true
		}
	}

	var out [][3]string // cada fileira: o desenho, o chão e os apelidos
	i := 0
	rs := rows(len(ms), per)
	if slices.Max(append(rs, 0)) > max(per, 1) { // a tabela do tema é mais larga que a tela: vira fileira
		rs = fitRows(len(ms), per)
	}
	fixo := -1 // a fileira da peça fixa, e onde ela entra: na frente ou no meio
	at := func(int) int { return 0 }
	switch {
	case o.fixo > 0 && o.centro:
		fixo, at = len(rs)/2, func(n int) int { return n / 2 }
	case o.fixo > 0:
		fixo = 0
	}
	if sozinha {
		rs = slices.Insert(rs, fixo, 0)
	}
	for r, n := range rs {
		// cada célula é uma peça com o seu rótulo; a fixa não é de ninguém e não leva nome. O desenho
		// começa com meio vão e o apelido não: ele já é centrado na peça mais o vão inteiro dela
		arts, names := []string{gapBlock(vao/2, nil)}, []string{}
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
		if r == fixo && n == 0 { // a fileira só da peça fixa
			add(m.fixed(o), "")
		}
		arts = append(arts, gapBlock(vao-vao/2, nil))
		out = append(out, [3]string{
			lipgloss.JoinHorizontal(lipgloss.Top, arts...),
			lipgloss.NewStyle().Foreground(lipgloss.Color(o.chao[0])).Background(lipgloss.Color(o.chao[1])).
				Reverse(true).Underline(true).Render(strings.Repeat("▄", max(floor, 1))), // ver pet.Pixels
			strings.Join(names, "")})
		i += n
	}
	// o arranjo de tabela é um lugar, não uma lista: as fileiras ficam centradas umas nas outras,
	// mantendo a fogueira no meio da sua fileira, mesmo quando a última leva menos peças
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
	return m.obraHead(s, cols) + "\n\n" + strings.Join(lines, "\n")
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
		lines[o.linkY] = lipgloss.NewStyle().Foreground(lipgloss.Color(o.engate)).Underline(true).
			Render(strings.Repeat("▄", w)) // sublinhado: ver pet.Pixels
	}
	return strings.Join(lines, "\n")
}

// obraView: a obra sozinha na tela, na tecla o. A tela da sala encolhe até a obra sair, e num
// terminal pequeno ela sai sempre; aqui ela não divide espaço com nada, então existe em qualquer
// tamanho. É o mesmo caminho do modo jardim, que também é uma tela inteira atrás de uma tecla.
// A tabela de foco vem em cima da obra, ou sozinha na sala sem obra; se as duas não cabem, a
// tabela fica e a obra sai, com um aviso (ADR 0010). Antes disso, a tabela fica enxuta.
func (m model) obraView(s rooms.Snapshot, w, h int) string {
	head := bold.Render(m.roomName(s))
	if m.obraOf(s) != nil {
		head += m.st.muted.Render(m.t(" · a obra de hoje"))
	}
	head += "\n\n"
	reset := m.obraOf(s) == nil // com a obra, o "reinicia em" é dela
	for _, tight := range []bool{false, true} {
		body := head
		if m.boardShown(s) {
			body += m.boardBlock(s, reset, tight, h-2) + "\n\n" // 2: o nome da sala e a linha em branco
		}
		if m.obraOf(s) == nil {
			if body = strings.TrimRight(body, "\n"); lipgloss.Height(body) <= h || tight {
				return body
			}
			continue
		}
		for _, flat := range []bool{false, true} { // a tabela do tema não cabe: vira uma fileira só
			if out := body + m.obraBlock(s, w, flat); lipgloss.Height(out) <= h || flat && !m.boardShown(s) {
				return out
			}
		}
	}
	return head + m.boardBlock(s, reset, true, h-4) + "\n\n" + // 4: o nome, o aviso e as linhas em branco
		m.st.muted.Render(m.t("a obra não coube junto da tabela: aumente a janela para ver as duas"))
}

// obraHead: o tema e o que cada peça custa. Sem o passo, o desenho não diz o que falta para
// crescer; com ele, quem está a dez minutos da próxima mudança sabe disso.
func (m model) obraHead(s rooms.Snapshot, cols int) string {
	return m.st.muted.Width(max(cols, 1)).Render(m.tf("%s · meta %s · a peça muda a cada %s · reinicia em %s",
		obraName(m.lang, s.Obra), dur(float64(s.Meta)), dur(float64(s.Meta/(rooms.ObraStages-1))), obraResetIn(m.now)))
}

// obraResetIn segue o dia local do servidor, como minutesToday. Construir a próxima data
// (em vez de somar 24 h) também respeita dias com mudança de horário de verão.
func obraResetIn(now time.Time) string {
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	// Arredondar para cima evita mostrar zero antes da virada do dia.
	seconds := int((next.Sub(now) + time.Second - 1) / time.Second)
	return fmt.Sprintf("%02d:%02d:%02d", seconds/3600, seconds/60%60, seconds%60)
}

// obraName: o apelido embaixo da peça, que é o que deixa a sala ver quem focou e quem não focou.
func (m model) obraName(x rooms.Member) string {
	if x.ID == m.id {
		return m.t("você")
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
	return drawn(o, i*o.w(), m.animY(o), o.w(), hex, true)
}

// fixed: a peça que não é de ninguém (a locomotiva, a fogueira), no fim da folha. Ela tem cores
// próprias e não é tingida.
func (m model) fixed(o *obra) string {
	return drawn(o, o.sheet.Bounds().Dx()-o.fixo, m.animY(o), o.fixo, "", false)
}

// animY: a faixa da folha do quadro de agora — a de cima, ou a do quadro B nos temas que se mexem.
func (m model) animY(o *obra) int {
	if o.anim > 0 && m.tick/o.anim%2 == 1 {
		return pieceH
	}
	return 0
}

// quadros: cada quadro já desenhado, por tema, posição e cor. O desenho não muda, e cada sessão
// redesenhava a obra inteira duas vezes a cada tick: numa sala de 10 eram 2 ms por sessão, o
// mesmo custo quadrático dos cards (IDEIAS.md, "Atualização em tempo real").
var quadros sync.Map

type quadroKey struct {
	o       *obra
	x, y, w int
	hex     string
}

func drawn(o *obra, x, y, w int, hex string, tinge bool) string {
	k := quadroKey{o, x, y, w, hex}
	if s, ok := quadros.Load(k); ok {
		return s.(string)
	}
	var swap map[uint32]color.Color
	if tinge {
		swap = pet.ArtworkColorMap(pieceBody[0], pieceBody[1], hex)
	}
	if o.nota {
		if swap == nil {
			swap = map[uint32]color.Color{}
		}
		swap[notaCor] = color.RGBA{} // o marcador some do desenho; quem aparece é o ♫ (cantando)
	}
	s := pet.Pixels(o.sheet, x, y, w, pieceH, 1, swap)
	if o.nota {
		s = cantando(o, s, x, y, w)
	}
	quadros.Store(k, s)
	return s
}

// Nos temas que cantam, os pixels de notaCor na folha não são desenho: marcam onde sai o ♫, na
// célula do pixel mais baixo da marca. Assim cada quadro diz onde fica a nota: no quadro A ela
// sai encostada no bico ou no instrumento, e no B sobe uma linha se afastando. É texto, e não
// pixel, porque a nota em pixel era um triângulo amarelo; e é o ♫, não o ♪, porque ele é o único
// símbolo musical de largura fixa: o ♪ ocupa duas colunas em terminal configurado para chinês
// ou japonês e desalinha a fileira.
const notaCor = 0xFFD166

var notaMusical = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD166")).Render("♫")

func cantando(o *obra, s string, x0, y0, w int) string {
	cx, cy := -1, -1
	for y := y0; y < y0+pieceH; y++ {
		for x := x0; x < x0+w; x++ {
			if r, g, b, a := o.sheet.At(x, y).RGBA(); a > 0 && r>>8<<16|g>>8<<8|b>>8 == notaCor && y > cy {
				cx, cy = x-x0, y // o primeiro da esquerda na linha mais baixa
			}
		}
	}
	if cy < 0 { // este quadro não canta
		return s
	}
	l := strings.Split(s, "\n")
	i := (cy - y0) / 2
	l[i] = ansi.Truncate(l[i], cx, "") + notaMusical + ansi.TruncateLeft(l[i], cx+1, "")
	return strings.Join(l, "\n")
}

// center: o desenho ou o apelido no meio de uma coluna de col células.
func center(s string, col int) string {
	return lipgloss.NewStyle().Width(col).Align(lipgloss.Center).Render(s)
}
