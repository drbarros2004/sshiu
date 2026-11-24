package pet

import "testing"

func TestCoinReward(t *testing.T) {
	old := Coins
	t.Cleanup(func() { Coins = old })
	Coins = CoinRules{SecondsPerReward: 60, CoinsPerReward: 1, Maximum: 9999}
	for _, tc := range []struct {
		name                                string
		balance, carry, seconds, want, rest int
	}{
		{"short_task", 0, 0, 23, 0, 23},
		{"carry_completes_minute", 0, 23, 37, 1, 0},
		{"multiple_intervals", 0, 0, 263, 4, 23},
		{"cap_clips_reward", 9998, 0, 180, 1, 0},
		{"cap_discards_reward", 9999, 0, 180, 0, 0},
		{"rate_change_discards_old_intervals", 4, 112, 0, 0, 52},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n, r := CoinReward(tc.balance, tc.carry, tc.seconds)
			if n != tc.want || r != tc.rest {
				t.Fatalf("recebido %d moedas e %ds, esperado %d e %ds", n, r, tc.want, tc.rest)
			}
		})
	}
	Coins.SecondsPerReward = 180
	if n, r := CoinReward(4, 23, 157); n != 1 || r != 0 {
		t.Fatalf("nova taxa: %d moedas, %ds", n, r)
	}
}
