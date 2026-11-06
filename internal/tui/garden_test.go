package tui

import (
	"slices"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
	"strings"
	"testing"
	"time"
)

func TestSplitTask(t *testing.T) {
	a40 := strings.Repeat("a", 40)
	cases := map[string][]string{
		"revisar a lista 3": {"revisar a lista 3"},
		a40:                 {a40},
		a40 + "bb":          {strings.Repeat("a", 21) + "-", strings.Repeat("a", 19) + "bb"},
		"fazer os exercícios de 1 a 12 da lista 3 de cálculo": {"fazer os exercícios de 1", "a 12 da lista 3 de cálculo"},
	}
	for s, want := range cases {
		if got := splitTask(s); !slices.Equal(got, want) {
			t.Errorf("%q: got %q, want %q", s, got, want)
		}
	}
}

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
	if i := wallpaperFor(store.Garden{Wallpaper: "rain"}, rooms.Focus{}); wallpapers[i].id != "rain" {
		t.Fatalf("escolhido pela aba visual: %d", i)
	}
}

func TestSwapLastsForTheTask(t *testing.T) {
	t0 := time.Now()
	m := model{id: "a", now: t0, w: rooms.World{Users: map[string]rooms.Member{"a": {Focus: rooms.Focus{Start: t0}}}}}
	m, _ = m.openGarden(false)
	shown := shownWallpapers(m.vis)
	want := shown[(slices.Index(shown, m.garden.wall)+1)%len(shown)]
	m, _ = m.swapWallpaper(m.garden, 1)
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
	for _, g := range []store.Garden{{}, {Joke: true}} {
		shown := shownWallpapers(g)
		last := wallpapers[shown[len(shown)-1]].id
		if got := cycleWallpaper(g, 1); got != wallpapers[0].id {
			t.Errorf("%+v: depois do sorteado: %q", g, got)
		}
		if got := cycleWallpaper(store.Garden{Wallpaper: last, Joke: g.Joke}, 1); got != "" {
			t.Errorf("%+v: depois do último: %q, want o sorteado", g, got)
		}
		if got := cycleWallpaper(g, -1); got != last {
			t.Errorf("%+v: antes do sorteado: %q", g, got)
		}
	}
}

func TestJokeWallpapers(t *testing.T) { // o brainrot só para quem pede (#36)
	joke := wallpaperIndex("brainrot")
	if !wallpapers[joke].joke {
		t.Fatal("o brainrot é de piada")
	}
	if slices.Contains(shownWallpapers(store.Garden{}), joke) || !slices.Contains(shownWallpapers(store.Garden{Joke: true}), joke) {
		t.Fatal("o brainrot aparece só com Joke")
	}
	if i := wallpaperFor(store.Garden{Wallpaper: "brainrot"}, rooms.Focus{Start: time.Now()}); i == joke {
		t.Error("escolhido mas não permitido: devia sortear outro")
	}
	t0 := time.UnixMilli(0)
	for k := range len(wallpapers) * 3 { // o sorteio por tarefa passa por todos os permitidos
		f := rooms.Focus{Start: t0.Add(time.Duration(k) * time.Millisecond)}
		if wallpaperFor(store.Garden{}, f) == joke {
			t.Fatal("o brainrot caiu no sorteio de quem não pediu")
		}
	}
}

func TestShadeAndPut(t *testing.T) {
	ramp := []uint8{0, 17, 60}
	if shade(ramp, 0, 3, 1) != 0 || shade(ramp, 1, 3, 1) != 60 || shade(ramp, 2, 0, 0) != 60 {
		t.Fatal("as pontas da rampa")
	}
	p := newPix(2, 2)
	p.set(0, 0, 17)
	p.set(0, 1, 60)
	p.set(1, 1, shade(ramp, 0, 1, 1)) // 0 é transparente: não pinta
	c := newCanvas(2, 1)
	c.put(p)
	if want := "\x1b[0;38;5;17;48;5;60m▀\x1b[0m "; c.String() != want {
		t.Errorf("got %q, want %q", c.String(), want)
	}
	c.ink(0, 0, "x ", "255") // a letra fica no fundo de baixo; o espaço não apaga nada
	if c.cells[0] != (cell{'x', "255", "60"}) || c.cells[1] != (cell{r: ' '}) {
		t.Errorf("ink: %+v", c.cells)
	}
}
