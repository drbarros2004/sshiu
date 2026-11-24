package tui

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"sshiu/internal/pet"
)

// O terminal informa células, não pixels da fonte. O mínimo evita uma caixa
// minúscula; a fração da janela dá espaço em terminais grandes, sem ocupar tudo.
func (m model) shopDialogView(w, h int) (string, int) {
	h = max(1, h-1) // a casca reserva a última linha para avisos
	if w < 12 || h < 5 {
		return fit(m.shopText("yes_no"), w, h), 0
	}
	layout := pet.Shop.Confirmation
	width := min(w, max(layout.MinimumWidth, w*layout.WidthPercent/100))
	padX := min(layout.PaddingHorizontal, max(0, (width-12)/2))
	contentWidth := max(1, width-2-2*padX)
	text, muted := shopStyle("confirm_text"), shopStyle("confirm_muted")
	questionStyle := shopStyle("confirm_question").Bold(true)
	border := shopStyle("confirm_border")
	green, red := shopStyle("confirm_balance"), shopStyle("confirm_price")
	title := m.shopText("confirm_title")
	var rows []string
	appendText := func(value string) {
		rows = append(rows, strings.Split(value, "\n")...)
	}
	if purchase := m.shop.Confirm; purchase != nil {
		question := m.shopText("confirm_all", purchase.Total)
		if !m.shop.Batch {
			item := pet.ShopItemByKey(m.shop.Keys[0])
			question = m.shopText("confirm_one", m.shopText(item.Kind), item.Name.Text(string(m.lang)), purchase.Total)
		}
		appendText(questionStyle.Render(ansi.Wrap(question, contentWidth, "")))
		rows = append(rows, border.Render(strings.Repeat("─", contentWidth)))
		balance := m.shopMoney(m.shopEconomy().Balance)
		remaining := m.shopMoney(m.shopEconomy().Balance - purchase.Total)
		rightWidth := min(contentWidth/2, max(lipgloss.Width(balance), lipgloss.Width(remaining), lipgloss.Width(m.shopText("included"))))
		leftWidth := max(1, contentWidth-rightWidth-2)
		pair := func(label, value string, labelStyle, valueStyle lipgloss.Style, indent int) string {
			left := labelStyle.Width(leftWidth).PaddingLeft(min(indent, max(0, leftWidth-1))).Render(label)
			right := valueStyle.Width(rightWidth).Align(lipgloss.Right).Render(value)
			return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
		}
		appendText(pair(m.shopText("balance"), balance, muted, green, 0))
		rows = append(rows, "")
		appendText(text.Bold(true).Render(m.shopText("pay_for")))
		for i, line := range purchase.Lines {
			// A opção incluída fica junto da base; outra compra abre um novo grupo.
			if i > 0 && !line.Included {
				rows = append(rows, "")
			}
			item := pet.ShopItemByKey(line.Key)
			category := item.Kind
			if category == "color" || category == "model" {
				category = "confirm_" + category
			}
			label := m.shopText(category) + " " + item.Name.Text(string(m.lang))
			if line.Included {
				appendText(pair(label, m.shopText("included"), muted, muted, layout.IncludedIndent))
			} else {
				appendText(pair(label, "-"+strconv.Itoa(line.Cost), text, red, layout.ItemIndent))
			}
		}
		rows = append(rows, border.Render(strings.Repeat("─", contentWidth)))
		appendText(pair(m.shopText("remaining"), remaining, muted, green, 0))
	} else {
		title = m.shopText("equip_title")
		question := m.shopText("equip_all")
		if !m.shop.Batch {
			a := m.shop.Purchased
			key := ""
			switch m.shop.PurchasedKind {
			case "character":
				key = "character/" + a.Character
			case "color":
				key = "color/" + a.Character + "/" + a.Variant
			case "hat":
				key = "hat/" + a.Hat
			case "model":
				key = "model/" + a.Hat + "/" + a.HatVariant
			}
			item := pet.ShopItemByKey(key)
			question = m.shopText("equip_one", item.Name.Text(string(m.lang)), m.shopText(item.Kind))
		}
		appendText(text.Render(ansi.Wrap(question, contentWidth, "")))
	}
	padY := layout.PaddingVertical
	naturalHeight := len(rows) + 2 + 2*padY + 2 // controles, espaço e bordas
	height := min(h, naturalHeight)
	if naturalHeight > height {
		padY = 0
	}
	visibleRows := max(1, height-2-2*padY-2)
	maxScroll := max(0, len(rows)-visibleRows)
	offset := min(m.shop.DialogScroll, maxScroll)
	shown := append([]string{}, rows[offset:min(len(rows), offset+visibleRows)]...)
	controls := m.shopText("yes_no")
	if maxScroll > 0 {
		controls += " · " + m.shopText("dialog_scroll")
	}
	shown = append(shown, "", muted.Render(ansi.Truncate(controls, contentWidth, "…")))
	shown = append(make([]string, padY), shown...)
	shown = append(shown, make([]string, padY)...)
	caption := " " + ansi.Truncate(title, width-5, "…") + " "
	var box strings.Builder
	box.WriteString(border.Render("╭─") + text.Render(caption) + border.Render(strings.Repeat("─", max(0, width-3-lipgloss.Width(caption)))+"╮") + "\n")
	for _, row := range shown {
		box.WriteString(border.Render("│") + strings.Repeat(" ", padX) + row + strings.Repeat(" ", max(0, contentWidth-lipgloss.Width(row))+padX) + border.Render("│") + "\n")
	}
	box.WriteString(border.Render("╰" + strings.Repeat("─", width-2) + "╯"))
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box.String()), maxScroll
}

func (m model) shopMoney(amount int) string {
	digits := strconv.Itoa(amount)
	separator := ","
	if string(m.lang) == "pt" {
		separator = "."
	}
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + separator + digits[i:]
	}
	return digits + " " + m.shopText("coins")
}
