package tui

import (
	"plantao/internal/rooms"
	"plantao/internal/store"
	"testing"
	"time"
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

func TestWallpaperForWithoutTask(t *testing.T) { // a prévia sem tarefa, em sorteado, derrubava a sessão
	for range 50 {
		if i := wallpaperFor(store.Garden{}, rooms.Focus{}); i < 0 || i >= len(wallpapers) {
			t.Fatalf("wallpaperFor sem tarefa = %d", i)
		}
	}
}

func TestSwapLastsForTheTask(t *testing.T) {
	t0 := time.Now()
	m := model{id: "a", now: t0, w: rooms.World{Users: map[string]rooms.Member{"a": {Focus: rooms.Focus{Start: t0}}}}}
	m, _ = m.openGarden(false)
	want := (m.garden.wall + 1) % len(wallpapers)
	m, _ = m.swapWallpaper(1)
	m.garden = nil
	if m, _ = m.openGarden(false); m.garden.wall != want {
		t.Errorf("reabrindo na mesma tarefa: %d, want %d", m.garden.wall, want)
	}
	m.w.Users["a"] = rooms.Member{Focus: rooms.Focus{Start: t0.Add(time.Hour)}}
	if got, def := m.gardenWall(false), wallpaperFor(m.vis, m.me().Focus); got != def {
		t.Errorf("na tarefa seguinte: %d, want o padrão %d", got, def)
	}
}

func TestCycleWallpaper(t *testing.T) {
	last := wallpapers[len(wallpapers)-1].id
	if got := cycleWallpaper("", 1); got != wallpapers[0].id {
		t.Errorf("depois do sorteado: %q", got)
	}
	if got := cycleWallpaper(last, 1); got != "" {
		t.Errorf("depois do último: %q, want o sorteado", got)
	}
	if got := cycleWallpaper("", -1); got != last {
		t.Errorf("antes do sorteado: %q", got)
	}
}
