package tui

import "testing"

func TestFit(t *testing.T) {
	cases := map[string]string{
		"a":          "a\n\n",     // completa até a altura
		"a\nb\nc\nd": "a\nb\nc",   // corta o que passa da altura
		"abcdefgh":   "abcd…\n\n", // corta a largura com reticências
		"ab        ": "ab   \n\n", // espaço sobrando sai sem reticências
	}
	for in, want := range cases {
		if got := fit(in, 5, 3); got != want {
			t.Errorf("fit(%q) = %q, want %q", in, got, want)
		}
	}
}

// A tarefa comprida quebra nas palavras e avisa com "…" que continua; a que cabe fica inteira.
func TestEllipsize(t *testing.T) {
	cases := map[string]string{
		"ler o cap. 4":                         "ler o cap. 4",
		"implementar designs para o trabalho":  "implementar\ndesigns para o…",
		"abcdefghijklmnopqrstuvwxyz0123456789": "abcdefghijklmnop\nqrstuvwxyz01234…", // palavra sem fim corta no meio
	}
	for in, want := range cases {
		if got := ellipsize(in, 16, 2); got != want {
			t.Errorf("ellipsize(%q) = %q, want %q", in, got, want)
		}
	}
}
