package store

import (
	"errors"
	"path/filepath"
	"sshiu/internal/pet"
	"testing"
	"time"
)

func shopTestUser(t *testing.T) Economy {
	t.Helper()
	if err := Open(filepath.Join(t.TempDir(), "coins.db")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { Close() })
	if err := PutUser("coins-user", "coins-user", 50); err != nil {
		t.Fatal(err)
	}
	e, err := ReadEconomy("coins-user")
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func TestCoinSettlement(t *testing.T) {
	shopTestUser(t)
	start := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	record := func(offset, seconds int, result pet.Result) TaskRow {
		t.Helper()
		s := start.Add(time.Duration(offset) * time.Second)
		row := TaskRow{Start: s, End: s.Add(time.Duration(seconds) * time.Second), Task: "foco", DeclaredMin: 5, Focused: time.Duration(seconds) * time.Second, Result: result}
		if _, err := RecordTask("coins-user", row); err != nil {
			t.Fatal(err)
		}
		return row
	}
	record(0, 23, pet.Done)
	record(30, 600, pet.NotDone)
	last := record(700, 37, pet.Done)
	if _, err := RecordTask("coins-user", last); err != nil {
		t.Fatal(err)
	}
	e, err := ReadEconomy("coins-user")
	if err != nil || e.Balance != 1 || e.Remainder != 0 {
		t.Fatalf("carteira %+v, erro %v", e, err)
	}
	if n, ok := TaskCoins("coins-user", last); !ok || n != 1 {
		t.Fatalf("recibo %d, presente %v", n, ok)
	}
}
func TestCoinMigrationDoesNotRegrantItems(t *testing.T) {
	e := shopTestUser(t)
	if e.Balance != 0 || e.State("character/fruit") != ItemOwned || e.State("color/fruit/coral") != ItemOwned {
		t.Fatalf("início inesperado %+v", e)
	}
	_, err := db.Exec(`UPDATE inventory SET state='available' WHERE user_id=? AND item=?`, "coins-user", "color/fruit/coral")
	if err != nil {
		t.Fatal(err)
	}
	e, err = ReadEconomy("coins-user")
	if err != nil || e.State("color/fruit/coral") != ItemAvailable {
		t.Fatal("migração repetiu concessão", err)
	}
}
func TestShopPurchaseAndStaleConfirmation(t *testing.T) {
	e := shopTestUser(t)
	if _, err := db.Exec(`UPDATE wallets SET balance=1000 WHERE user_id='coins-user'`); err != nil {
		t.Fatal(err)
	}
	e, err := ReadEconomy("coins-user")
	if err != nil {
		t.Fatal(err)
	}
	keys := []string{"character/bunny", "color/bunny/blue"}
	p, err := Quote(e, keys, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Lines) != 2 || !p.Lines[1].Included || p.Lines[1].Cost != 0 {
		t.Fatalf("pacote %+v", p)
	}
	if _, err = db.Exec(`UPDATE wallets SET balance=999 WHERE user_id='coins-user'`); err != nil {
		t.Fatal(err)
	}
	if _, err = Buy("coins-user", keys, true, p.Stamp); err == nil {
		t.Fatal("aceitou confirmação antiga")
	}
	e, err = ReadEconomy("coins-user")
	if err != nil {
		t.Fatal(err)
	}
	p, err = Quote(e, keys, true)
	if err != nil {
		t.Fatal(err)
	}
	e, err = Buy("coins-user", keys, true, p.Stamp)
	if err != nil || e.Balance != 999-p.Total || e.State(keys[0]) != ItemOwned || e.State(keys[1]) != ItemOwned {
		t.Fatalf("compra %+v %v", e, err)
	}
}
func TestShopRejectsUnavailablePurchases(t *testing.T) {
	base := shopTestUser(t)
	for _, tc := range []struct {
		name, key, state, want string
		balance                int
	}{
		{"locked", "character/bunny", ItemLocked, "locked_help", 1000},
		{"owned", "character/fruit", ItemOwned, "owned_help", 1000},
		{"missing_base", "color/bunny/green", ItemAvailable, "missing_base", 1000},
		{"included", "color/bunny/blue", ItemAvailable, "included_error", 1000},
		{"insufficient", "character/bunny", ItemAvailable, "insufficient", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := Economy{Balance: tc.balance, States: map[string]string{}, Prices: base.Prices}
			e.States[tc.key] = tc.state
			_, err := Quote(e, []string{tc.key}, false)
			var failure *ShopError
			if !errors.As(err, &failure) || failure.Kind != tc.want {
				t.Fatalf("erro recebido %v, esperado %s", err, tc.want)
			}
		})
	}
}
func TestDevHistoryIsolatesCarry(t *testing.T) {
	shopTestUser(t)
	if _, err := db.Exec(`UPDATE wallets SET remainder_sec=45 WHERE user_id='coins-user'`); err != nil {
		t.Fatal(err)
	}
	_, p, _, err := GetUser("coins-user")
	if err != nil {
		t.Fatal(err)
	}
	end := time.Now().Add(-time.Minute)
	h := TaskRow{Start: end.Add(-23 * time.Second), End: end, Task: "histórico", Focused: 23 * time.Second, Result: pet.Done}
	if err = SaveDevScene([]DevWrite{{ID: "coins-user", Nick: "coins-user", Pet: p, History: &h}}, nil, nil, map[string]string{}); err != nil {
		t.Fatal(err)
	}
	e, err := ReadEconomy("coins-user")
	if err != nil || e.Balance != 0 || e.Remainder != 45 {
		t.Fatalf("histórico alterou carry: %+v %v", e, err)
	}
}
