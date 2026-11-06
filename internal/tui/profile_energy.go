package tui

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sshiu/internal/pet"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

//go:embed energy_colors.json
var energyColorsJSON []byte

// Uma cor de energia aceita a mesma escrita das paletas e usa a tabela xterm.
type energyColor ansi.IndexedColor

func (c *energyColor) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	parsed, err := pet.ParsePaletteColor(value)
	if err != nil {
		return err
	}
	*c = energyColor(ansi.Convert256(parsed))
	return nil
}

type energyPalette struct {
	Bar      map[string]*energyColor `json:"bar"`
	Unfilled *energyColor            `json:"unfilled"`
	Markers  map[string]*energyColor `json:"markers"`
}

var energyColors = loadEnergyColors()

func loadEnergyColors() energyPalette {
	var colors energyPalette
	decoder := json.NewDecoder(bytes.NewReader(energyColorsJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&colors); err != nil {
		panic(fmt.Errorf("energy_colors.json: %w", err))
	}
	if decoder.Decode(new(any)) != io.EOF {
		panic("energy_colors.json: conteúdo extra depois do objeto principal")
	}
	moods := pet.MoodName[:pet.Radiant+1]
	for name := range colors.Bar {
		if name != "all" && !slices.Contains(moods, name) {
			panic(fmt.Sprintf("energy_colors.json: bar.%s não é um humor", name))
		}
	}
	if colors.Bar["all"] == nil {
		for _, name := range moods {
			if colors.Bar[name] == nil {
				panic(fmt.Sprintf("energy_colors.json: falta a cor de bar.%s", name))
			}
		}
	}
	for name := range colors.Markers {
		mood, ok := strings.CutPrefix(name, "before_")
		if name != "all" && (!ok || !slices.Contains(moods[1:], mood)) {
			panic(fmt.Sprintf("energy_colors.json: markers.%s deve indicar um humor de destino, como before_radiante", name))
		}
	}
	return colors
}

func (p energyPalette) barColor(mood pet.Mood) ansi.IndexedColor {
	if color := p.Bar["all"]; color != nil {
		return ansi.IndexedColor(*color)
	}
	return ansi.IndexedColor(*p.Bar[pet.MoodName[mood]])
}

func (p energyPalette) markerColor(next pet.Mood) ansi.IndexedColor {
	if color := p.Markers["all"]; color != nil {
		return ansi.IndexedColor(*color)
	}
	if color := p.Markers["before_"+pet.MoodName[next]]; color != nil {
		return ansi.IndexedColor(*color)
	}
	return p.barColor(next)
}

const energyMarkerWidth = 2                     // Em oitavos de uma célula, de 1 a 7.
const energyMarkerGlyph = '█' + energyMarkerWidth // Cada caractere seguinte cobre um oitavo a menos à esquerda.

func (m model) energyBar(energy float64, mood pet.Mood, fill lipgloss.Style) string {
	const cells = 20
	const pointsPerCell = 100 / cells

	points := int(energy)
	filled := points / pointsPerCell
	unfilledColor := ansi.Convert256(m.st.off)
	if energyColors.Unfilled != nil {
		unfilledColor = ansi.IndexedColor(*energyColors.Unfilled)
	}
	unfilled := lipgloss.NewStyle().Background(unfilledColor)
	remaining := unfilled.Render(strings.Repeat(" ", cells-filled))
	caption := "humor máximo"

	if next, limit, ok := pet.NextMood(mood); ok {
		// O traço fica no fim da célula que o preenchimento completa ao chegar ao limite.
		markerCell := limit/pointsPerCell - 1
		// O bloco cinza cobre a esquerda. A cor do fundo aparece no traço à direita.
		marker := unfilled.Foreground(unfilledColor).Background(energyColors.markerColor(next)).Render(string(energyMarkerGlyph))
		remaining = unfilled.Render(strings.Repeat(" ", markerCell-filled)) +
			marker + unfilled.Render(strings.Repeat(" ", cells-markerCell-1))
		pointsLeft := limit - points
		caption = fmt.Sprintf("próximo humor em %d · %s %d", limit, plural(pointsLeft, "falta", "faltam"), pointsLeft)
	}

	return "energia  " + fill.Render(strings.Repeat("█", filled)) + remaining + fmt.Sprintf(" %d\n", points) +
		"         " + m.st.muted.Render(caption)
}
