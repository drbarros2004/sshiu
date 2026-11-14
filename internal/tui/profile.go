package tui

// Abas "perfil" (o pet, as stats e as últimas tarefas) e "visual" (tema, idioma, stats e modo jardim).

import (
	"fmt"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/charmbracelet/x/ansi"
)

var resultLabel = map[pet.Result]string{pet.Done: "✓ cumprida", pet.NotDone: "não cumprida", pet.Interrupted: "interrompida"}

// profileView: o pet, o estado e as stats em w×h, em painéis (o padrão) ou em lista, como a aba
// visual escolhe (#45). Quando falta altura, as seções saem de baixo para cima: as últimas
// tarefas encurtam primeiro, e o pet, o estado e o calendário são os últimos a sair.
func (m model) profileView(w, h int) string {
	if m.pet.LastActive.IsZero() {
		return m.st.muted.Render(m.t("carregando…"))
	}
	e := m.pet.EnergyAt(m.now)
	mood := pet.MoodOf(e)
	moodSt := lipgloss.NewStyle().Foreground(energyColors.barColor(mood))
	first, gap := "", "\n" // gap: o que separa o apelido do humor na lista
	if m.first {           // o primeiro acesso: o apelido foi o primeiro passo, este é o último
		first = m.st.title.Render(m.t("último passo: a cor do pet")) + "\n" +
			m.st.muted.Render(m.t("enter salva e abre a aba sala")) + "\n\n"
		gap = first
	}
	humor := fmt.Sprintf("%-9s", m.t("humor")) + moodSt.Render(m.t(pet.MoodName[mood])) + "\n" + m.energyBar(e, mood, moodSt) + "\n"
	drawing := func(big bool) string {
		render := pet.Render
		if big {
			render = pet.RenderBig
		}
		if !m.custom {
			return render(mood, m.tick, m.pixel, m.myColor())
		}
		return bold.Render(m.t("prévia do pet")) + "\n\n" + render(m.previewMood, m.tick, m.pixel, m.myColor()) + "\n\n" +
			m.st.muted.Render("↑/↓ ") + m.t(pet.MoodName[m.previewMood]) + "\n" +
			m.st.muted.Render(m.t("apenas para visualização")+"\n"+m.t("(não altera humor ou energia)"))
	}
	tasks := []string{m.st.muted.Render(m.t("nenhuma ainda"))}
	if len(m.tasks) > 0 {
		tasks = nil
	}
	for _, t := range m.tasks {
		label := fmt.Sprintf("%-13s", m.t(resultLabel[t.Result]))
		if t.Result == pet.Done {
			label = m.st.ok.Render(label)
		} else {
			label = m.st.muted.Render(label)
		}
		line := fmt.Sprintf("%s  %s %3d min  %s", t.Start.Format("02/01 15:04"), label, int(t.Focused.Minutes()), t.Task)
		if r, ok := m.w.Rooms[t.Room]; ok { // a sala pode ter sumido desde então: aí não há nome a mostrar
			line += m.st.muted.Render(" · " + m.roomName(r))
		}
		tasks = append(tasks, line)
	}
	month := m.shownMonth()
	title, calG, legend := m.monthTitle(month, true), m.calendar(month), m.heatLegend()
	nums := m.rows(m.monthRows(month), 15)
	hoursT, bandsT := m.st.title.Render(m.t("quando você foca")), m.st.title.Render(m.t("o que você cumpre"))+m.st.muted.Render(m.t(" · pelo tempo declarado"))
	hours, bands := "", ""
	if m.totals.Tasks > 0 { // sem tarefa encerrada não há hora nem faixa a mostrar
		hours, bands = m.hoursBody(), m.bandsBody()
	}

	if m.statsList {
		d := drawing(h >= 23) // baixo (80×24): o pet menor, para o calendário caber embaixo do estado
		iw := w - lipgloss.Width(d) - 4
		cal := title + "\n" + calG + "\n" + legend
		monthSec, numsSec := cal, nums // estreito: os números descem para depois do calendário, e saem antes dele
		switch {
		case iw >= lipgloss.Width(cal)+4+lipgloss.Width(nums): // largo: ao lado do calendário, na altura dos dias
			monthSec, numsSec = lipgloss.JoinHorizontal(lipgloss.Top, cal, "    ", "\n\n"+nums), ""
		case lipgloss.Width(title)+3+lipgloss.Width(legend) <= iw: // a legenda sobe para a linha do mês: uma linha a menos
			monthSec = title + "   " + legend + "\n" + calG
		}
		sec := func(title, body string) string {
			if body == "" {
				return ""
			}
			return title + "\n" + body
		}
		info := stack(h, "\n\n", []string{
			bold.Render(m.nick) + "\n" + gap + humor + fmt.Sprintf("%-9s", m.t("streak")) + m.streakText() + "\n" + m.colorRow(),
			monthSec, numsSec, sec(hoursT, hours), sec(bandsT, bands),
		}, func(k int) string { return bold.Render(m.t("últimas tarefas")) + "\n" + strings.Join(tasks[:k], "\n") }, len(tasks))
		return lipgloss.JoinHorizontal(lipgloss.Top, d, "    ", info)
	}

	// a legenda da energia sem o recuo: com ele o painel do pet fica largo demais para o
	// calendário caber ao lado em 80 colunas
	state := strings.Replace(humor, "\n"+strings.Repeat(" ", 9), "\n", 1) + fmt.Sprintf("%-9s", m.t("streak")) + m.streakDays() + "\n" +
		fmt.Sprintf("%-9s", m.t("recorde")) + m.recordText() + "\n" +
		m.colorRow()
	left := func(big bool) string {
		d := drawing(big)
		in := max(lipgloss.Width(d), lipgloss.Width(state), lipgloss.Width(first))
		return m.box(bold.Render(m.nick), first+lipgloss.PlaceHorizontal(in, lipgloss.Center, d)+"\n\n"+state, in+4, 0)
	}
	recent := func(k int) string {
		return m.box(m.st.title.Render(m.t("últimas tarefas")), strings.Join(tasks[:k], "\n"), w, 0)
	}
	calBody := calG + "\n\n" + legend
	calB := m.box(title, calBody, calCols+4, 0)
	monthT := m.st.title.Render(m.t("o mês"))
	// o mês ao lado do calendário só com largura para os números inteiros, sem o "…" do box
	beside := calCols + 4 + 1 + lipgloss.Width(nums) + 4

	// largo e alto: o pet grande à esquerda, e à direita o mês, a hora e as faixas
	l := left(true)
	if rw := w - lipgloss.Width(l) - 1; rw >= calCols+4+1+30 {
		calRow := lipgloss.JoinHorizontal(lipgloss.Top, calB, " ", m.box(monthT, nums, rw-calCols-5, lipgloss.Height(calB)-2))
		if rw < beside { // os números não cabem inteiros ao lado: descem para baixo do calendário
			calRow = m.box(title, calBody, rw, 0) + "\n" + m.box(monthT, nums, rw, 0)
		}
		right := lipgloss.JoinVertical(lipgloss.Left, calRow, m.box(hoursT, hours, rw, 0), m.box(bandsT, bands, rw, 0))
		if top := lipgloss.JoinHorizontal(lipgloss.Top, l, " ", right); lipgloss.Height(top) <= h {
			return stack(h, "\n", []string{top}, recent, len(tasks))
		}
	}
	// pequeno: o pet menor, e em cima só o pet, o estado e o calendário; o resto desce
	l = left(false)
	lh := lipgloss.Height(l) - 2
	secs := []string{l, m.box(title, calBody, w, 0), m.box(monthT, nums, w, 0)}
	switch rw := w - lipgloss.Width(l) - 1; {
	case rw >= beside: // sobra largura: o mês também sobe, ao lado do calendário
		secs = []string{lipgloss.JoinHorizontal(lipgloss.Top, l, " ", m.box(title, calBody, calCols+4, lh),
			" ", m.box(monthT, nums, rw-calCols-5, lh))}
	case rw >= calCols+4:
		secs = []string{lipgloss.JoinHorizontal(lipgloss.Top, l, " ", m.box(title, calBody, rw, lh)),
			m.box(monthT, nums, w, 0)}
	}
	secs = append(secs, m.box(hoursT, hours, w, 0), m.box(bandsT, bands, w, 0))
	return stack(h, "\n", secs, recent, len(tasks))
}

// myColor: a cor do pet em prévia — a salva, fora do modo personalizar.
func (m model) myColor() string { return pet.Colors[m.pcolor].ID }

// colorRow: a linha da cor do pet. Fora do modo personalizar ela só diz qual é e como mexer;
// dentro, é a linha que as setas mudam, com a prévia no pet grande ao lado.
func (m model) colorRow() string {
	c := pet.Colors[m.pcolor]
	name := m.t(c.Name) + " " + lipgloss.NewStyle().Foreground(c.Body).Render("██")
	if !m.custom {
		row := fmt.Sprintf("%-9s", m.t("cor")) + m.st.muted.Render(name)
		if m.notice != "" { // o "salvo" da cor, na própria linha: quando ele some, nada sobe
			row += "  " + m.noticeStyle().Render(m.notice)
		}
		return row
	}
	return m.st.sel.Render(fmt.Sprintf("%-9s", m.t("cor"))) + m.st.muted.Render("‹ ") + name + m.st.muted.Render(" ›")
}

func (m model) profileHelp() string {
	if m.custom && m.first {
		return m.t("←/→ cor · ↑/↓ prévia do humor · enter salva e abre a sala")
	}
	if m.custom {
		return m.t("←/→ cor · ↑/↓ prévia do humor · enter salva a cor · esc volta")
	}
	return m.t("←/→ mês · c personaliza o pet")
}

// profileKey: fora do personalizar, ←/→ trocam o mês do calendário. A cor é salva; o humor
// escolhido só troca o desenho da prévia.
func (m model) profileKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if !m.custom {
		switch k {
		case "c":
			m.custom, m.notice = true, ""
			m.previewMood = pet.MoodOf(m.pet.EnergyAt(m.now))
		case "left":
			m = m.shiftMonth(-1)
		case "right":
			m = m.shiftMonth(1)
		}
		return m, nil
	}
	switch k {
	case "up":
		m.previewMood = (m.previewMood + pet.Sleeping) % (pet.Sleeping + 1)
	case "down":
		m.previewMood = (m.previewMood + 1) % (pet.Sleeping + 1)
	case "left":
		m.pcolor = (m.pcolor + len(pet.Colors) - 1) % len(pet.Colors)
	case "right":
		m.pcolor = (m.pcolor + 1) % len(pet.Colors)
	case "esc", "c":
		m.pcolor, m.custom = m.pcSaved, false
		if m.first { // o primeiro acesso termina na sala, com a cor que veio
			m.first = false
			return m.switchTab(tabRoom)
		}
	case "enter":
		id, paletteID := m.id, m.myColor()
		m.pcSaved, m.custom, m.notice = m.pcolor, false, m.t("salvo")
		save := act(func() {
			if err := store.SetPetColor(id, paletteID); err != nil {
				log.Error("setPetColor", "err", err)
				return
			}
			rooms.SetColor(id, paletteID)
		})
		if m.first {
			m.first = false
			next, cmd := m.switchTab(tabRoom)
			return next, tea.Batch(save, cmd)
		}
		return m, save
	}
	return m, nil
}

// firstGardenRow: a primeira linha do modo jardim na aba visual, depois do tema, do idioma e das stats.
const firstGardenRow = 3

// statsLayouts: os dois jeitos das stats na aba visual; o padrão são os painéis (#45).
// ponytail: as duas versões convivem até a turma escolher uma; aí sai a outra e esta linha.
var statsLayouts = map[bool]string{false: "painéis", true: "lista"}

// visualRows: as linhas da aba visual, cada uma com o valor em prévia.
func (m model) visualRows() [][2]string {
	wp := m.t("sorteado a cada tarefa")
	if i := wallpaperIndex(m.vis.Wallpaper); i >= 0 {
		wp = m.t(wallpapers[i].name)
	}
	shown := map[bool]string{true: m.t("aparece"), false: m.t("não aparece")}
	return [][2]string{
		{m.t("tema"), m.t(themes[m.theme].name)},
		// o rótulo nos dois idiomas: quem não lê o da tela acha a linha pela palavra do outro
		{m.t("idioma") + " · " + map[bool]string{true: "idioma", false: "language"}[m.lang == "en"], langs[langIndex(m.lang)].name},
		{m.t("estatísticas"), m.t(statsLayouts[m.statsList])},
		{m.t("wallpaper"), wp},
		{m.t("relógio"), m.t(clockModes[m.vis.Clock])},
		{m.t("tarefa"), shown[m.vis.Task]},
		{m.t("quem mais está focando"), shown[m.vis.Room]},
		{m.t("liga sozinho"), map[bool]string{true: m.tf("depois de %.0f s", gardenIdle.Seconds()), false: m.t("nunca")}[m.vis.Auto]},
		{m.t("permitir wallpapers para devs de verdade?"), map[bool]string{true: m.t("sim"), false: m.t("não")}[m.vis.Joke]},
	}
}

func (m model) visualKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	step := 0
	switch msg.String() {
	case "up", "k":
		m.vrow = max(0, m.vrow-1)
	case "down":
		m.vrow = min(len(m.visualRows())-1, m.vrow+1)
	case "left":
		step = -1
	case "right":
		step = 1
	case "j":
		return m.openGarden(gardenPreview)
	case "v":
		return m.openGarden(gardenEditor)
	case "enter":
		id, name, g, l, list := m.id, themes[m.theme].name, m.vis, m.lang, m.statsList
		m.saved, m.visSaved, m.langSaved, m.listSaved = m.theme, m.vis, m.lang, m.statsList
		m.notice = m.t("salvo")
		return m, act(func() {
			if err := store.SetTheme(id, name); err != nil {
				log.Error("setTheme", "err", err)
			}
			if err := store.SetLang(id, string(l)); err != nil {
				log.Error("setLang", "err", err)
			}
			if err := store.SetStatsList(id, list); err != nil {
				log.Error("setStatsList", "err", err)
			}
			if err := store.SetGarden(id, g); err != nil {
				log.Error("setGarden", "err", err)
			}
		})
	}
	if step == 0 {
		return m, nil
	}
	m.notice = ""
	cycle := func(i, n int) int { return (i + step + n) % n }
	switch m.vrow {
	case 0:
		m.theme = cycle(m.theme, len(themes))
		m.st = newStyles(themes[m.theme]) // prévia ao vivo
	case 1: // a tela toda já muda de idioma, como o tema muda de cor
		m.lang = langs[cycle(langIndex(m.lang), len(langs))].id
	case 2:
		m.statsList = !m.statsList
	case 3:
		m.vis.Wallpaper = cycleWallpaper(m.vis, step)
	case 4:
		m.vis.Clock = cycle(m.vis.Clock, len(clockModes))
	case 5:
		m.vis.Task = !m.vis.Task
	case 6:
		m.vis.Room = !m.vis.Room
	case 7:
		m.vis.Auto = !m.vis.Auto
	case 8:
		m.vis.Joke = !m.vis.Joke
		if i := wallpaperIndex(m.vis.Wallpaper); i >= 0 && wallpapers[i].joke && !m.vis.Joke {
			m.vis.Wallpaper = "" // o de piada escolhido some da lista: volta para o sorteado
		}
	}
	return m, nil
}

func (m model) visualView() string {
	var b strings.Builder
	for i, r := range m.visualRows() {
		if i == firstGardenRow {
			b.WriteString("\n" + bold.Render(m.t("modo jardim")) + m.st.muted.Render(m.t(" · a tela de descanso durante o foco")) + "\n")
		}
		// o rótulo comprido (o dos wallpapers de piada) empurra só o próprio valor
		cursor, label, value := "  ", r[0]+"  "+strings.Repeat(" ", max(0, 22-ansi.StringWidth(r[0]))), m.st.muted.Render("‹ ")+r[1]+m.st.muted.Render(" ›")
		if i == m.vrow {
			cursor, label = m.st.sel.Render("▸ "), m.st.sel.Render(label)
		}
		if i == 0 {
			for _, c := range []string{themes[m.theme].accent, themes[m.theme].good, themes[m.theme].bad, themes[m.theme].muted} {
				value += " " + lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render("██")
			}
		}
		b.WriteString(cursor + label + value + "\n")
	}
	if m.notice != "" {
		b.WriteString("\n" + m.st.ok.Render(m.notice))
	}
	return strings.TrimRight(b.String(), "\n")
}
