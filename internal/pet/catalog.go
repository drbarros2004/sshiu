package pet

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"
	"sync"

	"sshiu/assets"
)

// Appearance usa IDs estáveis, independentes dos nomes exibidos e do idioma.
// Hat vazio significa sem chapéu; nesse caso HatVariant também fica vazio.
type Appearance struct {
	Character  string `json:"character"`
	Variant    string `json:"variant"`
	Hat        string `json:"hat,omitempty"`
	HatVariant string `json:"hatVariant,omitempty"`
}

type DisplayName map[string]string

func (n DisplayName) Text(language string) string {
	if language == "" {
		language = "pt"
	}
	if text := n[language]; text != "" {
		return text
	}
	return n["en"]
}

type SpriteLayout struct {
	FrameWidth      int      `json:"frameWidth"`
	FrameHeight     int      `json:"frameHeight"`
	MoodOrder       []string `json:"moodOrder"`
	FrameCounts     []int    `json:"frameCounts"`
	FrameDurationMs []int    `json:"frameDurationMs"`
	HatY            [][]int  `json:"hatY,omitempty"`
}

type CharacterVariant struct {
	ID           string            `json:"id"`
	DisplayName  DisplayName       `json:"displayName"`
	DisplayColor string            `json:"displayColor"`
	ArtworkColor string            `json:"artworkColor"`
	SpriteSheet  string            `json:"spriteSheet"`
	OtherAssets  map[string]string `json:"otherAssets,omitempty"`
	Order        int               `json:"order"`
	swatch       color.Color
}

type Character struct {
	ID           string      `json:"id"`
	DisplayName  DisplayName `json:"displayName"`
	Order        int         `json:"order"`
	SpriteLayout string      `json:"spriteLayout"`
	Moods        string      `json:"moods"`
	Composition  struct {
		WithoutHat []string `json:"withoutHat"`
		WithHat    []string `json:"withHat"`
	} `json:"composition"`
	Variants  []CharacterVariant `json:"variants"`
	layout    SpriteLayout
	moodNames map[string]DisplayName
	headroom  int
}

type HatVariant struct {
	ID          string      `json:"id"`
	DisplayName DisplayName `json:"displayName"`
	Image       string      `json:"image"`
	Order       int         `json:"order"`
}

type Hat struct {
	ID          string       `json:"id"`
	DisplayName DisplayName  `json:"displayName"`
	Order       int          `json:"order"`
	Variants    []HatVariant `json:"variants"`
}

type artworkColor struct {
	ID          string      `json:"id"`
	DisplayName DisplayName `json:"displayName"`
	Order       int         `json:"order"`
	Color       string      `json:"color"`
}

type Catalog struct {
	Characters          []Character
	Hats                []Hat
	HatWidth, HatHeight int
	moodNames           map[string]DisplayName
	artworks            map[string]artworkColor
	images              map[string]image.Image
	mainPaths           map[string]string
	cacheMu             sync.Mutex
	renderCache         map[appearanceRenderKey]string
}

var moodIDs = [...]string{"wilted", "tired", "normal", "good", "radiant", "sleeping"}

// LoadCatalog valida tudo antes de abrir sessões. Um caminho ou grupo incorreto
// não vira uma camada invisível: o erro aponta o catálogo e a referência inválida.
func LoadCatalog(files fs.FS) (*Catalog, error) {
	c := &Catalog{images: map[string]image.Image{}, artworks: map[string]artworkColor{}, renderCache: map[appearanceRenderKey]string{}}
	var file struct {
		ArtworkColors string      `json:"artworkColors"`
		SpriteLayout  string      `json:"spriteLayout"`
		MainAssets    string      `json:"mainAssets"`
		Characters    []Character `json:"characters"`
		Moods         string      `json:"moods"`
	}
	if err := readCatalogJSON(files, "pets/characters.json", &file); err != nil {
		return nil, err
	}
	c.Characters = file.Characters
	if err := ordered(c.Characters, func(v Character) (string, int, DisplayName) { return v.ID, v.Order, v.DisplayName }); err != nil {
		return nil, fmt.Errorf("characters.json: %w", err)
	}
	var defaults SpriteLayout
	if err := readCatalogJSON(files, catalogPath(file.SpriteLayout), &defaults); err != nil {
		return nil, err
	}
	var artworkFile struct {
		Colors []artworkColor `json:"colors"`
	}
	if err := readCatalogJSON(files, catalogPath(file.ArtworkColors), &artworkFile); err != nil {
		return nil, err
	}
	if err := ordered(artworkFile.Colors, func(v artworkColor) (string, int, DisplayName) { return v.ID, v.Order, v.DisplayName }); err != nil {
		return nil, fmt.Errorf("artworks/colors.json: %w", err)
	}
	for _, v := range artworkFile.Colors {
		if _, err := ParsePaletteColor(v.Color); err != nil {
			return nil, fmt.Errorf("obra %s: %w", v.ID, err)
		}
		c.artworks[v.ID] = v
	}
	var err error
	c.moodNames, err = loadMoodNames(files, file.Moods, false)
	if err != nil {
		return nil, err
	}
	var hatFile struct {
		Width  int   `json:"width"`
		Height int   `json:"height"`
		Hats   []Hat `json:"hats"`
	}
	if err := readCatalogJSON(files, "pets/hats.json", &hatFile); err != nil {
		return nil, err
	}
	c.Hats, c.HatWidth, c.HatHeight = hatFile.Hats, hatFile.Width, hatFile.Height
	if c.HatWidth <= 0 || c.HatHeight <= 0 {
		return nil, fmt.Errorf("hats.json: width e height devem ser positivos")
	}
	if err := ordered(c.Hats, func(v Hat) (string, int, DisplayName) { return v.ID, v.Order, v.DisplayName }); err != nil {
		return nil, fmt.Errorf("hats.json: %w", err)
	}
	for _, h := range c.Hats {
		if err := ordered(h.Variants, func(v HatVariant) (string, int, DisplayName) { return v.ID, v.Order, v.DisplayName }); err != nil {
			return nil, fmt.Errorf("chapéu %s: %w", h.ID, err)
		}
		for _, v := range h.Variants {
			if err := c.loadImage(files, v.Image, c.HatWidth, c.HatHeight); err != nil {
				return nil, err
			}
		}
	}
	var mainFile struct {
		Assets []struct {
			ID          string `json:"id"`
			SpriteSheet string `json:"spriteSheet"`
		} `json:"assets"`
	}
	if err := readCatalogJSON(files, catalogPath(file.MainAssets), &mainFile); err != nil {
		return nil, err
	}
	mainPaths := map[string]string{}
	c.mainPaths = mainPaths
	for _, v := range mainFile.Assets {
		if v.ID == "" || mainPaths[v.ID] != "" {
			return nil, fmt.Errorf("main-assets.json: ID vazio ou repetido %q", v.ID)
		}
		mainPaths[v.ID] = v.SpriteSheet
	}
	for i := range c.Characters {
		ch := &c.Characters[i]
		// Decodificar sobre uma cópia herda os campos ausentes; listas locais substituem listas gerais.
		ch.layout = defaults
		ch.layout.MoodOrder = slices.Clone(defaults.MoodOrder)
		ch.layout.FrameCounts = slices.Clone(defaults.FrameCounts)
		ch.layout.FrameDurationMs = slices.Clone(defaults.FrameDurationMs)
		ch.layout.HatY = make([][]int, len(defaults.HatY))
		for row := range defaults.HatY {
			ch.layout.HatY[row] = slices.Clone(defaults.HatY[row])
		}
		if err := readCatalogJSON(files, catalogPath(ch.SpriteLayout), &ch.layout); err != nil {
			return nil, err
		}
		if err := validateLayout(ch.layout); err != nil {
			return nil, fmt.Errorf("%s (%s): %w", ch.SpriteLayout, ch.ID, err)
		}
		if ch.layout.FrameWidth != c.HatWidth {
			return nil, fmt.Errorf("%s: largura do quadro deve corresponder à largura do chapéu", ch.ID)
		}
		ch.moodNames, err = loadMoodNames(files, ch.Moods, true)
		if err != nil {
			return nil, err
		}
		if err := ordered(ch.Variants, func(v CharacterVariant) (string, int, DisplayName) { return v.ID, v.Order, v.DisplayName }); err != nil {
			return nil, fmt.Errorf("personagem %s: %w", ch.ID, err)
		}
		width := slices.Max(ch.layout.FrameCounts) * ch.layout.FrameWidth
		height := len(ch.layout.MoodOrder) * ch.layout.FrameHeight
		for _, anchors := range ch.layout.HatY {
			for _, y := range anchors {
				ch.headroom = max(ch.headroom, y+c.HatHeight-1-ch.layout.FrameHeight)
			}
		}
		for j := range ch.Variants {
			v := &ch.Variants[j]
			if _, ok := c.artworks[v.ArtworkColor]; !ok {
				return nil, fmt.Errorf("%s/%s: artworkColor desconhecido %q", ch.ID, v.ID, v.ArtworkColor)
			}
			v.swatch, err = ParsePaletteColor(v.DisplayColor)
			if err != nil {
				return nil, fmt.Errorf("%s/%s displayColor: %w", ch.ID, v.ID, err)
			}
			for _, layers := range [][]string{ch.Composition.WithoutHat, ch.Composition.WithHat} {
				if len(layers) == 0 {
					return nil, fmt.Errorf("%s: composição vazia", ch.ID)
				}
				seen := map[string]bool{}
				for _, layer := range layers {
					if seen[layer] {
						return nil, fmt.Errorf("%s: camada repetida %s", ch.ID, layer)
					}
					seen[layer] = true
					if layer == "hats/selected" {
						continue
					}
					var name string
					switch {
					case layer == "characters/body":
						name = v.SpriteSheet
					case strings.HasPrefix(layer, "characters/other_assets/"):
						name = v.OtherAssets[strings.TrimPrefix(layer, "characters/other_assets/")]
					case strings.HasPrefix(layer, "main_assets/"):
						name = mainPaths[strings.TrimPrefix(layer, "main_assets/")]
					}
					if name == "" {
						return nil, fmt.Errorf("%s/%s: camada sem caminho %q", ch.ID, v.ID, layer)
					}
					if err := c.loadImage(files, name, width, height); err != nil {
						return nil, fmt.Errorf("%s/%s: %w", ch.ID, v.ID, err)
					}
				}
			}
		}
	}
	return c, nil
}

func catalogPath(name string) string { return path.Join("pets", name) }

func readCatalogJSON(files fs.FS, name string, dst any) error {
	f, err := files.Open(name)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("%s: conteúdo extra após JSON", name)
	}
	return nil
}

func ordered[T any](list []T, fields func(T) (string, int, DisplayName)) error {
	if len(list) == 0 {
		return fmt.Errorf("lista vazia")
	}
	ids, orders := map[string]bool{}, map[int]bool{}
	for _, v := range list {
		id, order, name := fields(v)
		if id == "" || ids[id] || order < 0 || order >= len(list) || orders[order] || name["en"] == "" {
			return fmt.Errorf("ID/order/nome inválido em %q (order %d)", id, order)
		}
		ids[id], orders[order] = true, true
	}
	slices.SortFunc(list, func(a, b T) int { _, x, _ := fields(a); _, y, _ := fields(b); return x - y })
	return nil
}

func loadMoodNames(files fs.FS, name string, overrides bool) (map[string]DisplayName, error) {
	var f struct {
		Moods []struct {
			ID          string      `json:"id"`
			DisplayName DisplayName `json:"displayName"`
		} `json:"moods"`
	}
	if err := readCatalogJSON(files, catalogPath(name), &f); err != nil {
		return nil, err
	}
	names := map[string]DisplayName{}
	for _, m := range f.Moods {
		if !slices.Contains(moodIDs[:], m.ID) || names[m.ID] != nil || (!overrides && m.DisplayName["en"] == "") {
			return nil, fmt.Errorf("%s: humor inválido/repetido %q", name, m.ID)
		}
		names[m.ID] = m.DisplayName
	}
	if !overrides && len(names) != len(moodIDs) {
		return nil, fmt.Errorf("%s: faltam humores padrão", name)
	}
	return names, nil
}

func validateLayout(l SpriteLayout) error {
	if l.FrameWidth <= 0 || l.FrameHeight <= 0 || len(l.MoodOrder) != len(moodIDs) || len(l.FrameCounts) != len(l.MoodOrder) || len(l.FrameDurationMs) != len(l.MoodOrder) || len(l.HatY) != len(l.MoodOrder) {
		return fmt.Errorf("tamanho ou listas do layout inválidos")
	}
	seen := map[string]bool{}
	for row, id := range l.MoodOrder {
		if !slices.Contains(moodIDs[:], id) || seen[id] || l.FrameCounts[row] <= 0 || l.FrameDurationMs[row] < 500 || len(l.HatY[row]) != l.FrameCounts[row] {
			return fmt.Errorf("linha %d (%s): humor, quadros ou hatY inválidos; duração mínima de 500ms por quadro", row+1, id)
		}
		seen[id] = true
		for frame, y := range l.HatY[row] {
			if y < 1 || y > l.FrameHeight {
				return fmt.Errorf("linha %d quadro %d: hatY %d fora de 1..%d", row+1, frame+1, y, l.FrameHeight)
			}
		}
	}
	return nil
}

func (c *Catalog) loadImage(files fs.FS, name string, width, height int) error {
	if name == "" || !fs.ValidPath(name) {
		return fmt.Errorf("caminho de PNG inválido %q", name)
	}
	img := c.images[name]
	if img == nil {
		f, err := files.Open(catalogPath(name))
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		img, err = png.Decode(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		c.images[name] = img
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				_, _, _, a := img.At(x, y).RGBA()
				if a != 0 && a != 65535 {
					return fmt.Errorf("%s (%d,%d): pixel semitransparente", name, x, y)
				}
			}
		}
	}
	if img.Bounds().Dx() != width || img.Bounds().Dy() != height {
		return fmt.Errorf("%s: tamanho %dx%d; esperado %dx%d", name, img.Bounds().Dx(), img.Bounds().Dy(), width, height)
	}
	return nil
}

var Pets = func() *Catalog {
	c, err := LoadCatalog(assets.Pets)
	if err != nil {
		panic(err)
	}
	return c
}()
