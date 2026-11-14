package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// Só a resposta escolhida aparece, a lista rola até ela e nada passa da largura.
func TestAboutView(t *testing.T) {
	m := model{st: newStyles(themes[0]), lang: "en"}
	for _, size := range [][2]int{{80, 40}, {36, 12}} {
		w, h := size[0], size[1]
		for sel := range faqs {
			m.aboutSel = sel
			out := ansi.Strip(m.aboutView(w, h))
			if n := strings.Count(out, "\n") + 1; n > h {
				t.Errorf("%dx%d sel %d: %d linhas", w, h, sel, n)
			}
			for _, l := range strings.Split(out, "\n") {
				if ansi.StringWidth(l) > w {
					t.Errorf("%dx%d sel %d: linha larga %q", w, h, sel, l)
				}
			}
			flat := strings.Join(strings.Fields(out), " ")
			first := strings.Fields(m.t(faqs[sel].a))[0] + " " + strings.Fields(m.t(faqs[sel].a))[1]
			if !strings.Contains(flat, first) {
				t.Errorf("%dx%d sel %d: sem a resposta %q", w, h, sel, first)
			}
		}
	}
	m.aboutSel = 9 // --conexao: o número sai da tolerância de reconexão
	if out := m.aboutView(80, 40); !strings.Contains(out, "2 minutes") || strings.Contains(out, m.t(faqs[1].a)) {
		t.Errorf("a resposta da conexão sem o número, ou com outra resposta aberta:\n%s", out)
	}
}
