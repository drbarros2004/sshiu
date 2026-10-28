package tui

// A casca da TUI, no padrão Elm do Bubble Tea: tudo que vem de fora (salas, banco) chega como
// mensagem no Update, efeitos saem como tea.Cmd, e o View só lê o modelo.
// Cada aba tem seu arquivo: room.go (+ picker.go), friends.go, profile.go (perfil + tema).

import (
	"fmt"
	"image/color"
	"plantao/internal/pet"
	"plantao/internal/rooms"
	"plantao/internal/store"
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
	title, muted, bad, ok, card, lock, sel, tabOn, tabOff lipgloss.Style
}

func newStyles(t theme) styles {
	c := lipgloss.Color
	return styles{
		accent: c(t.accent), good: c(t.good),
		title:  lipgloss.NewStyle().Bold(true).Foreground(c(t.accent)),
		muted:  lipgloss.NewStyle().Foreground(c(t.muted)),
		bad:    lipgloss.NewStyle().Foreground(c(t.bad)),
		ok:     lipgloss.NewStyle().Foreground(c(t.good)),
		card:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c(t.border)).Width(20).Align(lipgloss.Center),
		lock:   lipgloss.NewStyle().Italic(true).Width(18).Height(2).MaxHeight(2).Align(lipgloss.Center),
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

var resultLabel = map[pet.Result]string{pet.Done: "✓ cumpriu", pet.NotDone: "✗ não cumpriu", pet.Abandoned: "saiu no meio", pet.Missed: "faltou"}

type tab int

const (
	tabRoom tab = iota
	tabFriends
	tabProfile
	tabTheme
)

var tabNames = [...]string{"sala", "amigos", "perfil", "tema"}

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
	friendsMsg []store.Friend
	profileMsg struct {
		pet  pet.Pet
		hist []store.SprintRow
	}
	errMsg   struct{ error }
	roomsMsg []store.RoomRow // minhas salas, para o seletor
	roomMsg  struct {        // entrei numa sala (criada agora ou por código)
		r       store.RoomRow
		created bool
		w       rooms.World // retrato já com a sala nova
	}
)

func doTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg{t, rooms.SnapshotWorld()} })
}

// act roda um efeito nas salas/banco fora do Update e devolve o retrato atualizado.
func act(f func()) tea.Cmd {
	return func() tea.Msg { f(); return worldMsg(rooms.SnapshotWorld()) }
}

func loadFriends(id string) tea.Cmd {
	return func() tea.Msg {
		fs, err := store.FriendsOf(id)
		if err != nil {
			return errMsg{err}
		}
		return friendsMsg(fs)
	}
}

type model struct {
	id, nick     string // nick vazio = primeiro acesso
	w            rooms.World
	room         string         // código da sala que você está vendo
	snap         rooms.Snapshot // w.rooms[room]
	now          time.Time
	tab          tab
	tick, width  int
	pixel        bool // terminal com 256+ cores: pet em pixel art
	err          string
	st           styles
	theme, saved int // tema em prévia e tema salvo

	in       textinput.Model // o campo ativo: apelido, Lock ou amigo novo
	chat     textinput.Model
	entering bool // escrevendo o Lock
	confirm  bool // confirmando abandono
	friends  []store.Friend
	cur      int     // amigo selecionado
	adding   bool    // digitando apelido de amigo
	pet      pet.Pet // perfil, carregado ao abrir a aba
	hist     []store.SprintRow
	picking  bool            // seletor de salas aberto
	pcur     int             // item selecionado no seletor
	myRooms  []store.RoomRow // salas privadas em que você entrou
	form     form            // formulário de sala em andamento
	draft    string          // nome da sala sendo criada
	notice   string          // aviso positivo (ex.: código da sala nova)
}

// setWorld guarda o retrato e aponta snap para a sala que você está vendo.
func (m *model) setWorld(w rooms.World) {
	m.w = w
	if _, ok := w.Rooms[m.room]; !ok {
		m.room = rooms.PublicCode
	}
	m.snap = w.Rooms[m.room]
}

func New(id string) model {
	m := model{id: id, width: 80, now: time.Now(), room: rooms.PublicCode, chat: newInput("mensagem…", 200)}
	m.setWorld(rooms.SnapshotWorld())
	nick, _, ok, err := store.GetUser(id)
	if err != nil {
		log.Error("getUser", "err", err)
	}
	m.nick = nick
	m.theme = themeIndex(store.Theme(id))
	m.saved, m.st = m.theme, newStyles(themes[m.theme])
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

func (m model) Init() tea.Cmd { return tea.Batch(doTick(), loadFriends(m.id)) }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.ColorProfileMsg:
		m.pixel = msg.Profile >= colorprofile.ANSI256
	case tickMsg:
		m.tick++
		m.now = msg.now
		m.setWorld(msg.w)
		return m, doTick()
	case worldMsg:
		m.setWorld(rooms.World(msg))
	case roomsMsg:
		m.myRooms = msg
	case roomMsg:
		m.room, m.picking, m.form, m.err = msg.r.Code, false, formNone, ""
		m.notice = "você entrou em " + msg.r.Name
		if msg.created {
			m.notice = "sala criada! passe o código " + msg.r.Code + " para os amigos"
		}
		m.setWorld(msg.w)
	case nickMsg:
		switch {
		case msg.err != nil && strings.Contains(msg.err.Error(), "UNIQUE"):
			m.err = "esse apelido já tem dono"
		case msg.err != nil:
			m.err = msg.err.Error()
		default:
			m.nick, m.err = msg.nick, ""
		}
	case friendsMsg:
		m.friends = msg
		m.cur = max(0, min(m.cur, len(m.friends)-1))
	case profileMsg:
		m.pet, m.hist = msg.pet, msg.hist
	case errMsg:
		m.err = msg.Error()
	case tea.KeyPressMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m model) key(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	switch {
	case k == "ctrl+c":
		return m, tea.Quit
	case m.nick == "":
		return m.nickKey(msg)
	case k == "tab":
		return m.switchTab((m.tab + 1) % tab(len(tabNames)))
	case k == "shift+tab":
		return m.switchTab((m.tab + tab(len(tabNames)) - 1) % tab(len(tabNames)))
	case !m.typing() && len(k) == 1 && k >= "1" && k <= "4":
		return m.switchTab(tab(k[0] - '1'))
	case !m.typing() && k == "q":
		return m, tea.Quit
	}
	switch m.tab {
	case tabRoom:
		return m.roomKey(msg)
	case tabFriends:
		return m.friendsKey(msg)
	case tabTheme:
		return m.themeKey(msg)
	}
	return m, nil
}

func (m model) switchTab(t tab) (tea.Model, tea.Cmd) {
	if m.tab == tabTheme && t != tabTheme { // saiu sem salvar: desfaz a prévia
		m.theme, m.st = m.saved, newStyles(themes[m.saved])
	}
	m.tab, m.err, m.notice, m.entering, m.adding, m.confirm, m.picking, m.form = t, "", "", false, false, false, false, formNone
	switch t {
	case tabFriends:
		return m, loadFriends(m.id)
	case tabProfile:
		id := m.id
		return m, func() tea.Msg {
			_, p, _, err := store.GetUser(id)
			if err != nil {
				return errMsg{err}
			}
			hist, err := store.History(id)
			if err != nil {
				return errMsg{err}
			}
			return profileMsg{p, hist}
		}
	}
	return m, nil
}

// typing: algum campo de texto está recebendo as teclas (então "q" e "1".."4" são texto).
func (m model) typing() bool { return m.entering || m.adding || m.form != formNone || m.chatOpen() }

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
	return m, func() tea.Msg { return nickMsg{nick, store.PutUser(id, nick, pet.EnergyStart)} }
}

func (m model) View() tea.View {
	var body string
	if m.nick == "" {
		body = m.st.title.Render("PLANTÃO") + "\n\nPrimeira vez por aqui. Como quer ser chamado?\n\n" + m.in.View() +
			m.errLine() + "\n\n" + m.st.muted.Render("enter confirma · sua chave SSH é o seu login")
	} else {
		var content, help string
		switch m.tab {
		case tabRoom:
			content, help = m.roomView(), m.roomHelp()
		case tabFriends:
			content, help = m.friendsView(), "a adiciona · x remove · ↑/↓ escolhe"
		case tabProfile:
			content = m.profileView()
		case tabTheme:
			content, help = m.themeView(), "↑/↓ prévia ao vivo · enter salva"
		}
		quit := "q sai"
		if m.typing() {
			quit = "ctrl+c sai"
		}
		if help != "" {
			help += " · "
		}
		body = m.header() + "\n\n" + content + m.errLine() + "\n\n" + m.st.muted.Render(help+"tab troca de aba · "+quit)
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

// header: abas à esquerda, estado do sprint à direita, sempre visível.
func (m model) header() string {
	parts := []string{m.st.title.Render("PLANTÃO") + " "}
	for i, name := range tabNames {
		label := fmt.Sprintf("%d %s", i+1, name)
		if tab(i) == tabFriends {
			if n := m.friendsActive(); n > 0 {
				label += fmt.Sprintf(" ●%d", n)
			}
		}
		if tab(i) == m.tab {
			parts = append(parts, m.st.tabOn.Render(label))
		} else {
			parts = append(parts, m.st.tabOff.Render(label))
		}
	}
	left := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	s, status := m.snap, ""
	switch s.Phase {
	case rooms.Focus:
		status = "foco · " + mmss(s.Ends.Sub(m.now))
	case rooms.Break:
		status = "intervalo · " + mmss(s.Ends.Sub(m.now))
	default:
		status = "próximo sprint " + s.Next.Format("15:04") + " · em " + mmss(s.Next.Sub(m.now))
	}
	right := m.st.muted.Render(status)
	inner := max(m.width-4, lipgloss.Width(left)+lipgloss.Width(right)+1)
	gap := inner - lipgloss.Width(left) - lipgloss.Width(right)
	return left + strings.Repeat(" ", gap) + right + "\n" + m.st.muted.Render(strings.Repeat("─", inner))
}

func mmss(d time.Duration) string {
	d = max(d, 0).Round(time.Second)
	if d >= time.Hour {
		return fmt.Sprintf("%d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
	}
	return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}

// Relógio grande em blocos, 3 linhas de altura.
var digits = [10][3]string{
	{"█▀█", "█ █", "▀▀▀"}, {"▀█ ", " █ ", "▀▀▀"}, {"▀▀█", "█▀▀", "▀▀▀"}, {"▀▀█", " ▀█", "▀▀▀"}, {"█ █", "▀▀█", "  ▀"},
	{"█▀▀", "▀▀█", "▀▀▀"}, {"█▀▀", "█▀█", "▀▀▀"}, {"▀▀█", "  █", "  ▀"}, {"█▀█", "█▀█", "▀▀▀"}, {"█▀█", "▀▀█", "▀▀▀"},
}

func bigClock(d time.Duration) string {
	var rows [3]string
	for _, c := range mmss(d) {
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
