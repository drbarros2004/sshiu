package pet

import (
	"testing"
	"time"
)

func TestMood(t *testing.T) {
	now := time.Date(2026, 10, 27, 14, 0, 0, 0, time.Local)
	p := Pet{Energy: EnergyStart, LastActive: now}

	p.Apply(Done, now)
	if p.Energy != 65 || p.Streak != 1 || MoodOf(p.EnergyAt(now)) != Good {
		t.Fatalf("done: %+v", p)
	}
	p.Apply(Done, now.AddDate(0, 0, 1)) // dia seguinte: decai 1 dia, ganha 15, streak sobe
	if p.Streak != 2 || p.Energy != 72 {
		t.Fatalf("streak: %+v", p)
	}
	later := now.AddDate(0, 0, 5) // 4 dias parado: decai e perde o streak
	if got := p.EnergyAt(later); got != 72-4*decayPerDay {
		t.Fatalf("decay: %v", got)
	}
	p.Apply(Abandoned, later)
	if p.StreakAt(later) != 0 || MoodOf(p.EnergyAt(later)) != Tired {
		t.Fatalf("abandoned: %+v", p)
	}
	if MoodOf(0) != Wilted || MoodOf(100) != Radiant || MoodOf(40) != Okay {
		t.Fatal("thresholds")
	}
}
