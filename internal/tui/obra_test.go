package tui

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"strings"
	"testing"
	"time"
	_ "time/tzdata"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestObraResetIn(t *testing.T) {
	zone := time.FixedZone("servidor", -3*60*60)
	for _, tc := range []struct{ at, want string }{
		{"2026-11-04T20:17:42", "03:42:18"},
		{"2026-11-04T23:59:58", "00:00:02"},
		{"2026-11-04T23:59:59.5", "00:00:01"},
		{"2026-11-05T00:00:00", "24:00:00"},
		{"2026-11-05T00:00:01", "23:59:59"},
		{"2026-12-31T23:59:59", "00:00:01"},
	} {
		now, err := time.ParseInLocation("2006-01-02T15:04:05", tc.at, zone)
		if err != nil {
			t.Fatal(err)
		}
		if got := obraResetIn(now); got != tc.want {
			t.Errorf("%s: %s, queria %s", tc.at, got, tc.want)
		}
	}
	// O dia pode ter 23 ou 25 horas; não se pode assumir uma duração fixa de 24 h.
	dst, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		month time.Month
		day   int
		want  string
	}{
		{time.March, 8, "23:00:00"}, {time.November, 1, "25:00:00"},
	} {
		now := time.Date(2026, tc.month, tc.day, 0, 0, 0, 0, dst)
		if got := obraResetIn(now); got != tc.want {
			t.Errorf("horário de verão: %s, queria %s", got, tc.want)
		}
	}
}

func TestObraHeadCountdown(t *testing.T) {
	m := model{now: time.Date(2026, 11, 4, 23, 59, 59, 0, time.Local), st: newStyles(themes[0])}
	s := rooms.Snapshot{Obra: "trem", Meta: 60}
	if text := ansi.Strip(m.obraHead(s, 120)); !strings.Contains(text, "a peça muda a cada 20 min · reinicia em 00:00:01") {
		t.Fatalf("cabeçalho: %s", text)
	}
	// Em janelas estreitas o texto deve quebrar, em vez de cortar o contador.
	text := m.obraHead(s, 36)
	if lipgloss.Width(text) > 36 || !strings.Contains(ansi.Strip(text), "00:00:01") {
		t.Fatalf("contador cortado na janela estreita: %s", text)
	}
	m.now = m.now.Add(time.Second)
	if !strings.Contains(ansi.Strip(m.obraHead(s, 120)), "reinicia em 24:00:00") {
		t.Fatal("o cabeçalho não acompanhou a virada")
	}
}

// O arranjo tem que dar conta de todo mundo: uma peça por pessoa, nem a mais nem a menos.
func TestObraRows(t *testing.T) {
	for n := 1; n <= rooms.MaxMembers; n++ {
		for nome, o := range obras {
			rows := o.rows(n, 120)
			soma := 0
			for _, k := range rows {
				soma += k
			}
			if soma != n {
				t.Errorf("%s com %d pessoas: %v soma %d", nome, n, rows, soma)
			}
		}
	}
}

// O trem quebra de linha quando não cabe, e nunca some: numa janela estreita sobra uma peça
// por fileira em vez de nenhuma.
func TestTremQuebra(t *testing.T) {
	if got := fitRows(6, 8); !slices.Equal(got, []int{6}) {
		t.Errorf("cabe numa fileira só: %v", got)
	}
	if got := fitRows(10, 5); !slices.Equal(got, []int{5, 5}) {
		t.Errorf("cinco por fileira: %v", got)
	}
	if got := fitRows(3, 0); !slices.Equal(got, []int{1, 1, 1}) {
		t.Errorf("janela minúscula: %v", got)
	}
	if got := fitRows(5, 4); !slices.Equal(got, []int{3, 2}) { // e não 4 e um vagão solto
		t.Errorf("dividido por igual: %v", got)
	}
}

// Os quatro estágios existem na folha de cada tema, senão a peça pronta sai em branco.
func TestFolhas(t *testing.T) {
	for nome, o := range obras {
		quadros := rooms.ObraStages
		if o.inst { // cadeira e estante, e um par de quadros por cor
			quadros = 2 + 2*len(pet.Colors)
		}
		want := o.w()*quadros + o.fixo // a peça fixa vem depois, com a largura dela
		if w := o.sheet.Bounds().Dx(); w != want {
			t.Errorf("%s: folha de %d px, queria %d", nome, w, want)
		}
		faixas := 1
		if o.anim > 0 { // o tema que se mexe tem a faixa do quadro B embaixo
			faixas = 2
		}
		if h := o.sheet.Bounds().Dy(); h != pieceH*faixas {
			t.Errorf("%s: folha de %d px de altura, queria %d", nome, h, pieceH*faixas)
		}
	}
}

// O tema que se mexe mexe de verdade: o quadro B da peça pronta é outro desenho, e o tick troca
// um pelo outro. Os parados dão o mesmo desenho em qualquer tick.
func TestMovimento(t *testing.T) {
	for nome, o := range obras {
		m, s := salaCom(nome, 4, 3)
		a := m.obraBlock(s, 120, false)
		m.tick = max(o.anim, 1)
		if b := m.obraBlock(s, 120, false); (a != b) != (o.anim > 0) {
			t.Errorf("%s: anim=%d, mas o desenho mudou=%v", nome, o.anim, a != b)
		}
	}
}

// O quadro B só mexe no desenho, não pinta: toda cor dele já está no quadro A da mesma peça. Cor
// nova ali é pixel escrito na mão com a paleta errada: o passarinho se abaixaria com as cores
// de outro tema e perderia o tom do pet.
func TestQuadroBNaoInventaCor(t *testing.T) {
	cores := func(o *obra, x0, w, y0 int) map[color.RGBA]bool {
		c := map[color.RGBA]bool{}
		for x := x0; x < x0+w; x++ {
			for y := y0; y < y0+pieceH; y++ {
				if p := color.RGBAModel.Convert(o.sheet.At(x, y)).(color.RGBA); p.A > 0 {
					c[p] = true
				}
			}
		}
		return c
	}
	for nome, o := range obras {
		if o.anim == 0 {
			continue
		}
		dx := o.sheet.Bounds().Dx()
		var pecas [][2]int // x e largura de cada peça na folha; a fixa vem no fim
		for x := 0; x < dx-o.fixo; x += o.w() {
			pecas = append(pecas, [2]int{x, o.w()})
		}
		if o.fixo > 0 {
			pecas = append(pecas, [2]int{dx - o.fixo, o.fixo})
		}
		for _, p := range pecas {
			a := cores(o, p[0], p[1], 0)
			for c := range cores(o, p[0], p[1], pieceH) {
				if !a[c] {
					t.Errorf("%s, peça em x=%d: o quadro B tem #%02X%02X%02X, que o A não tem", nome, p[0], c.R, c.G, c.B)
				}
			}
		}
	}
}

// O passarinho se abaixa em todo estágio em que já existe. As duas primeiras linhas ficam de
// fora: lá mora a nota do estágio 3, que sozinha já mudava o desenho enquanto o pássaro daquele
// estágio continuava em pé, e o TestMovimento não via a diferença.
func TestPassarinhoSeAbaixa(t *testing.T) {
	aves, cor := obras["passarinhos"], pet.Colors[0].ID
	parado, mexendo := model{}, model{tick: aves.anim}
	corpo := func(s string) string { return strings.Join(strings.Split(s, "\n")[2:], "\n") }
	for stage := 1; stage < rooms.ObraStages; stage++ { // no estágio 0 o fio ainda está vazio
		if corpo(parado.frame(aves, stage, cor)) == corpo(mexendo.frame(aves, stage, cor)) {
			t.Errorf("o pássaro do estágio %d não se abaixa", stage)
		}
	}
}

// O ♫ sai só de quem toca, onde a folha marca: encostado no instrumento (ou no bico do pássaro)
// no quadro A, e uma linha acima, se afastando, no B. Quadro fora da tabela não canta.
func TestNota(t *testing.T) {
	type pos struct{ x, y int } // coluna e linha do ♫; {-1, -1} é sem nota
	sem := pos{-1, -1}
	casos := []struct {
		tema   string
		quadro map[int][2]pos // quadro da folha: onde fica o ♫ no A e no B
	}{
		{"passarinhos", map[int][2]pos{rooms.ObraStages - 1: {{0, 1}, {0, 0}}}},
		{"orquestra", map[int][2]pos{
			3:  {{4, 2}, {3, 1}},   // violino: sai das cordas e sobe pelo arco
			5:  {{9, 2}, {10, 1}},  // violão: do ombro do corpo
			7:  {{9, 1}, {10, 0}},  // trompete: da campana
			9:  {{5, 1}, {5, 0}},   // tambor: do couro, entre as baquetas
			11: {{10, 2}, {11, 1}}, // violoncelo: do ombro do corpo
			13: {{7, 2}, {8, 1}},   // harpa: das cordas, por cima do braço
			15: {{10, 2}, {10, 1}}, // saxofone: da campana, que abre para cima
			17: {{11, 1}, {11, 0}}, // piano: pelo lado do tampo
		}},
	}
	onde := func(s string) pos {
		for y, l := range strings.Split(ansi.Strip(s), "\n") {
			if x := slices.Index([]rune(l), '♫'); x >= 0 {
				return pos{x, y}
			}
		}
		return sem
	}
	for _, c := range casos {
		o, cor := obras[c.tema], pet.Colors[0].ID
		for q := range (o.sheet.Bounds().Dx() - o.fixo) / o.w() {
			quer, canta := c.quadro[q]
			if !canta {
				quer = [2]pos{sem, sem}
			}
			for i, tick := range []int{0, o.anim} {
				if got := onde((model{tick: tick}).frame(o, q, cor)); got != quer[i] {
					t.Errorf("%s, quadro %d, tick %d: ♫ em %v, queria %v", c.tema, q, tick, got, quer[i])
				}
			}
		}
	}
}

// Toda cor das folhas, e o corpo tingido em cada cor do pet, sobrevive ao terminal de 256 cores:
// não vira cinza nem troca de matiz. Era assim que a pétala rosa saía cinza e a sombra da
// estante saía marrom. Os cinzas de propósito (trilho, prata, teto) ficam de fora.
func TestCoresEm256(t *testing.T) {
	check := func(onde string, c color.Color) {
		got := colorprofile.ANSI256.Convert(c)
		h1, s1 := hueSat(c)
		h2, s2 := hueSat(got)
		d := math.Abs(h1 - h2)
		d = math.Min(d, 360-d)
		r, g, b, _ := c.RGBA()
		escura := max(r, g, b)>>8 < 0x50 // o vão de porta e tecla: quase preto, cinza tanto faz
		if s1 >= 0.25 && !escura && (s2 < 0.1 || d > 40) {
			t.Errorf("%s: #%02X%02X%02X vira %v em 256 cores (matiz %.0f -> %.0f)", onde, r>>8, g>>8, b>>8, got, h1, h2)
		}
	}
	for nome, o := range obras {
		seen := map[color.RGBA]bool{}
		b := o.sheet.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := color.RGBAModel.Convert(o.sheet.At(x, y)).(color.RGBA)
				if c.A > 0 && !seen[c] {
					seen[c] = true
					check(nome, c)
				}
			}
		}
	}
	for _, pc := range pet.Colors {
		for _, c := range pet.ArtworkColorMap(pieceBody[0], pieceBody[1], pc.ID) {
			check("corpo "+pc.Name, c)
		}
	}
}

// hueSat: o matiz em graus e a saturação (HSV) de uma cor.
func hueSat(c color.Color) (float64, float64) {
	r, g, b, _ := c.RGBA()
	R, G, B := float64(r>>8), float64(g>>8), float64(b>>8)
	mx, mn := math.Max(R, math.Max(G, B)), math.Min(R, math.Min(G, B))
	if mx == 0 || mx == mn {
		return 0, 0
	}
	var h float64
	switch mx {
	case R:
		h = math.Mod((G-B)/(mx-mn)+6, 6)
	case G:
		h = (B-R)/(mx-mn) + 2
	default:
		h = (R-G)/(mx-mn) + 4
	}
	return h * 60, (mx - mn) / mx
}

// Na orquestra o instrumento é da cor: cada cor cai no seu par de quadros, e quem nunca escolheu
// toca o instrumento da primeira cor. Os outros temas usam o estágio direto.
func TestStageFrame(t *testing.T) {
	o := obras["orquestra"]
	cases := []struct {
		stage int
		hex   string
		want  int
	}{
		{0, pet.Colors[3].ID, 0},
		{1, pet.Colors[3].ID, 1},
		{2, "", 2},
		{3, "", 3},
		{2, pet.Colors[1].ID, 4},
		{3, pet.Colors[7].ID, 17},
	}
	for _, c := range cases {
		if got := stageFrame(o, c.stage, c.hex); got != c.want {
			t.Errorf("estágio %d, cor %q: quadro %d, queria %d", c.stage, c.hex, got, c.want)
		}
	}
	if got := stageFrame(obras["horta"], 3, pet.Colors[7].ID); got != 3 {
		t.Errorf("horta: quadro %d, queria 3", got)
	}
}

// A fogueira entra no meio da fileira do meio, e cada barraca continua com o seu nome.
func TestFogueiraNoMeio(t *testing.T) {
	sala := rooms.Snapshot{Obra: "acampamento", Meta: 60, Max: rooms.MaxMembers, Members: []rooms.Member{
		{ID: "eu", Nick: "eu"}, {ID: "a", Nick: "ana"}, {ID: "b", Nick: "bia"}, {ID: "c", Nick: "caio"},
	}}
	m := model{id: "eu", room: "s", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": sala}}}
	lines := strings.Split(m.obraBlock(sala, 120, false), "\n")
	var nomes []string
	for _, l := range lines {
		if f := strings.Fields(l); len(f) > 0 && !strings.ContainsAny(l, "▀▄█·") {
			nomes = append(nomes, strings.Join(f, " "))
		}
	}
	// 1-2-1: a fogueira não tem nome, então a fileira do meio mostra só os dois de cada lado
	if want := []string{"você", "ana bia", "caio"}; !slices.Equal(nomes, want) {
		t.Errorf("fileiras %q, queria %q", nomes, want)
	}
}

// salaCom: uma sala de n pessoas com apelidos de tamanho nick, em todos os estágios.
func salaCom(obra string, n, nick int) (model, rooms.Snapshot) {
	s := rooms.Snapshot{Obra: obra, Meta: 60, Max: rooms.MaxMembers}
	for i := range n {
		s.Members = append(s.Members, rooms.Member{ID: fmt.Sprint("m", i),
			Nick: strings.Repeat(string(rune('a'+i)), nick), TodayMin: float64(i%4) * 20,
			Color: pet.Colors[i%len(pet.Colors)].ID})
	}
	return model{id: "m0", room: "s", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": s}}}, s
}

// Nenhuma obra passa da largura que recebeu, em nenhum tamanho de sala, com apelido curto ou com
// o mais comprido: a tabela da cidade e do acampamento cede para a fileira quando não cabe. Linha
// que passa da tela o fit corta com "…" no meio do desenho.
func TestObraCabeNaLargura(t *testing.T) {
	for nome := range obras {
		for n := 2; n <= rooms.MaxMembers; n++ {
			for _, nick := range []int{3, 16} {
				for _, w := range []int{30, 40, 50, 60, 80, 100, 120, 160} {
					for _, flat := range []bool{false, true} {
						m, s := salaCom(nome, n, nick)
						for _, l := range strings.Split(m.obraBlock(s, w, flat), "\n")[1:] { // o cabeçalho é texto
							if got := lipgloss.Width(l); got > w {
								t.Fatalf("%s, %d pessoas, apelido %d, %d colunas, flat=%v: linha de %d",
									nome, n, nick, w, flat, got)
							}
						}
					}
				}
			}
		}
	}
}

// O apelido fica embaixo da peça dele, e não deslocado para a direita: com um apelido comprido o
// vão cresce, e era aí que os nomes escorregavam.
func TestApelidoEmbaixoDaPeca(t *testing.T) {
	m, s := salaCom("horta", 2, 16)
	lines := strings.Split(ansi.Strip(m.obraBlock(s, 120, true)), "\n")
	art := []rune(lines[len(lines)-3]) // a última linha do desenho
	names := lines[len(lines)-1]
	var centros []float64 // o meio de cada vaso
	for i := 0; i < len(art); {
		if art[i] == ' ' {
			i++
			continue
		}
		j := i
		for j < len(art) && art[j] != ' ' {
			j++
		}
		centros = append(centros, float64(i+j-1)/2)
		i = j
	}
	for k, nome := range []string{"você", strings.Repeat("b", 16)} {
		ini := strings.Index(names, nome)
		meio := float64(len([]rune(names[:ini]))) + float64(len([]rune(nome))-1)/2
		if k >= len(centros) || math.Abs(meio-centros[k]) > 1 {
			t.Errorf("%q no meio %.1f, peça no meio %v", nome, meio, centros)
		}
	}
}

// A tela abre mão das coisas na ordem certa quando falta altura (ADR 0012): a faixa encolhe com os
// cards cheios, os cards viram médios antes de a faixa sair, e os pets só viram ponto com a faixa
// já fora da tela.
func TestFitSteps(t *testing.T) {
	sala := rooms.Snapshot{Obra: "cidade", Meta: 60, Max: rooms.MaxMembers, Members: []rooms.Member{
		{ID: "eu", Nick: "eu"}, {ID: "a", Nick: "ana"}, {ID: "b", Nick: "bruno"},
	}}
	m := model{id: "eu", room: "s", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": sala}}}
	steps := m.fitSteps(sala)
	if steps[0] != (fitStep{cardFull, bandArranjo}) || steps[len(steps)-1] != (fitStep{cardDot, bandFora}) {
		t.Fatalf("a escada tem que ir da tela cheia aos pontos sem a faixa: %v", steps)
	}
	medio, fora := slices.IndexFunc(steps, func(s fitStep) bool { return s.card == cardMid }),
		slices.IndexFunc(steps, func(s fitStep) bool { return s.band == bandFora })
	if medio < 0 || fora < medio {
		t.Errorf("a faixa saiu antes de os cards virarem médios: médio=%d, fora=%d", medio, fora)
	}
	for _, st := range steps {
		if st.card == cardDot && st.band != bandFora {
			t.Errorf("os pets viraram ponto com a faixa ainda na tela: %v", steps)
		}
		if st.card == cardFull && st.band == bandFora {
			t.Errorf("a faixa saiu com os cards ainda cheios: %v", steps)
		}
	}
}

// A sala sem obra e sem tabela aberta só encolhe os cards: não há faixa para tirar.
func TestFitStepsSemObra(t *testing.T) {
	sala := rooms.Snapshot{Max: rooms.MaxMembers, Members: []rooms.Member{{ID: "eu"}}}
	m := model{id: "eu", room: "s", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": sala}}}
	want := []fitStep{{cardFull, bandFora}, {cardMid, bandFora}, {cardDot, bandFora}}
	if got := m.fitSteps(sala); !slices.Equal(got, want) {
		t.Errorf("passos %v, want %v", got, want)
	}
}

// Todo tema que o dono pode escolher tem desenho (ADR 0008).
func TestTemasTemArte(t *testing.T) {
	for _, nome := range rooms.Obras {
		if nome == "" {
			continue
		}
		if obras[nome] == nil {
			t.Errorf("tema %q sem folha", nome)
		}
	}
}
