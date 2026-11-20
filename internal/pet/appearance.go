package pet

import (
	"charm.land/lipgloss/v2"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"slices"
	"strings"
)

// Look mantém compatibilidade com perfis anteriores ao catálogo de personagens.
func (p Pet) Look() Appearance {
	if p.Appearance.Character == "" {
		return Pets.LegacyAppearance(p.Color)
	}
	return p.Appearance
}

func (c *Catalog) Character(id string) *Character {
	for i := range c.Characters {
		if c.Characters[i].ID == id {
			return &c.Characters[i]
		}
	}
	return nil
}

func (c *Character) Variant(id string) *CharacterVariant {
	for i := range c.Variants {
		if c.Variants[i].ID == id {
			return &c.Variants[i]
		}
	}
	return nil
}

func (c *Catalog) Hat(id string) *Hat {
	for i := range c.Hats {
		if c.Hats[i].ID == id {
			return &c.Hats[i]
		}
	}
	return nil
}

func (h *Hat) Variant(id string) *HatVariant {
	for i := range h.Variants {
		if h.Variants[i].ID == id {
			return &h.Variants[i]
		}
	}
	return nil
}

func (c *Catalog) ValidateAppearance(a Appearance) error {
	ch := c.Character(a.Character)
	if ch == nil {
		return fmt.Errorf("personagem desconhecido %q", a.Character)
	}
	if ch.Variant(a.Variant) == nil {
		return fmt.Errorf("%s: variante desconhecida %q", a.Character, a.Variant)
	}
	if a.Hat == "" {
		if a.HatVariant != "" {
			return fmt.Errorf("modelo de chapéu sem chapéu")
		}
		return nil
	}
	h := c.Hat(a.Hat)
	if h == nil || h.Variant(a.HatVariant) == nil {
		return fmt.Errorf("chapéu/modelo desconhecido %q/%q", a.Hat, a.HatVariant)
	}
	return nil
}

// LegacyAppearance migra apenas a aparência antiga: fruta, cor correspondente e sem chapéu.
func (c *Catalog) LegacyAppearance(oldColor string) Appearance {
	a := Appearance{Character: "fruit", Variant: c.Character("fruit").Variants[0].ID}
	for _, v := range c.Character("fruit").Variants {
		if strings.EqualFold(c.artworks[v.ArtworkColor].Color, oldColor) || v.ID == oldColor {
			a.Variant = v.ID
			break
		}
	}
	return a
}

func (c *Catalog) ArtworkColor(a Appearance) string {
	return c.artworks[c.Character(a.Character).Variant(a.Variant).ArtworkColor].Color
}

func (c *Catalog) Swatch(a Appearance) color.Color {
	return c.Character(a.Character).Variant(a.Variant).swatch
}

func (c *Catalog) MoodName(a Appearance, mood Mood, language string) string {
	if language == "" {
		language = "pt"
	}
	id := moodIDs[mood]
	local, defaults := c.Character(a.Character).moodNames[id], c.moodNames[id]
	for _, text := range []string{local[language], defaults[language], local["en"], defaults["en"]} {
		if text != "" {
			return text
		}
	}
	return id
}

// CanvasSize reserva a maior extensão do chapéu entre todos os quadros do personagem.
// O espaço não oscila quando o pet respira ou muda de humor.
func (c *Catalog) CanvasSize(a Appearance) (width, height int) {
	ch := c.Character(a.Character)
	height = ch.layout.FrameHeight
	if a.Hat != "" {
		height += ch.headroom
	}
	return ch.layout.FrameWidth, height
}

// Compose monta o quadro original em pixels. As mesmas camadas servem ao perfil,
// à sala e à prévia; nada depende da ordem física das pastas ou da cor do marcador.
func (c *Catalog) Compose(a Appearance, mood Mood, tick int) *image.NRGBA {
	ch := c.Character(a.Character)
	l := ch.layout
	row := slices.Index(l.MoodOrder, moodIDs[mood])
	frame := max(0, tick) * 500 / l.FrameDurationMs[row] % l.FrameCounts[row]
	w, h := c.CanvasSize(a)
	headroom := h - l.FrameHeight
	canvas := image.NewNRGBA(image.Rect(0, 0, w, h))
	v := ch.Variant(a.Variant)
	layers := ch.Composition.WithoutHat
	if a.Hat != "" {
		layers = ch.Composition.WithHat
	}
	for _, layer := range layers {
		var img image.Image
		origin := image.Pt(frame*l.FrameWidth, row*l.FrameHeight)
		target := image.Rect(0, headroom, w, h)
		switch {
		case layer == "hats/selected":
			if a.Hat == "" {
				continue
			}
			img = c.images[c.Hat(a.Hat).Variant(a.HatVariant).Image]
			top := headroom + l.FrameHeight - l.HatY[row][frame] - (c.HatHeight - 1)
			target = image.Rect(0, top, c.HatWidth, top+c.HatHeight)
			origin = image.Point{}
		case layer == "characters/body":
			img = c.images[v.SpriteSheet]
		case strings.HasPrefix(layer, "characters/other_assets/"):
			img = c.images[v.OtherAssets[strings.TrimPrefix(layer, "characters/other_assets/")]]
		case strings.HasPrefix(layer, "main_assets/"):
			img = c.images[c.mainPaths[strings.TrimPrefix(layer, "main_assets/")]]
		}
		draw.Draw(canvas, target, img, origin, draw.Over)
	}
	return canvas
}

type appearanceRenderKey struct {
	Appearance
	Mood                 Mood
	Frame, Scale, Height int
	Pixel                bool
}

// RenderTerminal reaproveita quadros já desenhados. O fallback ASCII também vem
// da composição, para não mostrar uma fruta quando a escolha foi outro personagem.
func (c *Catalog) RenderTerminal(a Appearance, mood Mood, tick, scale, height int, pixel bool) string {
	l := c.Character(a.Character).layout
	row := slices.Index(l.MoodOrder, moodIDs[mood])
	frame := max(0, tick) * 500 / l.FrameDurationMs[row] % l.FrameCounts[row]
	key := appearanceRenderKey{a, mood, frame, scale, height, pixel}
	c.cacheMu.Lock()
	cached, ok := c.renderCache[key]
	c.cacheMu.Unlock()
	if ok {
		return cached
	}
	img := c.Compose(a, mood, tick)
	if height > img.Bounds().Dy() {
		padded := image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), height))
		draw.Draw(padded, image.Rect(0, height-img.Bounds().Dy(), img.Bounds().Dx(), height), img, image.Point{}, draw.Src)
		img = padded
	}
	text := ""
	if pixel {
		text = Pixels(img, 0, 0, img.Bounds().Dx(), img.Bounds().Dy(), scale, nil)
	} else {
		var lines []string
		for y := 0; y < img.Bounds().Dy()*scale; y += 2 {
			var line strings.Builder
			for x := 0; x < img.Bounds().Dx()*scale; x++ {
				top, bot := img.At(x/scale, y/scale), img.At(x/scale, (y+1)/scale)
				glyph, ink := " ", top
				switch {
				case opaque(top) && opaque(bot):
					glyph = "#"
				case opaque(top):
					glyph = "\""
				case opaque(bot):
					glyph, ink = "_", bot
				}
				if glyph == " " {
					line.WriteByte(' ')
				} else {
					line.WriteString(lipgloss.NewStyle().Foreground(ink).Render(glyph))
				}
			}
			lines = append(lines, line.String())
		}
		text = strings.Join(lines, "\n")
	}
	c.cacheMu.Lock()
	// ponytail: até 512 quadros; renovar o cache inteiro mantém a memória limitada.
	// Trocar por LRU apenas se alternar entre muitas aparências causar trabalho excessivo.
	if len(c.renderCache) >= 512 {
		clear(c.renderCache)
	}
	c.renderCache[key] = text
	c.cacheMu.Unlock()
	return text
}
