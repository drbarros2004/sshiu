package pet

import (
	"testing"
	"time"
)

func TestMood(t *testing.T) {
	now := time.Date(2026, 10, 27, 14, 0, 0, 0, time.Local)
	p := Pet{Energy: 72, LastActive: now, Streak: 2, StreakDay: day(now)}
	if got := p.EnergyAt(now); got != 76 { // + bônus de streak (2 por dia)
		t.Fatalf("streak: %v", got)
	}
	later := now.AddDate(0, 0, 5) // 5 dias parado: decai e perde o streak
	if got := p.EnergyAt(later); got != 72-5*decayPerDay || p.StreakAt(later) != 0 {
		t.Fatalf("decay: %v", got)
	}
	if MoodOf(0) != Wilted || MoodOf(100) != Radiant || MoodOf(40) != Okay {
		t.Fatal("thresholds")
	}
}

func TestFocus(t *testing.T) {
	now := time.Date(2026, 10, 29, 14, 0, 0, 0, time.Local)
	p := Pet{Energy: EnergyStart, LastActive: now}

	p.Focus(30, false, now) // 30 min sem cumprir: só os minutos (8), e o dia já conta (>= 15 min)
	if p.Energy != 58 || p.Streak != 1 {
		t.Fatalf("30 min: %+v", p)
	}
	p.Focus(5, true, now) // cumprida com 5 min: sem bônus (só a partir de 10 min)
	if got := p.Energy - 58; got < 1.33 || got > 1.34 {
		t.Fatalf("5 min cumprida: +%v", got)
	}

	q := Pet{Energy: EnergyStart, LastActive: now}
	q.Focus(15, true, now) // cumprida com 15 min: 4 + bônus 3
	if q.Energy != 57 || q.Streak != 1 {
		t.Fatalf("15 min cumprida: %+v", q)
	}

	r := Pet{Energy: EnergyStart, LastActive: now}
	r.Focus(10, false, now) // 10 min ainda não contam o dia...
	if r.Streak != 0 {
		t.Fatalf("10 min: %+v", r)
	}
	r.Focus(5, false, now) // ...somando 15 no mesmo dia, conta
	if r.Streak != 1 {
		t.Fatalf("15 min no dia: %+v", r)
	}
	tomorrow := now.AddDate(0, 0, 1)
	r.Focus(15, false, tomorrow) // dia seguinte: streak sobe; a soma do dia recomeça
	if r.Streak != 2 || r.TodayMin != 15 {
		t.Fatalf("dia seguinte: %+v", r)
	}
}
