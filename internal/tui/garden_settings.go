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
func (c *canvas) gardenSettingsSpot(wall string, protected image.Rectangle) image.Rectangle {
	const height = 5
	left, center, right := 0, 1, 2
	corners := [...]image.Point{{right, 0}, {left, 0}, {center, 0}, {left, 1}, {right, 1}, {center, 1}}
	switch wall {
	case "coffee":
		corners[0], corners[3] = corners[3], corners[0] // lousa à esquerda em cima, xícara à direita embaixo
	case "brainrot":
		corners[0], corners[2] = corners[2], corners[0] // o espaço entre os vídeos
	}
	for _, corner := range corners {
		for width := min(38, c.w); width >= 22; width-- {
			x := [...]int{1, (c.w - width) / 2, c.w - width - 1}[corner.X]
			y := corner.Y * (c.h - height)
			box := image.Rect(x, y, x+width, y+height)
			if box.Overlaps(protected) || c.overlapsVideo(box) {
				continue
			}
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
	box := c.gardenSettingsSpot(wall, protected)
	if box.Empty() {
		return
	}

	border := colorName[int(ansi.Convert256(m.st.accent))]
	background := colorName[int(ansi.Convert256(m.st.off))]
	textColor := "255"
	r, g, b, _ := m.st.off.RGBA()
	if r+g+b > 3*0x8000 {
		textColor = "233"
	}
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			c.cells[y*c.w+x] = cell{r: ' ', b: background}
		}
	}

	rows := m.visualRows()
	label, value := rows[m.vrow][0], rows[m.vrow][1]
	if m.vrow == 6 {
		label = "wallpapers de piada"
	}
	if m.vrow == 1 {
		switch m.vis.Wallpaper {
		case "":
			value = "sorteado"
		case "brainrot":
			value = "brainrot"
		}
	}
	inner := box.Dx() - 4
	title := ansi.Truncate(fmt.Sprintf(" %d/%d · %s ", m.vrow+1, len(rows), label), box.Dx()-3, "…")
	c.ink(box.Min.X, box.Min.Y, "╭"+title+strings.Repeat("─", box.Dx()-2-ansi.StringWidth(title))+"╮", border)
	help := "espaço oculta · esc volta"
	if inner < ansi.StringWidth(help) {
		help = "espaço · esc volta"
	}
	for i, line := range []string{"‹ " + value + " ›", "↑/↓ opção ←/→ muda", help} {
		y := box.Min.Y + i + 1
		c.ink(box.Min.X, y, "│", border)
		c.ink(box.Max.X-1, y, "│", border)
		c.ink(box.Min.X+2, y, ansi.Truncate(line, inner, "…"), textColor)
	}
	c.ink(box.Min.X, box.Max.Y-1, "╰"+strings.Repeat("─", box.Dx()-2)+"╯", border)
}
