// Package pet é o bichinho de cada pessoa (humor, energia, streak e os sprites).
// "pet" é um nome provisório: troque aqui, no pacote e nos comentários quando o nome definitivo sair.
package pet

// Humor do pet. Tudo que decide como ele reage está aqui: ajuste os números à vontade.

import "time"

type Result string

const (
	Done      Result = "done"      // ficou até o fim e cumpriu o Lock
	NotDone   Result = "not_done"  // ficou até o fim, mas não cumpriu
	Abandoned Result = "abandoned" // saiu no meio do foco
	Missed    Result = "missed"    // estava na fila e faltou
)

const (
	EnergyStart    = 50.0 // energia de quem acabou de chegar (0..100)
	decayPerDay    = 8.0  // perda por dia inteiro sem nenhum sprint
	streakBonus    = 2.0  // bônus por dia de streak...
	streakBonusMax = 10.0 // ...até este teto
)

var delta = map[Result]float64{Done: 15, NotDone: 4, Abandoned: -15, Missed: -10}

type Mood int

const (
	Wilted Mood = iota
	Tired
	Okay
	Good
	Radiant
)

var moodMin = [...]float64{Wilted: 0, Tired: 20, Okay: 40, Good: 60, Radiant: 80} // energia mínima de cada estado

func MoodOf(e float64) Mood {
	for m := Radiant; m > Wilted; m-- {
		if e >= moodMin[m] {
			return m
		}
	}
	return Wilted
}

type Pet struct {
	Energy     float64 // energia guardada, sem decaimento nem bônus
	LastActive time.Time
	Streak     int
	StreakDay  string // último dia que contou pro streak (YYYY-MM-DD)
}

// EnergyAt é o que o pet mostra agora: energia com decaimento + bônus do streak.
func (p Pet) EnergyAt(now time.Time) float64 {
	return clamp(p.decayed(now) + min(float64(p.StreakAt(now))*streakBonus, streakBonusMax))
}

func (p Pet) StreakAt(now time.Time) int {
	if p.StreakDay == day(now) || p.StreakDay == day(now.AddDate(0, 0, -1)) {
		return p.Streak
	}
	return 0
}

func (p *Pet) Apply(r Result, now time.Time) {
	p.Energy = clamp(p.decayed(now) + delta[r])
	p.LastActive = now
	if r != Done && r != NotDone { // só quem ficou até o fim conta pro streak
		return
	}
	switch p.StreakDay {
	case day(now):
	case day(now.AddDate(0, 0, -1)):
		p.Streak++
	default:
		p.Streak = 1
	}
	p.StreakDay = day(now)
}

func (p Pet) decayed(now time.Time) float64 {
	days := int(now.Sub(p.LastActive) / (24 * time.Hour))
	return clamp(p.Energy - float64(days)*decayPerDay)
}

func day(t time.Time) string  { return t.Format("2006-01-02") }
func clamp(e float64) float64 { return max(0, min(100, e)) }
