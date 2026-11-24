package rooms

import (
	"errors"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/store"
)

func ShopRead(id string) (store.Economy, error) {
	mu.Lock()
	defer mu.Unlock()
	return store.ReadEconomy(id)
}
func ShopBuy(id string, keys []string, batch bool, stamp string) (store.Economy, error) {
	mu.Lock()
	defer mu.Unlock()
	return store.Buy(id, keys, batch, stamp)
}
func ShopDevSave(id string, states map[string]string, prices map[string]int, stamp string) error {
	if !Dev {
		return errors.New("modo dev desativado")
	}
	mu.Lock()
	defer mu.Unlock()
	return store.DevShopSave(id, states, prices, stamp)
}
func EquipOwned(id string, a pet.Appearance) error {
	mu.Lock()
	defer mu.Unlock()
	e, err := store.ReadEconomy(id)
	if err != nil {
		return err
	}
	if !e.Owns(a) {
		return errors.New("aparência não possuída")
	}
	if err = store.SetAppearance(id, a); err != nil {
		return err
	}
	u := load(id)
	u.appearance = a
	u.color = pet.Pets.ArtworkColor(a)
	return nil
}

// O recibo pertence ao dono da tarefa; a TUI só o desenha para essa identidade.
func eventReceipt(u *user, t store.TaskRow) {
	n, ok := store.TaskCoins(u.id, t)
	for _, r := range rooms {
		if !slices.Contains(r.members, u.id) {
			continue
		}
		roomEvent(r, u, "cumpriu", t.Task)
		e := r.events[0]
		e.At = t.End
		if ok {
			v := n
			e.Coins = &v
		}
	}
}
