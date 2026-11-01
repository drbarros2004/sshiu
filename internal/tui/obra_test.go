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
		if w := o.sheet.Bounds().Dx(); w != pieceW*rooms.ObraStages {
			t.Errorf("%s: folha de %d px, queria %d", nome, w, pieceW*rooms.ObraStages)
		}
		if h := o.sheet.Bounds().Dy(); h != pieceH {
			t.Errorf("%s: folha de %d px de altura, queria %d", nome, h, pieceH)
		}
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
