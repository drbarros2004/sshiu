package tui

import (
	"fmt"
	"image"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

const gardenEditorSizeHint = "aumente o terminal para editar no jardim"

const (
	gardenFullPanelWidth  = 120
	gardenFullPanelHeight = 36
)

func (m model) gardenEditorFits() bool { return m.width >= 64 && m.height >= 20 }

func (m model) gardenEditorKey(msg tea.KeyPressMsg) (model, tea.Cmd) {
	g := m.garden
	if g.mode == gardenEditorHidden {
		g.mode = gardenEditor // a primeira tecla só traz o painel de volta
	} else if msg.Code == tea.KeySpace {
		g.mode = gardenEditorHidden
	} else {
		switch msg.String() {
		case "esc":
			m.garden = nil
			return m, nil
		case "up", "down", "left", "right":
			if msg.String() == "up" && m.vrow == 1 {
				return m, nil
			}
			wallpaper := m.vis.Wallpaper
			next, _ := m.visualKey(msg)
			m = next.(model)
			if m.vis.Wallpaper != wallpaper || !slices.Contains(shownWallpapers(m.vis), g.wall) {
				g.wall = m.gardenWall(true)
				g.scene, g.start, g.last = wallpapers[g.wall].new(), m.now, m.now
			}
		default:
			return m, nil
		}
	}
	g.render(m)
	return m, nil
}

// gardenSettingsSpot tenta primeiro o canto com menos elementos do wallpaper.
// O painel sempre fica fora do bloco de relógio/tarefa e dos vídeos.
func (c *canvas) gardenSettingsSpot(wall string, protected image.Rectangle, width, height, footer int) image.Rectangle {
	if width > c.w-2 || height+footer > c.h {
		return image.Rectangle{}
	}
	left, center, right := 0, 1, 2
	corners := [...]image.Point{{right, 0}, {left, 0}, {center, 0}, {left, 1}, {right, 1}, {center, 1}}
	switch wall {
	case "coffee":
		corners[0], corners[3] = corners[3], corners[0] // lousa à esquerda em cima, xícara à direita embaixo
	case "brainrot":
		corners[0], corners[2] = corners[2], corners[0] // o espaço entre os vídeos
	}
	for _, corner := range corners {
		x := [...]int{1, (c.w - width) / 2, c.w - width - 1}[corner.X]
		y := corner.Y * (c.h - height - footer)
		box := image.Rect(x, y, x+width, y+height)
		if !box.Overlaps(protected) && !c.overlapsVideo(box) {
			return box
		}
	}
	return image.Rectangle{}
}

func (c *canvas) overlapsVideo(box image.Rectangle) bool {
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for _, video := range c.raw[y] {
			if video.x < box.Max.X && video.x+video.w > box.Min.X {
				return true
			}
		}
	}
	return false
}

func (c *canvas) gardenSettingsPanel(m model, wall string, protected image.Rectangle) {
	rows := m.visualRows()[1:]
	title := " modo jardim "
	labels := make([]string, len(rows))
	width := ansi.StringWidth(title) + 3
	for i, row := range rows {
		cursor := "  "
		if i+1 == m.vrow {
			cursor = "▸ "
		}
		labels[i] = fmt.Sprintf("%s%-24s", cursor, row[0]+"  ")
		width = max(width, ansi.StringWidth(labels[i]+row[1])+8)
	}

	var box image.Rectangle
	// Nestes desenhos, a caixa grande só cabe sem esconder a arte quando
	// cabe num lado da tela (a árvore, a xícara e o carro ficam no outro).
	if c.w >= gardenFullPanelWidth && c.h >= gardenFullPanelHeight &&
		(width <= c.w/2-1 || wall != "tree" && wall != "coffee" && wall != "road") {
		box = c.gardenSettingsSpot(wall, protected, width, len(rows)+2, 1)
	}
	selected := m.vrow - 1
	compact := box.Empty()
	if compact {
		label, value := rows[selected][0], rows[selected][1]
		if selected == 3 {
			label = "colegas focando"
		}
		if selected == len(rows)-1 {
			label = "wallpapers de piada"
		}
		if selected == 0 {
			switch m.vis.Wallpaper {
			case "":
				value = "sorteado"
			case "brainrot":
				value = "brainrot"
			}
		}
		title = fmt.Sprintf(" %d/%d ", selected+1, len(rows))
		labels = []string{label + " "}
		rows = [][2]string{{label, value}}
		selected = -1
		for width := min(38, c.w-2); width >= 22; width-- {
			box = c.gardenSettingsSpot(wall, protected, width, 5, 0)
			if !box.Empty() {
				break
			}
		}
		if box.Empty() {
			return
		}
	}

	border := colorName[int(ansi.Convert256(m.st.accent))]
	muted := colorName[int(ansi.Convert256(m.st.muted.GetForeground()))]
	const background, textColor = "16", "255" // preto fixo e texto claro na paleta xterm
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			c.cells[y*c.w+x] = cell{r: ' ', b: background}
		}
	}

	inner := box.Dx() - 4
	titleWidth := ansi.StringWidth(title)
	c.ink(box.Min.X, box.Min.Y, "╭"+strings.Repeat(" ", titleWidth)+strings.Repeat("─", box.Dx()-2-titleWidth)+"╮", border)
	c.ink(box.Min.X+1, box.Min.Y, title, textColor)
	for y := box.Min.Y + 1; y < box.Max.Y-1; y++ {
		c.ink(box.Min.X, y, "│", border)
		c.ink(box.Max.X-1, y, "│", border)
	}
	for i, row := range rows {
		y := box.Min.Y + i + 1
		labelColor := textColor
		if i == selected {
			labelColor = border
		}
		label, value := labels[i], row[1]
		if compact && ansi.StringWidth(label+value)+4 > inner {
			label = ansi.Truncate(strings.TrimSpace(label), max(1, inner-5-ansi.StringWidth(value)), "…") + " "
			value = ansi.Truncate(value, inner-4-ansi.StringWidth(label), "…")
		}
		x := box.Min.X + 2
		c.ink(x, y, label, labelColor)
		x += ansi.StringWidth(label)
		c.ink(x, y, "‹ ", muted)
		c.ink(x+2, y, value, textColor)
		c.ink(x+2+ansi.StringWidth(value), y, " ›", muted)
	}
	c.ink(box.Min.X, box.Max.Y-1, "╰"+strings.Repeat("─", box.Dx()-2)+"╯", border)

	if compact {
		help := []string{"↑/↓ opção · ←/→ muda", "espaço oculta · esc volta"}
		if inner < ansi.StringWidth(help[1]) {
			help = []string{"↑/↓ opção ←/→ muda", "espaço · esc volta"}
		}
		for i, line := range help {
			c.ink(box.Min.X+2, box.Min.Y+i+2, line, muted)
		}
	} else {
		help := "↑/↓ opção · ←/→ muda · espaço oculta · esc volta"
		c.text((c.w-ansi.StringWidth(help))/2, c.h-1, help, muted)
	}
}
