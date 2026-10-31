package tui

import (
	"plantao/internal/rooms"
	"plantao/internal/store"
	"testing"
)

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

// Sem tarefa em andamento o sorteio não pode sair do intervalo (já deu panic na sessão).
func TestWallpaperFor(t *testing.T) {
	if i := wallpaperFor(store.Garden{}, rooms.Focus{}); i < 0 || i >= len(wallpapers) {
		t.Fatalf("sem tarefa: %d", i)
	}
	if i := wallpaperFor(store.Garden{Wallpaper: "rain"}, rooms.Focus{}); wallpapers[i].id != "rain" {
		t.Fatalf("escolhido pela aba visual: %d", i)
	}
}
