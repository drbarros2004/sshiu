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
