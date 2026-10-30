package tui

// A casca da TUI, no padrão Elm do Bubble Tea: tudo que vem de fora (salas, banco) chega como
// mensagem no Update, efeitos saem como tea.Cmd, e o View só lê o modelo.
// Cada aba tem seu arquivo: room.go (+ picker.go), profile.go (perfil e visual); o modo jardim em garden.go.

import (
	"fmt"
	"image/color"
	"plantao/internal/pet"
	"plantao/internal/rooms"
	"plantao/internal/store"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/charmbracelet/colorprofile"
)

// Temas: só as cores de destaque; o texto normal usa a cor do próprio terminal (funciona em fundo claro e escuro).
type theme struct{ name, accent, muted, border, good, bad string }

var themes = []theme{
	{"plantão", "#ff8700", "#7a7a7a", "#4a4a4a", "#5fd787", "#ff5f5f"},
	{"catppuccin", "#cba6f7", "#7f849c", "#45475a", "#a6e3a1", "#f38ba8"},
	{"dracula", "#bd93f9", "#6272a4", "#44475a", "#50fa7b", "#ff5555"},
	{"gruvbox", "#fe8019", "#928374", "#665c54", "#b8bb26", "#fb4934"},
	{"tokyo night", "#7aa2f7", "#565f89", "#3b4261", "#9ece6a", "#f7768e"},
	{"nord", "#88c0d0", "#616e88", "#4c566a", "#a3be8c", "#bf616a"},
	{"latte (claro)", "#8839ef", "#8c8fa1", "#bcc0cc", "#40a02b", "#d20f39"},
}

type styles struct {
	accent, good                                          color.Color
	title, muted, bad, ok, card, task, sel, tabOn, tabOff lipgloss.Style
}

func newStyles(t theme) styles {
	c := lipgloss.Color
	return styles{
		accent: c(t.accent), good: c(t.good),
		title:  lipgloss.NewStyle().Bold(true).Foreground(c(t.accent)),
		muted:  lipgloss.NewStyle().Foreground(c(t.muted)),
		bad:    lipgloss.NewStyle().Foreground(c(t.bad)),
		ok:     lipgloss.NewStyle().Foreground(c(t.good)),
		card:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c(t.border)).Width(18).Align(lipgloss.Center),
		task:   lipgloss.NewStyle().Italic(true).Width(16).Height(2).MaxHeight(2).Align(lipgloss.Center),
		sel:    lipgloss.NewStyle().Bold(true).Foreground(c(t.accent)),
		tabOn:  lipgloss.NewStyle().Bold(true).Foreground(c("#11111b")).Background(c(t.accent)).Padding(0, 1),
		tabOff: lipgloss.NewStyle().Foreground(c(t.muted)).Padding(0, 1),
	}
}

func themeIndex(name string) int {
	for i, t := range themes {
		if t.name == name {
			return i
		}
	}
	return 0
}

var bold = lipgloss.NewStyle().Bold(true)

type tab int

const (
	tabRoom tab = iota
	tabProfile
	tabVisual
)

var tabNames = [...]string{"sala", "perfil", "visual"}

// Mensagens: o único jeito de o mundo de fora mudar o modelo.
type (
	tickMsg struct { // a cada 500ms: anima e traz o retrato de todas as salas
		now time.Time
		w   rooms.World
	}
	worldMsg rooms.World // retrato extra logo após uma ação, sem esperar o tick
	nickMsg  struct {
		nick string
		err  error
	}
	profileMsg struct {
		pet   pet.Pet
		tasks []store.TaskRow
	}
	errMsg  struct{ error }
	roomMsg struct { // entrei numa sala (criada agora ou por código)
		code    string
		created bool
		w       rooms.World // retrato já com a sala
	}
)

func doTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg{t, rooms.SnapshotWorld()} })
}

// act roda um efeito nas salas/banco fora do Update e devolve o retrato atualizado.
func act(f func()) tea.Cmd {
	return func() tea.Msg { f(); return worldMsg(rooms.SnapshotWorld()) }
}

type model struct {
	id, nick      string // nick vazio = primeiro acesso
	w             rooms.World
	room          string // código da sala que você está vendo ("" = nenhuma: o menu de salas)
	now           time.Time
	tab           tab
	tick, width   int
	height        int
	pixel         bool // terminal com 256+ cores: pet em pixel art
	err, notice   string
	st            styles
	theme, saved  int          // tema em prévia e tema salvo
	vis, visSaved store.Garden // modo jardim em prévia e salvo
	vrow          int          // linha selecionada na aba visual
	garden        *garden      // modo jardim ligado (nil = desligado)
	lastKey       time.Time    // a última tecla: 30 s sem nenhuma liga o modo jardim

	in       textinput.Model // o campo ativo: apelido, tarefa, nome ou código de sala
	entering bool            // escrevendo a tarefa
	choosing bool            // escolhendo o tempo declarado
	cont     bool            // ...para continuar a tarefa esgotada
	mins     int             // índice em minuteChoices()
	sel      int             // evento selecionado (ID)
	picking  bool            // menu de salas aberto
	size     int             // tamanho escolhido para a sala nova
	pcur     int             // item selecionado no menu
	form     form            // formulário de sala em andamento
	pet      pet.Pet         // perfil, carregado ao abrir a aba
	tasks    []store.TaskRow
}

// me: você como a sala te vê (status e tarefa).
func (m model) me() rooms.Member { return m.w.Users[m.id] }

// setWorld guarda o retrato e garante que a sala vista é uma das suas.
func (m *model) setWorld(w rooms.World) {
	m.w = w
	mine := w.MyRooms(m.id)
	if !slices.Contains(mine, m.room) {
		m.room = ""
		if len(mine) > 0 {
			m.room = mine[0]
		}
	}
}

func New(id string) model {
	m := model{id: id, width: 80, height: 24, now: time.Now(), lastKey: time.Now(), mins: slices.Index(minuteChoices(), 25)}
	m.setWorld(rooms.SnapshotWorld())
	nick, _, ok, err := store.GetUser(id)
	if err != nil {
		log.Error("getUser", "err", err)
	}
	m.nick = nick
	m.theme = themeIndex(store.Theme(id))
	m.saved, m.st = m.theme, newStyles(themes[m.theme])
	m.vis = store.GetGarden(id)
	m.visSaved = m.vis
	if !ok {
		m.in = newInput("seu apelido", 16)
	}
	return m
}

func newInput(placeholder string, limit int) textinput.Model {
	in := textinput.New()
	in.Placeholder = placeholder
	in.CharLimit = limit
	in.SetWidth(50)
	in.Focus()
	return in
}

func (m model) Init() tea.Cmd { return doTick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.ColorProfileMsg:
		m.pixel = msg.Profile >= colorprofile.ANSI256
	case tickMsg:
		m.tick++
		m.now = msg.now
		m.setWorld(msg.w)
		me := m.me()
		active := me.Status == rooms.Focusing && !me.Focus.Expired(m.now)
		switch {
		case m.garden != nil && !m.garden.preview && !active: // o foco acabou ou esgotou: volta para responder
			m.garden = nil
		case m.garden == nil && active && m.nick != "" && m.now.Sub(m.lastKey) >= gardenIdle:
			var cmd tea.Cmd
			m, cmd = m.openGarden(false)
			return m, tea.Batch(doTick(), cmd)
		}
		return m, doTick()
	case gardenTickMsg:
		if m.garden != nil && msg.gen == m.garden.gen {
			m.now = msg.t
			m.garden.render(m)
			return m, gardenTick(msg.gen)
		}
	case worldMsg:
		m.setWorld(rooms.World(msg))
	case roomMsg:
		m.setWorld(msg.w)
		m.room, m.picking, m.form, m.err = msg.code, false, formNone, ""
		m.notice = "você entrou em " + m.w.Rooms[msg.code].Name
		if msg.created {
			m.notice = "sala criada! passe o código " + msg.code + " para os colegas"
		}
	case nickMsg:
		switch {
		case msg.err != nil && strings.Contains(msg.err.Error(), "UNIQUE"):
			m.err = "esse apelido já tem dono"
		case msg.err != nil:
			m.err = msg.err.Error()
		default:
			m.nick, m.err = msg.nick, ""
		}
	case profileMsg:
		m.pet, m.tasks = msg.pet, msg.tasks
	case errMsg:
		m.err = msg.Error()
	case tea.KeyPressMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m model) key(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	m.lastKey = time.Now()
	if m.garden != nil && k != "ctrl+c" { // qualquer tecla sai do modo jardim, sem executar nada
		m.garden = nil
		return m, nil
	}
	n := tab(len(tabNames))
	switch {
	case k == "ctrl+c":
		return m, tea.Quit
	case m.nick == "":
		return m.nickKey(msg)
	case k == "tab" || !m.typing() && k == "d":
		return m.switchTab((m.tab + 1) % n)
	case k == "shift+tab" || !m.typing() && k == "a":
		return m.switchTab((m.tab + n - 1) % n)
	case !m.typing() && k == "q":
		return m, tea.Quit
	}
	switch m.tab {
	case tabRoom:
		return m.roomKey(msg)
	case tabVisual:
		return m.visualKey(msg)
	}
	return m, nil
}

func (m model) switchTab(t tab) (tea.Model, tea.Cmd) {
	if m.tab == tabVisual && t != tabVisual { // saiu sem salvar: desfaz a prévia
		m.theme, m.st, m.vis = m.saved, newStyles(themes[m.saved]), m.visSaved
	}
	m.tab, m.err, m.notice, m.entering, m.choosing, m.picking, m.form = t, "", "", false, false, false, formNone
	if t == tabProfile {
		id := m.id
		return m, func() tea.Msg {
			_, p, _, err := store.GetUser(id)
			if err != nil {
				return errMsg{err}
			}
			ts, err := store.Tasks(id)
			if err != nil {
				return errMsg{err}
			}
			return profileMsg{p, ts}
		}
	}
	return m, nil
}

// typing: algum campo de texto está recebendo as teclas (então "q", "a" e "d" são texto).
func (m model) typing() bool { return m.entering || m.form != formNone }

func (m model) nickKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() != "enter" {
		var cmd tea.Cmd
		m.in, cmd = m.in.Update(msg)
		return m, cmd
	}
	nick := strings.TrimSpace(m.in.Value())
	if nick == "" || strings.ContainsAny(nick, " :") {
		m.err = "apelido sem espaços, por favor"
		return m, nil
	}
	id := m.id
	return m, func() tea.Msg {
		err := store.PutUser(id, nick, pet.EnergyStart)
		if err == nil {
			rooms.Created(id)
		}
		return nickMsg{nick, err}
	}
}

func (m model) View() tea.View {
	if m.garden != nil {
		v := tea.NewView(m.garden.frame)
		v.AltScreen = true
		return v
	}
	var body string
	if m.nick == "" {
		body = m.st.title.Render("PLANTÃO") + "\n\nPrimeira vez por aqui. Como quer ser chamado?\n\n" + m.in.View() +
			m.errLine() + "\n\n" + m.st.muted.Render("enter confirma · sua chave SSH é o seu login")
	} else {
		var content, help string
		switch m.tab {
		case tabRoom:
			content, help = m.roomView(), m.roomHelp()
		case tabProfile:
			content = m.profileView()
		case tabVisual:
			content, help = m.visualView(), "↑/↓ escolhe · ←/→ muda · j vê o modo jardim · enter salva"
		}
		nav := "a/d troca de aba · q sai"
		if m.typing() {
			nav = "tab troca de aba · ctrl+c sai"
		}
		if help != "" {
			help += " · "
		}
		body = m.header() + "\n\n" + content + m.errLine() + "\n\n" + m.st.muted.Render(help+nav)
	}
	v := tea.NewView(lipgloss.NewStyle().Padding(1, 2).Render(body))
	v.AltScreen = true
	return v
}

func (m model) errLine() string {
	if m.err == "" {
		return ""
	}
	return "\n" + m.st.bad.Render(m.err)
}

// header: abas à esquerda, o seu status à direita, sempre visível.
func (m model) header() string {
	parts := []string{m.st.title.Render("PLANTÃO") + " "}
	for i, name := range tabNames {
		if tab(i) == m.tab {
			parts = append(parts, m.st.tabOn.Render(name))
		} else {
			parts = append(parts, m.st.tabOff.Render(name))
		}
	}
	left := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	right := m.st.muted.Render(m.status(m.me()))
	inner := max(m.width-4, lipgloss.Width(left)+lipgloss.Width(right)+1)
	gap := inner - lipgloss.Width(left) - lipgloss.Width(right)
	return left + strings.Repeat(" ", gap) + right + "\n" + m.st.muted.Render(strings.Repeat("─", inner))
}

// status: o que a sala vê de um membro, com o tempo que falta.
func (m model) status(x rooms.Member) string {
	switch {
	case x.Status == rooms.Focusing && x.Focus.Expired(m.now):
		return m.st.bad.Render("tempo esgotado")
	case x.Status == rooms.Focusing:
		return "focando " + mmss(x.Focus.Ends.Sub(m.now))
	case x.Status == rooms.Idle:
		return m.st.muted.Render("parado")
	}
	return m.st.muted.Render("offline")
}

func mmss(d time.Duration) string {
	d = max(d, 0).Round(time.Second)
	if d >= time.Hour {
		return fmt.Sprintf("%d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
	}
	return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}

// Números grandes em blocos, 3 linhas de altura.
var digits = [10][3]string{
	{"█▀█", "█ █", "▀▀▀"}, {"▀█ ", " █ ", "▀▀▀"}, {"▀▀█", "█▀▀", "▀▀▀"}, {"▀▀█", " ▀█", "▀▀▀"}, {"█ █", "▀▀█", "  ▀"},
	{"█▀▀", "▀▀█", "▀▀▀"}, {"█▀▀", "█▀█", "▀▀▀"}, {"▀▀█", "  █", "  ▀"}, {"█▀█", "█▀█", "▀▀▀"}, {"█▀█", "▀▀█", "▀▀▀"},
}

// bigText desenha dígitos e ":" em blocos.
func bigText(s string) string {
	var rows [3]string
	for _, c := range s {
		for i := range rows {
			if c == ':' {
				rows[i] += [3]string{"  ", "▀ ", "▀ "}[i]
			} else {
				rows[i] += digits[c-'0'][i] + " "
			}
		}
	}
	return strings.Join(rows[:], "\n")
}
