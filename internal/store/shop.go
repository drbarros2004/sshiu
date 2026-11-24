package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sshiu/internal/pet"
)

const (
	ItemAvailable = "available"
	ItemOwned     = "owned"
	ItemLocked    = "locked"
)

type Economy struct {
	Balance, Remainder int
	States             map[string]string
	Prices             map[string]int
}

func (e Economy) State(key string) string {
	if s := e.States[key]; s != "" {
		return s
	}
	return ItemAvailable
}
func (e Economy) Owns(a pet.Appearance) bool {
	for _, k := range pet.AppearanceItems(a) {
		if e.State(k) != ItemOwned {
			return false
		}
	}
	return true
}

// A carteira marca a migração: nunca repor itens retirados no dev ao reconectar.
func initializeWallet(tx *sql.Tx, id string) error {
	var exists int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM wallets WHERE user_id=?`, id).Scan(&exists); err != nil {
		return err
	}
	if exists != 0 {
		return nil
	}
	_, p, ok, err := scanUser(tx.QueryRow(selectUser, id))
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("usuário inexistente")
	}
	if _, err = tx.Exec(`INSERT INTO wallets(user_id,balance) VALUES (?,?)`, id, pet.Coins.InitialBalance); err != nil {
		return err
	}
	keys := append([]string{}, pet.Shop.StarterItems...)
	keys = append(keys, pet.AppearanceItems(p.Look())...)
	for _, k := range keys {
		if err = grantItem(tx, id, k); err != nil {
			return err
		}
	}
	return nil
}
func grantItem(tx *sql.Tx, id, key string) error {
	if pet.ShopItemByKey(key) == nil {
		return fmt.Errorf("item desconhecido: %s", key)
	}
	if _, err := tx.Exec(`INSERT INTO inventory(user_id,item,state) VALUES (?,?,'owned') ON CONFLICT(user_id,item) DO UPDATE SET state='owned'`, id, key); err != nil {
		return err
	}
	for _, item := range pet.ShopItems {
		if item.Parent == key && item.Included {
			if _, err := tx.Exec(`INSERT INTO inventory(user_id,item,state) VALUES (?,?,'owned') ON CONFLICT(user_id,item) DO UPDATE SET state='owned'`, id, item.Key); err != nil {
				return err
			}
		}
	}
	return nil
}
func economy(tx *sql.Tx, id string) (Economy, error) {
	e := Economy{States: map[string]string{}, Prices: map[string]int{}}
	if err := initializeWallet(tx, id); err != nil {
		return e, err
	}
	if err := tx.QueryRow(`SELECT balance,remainder_sec FROM wallets WHERE user_id=?`, id).Scan(&e.Balance, &e.Remainder); err != nil {
		return e, err
	}
	// Trocar a taxa não cria moedas: descartar intervalos inteiros já na leitura.
	e.Remainder %= pet.Coins.SecondsPerReward
	if _, err := tx.Exec(`UPDATE wallets SET remainder_sec=? WHERE user_id=?`, e.Remainder, id); err != nil {
		return e, err
	}
	rows, err := tx.Query(`SELECT item,state FROM inventory WHERE user_id=?`, id)
	if err != nil {
		return e, err
	}
	for rows.Next() {
		var k, v string
		if err = rows.Scan(&k, &v); err != nil {
			rows.Close()
			return e, err
		}
		e.States[k] = v
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return e, err
	}
	for _, item := range pet.ShopItems {
		e.Prices[item.Key] = item.Price
	}
	rows, err = tx.Query(`SELECT item,price FROM shop_prices`)
	if err != nil {
		return e, err
	}
	for rows.Next() {
		var k string
		var v int
		if err = rows.Scan(&k, &v); err != nil {
			rows.Close()
			return e, err
		}
		e.Prices[k] = v
	}
	err = rows.Err()
	rows.Close()
	return e, err
}
func ReadEconomy(id string) (Economy, error) {
	tx, err := db.Begin()
	if err != nil {
		return Economy{}, err
	}
	defer tx.Rollback()
	e, err := economy(tx, id)
	if err != nil {
		return e, err
	}
	return e, tx.Commit()
}
func settleCoins(tx *sql.Tx, id string, taskID int64, seconds int, isolated bool) (int, error) {
	e, err := economy(tx, id)
	if err != nil {
		return 0, err
	}
	carry := e.Remainder
	if isolated {
		carry = 0
	}
	coins, rest := pet.CoinReward(e.Balance, carry, seconds)
	if isolated {
		rest = e.Remainder
	}
	if _, err = tx.Exec(`INSERT INTO coin_receipts(task_id,user_id,coins) VALUES (?,?,?)`, taskID, id, coins); err != nil {
		return 0, err
	}
	_, err = tx.Exec(`UPDATE wallets SET balance=?,remainder_sec=? WHERE user_id=?`, e.Balance+coins, rest, id)
	return coins, err
}
func TaskCoins(id string, t TaskRow) (int, bool) {
	var n int
	err := db.QueryRow(`SELECT r.coins FROM coin_receipts r JOIN tasks t ON t.id=r.task_id WHERE r.user_id=? AND t.started_at=? AND t.ended_at=? AND t.task=? ORDER BY t.id DESC LIMIT 1`, id, t.Start.Unix(), t.End.Unix(), t.Task).Scan(&n)
	return n, err == nil
}

type PurchaseLine struct {
	Key      string
	Cost     int
	Included bool
}
type Purchase struct {
	Lines []PurchaseLine
	Total int
	Stamp string
}
type ShopError struct {
	Kind, Key  string
	Need, Have int
}

func (e *ShopError) Error() string { return e.Kind }

// Quote não tem efeitos. A mesma proposta é reconstruída dentro da transação.
func Quote(e Economy, keys []string, batch bool) (Purchase, error) {
	p := Purchase{}
	selected := map[string]bool{}
	for _, k := range keys {
		selected[k] = true
	}
	seen := map[string]bool{}
	for _, key := range keys {
		item := pet.ShopItemByKey(key)
		if item == nil {
			return p, &ShopError{Kind: "no_hat_error"}
		}
		state := e.State(key)
		if state == ItemOwned {
			if batch {
				continue
			}
			return p, &ShopError{Kind: "owned_help", Key: key}
		}
		if state == ItemLocked {
			return p, &ShopError{Kind: "locked_help", Key: key}
		}
		if item.Included {
			if batch && selected[item.Parent] {
				continue
			}
			return p, &ShopError{Kind: "included_error", Key: key}
		}
		if item.Parent != "" && e.State(item.Parent) != ItemOwned && !(batch && selected[item.Parent]) {
			return p, &ShopError{Kind: "missing_base", Key: key}
		}
		if !seen[key] {
			p.Lines = append(p.Lines, PurchaseLine{Key: key, Cost: e.Prices[key]})
			p.Total += e.Prices[key]
			seen[key] = true
		}
		for _, child := range pet.ShopItems {
			if child.Parent == key && child.Included {
				p.Lines = append(p.Lines, PurchaseLine{Key: child.Key, Included: true})
				seen[child.Key] = true
			}
		}
	}
	if len(p.Lines) == 0 {
		return p, &ShopError{Kind: "batch_owned"}
	}
	if p.Total > e.Balance {
		return p, &ShopError{Kind: "insufficient", Need: p.Total, Have: e.Balance}
	}
	b, _ := json.Marshal(struct {
		Balance int
		Lines   []PurchaseLine
	}{e.Balance, p.Lines})
	p.Stamp = string(b)
	return p, nil
}
func Buy(id string, keys []string, batch bool, stamp string) (Economy, error) {
	tx, err := db.Begin()
	if err != nil {
		return Economy{}, err
	}
	defer tx.Rollback()
	e, err := economy(tx, id)
	if err != nil {
		return e, err
	}
	p, err := Quote(e, keys, batch)
	if err != nil {
		return e, err
	}
	if stamp != p.Stamp {
		return e, &ShopError{Kind: "stale"}
	}
	for _, line := range p.Lines {
		if err = grantItem(tx, id, line.Key); err != nil {
			return e, err
		}
		e.States[line.Key] = ItemOwned
	}
	if _, err = tx.Exec(`UPDATE wallets SET balance=balance-? WHERE user_id=?`, p.Total, id); err != nil {
		return e, err
	}
	e.Balance -= p.Total
	return e, tx.Commit()
}

// Alterações da loja no dev: preços globais, estados individuais, tudo ou nada.
func DevShopSave(id string, states map[string]string, prices map[string]int, expected string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	e, err := economy(tx, id)
	if err != nil {
		return err
	}
	if EconomyStamp(e) != expected {
		return &ShopError{Kind: "stale"}
	}
	_, p, _, err := scanUser(tx.QueryRow(selectUser, id))
	if err != nil {
		return err
	}
	worn := map[string]bool{}
	for _, k := range pet.AppearanceItems(p.Look()) {
		worn[k] = true
	}
	for k, v := range states {
		if pet.ShopItemByKey(k) == nil || v != ItemOwned && v != ItemAvailable && v != ItemLocked {
			return errors.New("item/estado inválido")
		}
		if worn[k] && v != ItemOwned {
			return &ShopError{Kind: "dev_equipped_item_error"}
		}
		e.States[k] = v
	}
	for k, v := range states {
		if v == ItemOwned {
			for _, child := range pet.ShopItems {
				if child.Parent == k && child.Included {
					e.States[child.Key] = ItemOwned
				}
			}
		}
	}
	for _, item := range pet.ShopItems {
		if item.Included && e.State(item.Parent) == ItemOwned && e.State(item.Key) != ItemOwned {
			return &ShopError{Kind: "dev_included_variant_error"}
		}
	}
	for k := range states {
		if _, err = tx.Exec(`INSERT INTO inventory(user_id,item,state) VALUES (?,?,?) ON CONFLICT(user_id,item) DO UPDATE SET state=excluded.state`, id, k, e.States[k]); err != nil {
			return err
		}
	}
	for k, v := range states {
		if v == ItemOwned {
			if err = grantItem(tx, id, k); err != nil {
				return err
			}
		}
	}
	for k, v := range prices {
		if pet.ShopItemByKey(k) == nil || v < 0 || v > 9999 {
			return errors.New("preço inválido")
		}
		if _, err = tx.Exec(`INSERT INTO shop_prices(item,price) VALUES (?,?) ON CONFLICT(item) DO UPDATE SET price=excluded.price`, k, v); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func EconomyStamp(e Economy) string { b, _ := json.Marshal(e); return string(b) }
