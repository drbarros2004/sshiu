package tui

import (
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"strconv"
	"strings"
)

func (m model) walletBalance() int {
	if d := m.inline(); d != nil {
		if d.field == "coins" {
			return int(d.number)
		}
		for _, c := range d.changes {
			if c.ID == m.id && c.Coins != nil {
				return *c.Coins
			}
		}
	}
	return m.economy.Balance
}
func (m model) coinNumber(key string, n int) string {
	if n == 1 && pet.Shop.Texts[key+"_one"] != nil {
		key += "_one"
	}
	return strings.Replace(m.shopText(key, n), strconv.Itoa(n), shopStyle("coin").Render(strconv.Itoa(n)), 1)
}
func (m model) coinGain(f rooms.Focus) string {
	if !m.economyReady {
		return ""
	}
	seconds := int(f.Focused(m.now).Seconds())
	n, _ := pet.CoinReward(m.walletBalance(), m.economy.Remainder, seconds)
	text := " " + m.coinNumber("gain", n)
	uncapped := (seconds + m.economy.Remainder) / pet.Coins.SecondsPerReward * pet.Coins.CoinsPerReward
	if n < uncapped {
		text += m.shopText("gain_limited", pet.Coins.Maximum)
	}
	return text
}
func (m model) receiptSuffix(e rooms.Event) string {
	if e.Who == m.id && e.Coins != nil {
		return " " + m.coinNumber("receipt", *e.Coins)
	}
	return ""
}
