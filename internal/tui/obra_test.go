package tui

import (
	"image/color"
	"math"
	"plantao/internal/pet"
	"plantao/internal/rooms"
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

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
		if h := o.sheet.Bounds().Dy(); h != pieceH {
			t.Errorf("%s: folha de %d px de altura, queria %d", nome, h, pieceH)
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
		for _, c := range pet.Tint(pieceBody[0], pieceBody[1], pc.Hex) {
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
		{0, pet.Colors[3].Hex, 0},
		{1, pet.Colors[3].Hex, 1},
		{2, "", 2},
		{3, "", 3},
		{2, pet.Colors[1].Hex, 4},
		{3, pet.Colors[7].Hex, 17},
	}
	for _, c := range cases {
		if got := stageFrame(o, c.stage, c.hex); got != c.want {
			t.Errorf("estágio %d, cor %q: quadro %d, queria %d", c.stage, c.hex, got, c.want)
		}
	}
	if got := stageFrame(obras["horta"], 3, pet.Colors[7].Hex); got != 3 {
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

// A tela abre mão das coisas na ordem certa quando falta altura: a obra encolhe antes de sair,
// e sai por último, porque sair primeiro era o mesmo que nunca aparecer (ADR 0008).
func TestFitSteps(t *testing.T) {
	sala := rooms.Snapshot{Obra: "cidade", Meta: 60, Max: rooms.MaxMembers, Members: []rooms.Member{
		{ID: "eu", Nick: "eu"}, {ID: "a", Nick: "ana"}, {ID: "b", Nick: "bruno"},
	}}
	m := model{id: "eu", room: "s", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": sala}}}
	steps := m.fitSteps(sala, 120)
	if len(steps) < 4 {
		t.Fatalf("passos de menos: %d", len(steps))
	}
	if steps[0].ob == "" || steps[0].compact {
		t.Error("o primeiro passo tem que ser a tela cheia, com a obra")
	}
	if steps[len(steps)-1].ob != "" {
		t.Error("o último passo tem que ser sem obra")
	}
	// a obra tem que encolher antes de os pets virarem ponto, e sumir depois deles
	var achatou, pontinho, saiu int
	for i, st := range steps {
		switch {
		case st.ob != "" && st.ob != steps[0].ob && !st.compact:
			achatou = i
		case st.ob != "" && st.compact:
			pontinho = i
		case st.ob == "" && saiu == 0:
			saiu = i
		}
	}
	if !(achatou < pontinho && pontinho < saiu) {
		t.Errorf("ordem errada: achatou=%d pontinho=%d saiu=%d", achatou, pontinho, saiu)
	}
}

// A sala sem obra não ganha passos a mais: a escada continua a de antes.
func TestFitStepsSemObra(t *testing.T) {
	sala := rooms.Snapshot{Max: rooms.MaxMembers, Members: []rooms.Member{{ID: "eu"}}}
	m := model{id: "eu", room: "s", w: rooms.World{Rooms: map[string]rooms.Snapshot{"s": sala}}}
	if got := m.fitSteps(sala, 120); len(got) != 2 || got[0].ob != "" {
		t.Errorf("passos %v", got)
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
