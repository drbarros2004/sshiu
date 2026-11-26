package pet

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// Dados sintéticos: as entradas e expectativas não usam os personagens reais.
// As folhas têm a estrutura dos seis humores, mas só a linha de normal tem pixels.
// Os testes chamam Compose diretamente, com as buscas, CanvasSize e draw.Draw reais.
const (
	composeTestWidth   = 4
	composeTestHeight  = 4
	composeTestMoodRow = 2
)

// Cada caractere representa um pixel: R vermelho, G verde, B azul, Y amarelo,
// C ciano, M magenta, W branco e . transparente (alpha zero).
// Sempre escreva os pontos, inclusive nas bordas; não use espaços nem omita linhas.
// A quantidade de caracteres define a largura; a quantidade de linhas, a altura.
// As entradas usam dois quadros 4x4 lado a lado; as saídas são um quadro 4x4
// sem chapéu ou 4x5 com chapéu, pois o catálogo reserva uma linha acima do corpo.
var composeTestColors = map[byte]color.NRGBA{
	'.': {},
	'R': {R: 255, A: 255},
	'G': {G: 255, A: 255},
	'B': {B: 255, A: 255},
	'Y': {R: 255, G: 255, A: 255},
	'C': {G: 255, B: 255, A: 255},
	'M': {R: 255, B: 255, A: 255},
	'W': {R: 255, G: 255, B: 255, A: 255},
}

// Converte uma grade escrita manualmente em pixels, sem chamar Compose.
func composeTestImage(t *testing.T, rows []string) *image.NRGBA {
	t.Helper()
	if len(rows) == 0 || len(rows[0]) == 0 {
		t.Fatal("a grade precisa ter pelo menos um pixel")
	}
	img := image.NewNRGBA(image.Rect(0, 0, len(rows[0]), len(rows)))
	for y, row := range rows {
		if len(row) != img.Bounds().Dx() {
			t.Fatalf("linha %d: largura %d, esperada %d; use pontos para pixels transparentes", y, len(row), img.Bounds().Dx())
		}
		for x := 0; x < len(row); x++ {
			ink, exists := composeTestColors[row[x]]
			if !exists {
				t.Fatalf("pixel (%d,%d): símbolo desconhecido %q", x, y, row[x])
			}
			img.SetNRGBA(x, y, ink)
		}
	}
	return img
}

// A grade contém apenas a linha de normal: quadro 0 à esquerda, quadro 1 à direita.
// As demais linhas de humores ficam transparentes na folha criada na memória.
func composeTestSheet(t *testing.T, rows []string) *image.NRGBA {
	t.Helper()
	frames := composeTestImage(t, rows)
	if frames.Bounds().Dx() != 2*composeTestWidth || frames.Bounds().Dy() != composeTestHeight {
		t.Fatal("a entrada precisa ter dois quadros 4x4 lado a lado: grade 8x4")
	}
	sheet := image.NewNRGBA(image.Rect(0, 0, 2*composeTestWidth, 6*composeTestHeight))
	for y := 0; y < frames.Bounds().Dy(); y++ {
		for x := 0; x < frames.Bounds().Dx(); x++ {
			sheet.SetNRGBA(x, composeTestMoodRow*composeTestHeight+y, frames.NRGBAAt(x, y))
		}
	}
	return sheet
}

func newComposeTestCatalog(t *testing.T) (*Catalog, Appearance) {
	t.Helper()
	character := Character{
		ID: "test_character",
		Variants: []CharacterVariant{{
			ID:          "test_color",
			SpriteSheet: "body",
			OtherAssets: map[string]string{"overlay": "overlay"},
		}},
		layout: SpriteLayout{
			FrameWidth:      composeTestWidth,
			FrameHeight:     composeTestHeight,
			MoodOrder:       []string{"wilted", "tired", "normal", "good", "radiant", "sleeping"},
			FrameCounts:     []int{2, 2, 2, 2, 2, 2},
			FrameDurationMs: []int{1000, 1000, 1000, 1000, 1000, 1000},
			HatY:            [][]int{{3, 4}, {3, 4}, {3, 4}, {3, 4}, {3, 4}, {3, 4}},
		},
		// Com âncora 4 e chapéu de altura 2, é necessária uma linha acima do corpo.
		// Este campo normalmente é preparado por LoadCatalog, não por Compose.
		headroom: 1,
	}
	character.Composition.WithoutHat = []string{"characters/body"}
	character.Composition.WithHat = []string{"characters/body", "hats/selected"}
	catalog := &Catalog{
		renderCache: map[appearanceRenderKey]string{},
		Characters:  []Character{character},
		Hats: []Hat{{
			ID:       "test_hat",
			Variants: []HatVariant{{ID: "test_model", Image: "hat"}},
		}},
		HatWidth:  composeTestWidth,
		HatHeight: 2,
		mainPaths: map[string]string{"particles": "particles"},
		images: map[string]image.Image{
			"body": composeTestSheet(t, []string{
				"RRRRGGGG",
				"RRRRGGGG",
				"RRRRGGGG",
				"RRRRGGGG",
			}),
			"overlay": composeTestSheet(t, []string{
				"YYYYCCCC",
				"YYYYCCCC",
				"YYYYCCCC",
				"YYYYCCCC",
			}),
			"particles": composeTestSheet(t, []string{
				"MMMMWWWW",
				"MMMMWWWW",
				"MMMMWWWW",
				"MMMMWWWW",
			}),
			"hat": composeTestImage(t, []string{
				"BBBB",
				"BBBB",
			}),
		},
	}
	return catalog, Appearance{Character: "test_character", Variant: "test_color"}
}

// As expectativas são grades independentes da implementação: cada letra é um pixel.
// Também confere os limites e a transparência, sem depender da conversão ao terminal.
func assertComposeImage(t *testing.T, got *image.NRGBA, want []string) {
	t.Helper()
	if got == nil {
		t.Fatal("Compose devolveu uma imagem nula")
	}
	expected := composeTestImage(t, want)
	wantBounds := expected.Bounds()
	if got.Bounds() != wantBounds {
		t.Fatalf("limites: recebido %v, esperado %v", got.Bounds(), wantBounds)
	}
	for y, row := range want {
		for x := 0; x < len(row); x++ {
			wantColor := expected.NRGBAAt(x, y)
			if pixel := got.NRGBAAt(x, y); pixel != wantColor {
				t.Errorf("pixel (%d,%d): recebido RGBA %v, esperado %c = %v", x, y, pixel, row[x], wantColor)
			}
		}
	}
}

// Cada caso identifica uma origem de imagem diferente no switch de Compose.
func TestComposeIndividualLayers(t *testing.T) {
	for _, tc := range []struct {
		name, layer string
		hat         bool
		want        []string
	}{
		{"body", "characters/body", false, []string{
			"RRRR",
			"RRRR",
			"RRRR",
			"RRRR",
		}},
		{"hat", "hats/selected", true, []string{
			"....",
			"BBBB",
			"BBBB",
			"....",
			"....",
		}},
		{"character_asset", "characters/other_assets/overlay", false, []string{
			"YYYY",
			"YYYY",
			"YYYY",
			"YYYY",
		}},
		{"main_asset", "main_assets/particles", false, []string{
			"MMMM",
			"MMMM",
			"MMMM",
			"MMMM",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog, appearance := newComposeTestCatalog(t)
			if tc.hat {
				appearance.Hat, appearance.HatVariant = "test_hat", "test_model"
				catalog.Characters[0].Composition.WithHat = []string{tc.layer}
			} else {
				catalog.Characters[0].Composition.WithoutHat = []string{tc.layer}
			}
			assertComposeImage(t, catalog.Compose(appearance, Okay, 0), tc.want)
		})
	}
}

func TestComposeHatSelection(t *testing.T) {
	for _, tc := range []struct {
		name           string
		hat, hatInList bool
		want           []string
	}{
		{"without_hat_composition", false, false, []string{
			"YYYY",
			"YYYY",
			"YYYY",
			"YYYY",
		}},
		{"with_hat_composition", true, false, []string{
			"....",
			"BBBB",
			"BBBB",
			"RRRR",
			"RRRR",
		}},
		{"skip_unselected_hat", false, true, []string{
			"RRRR",
			"RRRR",
			"RRRR",
			"RRRR",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog, appearance := newComposeTestCatalog(t)
			catalog.Characters[0].Composition.WithoutHat = []string{"characters/body", "characters/other_assets/overlay"}
			if tc.hat {
				appearance.Hat, appearance.HatVariant = "test_hat", "test_model"
			}
			if tc.hatInList {
				// Exercita a proteção interna: uma referência ao chapéu não basta para selecioná-lo.
				catalog.Characters[0].Composition.WithoutHat = []string{"characters/body", "hats/selected"}
			}
			assertComposeImage(t, catalog.Compose(appearance, Okay, 0), tc.want)
		})
	}
}

func TestComposeLayerOrder(t *testing.T) {
	for _, tc := range []struct {
		name   string
		layers []string
		want   []string
	}{
		{"hat_over_body", []string{"characters/body", "hats/selected"}, []string{
			"....",
			"BBBB",
			"BBBB",
			"RRRR",
			"RRRR",
		}},
		{"body_over_hat", []string{"hats/selected", "characters/body"}, []string{
			"....",
			"RRRR",
			"RRRR",
			"RRRR",
			"RRRR",
		}},
		{"four_layers_reversed_assets", []string{"characters/body", "hats/selected", "main_assets/particles", "characters/other_assets/overlay"}, []string{
			"....", "BYMB", "BBBB", "YRRR", "RRRM",
		}},
		{"four_layers", []string{"characters/body", "hats/selected", "characters/other_assets/overlay", "main_assets/particles"}, []string{
			"....",
			"BMMB",
			"BBBB",
			"YRRR",
			"RRRM",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog, appearance := newComposeTestCatalog(t)
			appearance.Hat, appearance.HatVariant = "test_hat", "test_model"
			catalog.Characters[0].Composition.WithHat = tc.layers
			if strings.HasPrefix(tc.name, "four_layers") {
				// Camadas esparsas: todas contribuem para a imagem e algumas disputam pixels.
				catalog.images["overlay"] = composeTestSheet(t, []string{
					".Y......",
					"........",
					"Y.......",
					"........",
				})
				catalog.images["particles"] = composeTestSheet(t, []string{
					".MM.....",
					"........",
					"........",
					"...M....",
				})
			}
			assertComposeImage(t, catalog.Compose(appearance, Okay, 0), tc.want)
		})
	}
}

func TestComposeTransparency(t *testing.T) {
	t.Run("character_asset_keeps_body_visible", func(t *testing.T) {
		catalog, appearance := newComposeTestCatalog(t)
		catalog.Characters[0].Composition.WithoutHat = []string{"characters/body", "characters/other_assets/overlay"}
		catalog.images["overlay"] = composeTestSheet(t, []string{
			"Y.......",
			"........",
			"........",
			"...Y....",
		})
		assertComposeImage(t, catalog.Compose(appearance, Okay, 0), []string{
			"YRRR",
			"RRRR",
			"RRRR",
			"RRRY",
		})
	})
	t.Run("hat_keeps_body_visible_through_hole", func(t *testing.T) {
		catalog, appearance := newComposeTestCatalog(t)
		appearance.Hat, appearance.HatVariant = "test_hat", "test_model"
		catalog.images["hat"] = composeTestImage(t, []string{
			"B.BB",
			"BBBB",
		})
		assertComposeImage(t, catalog.Compose(appearance, Okay, 0), []string{
			"....",
			"BRBB",
			"BBBB",
			"RRRR",
			"RRRR",
		})
	})
}

func TestComposeFrameTiming(t *testing.T) {
	for _, tc := range []struct {
		name           string
		tick, duration int
		want           []string
	}{
		{"negative_tick_uses_first_frame", -1, 1000, []string{
			"RRRR",
			"RRRR",
			"RRRR",
			"RRRR",
		}},
		{"first_frame", 0, 1000, []string{
			"RRRR",
			"RRRR",
			"RRRR",
			"RRRR",
		}},
		{"before_frame_change", 1, 1000, []string{
			"RRRR",
			"RRRR",
			"RRRR",
			"RRRR",
		}},
		{"at_frame_change", 2, 1000, []string{
			"GGGG",
			"GGGG",
			"GGGG",
			"GGGG",
		}},
		{"before_cycle_restart", 3, 1000, []string{
			"GGGG",
			"GGGG",
			"GGGG",
			"GGGG",
		}},
		{"at_cycle_restart", 4, 1000, []string{
			"RRRR",
			"RRRR",
			"RRRR",
			"RRRR",
		}},
		{"shorter_duration", 1, 500, []string{
			"GGGG",
			"GGGG",
			"GGGG",
			"GGGG",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog, appearance := newComposeTestCatalog(t)
			catalog.Characters[0].layout.FrameDurationMs[composeTestMoodRow] = tc.duration
			assertComposeImage(t, catalog.Compose(appearance, Okay, tc.tick), tc.want)
		})
	}
}

func TestComposeFrameHatPosition(t *testing.T) {
	for _, tc := range []struct {
		name string
		tick int
		want []string
	}{
		{"first_frame_anchor", 0, []string{
			"....",
			"BBBB",
			"BBBB",
			"RRRR",
			"RRRR",
		}},
		{"second_frame_anchor", 2, []string{
			"BBBB",
			"BBBB",
			"GGGG",
			"GGGG",
			"GGGG",
		}},
		{"anchor_after_cycle_restart", 4, []string{
			"....",
			"BBBB",
			"BBBB",
			"RRRR",
			"RRRR",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog, appearance := newComposeTestCatalog(t)
			appearance.Hat, appearance.HatVariant = "test_hat", "test_model"
			assertComposeImage(t, catalog.Compose(appearance, Okay, tc.tick), tc.want)
		})
	}
}

// Cada caso muda somente um ID da aparência, mantendo humor e tick iguais.
func TestComposeAppearanceSelection(t *testing.T) {
	for _, field := range []string{"character", "variant", "hat", "hat_variant"} {
		t.Run(field, func(t *testing.T) {
			catalog, appearance := newComposeTestCatalog(t)
			catalog.images["alternative_body"] = composeTestSheet(t, []string{
				"GGGGGGGG", "GGGGGGGG", "GGGGGGGG", "GGGGGGGG",
			})
			catalog.images["alternative_hat"] = composeTestImage(t, []string{"YYYY", "YYYY"})
			want := []string{"GGGG", "GGGG", "GGGG", "GGGG"}
			switch field {
			case "character":
				other := catalog.Characters[0]
				other.ID = "other_character"
				other.Variants = []CharacterVariant{{ID: "test_color", SpriteSheet: "alternative_body"}}
				catalog.Characters = append(catalog.Characters, other)
				appearance.Character = other.ID
			case "variant":
				catalog.Characters[0].Variants = append(catalog.Characters[0].Variants,
					CharacterVariant{ID: "other_color", SpriteSheet: "alternative_body"})
				appearance.Variant = "other_color"
			case "hat":
				catalog.Hats = append(catalog.Hats, Hat{ID: "other_hat", Variants: []HatVariant{
					{ID: "test_model", Image: "alternative_hat"},
				}})
				appearance.Hat, appearance.HatVariant = "other_hat", "test_model"
			case "hat_variant":
				catalog.Hats[0].Variants = append(catalog.Hats[0].Variants,
					HatVariant{ID: "other_model", Image: "alternative_hat"})
				appearance.Hat, appearance.HatVariant = "test_hat", "other_model"
			}
			if appearance.Hat != "" {
				want = []string{"....", "YYYY", "YYYY", "RRRR", "RRRR"}
			}
			assertComposeImage(t, catalog.Compose(appearance, Okay, 0), want)
		})
	}
}

func TestComposeMoodSelection(t *testing.T) {
	for _, tc := range []struct {
		name      string
		mood      Mood
		reordered bool
		want      []string
	}{
		{"first_mood", Wilted, false, []string{"YYYY", "YYYY", "YYYY", "YYYY"}},
		{"last_mood", Sleeping, false, []string{"CCCC", "CCCC", "CCCC", "CCCC"}},
		{"mood_id_after_reordering", Wilted, true, []string{"CCCC", "CCCC", "CCCC", "CCCC"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog, appearance := newComposeTestCatalog(t)
			// Seis linhas de humores, cada uma com dois quadros; extremos distintos.
			catalog.images["body"] = composeTestImage(t, []string{
				"YYYYYYYY", "YYYYYYYY", "YYYYYYYY", "YYYYYYYY",
				"........", "........", "........", "........",
				"RRRRGGGG", "RRRRGGGG", "RRRRGGGG", "RRRRGGGG",
				"........", "........", "........", "........",
				"........", "........", "........", "........",
				"CCCCCCCC", "CCCCCCCC", "CCCCCCCC", "CCCCCCCC",
			})
			if tc.reordered {
				order := catalog.Characters[0].layout.MoodOrder
				order[0], order[5] = order[5], order[0]
			}
			assertComposeImage(t, catalog.Compose(appearance, tc.mood, 0), tc.want)
		})
	}
}

// No segundo quadro, corpo e ambos os assets devem avançar juntos.
func TestComposeFrameAssets(t *testing.T) {
	catalog, appearance := newComposeTestCatalog(t)
	catalog.Characters[0].Composition.WithoutHat = []string{
		"characters/body", "characters/other_assets/overlay", "main_assets/particles",
	}
	catalog.images["overlay"] = composeTestSheet(t, []string{
		"Y...C...", "........", "........", "........",
	})
	catalog.images["particles"] = composeTestSheet(t, []string{
		"........", "........", "........", "...M...W",
	})
	assertComposeImage(t, catalog.Compose(appearance, Okay, 2), []string{
		"CGGG", "GGGG", "GGGG", "GGGW",
	})
}

func TestLook(t *testing.T) {
	// Look consulta o catálogo global para perfis antigos. Restaurá-lo mantém
	// este teste isolado; não usar t.Parallel enquanto esse catálogo for trocado.
	original := Pets
	Pets = &Catalog{
		Characters: []Character{{ID: "fruit", Variants: []CharacterVariant{
			{ID: "default", ArtworkColor: "blue"}, {ID: "coral", ArtworkColor: "coral"},
		}}},
		artworks: map[string]artworkColor{"blue": {Color: "#0000FF"}, "coral": {Color: "#FF8787"}},
	}
	t.Cleanup(func() { Pets = original })
	for _, tc := range []struct {
		name string
		pet  Pet
		want Appearance
	}{
		{"saved_appearance", Pet{Appearance: Appearance{"bunny", "purple", "crown", "gold"}, Color: "#FF8787"}, Appearance{"bunny", "purple", "crown", "gold"}},
		{"legacy_color", Pet{Color: "#FF8787"}, Appearance{Character: "fruit", Variant: "coral"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pet.Look(); got != tc.want {
				t.Fatalf("aparência recebida %+v, esperada %+v", got, tc.want)
			}
		})
	}
}

func TestCatalogCharacter(t *testing.T) {
	catalog := &Catalog{
		Characters: []Character{{ID: "first"}, {ID: "target", Variants: []CharacterVariant{{ID: "first"}, {ID: "target"}}}},
		Hats:       []Hat{{ID: "first"}, {ID: "target", Variants: []HatVariant{{ID: "first"}, {ID: "target"}}}},
	}
	// O ponteiro deve apontar para a entrada certa, não para uma cópia ou a primeira.
	if catalog.Character("target") != &catalog.Characters[1] || catalog.Character("missing") != nil {
		t.Fatal("busca de personagem incorreta")
	}
}

func TestCharacterVariant(t *testing.T) {
	catalog := &Catalog{
		Characters: []Character{{ID: "first"}, {ID: "target", Variants: []CharacterVariant{{ID: "first"}, {ID: "target"}}}},
		Hats:       []Hat{{ID: "first"}, {ID: "target", Variants: []HatVariant{{ID: "first"}, {ID: "target"}}}},
	}
	// O ponteiro deve apontar para a entrada certa, não para uma cópia ou a primeira.
	character := &catalog.Characters[1]
	if character.Variant("target") != &character.Variants[1] || character.Variant("missing") != nil {
		t.Fatal("busca de variante do personagem incorreta")
	}
}

func TestCatalogHat(t *testing.T) {
	catalog := &Catalog{
		Characters: []Character{{ID: "first"}, {ID: "target", Variants: []CharacterVariant{{ID: "first"}, {ID: "target"}}}},
		Hats:       []Hat{{ID: "first"}, {ID: "target", Variants: []HatVariant{{ID: "first"}, {ID: "target"}}}},
	}
	// O ponteiro deve apontar para a entrada certa, não para uma cópia ou a primeira.
	if catalog.Hat("target") != &catalog.Hats[1] || catalog.Hat("missing") != nil {
		t.Fatal("busca de chapéu incorreta")
	}
}

func TestHatVariant(t *testing.T) {
	catalog := &Catalog{
		Characters: []Character{{ID: "first"}, {ID: "target", Variants: []CharacterVariant{{ID: "first"}, {ID: "target"}}}},
		Hats:       []Hat{{ID: "first"}, {ID: "target", Variants: []HatVariant{{ID: "first"}, {ID: "target"}}}},
	}
	// O ponteiro deve apontar para a entrada certa, não para uma cópia ou a primeira.
	hat := &catalog.Hats[1]
	if hat.Variant("target") != &hat.Variants[1] || hat.Variant("missing") != nil {
		t.Fatal("busca de modelo do chapéu incorreta")
	}
}

func TestValidateAppearance(t *testing.T) {
	catalog, base := newComposeTestCatalog(t)
	for _, tc := range []struct {
		name       string
		appearance Appearance
		invalid    bool
	}{
		{"without_hat", base, false},
		{"with_hat", Appearance{"test_character", "test_color", "test_hat", "test_model"}, false},
		{"unknown_character", Appearance{Character: "missing", Variant: "test_color"}, true},
		{"unknown_variant", Appearance{Character: "test_character", Variant: "missing"}, true},
		{"model_without_hat", Appearance{"test_character", "test_color", "", "test_model"}, true},
		{"unknown_hat", Appearance{"test_character", "test_color", "missing", "test_model"}, true},
		{"unknown_hat_model", Appearance{"test_character", "test_color", "test_hat", "missing"}, true},
		{"missing_hat_model", Appearance{"test_character", "test_color", "test_hat", ""}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := catalog.ValidateAppearance(tc.appearance)
			if (err != nil) != tc.invalid {
				t.Fatalf("erro recebido %v; deveria rejeitar: %v", err, tc.invalid)
			}
		})
	}
}

func TestLegacyAppearance(t *testing.T) {
	catalog := &Catalog{
		Characters: []Character{{ID: "fruit", Variants: []CharacterVariant{
			{ID: "default", ArtworkColor: "blue"}, {ID: "coral", ArtworkColor: "pink"},
		}}},
		artworks: map[string]artworkColor{"blue": {Color: "#0000FF"}, "pink": {Color: "#FF8787"}},
	}
	for _, tc := range []struct{ name, value, variant string }{
		{"color_case_insensitive", "#ff8787", "coral"},
		{"variant_id", "coral", "coral"},
		{"unknown_color", "#123456", "default"},
		{"empty_color", "", "default"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := Appearance{Character: "fruit", Variant: tc.variant}
			if got := catalog.LegacyAppearance(tc.value); got != want {
				t.Fatalf("aparência recebida %+v, esperada %+v", got, want)
			}
		})
	}
}

func TestArtworkColor(t *testing.T) {
	catalog := &Catalog{
		Characters: []Character{
			{ID: "first", Variants: []CharacterVariant{{ID: "first", ArtworkColor: "blue", swatch: composeTestColors['B']}}},
			{ID: "target", Variants: []CharacterVariant{
				{ID: "first", ArtworkColor: "blue", swatch: composeTestColors['B']},
				{ID: "target", ArtworkColor: "red", swatch: composeTestColors['R']},
			}},
		},
		artworks: map[string]artworkColor{"blue": {Color: "#0000FF"}, "red": {Color: "#FF0000"}},
	}
	appearance := Appearance{Character: "target", Variant: "target"}
	if got := catalog.ArtworkColor(appearance); got != "#FF0000" {
		t.Fatalf("cor da obra recebida %q", got)
	}
}

func TestSwatch(t *testing.T) {
	catalog := &Catalog{
		Characters: []Character{
			{ID: "first", Variants: []CharacterVariant{{ID: "first", ArtworkColor: "blue", swatch: composeTestColors['B']}}},
			{ID: "target", Variants: []CharacterVariant{
				{ID: "first", ArtworkColor: "blue", swatch: composeTestColors['B']},
				{ID: "target", ArtworkColor: "red", swatch: composeTestColors['R']},
			}},
		},
		artworks: map[string]artworkColor{"blue": {Color: "#0000FF"}, "red": {Color: "#FF0000"}},
	}
	appearance := Appearance{Character: "target", Variant: "target"}
	if got := color.NRGBAModel.Convert(catalog.Swatch(appearance)); got != composeTestColors['R'] {
		t.Fatalf("amostra recebida %v", got)
	}
}

func TestMoodName(t *testing.T) {
	for _, tc := range []struct {
		name, language  string
		local, defaults DisplayName
		want            string
	}{
		{"character_translation_first", "pt", DisplayName{"pt": "local", "en": "local en"}, DisplayName{"pt": "geral", "en": "default en"}, "local"},
		{"default_translation_before_local_english", "pt", DisplayName{"en": "local en"}, DisplayName{"pt": "geral", "en": "default en"}, "geral"},
		{"local_english_fallback", "fr", DisplayName{"en": "local en"}, DisplayName{"en": "default en"}, "local en"},
		{"default_english_fallback", "fr", nil, DisplayName{"en": "default en"}, "default en"},
		{"id_when_no_translation", "pt", nil, nil, "normal"},
		{"empty_language_is_portuguese", "", DisplayName{"pt": "português", "en": "english"}, nil, "português"},
		{"explicit_english", "en", DisplayName{"pt": "português", "en": "english"}, nil, "english"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog := &Catalog{
				Characters: []Character{{ID: "test", moodNames: map[string]DisplayName{"normal": tc.local}}},
				moodNames:  map[string]DisplayName{"normal": tc.defaults},
			}
			if got := catalog.MoodName(Appearance{Character: "test"}, Okay, tc.language); got != tc.want {
				t.Fatalf("nome recebido %q, esperado %q", got, tc.want)
			}
		})
	}
}

func TestCanvasSize(t *testing.T) {
	for _, tc := range []struct {
		name             string
		hat              string
		headroom, height int
	}{
		{"without_hat_ignores_headroom", "", 3, 4},
		{"with_hat_reserves_headroom", "test_hat", 3, 7},
		{"hat_fits_inside_body", "test_hat", 0, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog, appearance := newComposeTestCatalog(t)
			appearance.Hat = tc.hat
			if tc.hat != "" {
				appearance.HatVariant = "test_model"
			}
			catalog.Characters[0].headroom = tc.headroom
			width, height := catalog.CanvasSize(appearance)
			if width != 4 || height != tc.height {
				t.Fatalf("tamanho recebido %dx%d, esperado 4x%d", width, height, tc.height)
			}
		})
	}
}

func TestRenderTerminal(t *testing.T) {
	for _, tc := range []struct {
		name          string
		pixel         bool
		scale, height int
		rows          []string
		want          string
	}{
		{"ascii_transparency", false, 1, 0, []string{"RR..RR..", "R.R.R.R.", "........", "........"}, "#\"_ \n    "},
		{"pixel_mode", true, 1, 0, []string{"RR..RR..", "R.R.R.R.", "........", "........"}, "▄▄▄ \n    "},
		{"padding_above_body", false, 1, 6, []string{"RR..RR..", "R.R.R.R.", "........", "........"}, "    \n#\"_ \n    "},
		{"smaller_height_does_not_crop", false, 1, 2, []string{"RR..RR..", "R.R.R.R.", "........", "........"}, "#\"_ \n    "},
		{"double_scale", false, 2, 0, []string{"RRRRRRRR", "RRRRRRRR", "RRRRRRRR", "RRRRRRRR"}, "########\n########\n########\n########"},
		{"odd_height", false, 1, 5, []string{"RRRRRRRR", "RRRRRRRR", "RRRRRRRR", "RRRRRRRR"}, "____\n####\n\"\"\"\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog, appearance := newComposeTestCatalog(t)
			catalog.images["body"] = composeTestSheet(t, tc.rows)
			got := ansi.Strip(catalog.RenderTerminal(appearance, Okay, 0, tc.scale, tc.height, tc.pixel))
			if got != tc.want {
				t.Fatalf("texto recebido %q, esperado %q", got, tc.want)
			}
		})
	}
}

func TestRenderTerminalCache(t *testing.T) {
	t.Run("same_frame_reuses_cached_text", func(t *testing.T) {
		catalog, appearance := newComposeTestCatalog(t)
		first := catalog.RenderTerminal(appearance, Okay, 0, 1, 0, false)
		// Retirar a imagem após renderizar prova que não há uma nova composição.
		delete(catalog.images, "body")
		if got := catalog.RenderTerminal(appearance, Okay, 1, 1, 0, false); got != first || len(catalog.renderCache) != 1 {
			t.Fatal("o mesmo quadro deveria reutilizar o texto já renderizado")
		}
	})
	t.Run("different_render_options_have_separate_entries", func(t *testing.T) {
		catalog, appearance := newComposeTestCatalog(t)
		other := catalog.Characters[0]
		other.ID = "other"
		catalog.Characters = append(catalog.Characters, other)
		otherAppearance := appearance
		otherAppearance.Character = "other"
		for index, tc := range []struct {
			appearance          Appearance
			mood                Mood
			tick, scale, height int
			pixel               bool
		}{
			{appearance, Okay, 0, 1, 0, false},
			{appearance, Okay, 2, 1, 0, false},
			{appearance, Wilted, 0, 1, 0, false},
			{appearance, Okay, 0, 2, 0, false},
			{appearance, Okay, 0, 1, 6, false},
			{appearance, Okay, 0, 1, 0, true},
			{otherAppearance, Okay, 0, 1, 0, false},
		} {
			catalog.RenderTerminal(tc.appearance, tc.mood, tc.tick, tc.scale, tc.height, tc.pixel)
			if len(catalog.renderCache) != index+1 {
				t.Fatalf("opção %d não criou uma entrada independente", index)
			}
		}
	})
	t.Run("cache_limit_discards_old_entries", func(t *testing.T) {
		catalog, appearance := newComposeTestCatalog(t)
		// Chaves artificiais evitam renderizar 512 imagens apenas para encher o mapa.
		for frame := 10; frame < 522; frame++ {
			catalog.renderCache[appearanceRenderKey{Appearance: appearance, Frame: frame}] = "old"
		}
		got := catalog.RenderTerminal(appearance, Okay, 0, 1, 0, false)
		if len(catalog.renderCache) != 1 || ansi.Strip(got) != "####\n####" {
			t.Fatal("o cache cheio deveria ser renovado e guardar o novo quadro")
		}
	})
}
