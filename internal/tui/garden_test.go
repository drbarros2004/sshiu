package tui

import "testing"

func TestFocusingLine(t *testing.T) {
	cases := map[string][]string{
		"ninguém mais está focando agora":                nil,
		"ana também está focando":                        {"ana"},
		"ana e bruno também estão focando":               {"ana", "bruno"},
		"ana, bruno, duda e mais 2 também estão focando": {"ana", "bruno", "duda", "edu", "caio"},
	}
	for want, nicks := range cases {
		if got := focusingLine(nicks); got != want {
			t.Errorf("%v: got %q, want %q", nicks, got, want)
		}
	}
}
