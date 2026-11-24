package tui

import (
	"github.com/charmbracelet/x/ansi"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
	"strings"
	"testing"
	"time"
)

func TestCoinReceiptPrivacy(t *testing.T) {
	m := model{id: "owner", lang: "en"}
	n := 3
	event := rooms.Event{Who: "other", Coins: &n}
	if m.receiptSuffix(event) != "" {
		t.Fatal("mostrou o recibo de outra identidade")
	}
	event.Who = "owner"
	if got := ansi.Strip(m.receiptSuffix(event)); !strings.Contains(got, "+3 coins gained") {
		t.Fatalf("recibo próprio: %q", got)
	}
}
func TestCoinGainShowsWalletCap(t *testing.T) {
	now := time.Now()
	m := model{lang: "en", now: now, economyReady: true, economy: store.Economy{Balance: 9998}}
	f := rooms.Focus{SegStart: now.Add(-3 * time.Minute), Ends: now.Add(time.Minute)}
	got := ansi.Strip(m.coinGain(f))
	if !strings.Contains(got, "+1 coin") || !strings.Contains(got, "9999-coin limit") {
		t.Fatalf("prévia limitada: %q", got)
	}
}
func TestOwnedAppearanceChoices(t *testing.T) {
	e := newAppearanceEditor(pet.Appearance{Character: "fruit", Variant: "blue"})
	e.Owned = map[string]string{"character/fruit": store.ItemOwned, "color/fruit/blue": store.ItemOwned, "color/fruit/coral": store.ItemOwned}
	mood := pet.Okay
	e.key("right", true, &mood)
	if e.Choice.Character != "fruit" {
		t.Fatal("permitiu personagem não possuído")
	}
	e.Row = 1
	e.key("right", true, &mood)
	if e.Choice.Variant != "coral" {
		t.Fatalf("não pulou cores não possuídas: %+v", e.Choice)
	}
}
func TestShopCancelDoesNotPurchase(t *testing.T) {
	e := store.Economy{Balance: 1000, States: map[string]string{}, Prices: map[string]int{}}
	for _, item := range pet.ShopItems {
		e.Prices[item.Key] = item.Price
	}
	quote, err := store.Quote(e, []string{"character/bunny"}, false)
	if err != nil {
		t.Fatal(err)
	}
	m := model{id: "owner", nick: "owner", lang: "en", tab: tabShop, economy: e, economyReady: true,
		pet: pet.Pet{Appearance: pet.Appearance{Character: "fruit", Variant: "blue"}}}
	m.shop.Confirm = &quote
	m, _ = pressInline(m, 'n', false)
	if m.shop.Confirm != nil || m.economy.Balance != 1000 || m.economy.State("character/bunny") == store.ItemOwned {
		t.Fatal("cancelamento alterou saldo/inventário")
	}
}

func TestShopDescriptionsFollowSelectedCategory(t *testing.T) {
	for _, height := range []int{24, 50} {
		m := model{lang: "en", economyReady: true}
		m.shop.Editor = newAppearanceEditor(pet.Appearance{Character: "fruit", Variant: "blue"})
		character := pet.ShopItemByKey("character/fruit").Description.Text("en")
		color := pet.ShopItemByKey("color/fruit/blue").Description.Text("en")
		for _, state := range []string{store.ItemOwned, store.ItemLocked, store.ItemAvailable} {
			m.economy.States = map[string]string{"character/fruit": state}
			m.shop.Editor.Row = 0
			screen := ansi.Strip(m.shopView(220, height))
			if !strings.Contains(screen, character) || strings.Contains(screen, color) {
				t.Fatal("apenas a categoria selecionada deve mostrar descrição, independente do estado ou altura")
			}
		}
		m.shop.Editor.Row = 1
		screen := ansi.Strip(m.shopView(220, height))
		if !strings.Contains(screen, color) || strings.Contains(screen, character) {
			t.Fatal("a descrição deve acompanhar a categoria selecionada")
		}
		m.shop.Editor.Row = 2
		screen = ansi.Strip(m.shopView(220, height))
		if !strings.Contains(screen, pet.Pets.NoHat.Description.Text("en")) {
			t.Fatal("sem chapéu deve usar a descrição do catálogo de chapéus")
		}
	}
}

func TestShopDescriptionFallback(t *testing.T) {
	for _, tc := range []struct {
		state, text string
		balance     int
	}{
		{store.ItemOwned, "description_owned", 0},
		{store.ItemLocked, "description_locked", 0},
		{store.ItemAvailable, "description_available", 40},
		{store.ItemAvailable, "description_insufficient", 0},
	} {
		t.Run(tc.text, func(t *testing.T) {
			m := model{lang: "en", economyReady: true, economy: store.Economy{
				Balance: tc.balance,
				States:  map[string]string{"color/fruit/green": tc.state},
				Prices:  map[string]int{"color/fruit/green": 40},
			}}
			m.shop.Editor = newAppearanceEditor(pet.Appearance{Character: "fruit", Variant: "green"})
			m.shop.Editor.Row = 1
			if !strings.Contains(ansi.Strip(m.shopView(220, 50)), m.shopText(tc.text)) {
				t.Fatal("item sem descrição deve usar o texto curto de seu estado")
			}
		})
	}
}
