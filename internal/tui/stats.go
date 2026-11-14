package tui

// Os gráficos do foco: o calendário do mês, as barras da semana e os números das tarefas.
// Aparecem na aba perfil e na coluna da sala solo enquanto não há tarefa.
// Tudo sai de duas consultas (store.FocusByDay e store.GetTotals), carregadas com o pet.

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

const barCols = 20 // largura das barras da semana

var (
	weekdays = [7]string{"seg", "ter", "qua", "qui", "sex", "sáb", "dom"}
	months   = [12]string{"janeiro", "fevereiro", "março", "abril", "maio", "junho",
		"julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
)

// weekIndex: segunda é 0 e domingo é 6; o time.Weekday começa no domingo.
func weekIndex(t time.Time) int { return (int(t.Weekday()) + 6) % 7 }

func dayKey(t time.Time) string { return t.Format("2006-01-02") }

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

// statsBlock: o que fica fixo na tela quando não há tarefa. As barras da semana só entram onde
// sobra altura (a coluna da sala solo); na aba perfil elas empurram as últimas tarefas para fora.
func (m model) statsBlock(week bool) string {
	parts := []string{m.calendar(), m.taskNumbers()}
	if week {
		parts = []string{m.calendar(), m.weekBars(), m.taskNumbers()}
	}
	return strings.Join(parts, "\n\n")
}

// dayStyle: quanto mais tempo focado no dia, mais aceso ele fica.
func (m model) dayStyle(min float64) lipgloss.Style {
	switch {
	case min >= 60:
		return m.st.title // o dia cheio, em destaque e negrito
	case min >= 25:
		return lipgloss.NewStyle().Foreground(m.st.accent)
	case min > 0:
		return lipgloss.NewStyle()
	}
	return m.st.muted
}

// calendar: o mês inteiro, um número por dia, aceso conforme o tempo focado.
func (m model) calendar() string {
	first := time.Date(m.now.Year(), m.now.Month(), 1, 0, 0, 0, 0, m.now.Location())
	var b strings.Builder
	b.WriteString(bold.Render(m.t(months[int(m.now.Month())-1])) + "\n")
	for _, d := range weekdays {
		b.WriteString(m.st.muted.Render(fmt.Sprintf("%2s ", m.t(d)[:1])))
	}
	b.WriteString("\n" + strings.Repeat("   ", weekIndex(first)))
	for d := first; d.Month() == first.Month(); d = d.AddDate(0, 0, 1) {
		cell := m.dayStyle(m.days[dayKey(d)]).Render(fmt.Sprintf("%2d", d.Day()))
		if dayKey(d) == dayKey(m.now) { // hoje fica sublinhado, para você se achar no mês
			cell = m.dayStyle(m.days[dayKey(d)]).Underline(true).Render(fmt.Sprintf("%2d", d.Day()))
		}
		b.WriteString(cell + " ")
		if weekIndex(d) == 6 {
			b.WriteString("\n")
		}
	}
	legend := m.st.muted.Render(m.t("sem foco")) + " · " + lipgloss.NewStyle().Render(m.t("até 25 min")) + " · " +
		lipgloss.NewStyle().Foreground(m.st.accent).Render("25+") + " · " + m.st.title.Render("1 h+")
	return strings.TrimRight(b.String(), " \n") + "\n" + legend
}

// weekBars: os minutos de cada dia desta semana, da segunda até hoje.
func (m model) weekBars() string {
	monday := m.now.AddDate(0, 0, -weekIndex(m.now))
	var top float64
	for i := 0; i <= weekIndex(m.now); i++ {
		top = max(top, m.days[dayKey(monday.AddDate(0, 0, i))])
	}
	var b strings.Builder
	b.WriteString(bold.Render(m.t("esta semana")) + "\n")
	for i := 0; i <= weekIndex(m.now); i++ {
		min := m.days[dayKey(monday.AddDate(0, 0, i))]
		n := 0
		if top > 0 {
			n = int(float64(barCols) * min / top)
		}
		b.WriteString(fmt.Sprintf("%s  %s %s\n", m.st.muted.Render(m.t(weekdays[i])), m.bar(n, barCols),
			m.st.muted.Render(fmt.Sprintf("%7s", dur(min)))))
	}
	return strings.TrimRight(b.String(), "\n")
}

// bar: uma barra de cols células com as n primeiras acesas.
func (m model) bar(n, cols int) string {
	return lipgloss.NewStyle().Foreground(m.st.accent).Render(strings.Repeat("█", n)) +
		m.st.muted.Render(strings.Repeat("░", cols-n))
}

// taskNumbers: o que as tarefas encerradas dizem sobre o seu foco.
func (m model) taskNumbers() string {
	t := m.totals
	if t.Tasks == 0 {
		return m.st.muted.Render(m.t("a sua primeira tarefa cumprida aparece aqui"))
	}
	month := 0.0
	for k, v := range m.days {
		if strings.HasPrefix(k, m.now.Format("2006-01")) {
			month += v
		}
	}
	row := func(label, value string) string {
		return m.st.muted.Render(fmt.Sprintf("%-11s", label)) + value + "\n"
	}
	foca := fmt.Sprintf("%d", int(t.Focused.Minutes())/t.Tasks)
	if t.Focused.Minutes() < float64(t.Tasks) { // a média abaixo de 1 min não é zero
		foca = "<1"
	}
	return strings.TrimRight(row(m.t("focado"), dur(month)+m.st.muted.Render(m.t(" no mês · "))+dur(t.Focused.Minutes())+m.st.muted.Render(m.t(" no total")))+
		row(m.t("tarefas"), m.tf("%d de %d cumpridas", t.Done, t.Tasks))+
		row(m.t("por tarefa"), m.tf("declara %d min, foca %s",
			int(t.Declared.Minutes())/t.Tasks, foca)), "\n")
}
