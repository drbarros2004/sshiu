package tui

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"sshiu/internal/pet"
)

// O recorde conta os dias com 15 min ou mais desde 2026 e nunca fica abaixo do streak do pet.
func TestRecord(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.Local)
	days := map[string]float64{}
	for d := 2; d <= 6; d++ { // 5 dias seguidos
		days[dayKey(time.Date(2026, 3, d, 0, 0, 0, 0, time.Local))] = 20
	}
	days["2026-03-07"] = 14 // abaixo dos 15: quebra
	days["2026-03-18"], days["2026-03-19"] = 15, 30
	m := model{now: now}
	if m.rec, m.recAt = longestRun(days, now); m.rec != 5 || m.recAt.Day() != 6 {
		t.Fatalf("recorde %d em %v", m.rec, m.recAt)
	}
	m.pet = pet.Pet{Streak: 9, StreakDay: "2026-03-19"} // o streak do pet vale mais que a conta
	if rec, _ := m.record(); rec != max(5, m.pet.StreakAt(now)) {
		t.Fatalf("recorde abaixo do streak: %d", rec)
	}
}

// No mês corrente, a comparação com o anterior vai até o mesmo dia; janeiro de 2026 não tem anterior.
func TestMonthRows(t *testing.T) {
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local)
	m := model{now: now, st: newStyles(themes[0]), days: map[string]float64{
		"2026-02-05": 30, "2026-02-20": 600, "2026-03-04": 90, "2026-03-09": 10,
	}}
	rows := m.monthRows(monthOf(now))
	got := func(i int) string { return ansi.Strip(rows[i][1]) }
	if got(0) != "1h 40" || !strings.HasPrefix(got(1), "2 de 10") || !strings.HasPrefix(got(2), "qua 4") ||
		!strings.HasPrefix(got(3), "+1h 10") { // 100 min contra os 30 até o dia 10 de fevereiro
		t.Fatalf("março: %q", rows)
	}
	if rows = m.monthRows(statsFrom); ansi.Strip(rows[3][1]) != "—" {
		t.Fatalf("janeiro sem mês anterior: %q", rows[3])
	}
	if m = m.shiftMonth(1); !m.shownMonth().Equal(monthOf(now)) { // não passa do mês de agora
		t.Fatal("andou para o futuro")
	}
	for range 5 {
		m = m.shiftMonth(-1)
	}
	if !m.shownMonth().Equal(statsFrom) { // nem de janeiro de 2026
		t.Fatalf("voltou para %v", m.shownMonth())
	}
	if m = m.shiftMonth(1).shiftMonth(1); !m.smonth.IsZero() { // de volta a março: acompanha a virada do mês
		t.Fatalf("preso em %v", m.smonth)
	}
}

// stack: as seções inteiras enquanto cabem, e as tarefas com quantas linhas couberem.
func TestStack(t *testing.T) {
	task := func(k int) string { return "t" + strings.Repeat("\nx", k) }
	if got := stack(6, "\n", []string{"a", "b\nb", ""}, task, 4); got != "a\nb\nb\nt\nx\nx" {
		t.Fatalf("%q", got)
	}
	if got := stack(5, "\n", []string{"a", "b\nb", "c\nc\nc"}, task, 4); got != "a\nb\nb" { // o que não cabe leva junto o que vem depois
		t.Fatalf("%q", got)
	}
	if got := stack(3, "\n", []string{"a\na\na\na"}, task, 4); got != "a\na\na\na" { // a primeira sempre entra
		t.Fatalf("%q", got)
	}
}

// O seletor traz o último tempo no lugar dele e o "outro" no fim; o campo aceita de 5 a 120.
func TestChoicesAndOther(t *testing.T) {
	m := model{lastMin: 50}
	if cs := m.choices(); !slices.Equal(cs, []int{5, 10, 15, 20, 25, 30, 45, 50, 60, 0}) {
		t.Fatalf("com 50: %v", cs)
	}
	if m = m.openChoice(false); m.choices()[m.mins] != 50 {
		t.Fatal("não abriu no último tempo")
	}
	if m.lastMin = 0; m.openChoice(false).choices()[m.openChoice(false).mins] != 25 {
		t.Fatal("sem último tempo, o padrão é 25")
	}
	m.lastMin = 25
	if cs := m.choices(); len(cs) != 9 { // o 25 já está na lista: não repete
		t.Fatalf("com 25: %v", cs)
	}
	m.other = true
	for _, tc := range []struct{ typed, err string }{{"", "digite os minutos"}, {"3", "o mínimo é 5 min"}, {"121", "o máximo é 120 min"}} {
		m.otherMin = tc.typed
		next, _ := m.otherKey("enter")
		if got := next.(model).err; got != tc.err {
			t.Fatalf("%q: erro %q", tc.typed, got)
		}
	}
	next, _ := model{other: true, otherMin: "7"}.otherKey("a") // só algarismos
	if next, _ = next.(model).otherKey("5"); next.(model).otherMin != "75" {
		t.Fatalf("digitado %q", next.(model).otherMin)
	}
}

// O calendário cabe nas suas 34 colunas em toda linha, em qualquer mês.
func TestCalendarWidth(t *testing.T) {
	m := model{now: time.Date(2026, 11, 13, 9, 0, 0, 0, time.Local), st: newStyles(themes[0]), days: map[string]float64{}}
	for mo := statsFrom; !mo.After(m.now); mo = mo.AddDate(0, 1, 0) {
		for _, l := range strings.Split(m.calendar(mo), "\n") {
			if w := ansi.StringWidth(l); w > calCols {
				t.Fatalf("%v: linha de %d colunas: %q", mo.Month(), w, ansi.Strip(l))
			}
		}
	}
}
