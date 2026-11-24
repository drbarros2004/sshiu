package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"sshiu/internal/store"
)

func purchaseDialogModel() model {
	m := model{lang: "en", economy: store.Economy{Balance: 3865}}
	m.shop.Batch = true
	m.shop.Confirm = &store.Purchase{Total: 120, Lines: []store.PurchaseLine{
		{Key: "character/bunny", Cost: 120},
		{Key: "color/bunny/blue", Included: true},
	}}
	return m
}

func TestShopPurchaseDialogLayout(t *testing.T) {
	for _, size := range []struct {
		name          string
		width, height int
	}{
		{"wide", 180, 50},
		{"normal", 84, 28},
		{"narrow", 42, 18},
	} {
		t.Run(size.name, func(t *testing.T) {
			m := purchaseDialogModel()
			screen, limit := m.shopDialogView(size.width, size.height)
			plain := ansi.Strip(screen)
			if limit != 0 {
				t.Fatalf("diálogo deveria caber sem rolagem: %d", limit)
			}
			for _, want := range []string{"confirm purchase", "buy these items for 120 coins?", "current balance", "3,865 coins", "items:", "character bunny", "color blue", "included", "balance after purchase", "3,745 coins", "y yes · n no"} {
				if !strings.Contains(plain, want) {
					t.Errorf("faltou %q na confirmação", want)
				}
			}
			if !strings.Contains(plain, "-120") || strings.Contains(plain, "-0") {
				t.Fatal("preço deve ter sinal de menos; opção incluída deve mostrar apenas included")
			}
			first, last := -1, -1
			balanceRow, controlsRow := -1, -1
			for i, line := range strings.Split(plain, "\n") {
				if strings.Contains(line, "balance after purchase") {
					balanceRow = i
				}
				if strings.Contains(line, "y yes · n no") {
					controlsRow = i
				}
				if ansi.StringWidth(line) > size.width {
					t.Fatalf("linha %d ultrapassou a janela", i)
				}
				if strings.Contains(line, "╭") {
					first = i
					left := ansi.StringWidth(strings.SplitN(line, "╭", 2)[0])
					right := size.width - ansi.StringWidth(strings.TrimRight(line, " "))
					if left-right > 1 || right-left > 1 {
						t.Fatal("caixa não está centralizada horizontalmente")
					}
				}
				if strings.Contains(line, "╰") {
					last = i
				}
			}
			if balanceRow < 0 || controlsRow-balanceRow != 2 {
				t.Fatal("saldo final e controles devem ter apenas uma linha vazia entre eles")
			}
			bottom := size.height - 2 - last
			if first < 0 || last < first || first-bottom > 1 || bottom-first > 1 {
				t.Fatal("caixa não está centralizada verticalmente")
			}
		})
	}
}

func TestShopDialogScrollKeepsControls(t *testing.T) {
	m := purchaseDialogModel()
	_, limit := m.shopDialogView(42, 10)
	if limit <= 0 {
		t.Fatal("janela baixa deveria permitir rolar")
	}
	m.shop.DialogScroll = limit
	screen, _ := m.shopDialogView(42, 10)
	plain := ansi.Strip(screen)
	if !strings.Contains(plain, "3,745 coins") || !strings.Contains(plain, "y yes · n no") {
		t.Fatal("rolagem perdeu o saldo final ou a resposta")
	}
}

func TestShopPurchaseDialogItemGroups(t *testing.T) {
	m := purchaseDialogModel()
	m.shop.Confirm.Lines = append(m.shop.Confirm.Lines,
		store.PurchaseLine{Key: "hat/top_hat", Cost: 60},
		store.PurchaseLine{Key: "model/top_hat/normal", Included: true},
	)
	m.shop.Confirm.Total = 180
	screen, _ := m.shopDialogView(180, 50)
	rows := strings.Split(ansi.Strip(screen), "\n")
	find := func(label string) int {
		t.Helper()
		for i, row := range rows {
			if strings.Contains(row, label) {
				return i
			}
		}
		t.Fatalf("faltou %q", label)
		return -1
	}
	heading, body := find("items:"), find("character bunny")
	color, hat, model := find("color blue"), find("hat top hat"), find("model normal")
	if body != heading+1 || color != body+1 || hat != color+2 || model != hat+1 {
		t.Fatal("compras precisam de uma linha vazia entre grupos, mantendo opções incluídas junto da base")
	}
	if strings.Index(rows[body], "character bunny")-strings.Index(rows[heading], "items:") != 2 ||
		strings.Index(rows[color], "color blue")-strings.Index(rows[body], "character bunny") != 2 {
		t.Fatal("recuos devem ser de duas posições para compras e quatro para opções incluídas")
	}
	priceEdge := strings.Index(rows[body], "-120") + len("-120")
	if strings.Index(rows[hat], "-60")+len("-60") != priceEdge ||
		strings.Index(rows[color], "included")+len("included") != priceEdge {
		t.Fatal("preços e included devem terminar na mesma coluna")
	}
}

func TestShopEquipDialogUsesCenteredBox(t *testing.T) {
	m := model{lang: "en"}
	m.shop.Equip, m.shop.Batch = true, true
	screen, limit := m.shopDialogView(120, 32)
	plain := ansi.Strip(screen)
	if limit != 0 || !strings.Contains(plain, "equip items") || !strings.Contains(plain, "do you want to equip all selected items?") || !strings.Contains(plain, "y yes · n no") {
		t.Fatal("diálogo de equipamento incompleto")
	}
	if strings.Contains(plain, "current balance") {
		t.Fatal("equipar não deve repetir cobrança")
	}
}
