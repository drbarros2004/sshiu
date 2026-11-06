package pet

// O template diz a parte de cada pixel. A paleta diz sua cor final, igual em todos os humores.

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

//go:embed palettes.json
var palettesJSON []byte

// Palette mantém o ID já salvo no banco separado das cores usadas no desenho.
// Body também serve à amostra na interface e ao desenho ASCII.
type Palette struct {
	Name, ID string
	Body     color.Color
	parts    map[uint32]color.Color
}

var Colors []Palette // ordem de escolha; azul é o padrão

type paletteDefinition struct {
	Name   string            `json:"name"`
	ID     string            `json:"id"`
	Colors map[string]string `json:"colors"`
}

type paletteFile struct {
	Parts    map[string]string   `json:"parts"`
	Defaults map[string]string   `json:"defaults"`
	Palettes []paletteDefinition `json:"palettes"`
}

// readPalettes valida os arquivos antes de o servidor abrir. Um marcador novo nunca sai
// na tela por acidente: o erro informa a cor e onde cadastrar a parte que falta.
func readPalettes(data []byte, template image.Image) ([]Palette, error) {
	var file paletteFile
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("palettes.json: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("palettes.json: conteúdo extra depois do objeto principal")
	}
	if len(file.Palettes) == 0 || file.Parts["body"] == "" {
		return nil, fmt.Errorf("palettes.json: defina a parte body e pelo menos uma paleta")
	}
	markers := map[uint32]string{}
	for name, value := range file.Parts {
		if name == "" || len(value) != 7 || !strings.HasPrefix(value, "#") {
			return nil, fmt.Errorf("parts.%s: o marcador deve ser #RRGGBB", name)
		}
		c, err := ParsePaletteColor(value)
		if err != nil {
			return nil, fmt.Errorf("parts.%s: %w", name, err)
		}
		rgb := pixelRGB(c)
		if other, exists := markers[rgb]; exists {
			return nil, fmt.Errorf("parts: %s e %s usam o mesmo marcador %s", name, other, value)
		}
		markers[rgb] = name
	}
	for y := template.Bounds().Min.Y; y < template.Bounds().Max.Y; y++ {
		for x := template.Bounds().Min.X; x < template.Bounds().Max.X; x++ {
			c := template.At(x, y)
			_, _, _, a := c.RGBA()
			if a == 0 {
				continue
			}
			if a != 0xffff {
				return nil, fmt.Errorf("pet.png (%d,%d): use pixel opaco ou transparente, sem semitransparência", x, y)
			}
			if _, exists := markers[pixelRGB(c)]; !exists {
				return nil, fmt.Errorf("pet.png (%d,%d): marcador #%06X sem grupo; cadastre em palettes.json → parts", x, y, pixelRGB(c))
			}
		}
	}
	// Cada parte tem um padrão explícito; as paletas listam apenas suas diferenças.
	for name := range file.Defaults {
		if _, exists := file.Parts[name]; !exists {
			return nil, fmt.Errorf("defaults: parte %q não cadastrada em parts", name)
		}
	}
	for _, name := range markers {
		value, exists := file.Defaults[name]
		if !exists {
			return nil, fmt.Errorf("defaults: falta a cor de %s", name)
		}
		if _, err := ParsePaletteColor(value); err != nil {
			return nil, fmt.Errorf("defaults, %s: %w", name, err)
		}
	}
	var palettes []Palette
	ids := map[string]bool{}
	for _, def := range file.Palettes {
		if def.Name == "" || def.ID == "" || ids[def.ID] {
			return nil, fmt.Errorf("palettes.json: cada paleta precisa de name e id único (recebido %q)", def.ID)
		}
		ids[def.ID] = true
		for name := range def.Colors {
			if _, exists := file.Parts[name]; !exists {
				return nil, fmt.Errorf("paleta %s: parte %q não cadastrada em parts", def.Name, name)
			}
		}
		palette := Palette{Name: def.Name, ID: def.ID, parts: map[uint32]color.Color{}}
		for marker, name := range markers {
			value := file.Defaults[name]
			if override, exists := def.Colors[name]; exists {
				value = override
			}
			c, err := ParsePaletteColor(value)
			if err != nil {
				return nil, fmt.Errorf("paleta %s, %s: %w", def.Name, name, err)
			}
			palette.parts[marker] = c
			if name == "body" {
				palette.Body = c
			}
		}
		palettes = append(palettes, palette)
	}
	return palettes, nil
}

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
