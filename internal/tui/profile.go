package tui

// Abas "perfil" (o pet, as stats e as últimas tarefas) e "visual" (tema, idioma, stats e modo jardim).

import (
	"fmt"
	"slices"
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
	titleNick := bold.Render(m.nick)
	if m.inline() != nil && m.custom {
		titleNick += " · " + devAccent.Render(m.t("dev"))
	}
	if m.notice != "" {
		titleNick += " · " + m.noticeStyle().Render(m.notice)
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
	humor := fmt.Sprintf("%-9s", m.t("humor")) + moodSt.Render(pet.Pets.MoodName(m.profileAppearance(), mood, string(m.lang))) + "\n" + m.energyBar(e, mood, moodSt) + "\n"
	drawing := func(big bool) string {
		shownMood := mood
		if m.custom {
			shownMood = m.previewMood
		}
		scale := 1
		if big {
			scale = 2
		}
		drawing := pet.Pets.RenderTerminal(m.profileAppearance(), shownMood, m.tick, scale, 0, m.pixel)
		top, bottom := profilePetPaddingTop, profilePetPaddingBottom
		if h < 26 {
			top, bottom = 0, 0
		}
		return lipgloss.NewStyle().Padding(top, profilePetPaddingHorizontal, bottom).Render(drawing)
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
	// Em uma janela baixa, separar desenho/aparência dos números evita empilhar
	// tudo num cartão que ultrapassaria a tela. A prévia mantém o tamanho original.
	if h < 17 && w >= 70 {
		side := humor + "\n" + m.colorRow()
		if !m.custom {
			side = humor + m.t("streak") + "  " + m.streakDays() + "\n" + m.t("recorde") + "  " + m.recordText() + "\n\n" + m.colorRow()
		}
		body := lipgloss.JoinHorizontal(lipgloss.Top, drawing(false), " ", side)
		if lipgloss.Height(body)+2 > h {
			return body
		}
		return m.box(titleNick, body, w, 0)
	}
	if h < 26 && w >= 70 {
		if m.custom {
			body := lipgloss.JoinHorizontal(lipgloss.Top, drawing(false), "  ", m.colorRow())
			return m.box(titleNick, first+body+"\n"+humor, w, 0)
		}
		leftWidth := max(40, lipgloss.Width(m.colorRow())+4)
		if leftWidth+1+calCols > w {
			body := lipgloss.JoinHorizontal(lipgloss.Top, drawing(false), " ", humor+"\n"+m.colorRow())
			return m.box(titleNick, body, w, 0)
		}
		left := m.box(titleNick, drawing(false)+"\n"+m.colorRow(), leftWidth, 0)
		right := strings.ReplaceAll(humor, "\n"+strings.Repeat(" ", 9), "\n") + m.t("streak") + "  " + m.streakDays() + "\n" + m.t("recorde") + "  " + m.recordText() + "\n\n" + title + "\n" + calG + "\n" + legend
		return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
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
			titleNick + "\n" + gap + humor + fmt.Sprintf("%-9s", m.t("streak")) + m.streakText() + "\n\n" + m.colorRow(),
			monthSec, numsSec, sec(hoursT, hours), sec(bandsT, bands),
		}, func(k int) string { return bold.Render(m.t("últimas tarefas")) + "\n" + strings.Join(tasks[:k], "\n") }, len(tasks))
		return lipgloss.JoinHorizontal(lipgloss.Top, d, "    ", info)
	}

	// a legenda da energia sem o recuo: com ele o painel do pet fica largo demais para o
	// calendário caber ao lado em 80 colunas
	state := strings.Replace(humor, "\n"+strings.Repeat(" ", 9), "\n", 1) + fmt.Sprintf("%-9s", m.t("streak")) + m.streakDays() + "\n" +
		fmt.Sprintf("%-9s", m.t("recorde")) + m.recordText() + "\n\n" + m.st.muted.Render(strings.Repeat("─", 24)) + "\n" +
		m.colorRow()
	left := func(big bool) string {
		d := drawing(big)
		in := max(lipgloss.Width(d), lipgloss.Width(state), lipgloss.Width(first))
		return m.box(titleNick, first+lipgloss.PlaceHorizontal(in, lipgloss.Center, d)+"\n\n"+state, in+4, 0)
	}
	recent := func(k int) string {
		return m.box(m.st.title.Render(m.t("últimas tarefas")), strings.Join(tasks[:k], "\n"), w, 0)
	}
	calBody := calG + "\n\n" + legend
	calB := m.box(title, calBody, calCols+4, 0)
	monthT := m.st.title.Render(m.t("o mês"))
	// o mês ao lado do calendário só com largura para os números inteiros, sem o "…" do box
	beside := calCols + 4 + 1 + lipgloss.Width(nums) + 4

	// largo e alto: o pet grande à esquerda, e à direita o mês, a hora e as faixas. Os painéis
	// empilhados levam uma linha vazia entre eles, como a coluna vazia entre os lado a lado, só
	// quando cabe tudo assim; senão voltam a encostar, para o espaço não esconder painel.
	l := left(true)
	if rw := w - lipgloss.Width(l) - 1; rw >= calCols+4+1+30 {
		recentW := func(k int) string {
			return m.box(m.st.title.Render(m.t("últimas tarefas")), strings.Join(tasks[:k], "\n"), rw, 0)
		}
		column := func(sep string) string {
			calRow := lipgloss.JoinHorizontal(lipgloss.Top, calB, " ", m.box(monthT, nums, rw-calCols-5, lipgloss.Height(calB)-2))
			if rw < beside { // os números não cabem inteiros ao lado: descem para baixo do calendário
				calRow = m.box(title, calBody, rw, 0) + sep + m.box(monthT, nums, rw, 0)
			}
			return panels(sep, calRow, m.box(hoursT, hours, rw, 0), m.box(bandsT, bands, rw, 0))
		}
		if all := lipgloss.JoinHorizontal(lipgloss.Top, l, " ", panels("\n\n", column("\n\n"), recentW(len(tasks)))); lipgloss.Height(all) <= h {
			return all
		}
		right := column("\n")
		if top := lipgloss.JoinHorizontal(lipgloss.Top, l, " ", right); lipgloss.Height(top) <= h {
			for k := len(tasks); k > 0; k-- {
				if withTasks := right + "\n" + recentW(k); lipgloss.Height(withTasks) <= h {
					return lipgloss.JoinHorizontal(lipgloss.Top, l, " ", withTasks)
				}
			}
			return top
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
	if all := panels("\n\n", append(secs, recent(len(tasks)))...); lipgloss.Height(all) <= h {
		return all
	}
	return stack(h, "\n", secs, recent, len(tasks))
}

// panels: os painéis um embaixo do outro, separados por sep; o painel vazio (o box sem conteúdo,
// antes da primeira tarefa) some sem deixar linha.
func panels(sep string, ps ...string) string {
	return strings.Join(slices.DeleteFunc(slices.Clone(ps), func(s string) bool { return s == "" }), sep)
}

// Ajuste o respiro da imagem sem alterar seus quadros ou sua posição do chapéu.
const (
	profilePetPaddingHorizontal = 2
	profilePetPaddingTop        = 1
	profilePetPaddingBottom     = 1
)

func (m model) colorRow() string {
	var editor *appearanceEditor
	if m.custom {
		editor = m.appearanceEdit
	}
	return m.appearanceRows(m.profileAppearance(), editor, m.custom)
}

func (m model) profileHelp() string {
	if m.custom {
		return m.t("↑/↓ opção · ←/→ altera · enter salva · esc cancela")
	}
	return m.t("←/→ mês · c personaliza o pet")
}

func (m model) profileKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if m.pendingAppearance != nil {
		return m, nil
	}
	if !m.custom {
		switch k {
		case "c":
			m.custom, m.notice = true, ""
			a := m.pet.Look()
			if m.inline() != nil {
				a = m.inlinePerson().Pet.Look()
			}
			m.appearanceEdit = newAppearanceEditor(a)
			m.previewMood = pet.MoodOf(m.pet.EnergyAt(m.now))
		case "left":
			m = m.shiftMonth(-1)
		case "right":
			m = m.shiftMonth(1)
		}
		return m, nil
	}
	if m.appearanceEdit == nil {
		a := m.pet.Look()
		if m.inline() != nil {
			a = m.inlinePerson().Pet.Look()
		}
		m.appearanceEdit = newAppearanceEditor(a)
	}
	switch k {
	case "esc", "c":
		m.custom, m.appearanceEdit = false, nil
		if m.first {
			m.first = false
			return m.switchTab(tabRoom)
		}
	case "enter":
		id, a := m.id, m.appearanceEdit.Choice
		m.appearanceRevision++
		m.pet.Appearance, m.pet.Color = a, pet.Pets.ArtworkColor(a)
		m.pendingAppearance = &a
		revision := m.appearanceRevision
		m.custom, m.notice, m.appearanceEdit = false, m.t("salvando…"), nil
		save := func() tea.Msg {
			err := store.SetAppearance(id, a)
			if err != nil {
				return appearanceSavedMsg{id: id, revision: revision, err: err}
			}
			rooms.SetAppearance(id, a)
			return appearanceSavedMsg{id: id, revision: revision, world: rooms.SnapshotWorld()}
		}

		if m.first {
			m.first = false
			next, cmd := m.switchTab(tabRoom)
			return next, tea.Batch(save, cmd)
		}
		return m, save
	default:
		m.appearanceEdit.key(k, true, &m.previewMood)
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
