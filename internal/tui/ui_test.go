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
