// Package pet é o bichinho de cada usuário (humor, energia, streak e os sprites).
// "pet" é um nome provisório: troque aqui, no pacote e nos comentários quando o nome definitivo sair.
package pet

// Humor do pet. Tudo que decide como ele reage está aqui: ajuste os números à vontade.

import "time"

type Result string

// Resultado de uma tarefa (CONTEXT.md).
const (
	Done        Result = "done"        // cumprida ("feito")
	NotDone     Result = "not_done"    // trocada ou parada
	Interrupted Result = "interrupted" // desconectou e não voltou a tempo: equivale a parar
)

const (
	EnergyStart    = 50.0 // energia de quem acabou de chegar (0..100)
	decayPerDay    = 8.0  // perda por dia inteiro sem focar
	streakBonus    = 2.0  // bônus por dia de streak...
	streakBonusMax = 10.0 // ...até este teto
)

// A energia vem dos minutos focados de verdade, não do tempo declarado (docs/regras.md).
const (
	energyPer30Min   = 8.0  // 30 min de foco devolvem um dia parado
	doneBonus        = 3.0  // tarefa cumprida...
	doneBonusMinutes = 10.0 // ...com pelo menos estes minutos focados
	streakMinutes    = 15.0 // minutos focados no dia para o dia contar no streak
)

// Focus soma ao pet uma tarefa encerrada: os minutos que contaram e se foi cumprida.
func (p *Pet) Focus(minutes float64, done bool, now time.Time) {
	e := p.decayed(now) + minutes*energyPer30Min/30
	if done && minutes >= doneBonusMinutes {
		e += doneBonus
	}
	p.Energy, p.LastActive = clamp(e), now
	if p.Today != day(now) {
		p.Today, p.TodayMin = day(now), 0
	}
	p.TodayMin += minutes
	if p.TodayMin < streakMinutes {
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
	Today      string // dia de TodayMin (YYYY-MM-DD)
	TodayMin   float64
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

func (p Pet) decayed(now time.Time) float64 {
	days := int(now.Sub(p.LastActive) / (24 * time.Hour))
	return clamp(p.Energy - float64(days)*decayPerDay)
}

func day(t time.Time) string  { return t.Format("2006-01-02") }
func clamp(e float64) float64 { return max(0, min(100, e)) }
