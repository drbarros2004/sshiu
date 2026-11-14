package tui

// Abas "perfil" (o pet grande + as últimas tarefas) e "visual" (tema e modo jardim).

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

// profileView: o pet à esquerda e as informações à direita, em h linhas. Quando a altura não dá
// para tudo, as últimas tarefas encurtam; se nem uma cabe, o calendário sai (ele também está na
// sala solo) para os números e as tarefas ficarem. Seção sem nenhuma linha não aparece.
func (m model) profileView(_, h int) string {
	if m.pet.LastActive.IsZero() {
		return m.st.muted.Render(m.t("carregando…"))
	}
	e := m.pet.EnergyAt(m.now)
	mood := pet.MoodOf(e)
	moodSt := lipgloss.NewStyle().Foreground(energyColors.barColor(mood))
	head := bold.Render(m.nick) + "\n\n"
	if m.first { // o primeiro acesso: o apelido foi o primeiro passo, este é o último
		head = bold.Render(m.nick) + "\n" + m.st.title.Render(m.t("último passo: a cor do pet")) + "\n" +
			m.st.muted.Render(m.t("enter salva e abre a aba sala")) + "\n\n"
	}
	head += fmt.Sprintf("%-9s", m.t("humor")) + moodSt.Render(m.t(pet.MoodName[mood])) + "\n" +
		m.energyBar(e, mood, moodSt) + "\n" +
		fmt.Sprintf("%-9s%d %s", m.t("streak"), m.pet.StreakAt(m.now), plural(m.pet.StreakAt(m.now), m.t("dia seguido"), m.t("dias seguidos"))) + "\n" +
		m.colorRow()
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
	build := func(cal bool, k int) string {
		parts := []string{head}
		if cal {
			parts = append(parts, m.calendar())
		}
		parts = append(parts, m.taskNumbers())
		if k > 0 {
			parts = append(parts, bold.Render(m.t("últimas tarefas"))+"\n"+strings.Join(tasks[:k], "\n"))
		}
		return strings.Join(parts, "\n\n")
	}
	info := build(true, len(tasks))
	for _, cal := range []bool{true, false} {
		for k := len(tasks); k >= 1 && lipgloss.Height(info) > h; k-- {
			info = build(cal, k)
		}
	}
	if lipgloss.Height(info) > h { // nem uma tarefa cabe: os números ficam, com o calendário se der
		if info = build(true, 0); lipgloss.Height(info) > h {
			info = build(false, 0)
		}
	}
	drawing := pet.RenderBig(mood, m.tick, m.pixel, m.myColor())
	if m.custom {
		preview := pet.RenderBig(m.previewMood, m.tick, m.pixel, m.myColor())
		if m.height > 0 && m.height < 30 { // o desenho menor mantém os avisos visíveis em 80×24
			preview = pet.Render(m.previewMood, m.tick, m.pixel, m.myColor())
		}
		drawing = bold.Render(m.t("prévia do pet")) + "\n\n" +
			preview + "\n\n" +
			m.st.muted.Render("↑/↓ ") + m.t(pet.MoodName[m.previewMood]) + "\n" +
			m.st.muted.Render(m.t("apenas para visualização")+"\n"+m.t("(não altera humor ou energia)"))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, drawing, "    ", info)
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
	return m.t("c personaliza o pet")
}

// profileKey: a cor é salva; o humor escolhido só troca o desenho da prévia.
func (m model) profileKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if !m.custom {
		if k == "c" {
			m.custom, m.notice = true, ""
			m.previewMood = pet.MoodOf(m.pet.EnergyAt(m.now))
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

// firstGardenRow: a primeira linha do modo jardim na aba visual, depois do tema e do idioma.
const firstGardenRow = 2

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
		id, name, g, l := m.id, themes[m.theme].name, m.vis, m.lang
		m.saved, m.visSaved, m.langSaved = m.theme, m.vis, m.lang
		m.notice = m.t("salvo")
		return m, act(func() {
			if err := store.SetTheme(id, name); err != nil {
				log.Error("setTheme", "err", err)
			}
			if err := store.SetLang(id, string(l)); err != nil {
				log.Error("setLang", "err", err)
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
		m.vis.Wallpaper = cycleWallpaper(m.vis, step)
	case 3:
		m.vis.Clock = cycle(m.vis.Clock, len(clockModes))
	case 4:
		m.vis.Task = !m.vis.Task
	case 5:
		m.vis.Room = !m.vis.Room
	case 6:
		m.vis.Auto = !m.vis.Auto
	case 7:
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
