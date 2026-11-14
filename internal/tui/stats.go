package tui

// As stats do foco (#45): o calendário do mês em blocos de calor, os números do mês, a hora do dia
// em que você foca, o quanto você cumpre em cada faixa de tempo declarado e o recorde de streak.
// Aparecem na aba perfil e na coluna da sala solo parada, em painéis (o padrão) ou em lista, como
// cada um escolhe na aba visual. Tudo sai de store.FocusByDay, FocusByHour e GetTotals, lidos com o pet.

import (
	"fmt"
	"image/color"
	"math"
	"sshiu/internal/pet"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	barCols  = 16        // largura das barras da semana e das faixas
	calCols  = 34        // largura do calendário: 7 dias de 4 colunas, com 1 entre eles
	onAccent = "#11111b" // o texto em cima da cor de destaque (aba ativa, dia cheio no calendário)
)

// bandLabels: as faixas do tempo declarado, na ordem de store.Totals.Bands (docs/regras.md).
var bandLabels = [3][2]string{{"curtas", "até 15 min"}, {"médias", "até 30 min"}, {"longas", "mais de 30"}}

// statsFrom: o primeiro mês do calendário. O histórico do sshiu começa em 2026.
var statsFrom = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Local)

var (
	weekdays = [7]string{"seg", "ter", "qua", "qui", "sex", "sáb", "dom"}
	months   = [12]string{"janeiro", "fevereiro", "março", "abril", "maio", "junho",
		"julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
)

// weekIndex: segunda é 0 e domingo é 6; o time.Weekday começa no domingo.
func weekIndex(t time.Time) int { return (int(t.Weekday()) + 6) % 7 }

func dayKey(t time.Time) string { return t.Format("2006-01-02") }

func monthOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// dur: "45 min", "6h 20" ou "1h" na hora redonda.
func dur(min float64) string {
	switch {
	case min < 60:
		return fmt.Sprintf("%d min", int(min))
	case int(min)%60 == 0:
		return fmt.Sprintf("%dh", int(min)/60)
	}
	return fmt.Sprintf("%dh %02d", int(min)/60, int(min)%60)
}

// blend: a cor entre a e b ("#rrggbb"), t de 0 (a) a 1 (b).
func blend(a, b string, t float64) string {
	ch := func(h string, i int) float64 { v, _ := strconv.ParseUint(h[1+2*i:3+2*i], 16, 8); return float64(v) }
	out := "#"
	for i := range 3 {
		out += fmt.Sprintf("%02x", int(math.Round(ch(a, i)+(ch(b, i)-ch(a, i))*t)))
	}
	return out
}

// shownMonth: o mês do calendário do perfil; o zero é o mês de agora.
func (m model) shownMonth() time.Time {
	if m.smonth.IsZero() {
		return monthOf(m.now)
	}
	return m.smonth
}

// shiftMonth: ←/→ no perfil, de janeiro de 2026 até o mês de agora.
func (m model) shiftMonth(step int) model {
	if next := m.shownMonth().AddDate(0, step, 0); !next.Before(statsFrom) && !next.After(m.now) {
		m.smonth = next
		if next.Equal(monthOf(m.now)) { // de volta ao mês de agora: o zero acompanha a virada do mês
			m.smonth = time.Time{}
		}
	}
	return m
}

// heat: a cor do dia no calendário, do fundo apagado do tema (sem foco) ao destaque (1 h ou mais).
func (m model) heat(min float64) lipgloss.Style {
	th, c := themes[m.theme], lipgloss.Color
	st := lipgloss.NewStyle()
	switch {
	case min >= 60:
		return st.Background(c(th.accent)).Foreground(c(onAccent))
	case min >= 25:
		return st.Background(c(blend(th.off, th.accent, 0.6))).Foreground(c(onAccent))
	case min > 0:
		return st.Background(c(blend(th.off, th.accent, 0.3)))
	}
	return st.Background(c(th.off)).Foreground(c(th.muted))
}

// monthTitle: o nome do mês, com as setas para os outros meses quando nav (só no perfil).
func (m model) monthTitle(month time.Time, nav bool) string {
	name := bold.Render(m.t(months[month.Month()-1]) + " " + strconv.Itoa(month.Year()))
	if !nav {
		return name
	}
	l, r := "‹ ", " ›"
	if !month.After(statsFrom) {
		l = "  "
	}
	if !month.Before(monthOf(m.now)) {
		r = "  "
	}
	return m.st.muted.Render(l) + name + m.st.muted.Render(r)
}

// calendar: o mês em blocos, mais acesos quanto mais tempo focado. Hoje fica sublinhado, para você
// se achar no mês, e os dias que ainda não chegaram ficam apagados.
func (m model) calendar(month time.Time) string {
	var b strings.Builder
	for i, d := range weekdays {
		b.WriteString(m.st.muted.Render(fmt.Sprintf(" %2s ", m.t(d)[:1])) + map[bool]string{true: "\n", false: " "}[i == 6])
	}
	b.WriteString(strings.Repeat("     ", weekIndex(month)))
	for d := month; d.Month() == month.Month(); d = d.AddDate(0, 0, 1) {
		cell := m.st.muted.Faint(true).Render(fmt.Sprintf(" %2d ", d.Day()))
		if !d.After(m.now) {
			st := m.heat(m.days[dayKey(d)])
			if dayKey(d) == dayKey(m.now) {
				st = st.Bold(true).Underline(true)
			}
			cell = st.Render(fmt.Sprintf(" %2d ", d.Day()))
		}
		// o domingo fecha a linha sem o espaço do vão: com ele, a linha passava das 34 colunas
		b.WriteString(cell + map[bool]string{true: "\n", false: " "}[weekIndex(d) == 6])
	}
	return strings.TrimRight(b.String(), " \n")
}

// heatLegend: "menos ▪▪▪▪ mais", na ordem dos quatro níveis do calendário.
func (m model) heatLegend() string {
	out := m.st.muted.Render(m.t("menos") + " ")
	for _, min := range []float64{0, 1, 25, 60} {
		out += m.heat(min).Render("  ") + " "
	}
	return out + m.st.muted.Render(m.t("mais"))
}

// monthRows: os números do mês. No mês corrente, a comparação com o anterior vai até o mesmo dia:
// meio mês contra o anterior inteiro enganava.
func (m model) monthRows(month time.Time) [][2]string {
	var total, best, prev float64
	var bestDay time.Time
	active, days := 0, 0
	current := month.Equal(monthOf(m.now))
	for d := month; d.Month() == month.Month() && !d.After(m.now); d = d.AddDate(0, 0, 1) {
		v := m.days[dayKey(d)]
		total, days = total+v, days+1
		if v > 0 {
			active++
		}
		if v > best {
			best, bestDay = v, d
		}
	}
	p := month.AddDate(0, -1, 0)
	for d := p; d.Month() == p.Month() && (!current || d.Day() <= m.now.Day()); d = d.AddDate(0, 0, 1) {
		prev += m.days[dayKey(d)]
	}
	of, bestS, diff := m.tf("%d de %d", active, days), m.st.muted.Render("—"), m.st.muted.Render("—")
	if current {
		of += m.st.muted.Render(m.t(" até hoje"))
	}
	if best > 0 {
		bestS = m.t(weekdays[weekIndex(bestDay)]) + fmt.Sprintf(" %d", bestDay.Day()) + m.st.muted.Render(" · ") + dur(best)
	}
	if !p.Before(statsFrom) {
		sign, d := "+", total-prev
		if d < 0 {
			sign, d = "−", -d
		}
		diff = sign + dur(d)
		if current {
			diff += m.st.muted.Render(m.tf(" até o dia %d", m.now.Day()))
		}
	}
	return [][2]string{
		{m.t("focado"), m.st.title.Render(dur(total))},
		{m.t("dias com foco"), of},
		{m.t("melhor dia"), bestS},
		{m.t("mês passado"), diff},
	}
}

// rows: rótulo apagado numa coluna de w, valor ao lado, como a saída de um comando.
func (m model) rows(rs [][2]string, w int) string {
	var b strings.Builder
	for _, r := range rs {
		b.WriteString(m.st.muted.Render(r[0]+strings.Repeat(" ", max(1, w-ansi.StringWidth(r[0])))) + r[1] + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// longestRun: a maior sequência de dias com 15 min ou mais desde 2026 e o dia em que ela terminou.
// Percorre o ano inteiro, então sai uma vez no loadProfile, e não a cada quadro.
func longestRun(days map[string]float64, now time.Time) (best int, end time.Time) {
	run := 0
	for d := statsFrom; !d.After(now); d = d.AddDate(0, 0, 1) {
		run++
		if days[dayKey(d)] < pet.StreakMinutes {
			run = 0
		}
		if run > best {
			best, end = run, d
		}
	}
	return best, end
}

// record: o recorde de streak, que nunca fica abaixo do streak de agora (o do pet).
func (m model) record() (int, time.Time) {
	if s := m.pet.StreakAt(m.now); s > m.rec {
		return s, m.now
	}
	return m.rec, m.recAt
}

// recordText: "11 em março", o recorde e o mês em que ele terminou; de outro ano, o ano no lugar do
// mês ("11 em 2026"), que é mais curto que os dois juntos e cabe na linha do streak em 80 colunas.
// Sem recorde, só o 0.
func (m model) recordText() string {
	rec, at := m.record()
	if rec == 0 {
		return m.st.title.Render("0")
	}
	when := m.t(months[at.Month()-1])
	if at.Year() != m.now.Year() {
		when = strconv.Itoa(at.Year())
	}
	return m.st.title.Render(strconv.Itoa(rec)) + m.st.muted.Render(" "+m.tf("em %s", when))
}

// streakDays: "4 dias seguidos", o streak de agora (o do pet).
func (m model) streakDays() string {
	n := m.pet.StreakAt(m.now)
	return fmt.Sprintf("%d %s", n, plural(n, m.t("dia seguido"), m.t("dias seguidos")))
}

// streakText: "4 dias seguidos · recorde 11 em março".
func (m model) streakText() string {
	return m.streakDays() + m.st.muted.Render(m.t(" · recorde ")) + m.recordText()
}

// spark: os minutos focados em cada hora do dia numa linha, cada hora com cell colunas, e as horas
// embaixo. Vazio enquanto não há foco.
func (m model) spark(cell int) string {
	top := 0.0
	for _, v := range m.hours {
		top = max(top, v)
	}
	if top == 0 {
		return ""
	}
	th := themes[m.theme]
	bars := []rune("▁▂▃▄▅▆▇█")
	var b strings.Builder
	for _, v := range m.hours {
		ch := " "
		if v > 0 {
			ch = string(bars[int(math.Round(v/top*7))])
		}
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(blend(th.muted, th.accent, 0.25+0.75*v/top))).Render(strings.Repeat(ch, cell)))
	}
	axis := ""
	for _, h := range []int{0, 6, 12, 18} {
		axis += fmt.Sprintf("%-*s", 6*cell, fmt.Sprintf("%dh", h))
	}
	return b.String() + "\n" + m.st.muted.Render(strings.TrimRight(axis, " "))
}

// peak: as duas horas seguidas com mais foco ("9h às 11h") e a divisão em manhã (5h às 12h),
// tarde (12h às 18h) e noite (o resto).
func (m model) peak() (string, string) {
	best, at, total := 0.0, 0, 0.0
	var part [3]float64
	for h, v := range m.hours {
		if w := v + m.hours[(h+1)%24]; w > best {
			best, at = w, h
		}
		total += v
		switch {
		case h >= 5 && h < 12:
			part[0] += v
		case h >= 12 && h < 18:
			part[1] += v
		default:
			part[2] += v
		}
	}
	pct := func(v float64) string { return fmt.Sprintf("%.0f%%", 100*v/total) } // hoursBody só chama com foco
	sep := m.st.muted.Render(" · ")
	return m.tf("%dh às %dh", at, (at+2)%24),
		m.t("manhã") + " " + pct(part[0]) + sep + m.t("tarde") + " " + pct(part[1]) + sep + m.t("noite") + " " + pct(part[2])
}

// hoursBody: o horário com o pico e a divisão do dia, para o painel e para a lista. Vazio sem
// nenhum minuto focado: as tarefas encerradas em 0 min não têm pico a mostrar.
func (m model) hoursBody() string {
	sp := m.spark(2)
	if sp == "" {
		return ""
	}
	pk, split := m.peak()
	return sp + "\n" + m.st.muted.Render(m.t("pico ")) + m.st.title.Render(pk) + "   " + split
}

// bandsBody: quanto você cumpre em cada faixa de tempo declarado, a média e o total.
func (m model) bandsBody() string {
	var b strings.Builder
	for i, bd := range m.totals.Bands {
		b.WriteString(m.st.muted.Render(fmt.Sprintf("%-8s", m.t(bandLabels[i][0]))+fmt.Sprintf("%-12s", m.t(bandLabels[i][1]))) + " ")
		if bd[1] == 0 {
			b.WriteString(m.st.muted.Render(m.t("nenhuma ainda")) + "\n")
			continue
		}
		f := float64(bd[0]) / float64(bd[1])
		n, fill := int(math.Round(f*barCols)), m.st.accent
		if f < 0.5 {
			fill = m.st.muted.GetForeground()
		}
		b.WriteString(m.solid(fill, n) + fmt.Sprintf(" %3.0f%%", 100*f) + m.st.muted.Render(" "+m.tf("%d de %d", bd[0], bd[1])) + "\n")
	}
	t := m.totals
	foca := strconv.Itoa(int(t.Focused.Minutes()) / t.Tasks)
	if t.Focused.Minutes() < float64(t.Tasks) { // a média abaixo de 1 min não é zero
		foca = "<1"
	}
	b.WriteString(m.st.muted.Render(m.t("em média declara ")) + fmt.Sprintf("%d min", int(t.Declared.Minutes())/t.Tasks) +
		m.st.muted.Render(m.t(" e foca ")) + foca + "\n" +
		m.st.muted.Render(m.t("no total ")) + dur(t.Focused.Minutes()) + m.st.muted.Render(m.tf(", em %d tarefas", t.Tasks)))
	return b.String()
}

// solid: uma barra de barCols células, as n primeiras em fill e o resto na cor apagada do tema.
func (m model) solid(fill color.Color, n int) string {
	return lipgloss.NewStyle().Foreground(fill).Render(strings.Repeat("█", n)) +
		lipgloss.NewStyle().Foreground(m.st.off).Render(strings.Repeat("█", barCols-n))
}

// weekBars: os minutos de cada dia desta semana, da segunda até hoje.
func (m model) weekBars() string {
	monday := m.now.AddDate(0, 0, -weekIndex(m.now))
	var top float64
	for i := 0; i <= weekIndex(m.now); i++ {
		top = max(top, m.days[dayKey(monday.AddDate(0, 0, i))])
	}
	var b strings.Builder
	for i := 0; i <= weekIndex(m.now); i++ {
		min := m.days[dayKey(monday.AddDate(0, 0, i))]
		n := 0
		if top > 0 {
			n = int(float64(barCols) * min / top)
		}
		b.WriteString(m.st.muted.Render(fmt.Sprintf("%-5s", m.t(weekdays[i]))) + m.solid(m.st.accent, n) + m.st.muted.Render(fmt.Sprintf(" %7s", dur(min))) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// bar: uma barra de cols células com as n primeiras acesas (a tabela de foco).
func (m model) bar(n, cols int) string {
	return lipgloss.NewStyle().Foreground(m.st.accent).Render(strings.Repeat("█", n)) +
		m.st.muted.Render(strings.Repeat("░", cols-n))
}

// box: um painel de w colunas e pelo menos h linhas por dentro, com o título na borda de cima,
// como num monitor de terminal. O que passa da largura é cortado com "…". Sem corpo, não há painel
// (a seção some, e o stack a pula).
func (m model) box(title, body string, w, h int) string {
	if body == "" {
		return ""
	}
	bc := lipgloss.NewStyle().Foreground(lipgloss.Color(themes[m.theme].border))
	t := " " + title + " "
	lines := strings.Split(body, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	var b strings.Builder
	b.WriteString(bc.Render("╭─") + t + bc.Render(strings.Repeat("─", max(0, w-3-ansi.StringWidth(t)))+"╮") + "\n")
	for _, l := range lines {
		l = ansi.Truncate(l, w-4, "…")
		b.WriteString(bc.Render("│ ") + l + strings.Repeat(" ", max(0, w-4-ansi.StringWidth(l))) + bc.Render(" │") + "\n")
	}
	b.WriteString(bc.Render("╰" + strings.Repeat("─", w-2) + "╯"))
	return b.String()
}

// stack: as seções na ordem, inteiras, enquanto cabem em h linhas; a primeira sempre entra. As
// últimas tarefas vêm por último, com quantas linhas couberem (tasks(k) monta a seção com k),
// nunca só o título. A ordem é a prioridade: o que vem depois é o que sai primeiro da tela.
func stack(h int, sep string, secs []string, tasks func(k int) string, n int) string {
	out := secs[0]
	for _, s := range secs[1:] {
		if s == "" {
			continue
		}
		if lipgloss.Height(out+sep+s) > h {
			return out
		}
		out += sep + s
	}
	for k := n; k >= 1; k-- {
		if t := out + sep + tasks(k); lipgloss.Height(t) <= h {
			return t
		}
	}
	return out
}

// soloStats: a coluna da direita da sala solo parada, em até w colunas: o mês, a semana e a
// constância, em painéis ou em lista.
func (m model) soloStats(w int) string {
	month := monthOf(m.now)
	monday := m.now.AddDate(0, 0, -weekIndex(m.now))
	week, total := 0.0, 0.0
	for d := monday; !d.After(m.now); d = d.AddDate(0, 0, 1) {
		week += m.days[dayKey(d)]
	}
	for d := month; !d.After(m.now); d = d.AddDate(0, 0, 1) {
		total += m.days[dayKey(d)]
	}
	stats := m.rows([][2]string{
		{m.t("streak"), m.streakDays()},
		{m.t("recorde"), m.recordText()},
		{m.t("hoje"), dur(m.days[dayKey(m.now)])},
		{m.t("semana"), dur(week)},
		{m.t("mês"), dur(total)},
	}, 9)
	if m.statsList || w < calCols+4 {
		return m.monthTitle(month, false) + "\n" + m.calendar(month) + "\n\n" +
			bold.Render(m.t("esta semana")) + "\n" + m.weekBars() + "\n\n" + stats
	}
	bw := calCols + 4
	return lipgloss.JoinVertical(lipgloss.Left,
		m.box(m.monthTitle(month, false), m.calendar(month), bw, 0),
		m.box(m.st.title.Render(m.t("esta semana")), m.weekBars(), bw, 0),
		m.box(m.st.title.Render(m.t("constância")), stats, bw, 0))
}
