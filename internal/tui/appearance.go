package tui

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
)

type appearanceSavedMsg struct {
	id       string
	revision int
	err      error
	world    rooms.World
}

// Cada abertura tem sua própria memória de cores/modelos. Só Choice é salvo.
type appearanceEditor struct {
	Choice pet.Appearance
	Row    int
	Colors map[string]string
	Models map[string]string
}

func newAppearanceEditor(a pet.Appearance) *appearanceEditor {
	e := &appearanceEditor{Choice: a, Colors: map[string]string{a.Character: a.Variant}, Models: map[string]string{}}
	if a.Hat != "" {
		e.Models[a.Hat] = a.HatVariant
	}
	return e
}

func (e *appearanceEditor) rows(moodPreview bool) []string {
	rows := []string{"character", "color", "hat"}
	if e.Choice.Hat != "" {
		rows = append(rows, "model")
	}
	if moodPreview {
		rows = append(rows, "mood")
	}
	return rows
}

func (e *appearanceEditor) key(k string, moodPreview bool, mood *pet.Mood) {
	rows := e.rows(moodPreview)
	e.Row = min(e.Row, len(rows)-1)
	switch k {
	case "up":
		e.Row = (e.Row + len(rows) - 1) % len(rows)
		return
	case "down":
		e.Row = (e.Row + 1) % len(rows)
		return
	}
	delta := 0
	if k == "left" {
		delta = -1
	}
	if k == "right" {
		delta = 1
	}
	if delta == 0 {
		return
	}
	cycle := func(i, n int) int { return (i + delta + n) % n }
	a := &e.Choice
	switch rows[e.Row] {
	case "character":
		cs := pet.Pets.Characters
		i := slices.IndexFunc(cs, func(c pet.Character) bool { return c.ID == a.Character })
		ch := cs[cycle(i, len(cs))]
		a.Character = ch.ID
		a.Variant = e.Colors[ch.ID]
		if a.Variant == "" {
			a.Variant = ch.Variants[0].ID
			e.Colors[ch.ID] = a.Variant
		}
	case "color":
		vs := pet.Pets.Character(a.Character).Variants
		i := slices.IndexFunc(vs, func(v pet.CharacterVariant) bool { return v.ID == a.Variant })
		a.Variant = vs[cycle(i, len(vs))].ID
		e.Colors[a.Character] = a.Variant
	case "hat":
		hs := pet.Pets.Hats
		i := slices.IndexFunc(hs, func(h pet.Hat) bool { return h.ID == a.Hat }) + 1
		i = cycle(i, len(hs)+1)
		if i == 0 {
			a.Hat, a.HatVariant = "", ""
			return
		}
		h := hs[i-1]
		a.Hat = h.ID
		a.HatVariant = e.Models[h.ID]
		if a.HatVariant == "" {
			a.HatVariant = h.Variants[0].ID
			e.Models[h.ID] = a.HatVariant
		}
	case "model":
		vs := pet.Pets.Hat(a.Hat).Variants
		i := slices.IndexFunc(vs, func(v pet.HatVariant) bool { return v.ID == a.HatVariant })
		a.HatVariant = vs[cycle(i, len(vs))].ID
		e.Models[a.Hat] = a.HatVariant
	case "mood":
		*mood = pet.Mood(cycle(int(*mood), int(pet.Sleeping)+1))
	}
}

func (m model) profileAppearance() pet.Appearance {
	if m.custom && m.appearanceEdit != nil {
		return m.appearanceEdit.Choice
	}
	return m.pet.Look()
}

func (m model) appearanceRows(a pet.Appearance, e *appearanceEditor, preview bool) string {
	ch := pet.Pets.Character(a.Character)
	v := ch.Variant(a.Variant)
	labels := []string{m.t("personagem"), m.t("cor do personagem"), m.t("chapéu")}
	values := []string{ch.DisplayName.Text(string(m.lang)), v.DisplayName.Text(string(m.lang)) + " " + lipgloss.NewStyle().Foreground(pet.Pets.Swatch(a)).Render("██"), m.t("sem chapéu")}
	if a.Hat != "" {
		hat := pet.Pets.Hat(a.Hat)
		values[2] = hat.DisplayName.Text(string(m.lang))
		labels = append(labels, m.t("modelo do chapéu"))
		values = append(values, hat.Variant(a.HatVariant).DisplayName.Text(string(m.lang)))
	}
	if preview {
		labels = append(labels, m.t("prévia do humor"))
		values = append(values, pet.Pets.MoodName(a, m.previewMood, string(m.lang)))
	}
	var lines []string
	for i, label := range labels {
		prefix, value := "  ", values[i]
		label = fmt.Sprintf("%-18s", label)
		if e != nil {
			value = "‹ " + value + " ›"
			if e.Row == i {
				prefix = m.st.sel.Render("▸ ")
				label = m.st.sel.Render(label)
			}
		}
		lines = append(lines, prefix+label+" "+value)
	}
	if e != nil && preview && e.rows(true)[e.Row] == "mood" {
		lines = append(lines, m.st.muted.Render(m.t("apenas para visualização")+"\n"+m.t("(não altera humor ou energia)")))
	}
	return strings.Join(lines, "\n")
}
