package pet

// Catálogo da loja: dados editáveis, separados das regras de saldo e da TUI.
import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"sshiu/assets"
)

type CoinRules struct {
	SecondsPerReward int    `json:"secondsPerReward"`
	CoinsPerReward   int    `json:"coinsPerReward"`
	Maximum          int    `json:"maximum"`
	InitialBalance   int    `json:"initialBalance"`
	Sprite           string `json:"sprite"`
	Color            string `json:"color"`
}
type ShopDialogConfig struct {
	MinimumWidth      int `json:"minimumWidth"`
	WidthPercent      int `json:"widthPercent"`
	PaddingHorizontal int `json:"paddingHorizontal"`
	PaddingVertical   int `json:"paddingVertical"`
	ItemIndent        int `json:"itemIndent"`
	IncludedIndent    int `json:"includedIndent"`
}

type ShopConfig struct {
	Confirmation ShopDialogConfig       `json:"confirmation"`
	Colors       map[string]string      `json:"colors"`
	Texts        map[string]DisplayName `json:"texts"`
	StarterItems []string               `json:"starterItems"`
}
type ShopItem struct {
	Key, Kind, Parent string
	Name              DisplayName
	Description       DisplayName
	Price             int
	Included          bool
}

var Coins CoinRules
var Shop ShopConfig
var CoinImage image.Image
var ShopItems = loadShop()

func loadShop() []ShopItem {
	read := func(path string, out any) {
		b, err := assets.Pets.ReadFile(path)
		if err == nil {
			err = json.Unmarshal(b, out)
		}
		if err != nil {
			panic(fmt.Errorf("%s: %w", path, err))
		}
	}
	read("coins/coins.json", &Coins)
	read("shop/shop.json", &Shop)
	layout := Shop.Confirmation
	if layout.MinimumWidth < 12 || layout.WidthPercent < 1 || layout.WidthPercent > 100 || layout.PaddingHorizontal < 0 || layout.PaddingVertical < 0 || layout.ItemIndent < 0 || layout.IncludedIndent < 0 {
		panic("shop.json: dimensões da confirmação inválidas")
	}

	// Zero é um preço explícito, não uma consequência de esquecer o campo.
	type pricedBase struct {
		ID      string
		Price   *int `json:"price"`
		Default *int `json:"defaultVariantPrice"`
	}
	for _, path := range []string{"pets/characters.json", "pets/hats.json"} {
		var file struct {
			Characters []pricedBase `json:"characters"`
			Hats       []pricedBase `json:"hats"`
		}
		read(path, &file)
		for _, base := range append(file.Characters, file.Hats...) {
			if base.Price == nil || base.Default == nil || *base.Price < 0 || *base.Price > 9999 || *base.Default < 0 || *base.Default > 9999 {
				panic(fmt.Errorf("%s: %s precisa de price e defaultVariantPrice de 0 a 9999", path, base.ID))
			}
		}
	}
	for key, text := range Shop.Texts {
		if text["pt"] == "" || text["en"] == "" {
			panic("shop.json: traduções ausentes: " + key)
		}
	}
	if Coins.SecondsPerReward <= 0 || Coins.CoinsPerReward <= 0 || Coins.CoinsPerReward > 9999 || Coins.Maximum < 1 || Coins.Maximum > 9999 || Coins.InitialBalance < 0 || Coins.InitialBalance > Coins.Maximum {
		panic("coins.json: taxa ou limites inválidos")
	}
	for _, v := range append([]string{Coins.Color}, shopColorValues()...) {
		if _, err := ParsePaletteColor(v); err != nil {
			panic(err)
		}
	}
	f, err := assets.Pets.Open(Coins.Sprite)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	CoinImage, err = png.Decode(f)
	if err != nil || CoinImage.Bounds().Dx() != 2 || CoinImage.Bounds().Dy() != 2 {
		panic("coins.json: sprite precisa ser PNG 2x2")
	}
	var items []ShopItem
	for _, c := range Pets.Characters {
		key := "character/" + c.ID
		items = append(items, ShopItem{Key: key, Kind: "character", Name: c.DisplayName, Description: c.Description, Price: c.Price})
		for i, v := range c.Variants {
			price := c.DefaultVariantPrice
			if v.Price != nil {
				price = *v.Price
			}
			if i == 0 {
				price = 0
			}
			items = append(items, ShopItem{Key: "color/" + c.ID + "/" + v.ID, Kind: "color", Parent: key, Name: v.DisplayName, Description: v.Description, Price: price, Included: i == 0})
		}
	}
	for _, h := range Pets.Hats {
		key := "hat/" + h.ID
		items = append(items, ShopItem{Key: key, Kind: "hat", Name: h.DisplayName, Description: h.Description, Price: h.Price})
		for i, v := range h.Variants {
			price := h.DefaultVariantPrice
			if v.Price != nil {
				price = *v.Price
			}
			if i == 0 {
				price = 0
			}
			items = append(items, ShopItem{Key: "model/" + h.ID + "/" + v.ID, Kind: "model", Parent: key, Name: v.DisplayName, Description: v.Description, Price: price, Included: i == 0})
		}
	}
	seen := map[string]bool{}
	for _, item := range items {
		if seen[item.Key] || item.Price < 0 || item.Price > 9999 {
			panic("loja: ID duplicado ou preço inválido: " + item.Key)
		}
		seen[item.Key] = true
	}
	for _, key := range Shop.StarterItems {
		if !seen[key] {
			panic("shop.json: item inicial desconhecido: " + key)
		}
	}
	return items
}
func shopColorValues() []string {
	var values []string
	for _, v := range Shop.Colors {
		values = append(values, v)
	}
	return values
}
func ShopItemByKey(key string) *ShopItem {
	for i := range ShopItems {
		if ShopItems[i].Key == key {
			return &ShopItems[i]
		}
	}
	return nil
}
func AppearanceItems(a Appearance) []string {
	keys := []string{"character/" + a.Character, "color/" + a.Character + "/" + a.Variant}
	if a.Hat != "" {
		keys = append(keys, "hat/"+a.Hat, "model/"+a.Hat+"/"+a.HatVariant)
	}
	return keys
}

// CoinReward não muda o saldo: permite usar o mesmo cálculo na prévia e ao salvar.
func CoinReward(balance, remainder, seconds int) (credited, nextRemainder int) {
	seconds = max(0, seconds) + max(0, remainder)%Coins.SecondsPerReward
	return min(seconds/Coins.SecondsPerReward*Coins.CoinsPerReward, max(0, Coins.Maximum-balance)), seconds % Coins.SecondsPerReward
}
