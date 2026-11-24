package tui

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
)

type shopPanel struct {
	Editor             *appearanceEditor
	Mood               pet.Mood
	Confirm            *store.Purchase
	Keys               []string
	Batch, Busy, Equip bool
	Purchased          pet.Appearance
	PurchasedKind      string
	Dev                bool
	PriceEditing       bool
	OriginalPrice      int
	States             map[string]string
	Prices             map[string]int
	BaseStamp          string
}
type economyMsg struct {
	ID       string
	Revision int
	Value    store.Economy
	Err      error
}
type purchaseMsg struct {
	ID     string
	Value  store.Economy
	Err    error
	Batch  bool
	Choice pet.Appearance
	Kind   string
}
type shopSavedMsg struct {
	ID   string
	Err  error
	Exit bool
}

func (m model) shopText(key string, args ...any) string {
	text := pet.Shop.Texts[key].Text(string(m.lang))
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}
func shopStyle(key string) lipgloss.Style {
	value := pet.Shop.Colors[key]
	if key == "coin" && value == "" {
		value = pet.Coins.Color
	}
	c, _ := pet.ParsePaletteColor(value)
	return lipgloss.NewStyle().Foreground(c)
}
func (m model) coinDisplay(amount int) string {
	icon := pet.Pixels(pet.CoinImage, 0, 0, 5, 5, 1, nil)
	return lipgloss.JoinHorizontal(lipgloss.Center, icon, " ", shopStyle("coin").Render(strconv.Itoa(amount)))
}
func (m model) loadEconomy() tea.Cmd {
	id, revision := m.id, m.economyRevision
	if m.nick == "" {
		return nil
	}
	return func() tea.Msg { e, err := rooms.ShopRead(id); return economyMsg{id, revision, e, err} }
}
func (m model) updateShop(msg tea.Msg) (model, tea.Cmd, bool) {
	switch v := msg.(type) {
	case economyMsg:
		if v.ID != m.id || v.Revision != m.economyRevision {
			return m, nil, true
		}
		if v.Err != nil {
			m.err = v.Err.Error()
			return m, nil, true
		}
		m.economy, m.economyReady = v.Value, true
		if m.shop.BaseStamp == "reopen" {
			m.startShopDev()
		}
		return m, nil, true
	case purchaseMsg:
		if v.ID != m.id {
			return m, nil, true
		}
		m.shop.Busy = false
		m.shop.Confirm = nil
		if v.Err != nil {
			m.err = m.shopError(v.Err, v.Batch)
			return m, m.loadEconomy(), true
		}
		m.economyRevision++
		m.economy = v.Value
		m.economyReady = true
		m.shop.Equip = true
		m.shop.Batch = v.Batch
		m.shop.Purchased = v.Choice
		m.shop.PurchasedKind = v.Kind
		m.notice = m.shopText("bought")
		return m, nil, true
	case shopSavedMsg:
		if v.ID != m.id {
			return m, nil, true
		}
		m.shop.Busy = false
		if v.Err != nil {
			m.err = m.shopError(v.Err, false)
			return m, nil, true
		}
		m.economyRevision++
		m.shop.Dev = false
		m.shop.PriceEditing = false
		m.notice = m.shopText("saved")
		if !v.Exit {
			m.shop.BaseStamp = "reopen"
		}
		return m, m.loadEconomy(), true
	}
	return m, nil, false
}
func (m model) shopError(err error, batch bool) string {
	var e *store.ShopError
	if !errors.As(err, &e) {
		return m.lang.errText(err)
	}
	item := pet.ShopItemByKey(e.Key)
	category := ""
	if item != nil {
		category = m.shopText(item.Kind)
	}
	switch e.Kind {
	case "insufficient":
		if batch {
			return m.shopText("batch_insufficient", e.Need, e.Have)
		}
		return m.shopText("insufficient", category, e.Need, e.Have)
	case "missing_base":
		base := pet.ShopItemByKey(item.Parent)
		return m.shopText("missing_base", m.shopText(base.Kind), base.Name.Text(string(m.lang)), category)
	case "included_error":
		base := pet.ShopItemByKey(item.Parent)
		return m.shopText("included_error", category, base.Name.Text(string(m.lang)))
	case "locked_help":
		if batch {
			return m.shopText("batch_locked")
		}
		return m.shopText(e.Kind, category)
	case "owned_help":
		return m.shopText(e.Kind, category)
	default:
		return m.shopText(e.Kind)
	}
}
func (m *model) ensureShop() {
	if m.shop.Editor == nil {
		m.shop.Editor = newAppearanceEditor(m.pet.Look())
		m.shop.Mood = pet.MoodOf(m.pet.EnergyAt(m.now))
	}
	if m.shop.BaseStamp == "reopen" && m.economyReady {
		m.startShopDev()
	}
}
func (m *model) startShopDev() {
	m.shop.Dev = true
	m.shop.States = maps.Clone(m.economy.States)
	m.shop.Prices = maps.Clone(m.economy.Prices)
	m.shop.BaseStamp = store.EconomyStamp(m.economy)
}
func (m model) shopEconomy() store.Economy {
	e := m.economy
	if m.shop.Dev {
		e.States = m.shop.States
		e.Prices = m.shop.Prices
	}
	return e
}
func (m model) currentShopKey() string {
	e := m.shop.Editor
	a := e.Choice
	switch e.rows(true)[e.Row] {
	case "character":
		return "character/" + a.Character
	case "color":
		return "color/" + a.Character + "/" + a.Variant
	case "hat":
		if a.Hat != "" {
			return "hat/" + a.Hat
		}
	case "model":
		return "model/" + a.Hat + "/" + a.HatVariant
	}
	return ""
}
func (m model) shopKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.ensureShop()
	k := msg.String()
	m.err = ""
	if !m.economyReady || m.shop.Busy {
		return m, nil
	}
	if m.shop.Equip {
		if k == "n" || k == "esc" {
			m.shop.Equip = false
			return m, nil
		}
		if k != "y" {
			return m, nil
		}
		a := m.shop.Purchased
		if !m.shop.Batch {
			a = m.pet.Look()
			selected := m.shop.Purchased
			switch m.shop.PurchasedKind {
			case "character":
				a.Character = selected.Character
				a.Variant = pet.Pets.Character(a.Character).Variants[0].ID
			case "color":
				a.Character, a.Variant = selected.Character, selected.Variant
			case "hat":
				a.Hat = selected.Hat
				a.HatVariant = pet.Pets.Hat(a.Hat).Variants[0].ID
			case "model":
				a.Hat, a.HatVariant = selected.Hat, selected.HatVariant
			}
		}
		m.shop.Equip = false
		id := m.id
		return m, func() tea.Msg {
			err := rooms.EquipOwned(id, a)
			return appearanceSavedMsg{id: id, revision: m.appearanceRevision, err: err, world: rooms.SnapshotWorld()}
		}
	}
	if m.shop.Confirm != nil {
		if k == "n" || k == "esc" {
			m.shop.Confirm = nil
			return m, nil
		}
		if k != "y" {
			return m, nil
		}
		id, keys, batch, stamp, choice := m.id, slices.Clone(m.shop.Keys), m.shop.Batch, m.shop.Confirm.Stamp, m.shop.Editor.Choice
		kind := m.shop.Editor.rows(true)[m.shop.Editor.Row]
		m.shop.Busy = true
		return m, func() tea.Msg {
			e, err := rooms.ShopBuy(id, keys, batch, stamp)
			return purchaseMsg{id, e, err, batch, choice, kind}
		}
	}
	if m.shop.Dev {
		key := m.currentShopKey()
		if m.shop.PriceEditing {
			delta := map[string]int{"left": -1, "right": 1, "j": -10, "l": 10}[k]
			if delta != 0 {
				m.shop.Prices[key] = max(0, min(9999, m.shop.Prices[key]+delta))
			}
			if k == "esc" {
				m.shop.Prices[key] = m.shop.OriginalPrice
			}
			if k == "enter" || k == "esc" {
				m.shop.PriceEditing = false
			}
			return m, nil
		}
		switch k {
		case "ctrl+p":
			if key != "" {
				m.shop.PriceEditing = true
				m.shop.OriginalPrice = m.shop.Prices[key]
			}
			return m, nil
		case "ctrl+o":
			if key != "" {
				for _, worn := range pet.AppearanceItems(m.pet.Look()) {
					if worn == key {
						m.err = m.shopText("equipped_error")
						return m, nil
					}
				}
				state := m.shopEconomy().State(key)
				next := map[string]string{store.ItemAvailable: store.ItemOwned, store.ItemOwned: store.ItemLocked, store.ItemLocked: store.ItemAvailable}[state]
				item := pet.ShopItemByKey(key)
				if item.Included && m.shopEconomy().State(item.Parent) == store.ItemOwned && next != store.ItemOwned {
					m.err = m.shopText("dependency_error")
					return m, nil
				}
				m.shop.States[key] = next
				if next == store.ItemOwned {
					for _, child := range pet.ShopItems {
						if child.Parent == key && child.Included {
							m.shop.States[child.Key] = store.ItemOwned
						}
					}
				}
			}
			return m, nil
		case "x", "ctrl+s":
			id, states, prices, stamp := m.id, maps.Clone(m.shop.States), maps.Clone(m.shop.Prices), m.shop.BaseStamp
			exit := k == "x"
			m.shop.Busy = true
			return m, func() tea.Msg { return shopSavedMsg{id, rooms.ShopDevSave(id, states, prices, stamp), exit} }
		case "esc":
			m.shop.Dev = false
			m.shop.PriceEditing = false
			return m, nil
		case "enter", "ctrl+b":
			return m, nil
		}
	} else {
		switch k {
		case "x":
			if Dev && !m.hideDev {
				m.startShopDev()
			}
			return m, nil
		case "ctrl+r":
			m.shop.Editor = newAppearanceEditor(m.pet.Look())
			return m, nil
		case "enter", "ctrl+b":
			if m.shop.Editor.rows(true)[m.shop.Editor.Row] == "mood" && k == "enter" {
				return m, nil
			}
			keys := []string{m.currentShopKey()}
			batch := k == "ctrl+b"
			if batch {
				keys = pet.AppearanceItems(m.shop.Editor.Choice)
			}
			p, err := store.Quote(m.economy, keys, batch)
			if err != nil {
				m.err = m.shopError(err, batch)
				return m, nil
			}
			m.shop.Keys, m.shop.Batch, m.shop.Confirm = keys, batch, &p
			return m, nil
		}
	}
	m.shop.Editor.key(k, true, &m.shop.Mood)
	return m, nil
}
func (m model) shopHelp() string {
	if m.shop.Confirm != nil || m.shop.Equip {
		return m.shopText("yes_no")
	}
	if m.shop.PriceEditing {
		return m.shopText("number_help")
	}
	if m.shop.Dev {
		return m.shopText("dev_help")
	}
	s := m.shopText("help")
	if Dev && !m.hideDev {
		s += m.shopText("dev_hint")
	}
	return s
}
func (m model) shopView(w, h int) string {
	if !m.economyReady {
		return m.st.muted.Render(m.t("carregando…"))
	}
	m.ensureShop()
	e := m.shopEconomy()
	a := m.shop.Editor.Choice
	if p := m.shop.Confirm; p != nil {
		title := m.shopText("confirm_all", p.Total)
		if !m.shop.Batch {
			item := pet.ShopItemByKey(m.shop.Keys[0])
			title = m.shopText("confirm_one", m.shopText(item.Kind), item.Name.Text(string(m.lang)), p.Total)
		}
		lines := []string{title, "", shopStyle("balance").Render(m.shopText("balance", e.Balance)), m.shopText("pay_for")}
		for _, line := range p.Lines {
			item := pet.ShopItemByKey(line.Key)
			suffix := ""
			indent := "  "
			if line.Included {
				suffix = " · " + m.shopText("included")
				indent = "    "
			}
			lines = append(lines, indent+m.shopText(item.Kind)+" "+item.Name.Text(string(m.lang))+": "+shopStyle("deduction").Render(fmt.Sprintf("-%d", line.Cost))+suffix)
		}
		lines = append(lines, shopStyle("balance").Render(m.shopText("remaining", e.Balance-p.Total)), "", m.shopText("yes_no"))
		return m.box(m.shopText("shop"), strings.Join(lines, "\n"), w, 0)
	}
	if m.shop.Equip {
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
		return m.box(m.shopText("shop"), question+"\n\n"+m.shopText("yes_no"), w, 0)
	}
	rows := m.shop.Editor.rows(true)
	var lines []string
	for i, kind := range rows {
		if kind == "mood" {
			lines = append(lines, m.shopRow(i, m.shopText(kind), pet.Pets.MoodName(a, m.shop.Mood, string(m.lang)), ""))
			if i == m.shop.Editor.Row {
				lines = append(lines, m.st.muted.Render(m.shopText("preview")))
			}
			continue
		}
		var items []pet.ShopItem
		for _, item := range pet.ShopItems {
			if item.Kind == kind && (item.Parent == "" || kind == "color" && item.Parent == "character/"+a.Character || kind == "model" && item.Parent == "hat/"+a.Hat) {
				items = append(items, item)
			}
		}
		keys := pet.AppearanceItems(a)
		key := ""
		switch kind {
		case "character":
			key = keys[0]
		case "color":
			key = keys[1]
		case "hat":
			if a.Hat != "" {
				key = keys[2]
			}
		case "model":
			key = keys[3]
		}
		owned, index, total := 0, 0, len(items)
		for j, item := range items {
			if e.State(item.Key) == store.ItemOwned {
				owned++
			}
			if item.Key == key {
				index = j
			}
		}
		name, state, info := m.shopText("no_hat"), shopStyle("owned").Render(m.shopText("owned")), m.shopText("no_hat_help")
		if kind == "hat" {
			owned++
			total++
			if key != "" {
				index++
			}
		}
		if item := pet.ShopItemByKey(key); item != nil {
			name = item.Name.Text(string(m.lang))
			st := e.State(key)
			category := m.shopText(kind)
			switch {
			case st == store.ItemOwned:
				info = m.shopText("owned_help", category)
				if item.Included {
					info = m.shopText("included_owned", category, pet.ShopItemByKey(item.Parent).Name.Text(string(m.lang)))
				}
			case item.Included:
				base := pet.ShopItemByKey(item.Parent).Name.Text(string(m.lang))
				state = m.st.muted.Render(m.shopText("comes", base))
				info = m.shopText("included_help", category, base)
			case st == store.ItemLocked:
				state = shopStyle("locked").Render(m.shopText("locked"))
				info = m.shopText("locked_help", category)
			default:
				state = m.coinDisplay(e.Prices[key])
				info = m.shopText("can_buy", category, e.Prices[key])
				if e.Prices[key] > e.Balance {
					info = m.shopText("need_money", category, e.Prices[key])
				}
			}
		}
		if m.shop.PriceEditing && key == m.currentShopKey() {
			state = m.coinDisplay(e.Prices[key])
			info = m.shopText("number_help")
		}
		label := fmt.Sprintf("%s (%d/%d)", m.shopText(kind), index+1, total)
		row := m.shopRow(i, label, name, state)
		plural := map[string]string{"character": "characters", "color": "colors", "hat": "hats", "model": "models"}[kind]
		lines = append(lines, row)
		if i == m.shop.Editor.Row || h >= 38 {
			lines = append(lines, m.st.muted.Render(info), m.st.muted.Render(m.shopText("have", owned, total, m.shopText(plural))))
		}
	}
	rightWidth := max(20, w/3)
	leftWidth := max(28, w-rightWidth-2)
	title := m.shopText("shop")
	if m.shop.Dev {
		title += " · " + devAccent.Render(m.t("dev"))
	}
	left := m.box(title, strings.Join(lines, "\n"), leftWidth, 0)
	scale := 1
	if w >= 112 && h >= 30 {
		scale = 2
	}
	preview := pet.Pets.RenderTerminal(a, m.shop.Mood, m.tick, scale, 0, m.pixel)
	right := m.box(m.shopText("shop"), m.coinDisplay(e.Balance)+"\n\n"+preview, rightWidth, 0)
	if w < 58 {
		return left + "\n" + right
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}
func (m model) shopRow(i int, label, name, state string) string {
	prefix := "  "
	if i == m.shop.Editor.Row {
		prefix = m.st.sel.Render("▸ ")
		label = m.st.sel.Render(label)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, prefix+label+"  ‹ "+name+" ›  ", state)
}
