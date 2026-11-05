package pet

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// Os PNGs aprovados são a referência, não uma segunda cópia dos valores do JSON.
// Compara cada pixel e os quadros de terminal nos dois tamanhos, em todos os humores.
func TestPalettesMatchDesigns(t *testing.T) {
	template, err := png.Decode(bytes.NewReader(petPNG))
	if err != nil {
		t.Fatal(err)
	}
	for _, palette := range Colors {
		name := palette.Name
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/pet_" + name + ".png")
			if err != nil {
				t.Fatal(err)
			}
			design, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			for y := range template.Bounds().Dy() {
				for x := range template.Bounds().Dx() {
					got := color.NRGBAModel.Convert(swapped(template.At(x, y), palette.parts))
					want := color.NRGBAModel.Convert(design.At(x, y))
					if got != want {
						t.Fatalf("pixel (%d,%d): %v; PNG tem %v", x, y, got, want)
					}
				}
			}
			for mood, count := range moodFrames {
				for frame := range count {
					tick := frame * moodPeriod[mood]
					if Render(Mood(mood), tick, true, palette.ID) != Pixels(design, frame*16, mood*16, 16, 16, 1, nil) {
						t.Errorf("quadro pequeno: humor %d, quadro %d", mood, frame)
					}
					if RenderBig(Mood(mood), tick, true, palette.ID) != Pixels(design, frame*16, mood*16, 16, 16, 2, nil) {
						t.Errorf("quadro grande: humor %d, quadro %d", mood, frame)
					}
				}
			}
		})
	}
}

func TestPaletteValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*paletteFile, *image.NRGBA)
		want string
	}{
		{"unknown marker", func(_ *paletteFile, img *image.NRGBA) { img.Set(0, 0, color.NRGBA{1, 2, 3, 255}) }, "marcador #010203 sem grupo"},
		{"semi-transparent", func(_ *paletteFile, img *image.NRGBA) { img.Set(0, 0, color.NRGBA{1, 2, 3, 128}) }, "sem semitransparência"},
		{"unnamed palette part", func(f *paletteFile, _ *image.NRGBA) { f.Palettes[0].Colors["typo"] = "#FFFFFF" }, `parte "typo" não cadastrada`},
		{"missing colour", func(f *paletteFile, _ *image.NRGBA) { delete(f.Defaults, "body") }, "falta a cor de body"},
		{"unknown default part", func(f *paletteFile, _ *image.NRGBA) { f.Defaults["typo"] = "#FFFFFF" }, `defaults: parte "typo" não cadastrada`},
		{"invalid default", func(f *paletteFile, _ *image.NRGBA) { f.Defaults["body"] = "invalid" }, "defaults, body:"},
		{"new named part", func(f *paletteFile, _ *image.NRGBA) { f.Parts["hat"] = "#010203" }, "falta a cor de hat"},
		{"duplicate marker", func(f *paletteFile, _ *image.NRGBA) { f.Parts["hat"] = f.Parts["body"] }, "mesmo marcador"},
		{"invalid colour", func(f *paletteFile, _ *image.NRGBA) { f.Palettes[0].Colors["body"] = "xterm:256" }, "cor \"xterm:256\" inválida"},
		{"duplicate ID", func(f *paletteFile, _ *image.NRGBA) { f.Palettes[1].ID = f.Palettes[0].ID }, "id único"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var file paletteFile
			if err := json.Unmarshal(palettesJSON, &file); err != nil {
				t.Fatal(err)
			}
			img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
			tc.edit(&file, img)
			data, err := json.Marshal(file)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := readPalettes(data, img); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("erro = %v; esperado conter %q", err, tc.want)
			}
		})
	}
}

func TestPaletteColorFormats(t *testing.T) {
	for value, want := range map[string]uint32{"#AF5F5F": 0xaf5f5f, "#af5f5f": 0xaf5f5f, "xterm:95": 0x875f5f, "xterm:131": 0xaf5f5f, "xterm:255": 0xeeeeee, "xterm:0": 0x000000} {
		c, err := parsePaletteColor(value)
		if err != nil || pixelRGB(c) != want {
			t.Errorf("%q: %v, %v; esperado #%06X", value, c, err, want)
		}
	}
	for _, value := range []string{"", "#ABC", "#GGGGGG", "95", "xterm:-1", "xterm:256", "xterm:abc"} {
		if _, err := parsePaletteColor(value); err == nil {
			t.Errorf("aceitou cor inválida %q", value)
		}
	}
}

// As três aproximações conhecidas vêm dos PNGs; as demais cores permanecem iguais.
func TestPalette256Colors(t *testing.T) {
	conversions := map[uint32]uint32{
		0xffe066: 0xffd75f, // miolo da flor verde
		0x800000: 0x870000, // miolo da flor turquesa
		0xf7f7f7: 0xffffff, // estrelinhas do rosa
	}
	for _, palette := range Colors {
		for marker, c := range palette.parts {
			original := pixelRGB(c)
			want := original
			if converted, exists := conversions[original]; exists {
				want = converted
			}
			if got := pixelRGB(ansi.Convert256(c)); got != want {
				t.Errorf("%s, marcador #%06X: #%06X vira #%06X; esperado #%06X", palette.Name, marker, original, got, want)
			}
		}
	}
}

func TestPaletteJSONErrors(t *testing.T) {
	for _, data := range []string{
		"{",
		`{"parts": {}, "palettes": [], "unknown": true}`,
		string(palettesJSON) + " {}",
	} {
		if _, err := readPalettes([]byte(data), image.NewNRGBA(image.Rect(0, 0, 1, 1))); err == nil {
			t.Fatal("aceitou JSON incompleto, desconhecido ou com conteudo extra")
		}
	}
}

// Um padrão vale para todas as paletas, e uma exceção não altera as demais.
func TestPaletteDefaults(t *testing.T) {
	data := []byte(`{
		"parts": {"body": "#FFFF00", "eyes_mouth": "#FF0000"},
		"defaults": {"body": "#87D7FF", "eyes_mouth": "xterm:235"},
		"palettes": [
			{"name": "base", "id": "base", "colors": {}},
			{"name": "custom", "id": "custom", "colors": {"body": "#FF8787"}},
			{"name": "other", "id": "other", "colors": {"eyes_mouth": "#000000"}}
		]
	}`)
	palettes, err := readPalettes(data, image.NewNRGBA(image.Rect(0, 0, 1, 1)))
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []struct{ body, eyes uint32 }{
		{0x87d7ff, 0x262626},
		{0xff8787, 0x262626},
		{0x87d7ff, 0x000000},
	} {
		if got := pixelRGB(palettes[i].Body); got != want.body {
			t.Errorf("%s: corpo #%06X; esperado #%06X", palettes[i].Name, got, want.body)
		}
		if got := pixelRGB(palettes[i].parts[0xff0000]); got != want.eyes {
			t.Errorf("%s: olhos #%06X; esperado #%06X", palettes[i].Name, got, want.eyes)
		}
	}
}
