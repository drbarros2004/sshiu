package tui

import (
	"plantao/internal/rooms"
	"slices"
	"testing"
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
	if got := fitRows(6, 120); !slices.Equal(got, []int{6}) {
		t.Errorf("cabe numa fileira só: %v", got)
	}
	if got := fitRows(10, 76); !slices.Equal(got, []int{5, 5}) {
		t.Errorf("76 colunas: %v", got)
	}
	if got := fitRows(3, 4); !slices.Equal(got, []int{1, 1, 1}) {
		t.Errorf("janela minúscula: %v", got)
	}
}

// Os quatro estágios existem na folha de cada tema, senão a peça pronta sai em branco.
func TestFolhas(t *testing.T) {
	for nome, o := range obras {
		quadros := rooms.ObraStages
		if o.lead { // o tema com peça fixa na frente tem um quadro a mais
			quadros++
		}
		if w := o.sheet.Bounds().Dx(); w != pieceW*quadros {
			t.Errorf("%s: folha de %d px, queria %d", nome, w, pieceW*quadros)
		}
		if h := o.sheet.Bounds().Dy(); h != pieceH {
			t.Errorf("%s: folha de %d px de altura, queria %d", nome, h, pieceH)
		}
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
