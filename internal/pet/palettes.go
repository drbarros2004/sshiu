package pet

import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"image/color"
	"strconv"
	"strings"
)

// Palette mantém a ordem e os IDs antigos das cores das obras e dos bots.
// O desenho dos personagens usa exclusivamente os catálogos de assets/pets.
type Palette struct {
	Name, ID string
	Body     color.Color
}

var Colors = func() []Palette {
	colors := make([]Palette, len(Pets.artworks))
	for _, a := range Pets.artworks {
		swatch, _ := ParsePaletteColor(a.Color)
		colors[a.Order] = Palette{Name: a.DisplayName.Text("pt"), ID: a.Color, Body: swatch}
	}
	return colors
}()

// ParsePaletteColor aceita HEX ou um índice explícito, por exemplo "xterm:95".
func ParsePaletteColor(value string) (color.Color, error) {
	if strings.HasPrefix(value, "xterm:") {
		n, err := strconv.Atoi(strings.TrimPrefix(value, "xterm:"))
		if err == nil && n >= 0 && n <= 255 {
			return ansi.IndexedColor(n), nil
		}
	} else if len(value) == 7 && strings.HasPrefix(value, "#") {
		n, err := strconv.ParseUint(value[1:], 16, 24)
		if err == nil {
			return color.RGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 255}, nil
		}
	}
	return nil, fmt.Errorf("cor %q inválida; use #RRGGBB ou xterm:0 até xterm:255", value)
}

func pixelRGB(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	return r>>8<<16 | g>>8<<8 | b>>8
}

// ColorIndex mantém as escolhas existentes; IDs ausentes ou desconhecidos usam azul.
func ColorIndex(id string) int {
	for i, palette := range Colors {
		if palette.ID == id {
			return i
		}
	}
	return 0
}
