package tui

import (
	"slices"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestGardenEditorModes(t *testing.T) {
	settings := store.DefaultGarden
	settings.Wallpaper = "sky"
	m := model{nick: "ana", tab: tabVisual, width: 120, height: 36, now: time.Now(),
		vis: settings, visSaved: settings, st: newStyles(themes[0])}
	press := func(code rune) {
		t.Helper()
		next, _ := m.key(tea.KeyPressMsg{Code: code})
		m = next.(model)
	}
	panelVisible := func() bool {
		return strings.Contains(ansi.Strip(m.garden.frame), "espaço oculta")
	}
	press('j')
	if m.garden.mode != gardenPreview || panelVisible() {
		t.Fatal("j deveria abrir a prévia sem painel")
	}
	press(tea.KeyEscape)
	press('v')
	if m.garden.mode != gardenEditor || !panelVisible() {
		t.Fatal("v deveria abrir o editor com painel")
	}
	strip := ansi.Strip(m.garden.frame)
	if !strings.Contains(strip, "permitir wallpapers para devs de verdade?") || strings.Contains(strip, "tema") {
		t.Fatal("o painel completo deveria mostrar só as opções do jardim")
	}
	m.width--
	m.garden.render(m)
	strip = ansi.Strip(m.garden.frame)
	if !strings.Contains(strip, "wallpaper ‹ céu noturno ›") || strings.Contains(strip, "permitir wallpapers") {
		t.Fatal("abaixo da largura mínima deveria aparecer o painel compacto")
	}
	press(tea.KeyUp)
	press(tea.KeyRight)
	if m.theme != 0 || m.vis.Wallpaper != "rain" || m.visSaved != settings {
		t.Fatal("a primeira opção deveria mudar o wallpaper em prévia")
	}
	press(tea.KeySpace)
	if panelVisible() || m.garden.mode != gardenEditorHidden {
		t.Fatal("espaço deveria ocultar o painel")
	}
	press(tea.KeyRight)
	if !panelVisible() || m.vis.Wallpaper != "rain" {
		t.Fatal("a primeira tecla só deveria trazer o painel de volta")
	}
	for range 2 {
		press(tea.KeyRight)
	}
	strip = ansi.Strip(m.garden.frame)
	if !strings.Contains(strip, "1/6") || !strings.Contains(strip, "wallpaper ‹ café ›") {
		t.Fatal("a opção selecionada deveria aparecer no painel compacto")
	}
	press(tea.KeyEscape)
	if m.garden != nil || m.tab != tabVisual || m.vis.Wallpaper != "coffee" {
		t.Fatal("esc deveria voltar aos controles com a edição em prévia")
	}
}

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
		if got := focusingLine("", nicks, 80); got != want {
			t.Errorf("%v: got %q, want %q", nicks, got, want)
		}
	}
	// a linha nunca passa da largura, mesmo com nome largo (#85)
	long := []string{"田中さくら東京大学の学生", "🐱🐶🐭🐹🐰🦊🐻🐼", "abcdefghijklmnop"}
	for _, w := range []int{20, 30, 45, 60, 100} {
		if got := focusingLine("", long, w); ansi.StringWidth(got) > w {
			t.Errorf("largura %d: %q tem %d colunas", w, got, ansi.StringWidth(got))
		}
	}
	if got, want := focusingLine("", long, 60), "田中さくら東京大学の学生 e mais 2 também estão focando"; got != want {
		t.Errorf("largura 60: got %q", got)
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
	m, _ = m.openGarden(gardenFocus)
	shown := shownWallpapers(m.vis)
	want := shown[(slices.Index(shown, m.garden.wall)+1)%len(shown)]
	m, _ = m.swapWallpaper(m.garden, 1)
	m.garden = nil
	if m, _ = m.openGarden(gardenFocus); m.garden.wall != want {
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
	if want := "\x1b[0;7;4;38;5;17;48;5;60m▄\x1b[0m "; c.String() != want {
		t.Errorf("got %q, want %q", c.String(), want)
	}
	c.ink(0, 0, "x ", "255") // a letra fica no fundo de baixo; o espaço não apaga nada
	if c.cells[0] != (cell{r: 'x', c: "255", b: "60"}) || c.cells[1] != (cell{r: ' '}) {
		t.Errorf("ink: %+v", c.cells)
	}
}

// Os wallpapers com quadro (a lousa, as placas, o calendário) se desenham em qualquer tela, da
// minúscula à larga, e o quadro fica em volta do texto: o canto que a cena devolve já é o que o
// render usa (fitText não o move) e o quadro sai dele.
func TestTextBoardsAnySize(t *testing.T) {
	for _, id := range []string{"coffee", "tree", "lighthouse", "cat", "dog"} {
		for _, sz := range [][2]int{{20, 10}, {50, 14}, {80, 24}, {120, 36}, {250, 60}} {
			sc := wallpapers[wallpaperIndex(id)].new()
			tw, th := min(33, sz[0]-2), min(8, sz[1])
			x, y, _ := sc.(spot).spot(sz[0], sz[1], tw, th)
			if fx, fy := fitText(sz[0], sz[1], tw, th, x, y); fx != x || fy != y {
				t.Errorf("%s em %v: a cena pôs o texto em (%d, %d) e o render em (%d, %d)", id, sz, x, y, fx, fy)
			}
			var board [4]int
			switch s := sc.(type) {
			case *coffee:
				board = s.board
			case *tree:
				board = s.board
			case *lighthouse:
				board = s.board
			case *sleeper:
				board = s.board
			}
			if board != boardAt(x, y, tw, th) {
				t.Errorf("%s em %v: o quadro %v não está em volta do texto em (%d, %d)", id, sz, board, x, y)
			}
			sc.draw(newCanvas(sz[0], sz[1]), 7.3, 0.125)
		}
	}
}

// caractere largo, acento combinante e espaço de largura zero ocupam as colunas que o terminal
// usa, e a linha nunca passa da largura do jardim (#85).
func TestCanvasWideText(t *testing.T) {
	c := newCanvas(10, 1)
	c.text(0, 0, "日z̷​é🐱abcdef", "255")
	if got := ansi.Strip(c.String()); ansi.StringWidth(got) != 10 || !strings.HasPrefix(got, "日z̷é🐱") {
		t.Errorf("got %q (%d colunas)", got, ansi.StringWidth(got))
	}
	c = newCanvas(4, 1)
	c.text(0, 0, "日本", "255")
	c.set(1, 0, 'x', "") // o desenho apaga metade de cada caractere largo
	c.set(2, 0, 'y', "")
	if got := ansi.Strip(c.String()); got != " xy " {
		t.Errorf("metades apagadas: got %q", got)
	}
	c = newCanvas(3, 1)
	c.ink(2, 0, "日", "255") // não cabe na última coluna: fica de fora
	if got := ansi.Strip(c.String()); got != "   " {
		t.Errorf("na borda: got %q", got)
	}
}
