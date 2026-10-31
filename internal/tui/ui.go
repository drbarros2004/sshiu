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
	"github.com/charmbracelet/x/ansi"
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
		pet    pet.Pet
		tasks  []store.TaskRow
		days   map[string]float64 // minutos focados por dia, para o calendário
		totals store.Totals
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
	invite        string // o "ssh ..." com que um colega chega a este servidor
	pending       string // código que veio no comando do ssh, a entrar na primeira chance
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
	garden        *garden      // modo jardim em tela cheia (nil = desligado)
	side          *garden      // o jardim da coluna da sala solo (nil = não tem)
	lastKey       time.Time    // a última tecla: 30 s sem nenhuma liga o modo jardim
	wall          int          // o wallpaper trocado no jardim, que vale para a tarefa...
	wallFor       time.Time    // ...que começou neste instante

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
	joining  string          // código da sala na prévia, esperando o enter
	pet      pet.Pet         // perfil, carregado ao abrir a aba
	tasks    []store.TaskRow
	days     map[string]float64 // minutos focados por dia (uns 40 dias)
	totals   store.Totals
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

func New(id, invite, code string) model {
	m := model{id: id, invite: invite, pending: code, width: 80, height: 24, now: time.Now(),
		lastKey: time.Now(), mins: slices.Index(minuteChoices(), 25)}
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
	// o ctrl+v do textinput leria a área de transferência do servidor; quem cola é o terminal do
	// usuário, que manda o texto colado como tea.PasteMsg.
	in.KeyMap.Paste.SetEnabled(false)
	in.Focus()
	return in
}

func (m model) Init() tea.Cmd {
	if m.pending != "" && m.nick != "" { // "ssh <servidor> <código>": entra sem passar pelo menu
		return tea.Batch(doTick(), m.loadProfile(), m.join(m.pending))
	}
	return tea.Batch(doTick(), m.loadProfile())
}

// active: tarefa em andamento e dentro do tempo declarado.
func (m model) active() bool {
	me := m.me()
	return me.Status == rooms.Focusing && !me.Focus.Expired(m.now)
}

// syncSide mantém o jardim da coluna da sala solo: cria quando a tela é dela e ele cabe, descarta
// quando não, e refaz quando o terminal muda de tamanho. Quem o anima é o sideTickMsg.
func (m model) syncSide() model {
	w, h := 0, soloRows(m.height)
	if m.nick != "" && m.tab == tabRoom && !m.inPicker() && m.solo() && m.active() {
		w = soloGardenWidth(max(m.width-4, 40))
	}
	if w == 0 {
		m.side = nil
		return m
	}
	if m.side == nil || m.side.w != w || m.side.h != h {
		gen := 1
		if m.side != nil {
			gen = m.side.gen + 1
		}
		i := m.gardenWall(false) // a troca feita na tela cheia vale também na coluna
		m.side = &garden{scene: wallpapers[i].new(), start: m.now, last: m.now, gen: gen,
			bare: true, w: w, h: h}
	}
	if m.side.frame == "" {
		m.side.render(m) // o primeiro quadro; daí em diante quem anda é o sideTick
	}
	return m
}

// loadProfile lê o pet e as últimas tarefas do banco: a aba perfil e o painel da sala solo.
func (m model) loadProfile() tea.Cmd {
	id := m.id
	return func() tea.Msg {
		_, p, _, err := store.GetUser(id)
		if err != nil {
			return errMsg{err}
		}
		ts, err := store.Tasks(id)
		if err != nil {
			return errMsg{err}
		}
		days, err := store.FocusByDay(id, time.Now().AddDate(0, 0, -40)) // o mês inteiro, com folga
		if err != nil {
			return errMsg{err}
		}
		totals, err := store.GetTotals(id)
		if err != nil {
			return errMsg{err}
		}
		return profileMsg{p, ts, days, totals}
	}
}

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
		active := m.active()
		switch {
		case m.garden != nil && !m.garden.keep && !active: // o foco acabou ou esgotou: volta para responder
			m.garden = nil
		case m.garden == nil && active && m.nick != "" && m.now.Sub(m.lastKey) >= gardenIdle:
			var cmd tea.Cmd
			m, cmd = m.openGarden(false)
			return m, tea.Batch(doTick(), cmd)
		}
		m = m.syncSide()
		if m.side != nil && !m.side.ticking && m.garden == nil { // põe os 8 fps para andar
			m.side.ticking = true
			return m, tea.Batch(doTick(), sideTick(m.side.gen))
		}
		return m, doTick()
	case gardenTickMsg:
		if m.garden != nil && msg.gen == m.garden.gen {
			m.now = msg.t
			m.garden.render(m)
			return m, gardenTick(msg.gen)
		}
	case sideTickMsg:
		if m.side == nil || msg.gen != m.side.gen {
			return m, nil
		}
		if m.garden != nil { // a tela cheia tomou o lugar: a coluna para de desenhar
			m.side.ticking = false
			return m, nil
		}
		m.now = msg.t
		m.side.render(m)
		return m, sideTick(msg.gen)
	case worldMsg:
		m.setWorld(rooms.World(msg))
	case roomMsg:
		m.setWorld(msg.w)
		m.room, m.picking, m.form, m.err = msg.code, false, formNone, ""
		m.notice = "você entrou em " + m.w.Rooms[msg.code].Name
		if msg.created {
			m.notice = "" // o código está no título e o convite embaixo explica o que fazer
			if m.solo() {
				m.notice = "sala solo criada: só você entra aqui"
			}
		}
		return m, m.loadProfile()
	case nickMsg:
		switch {
		case msg.err != nil && strings.Contains(msg.err.Error(), "UNIQUE"):
			m.err = "esse apelido já tem dono"
		case msg.err != nil:
			m.err = msg.err.Error()
		default:
			m.nick, m.err = msg.nick, ""
			if m.pending != "" { // o código do comando esperava o apelido do primeiro acesso
				code := m.pending
				m.pending = ""
				return m, m.join(code)
			}
		}
	case profileMsg:
		m.pet, m.tasks, m.days, m.totals = msg.pet, msg.tasks, msg.days, msg.totals
	case errMsg:
		m.err = msg.Error()
	case tea.PasteMsg: // colar vai para o campo ativo; o texto vem do terminal (bracketed paste)
		if m.nick == "" || m.typing() {
			var cmd tea.Cmd
			m.in, cmd = m.in.Update(msg)
			return m, cmd
		}
	case tea.KeyPressMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m model) key(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	m.lastKey = time.Now()
	if m.garden != nil && k != "ctrl+c" { // ←/→ trocam o wallpaper; qualquer outra tecla sai, sem executar nada
		switch k {
		case "left":
			return m.swapWallpaper(-1)
		case "right":
			return m.swapWallpaper(1)
		}
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
		return m, m.loadProfile()
	}
	return m, nil
}

// typing: algum campo de texto está recebendo as teclas (então "q", "a" e "d" são texto).
// Os formulários sem campo (o tamanho da sala, a confirmação de saída) não entram: neles as
// teclas de sempre valem, inclusive o q.
func (m model) typing() bool { return m.entering || m.form == formName || m.form == formCode }

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
		body = m.st.title.Render("PLANTÃO") + "\n\nprimeira vez por aqui. qual vai ser o seu apelido?\n\n" + m.in.View() +
			m.errLine() + "\n\n" + m.st.muted.Render("enter confirma · sua chave SSH é o seu login")
	} else {
		var help string
		content := m.roomView // o conteúdo recebe o espaço que sobra (largura, altura)
		switch m.tab {
		case tabRoom:
			help = m.roomHelp()
		case tabProfile:
			content = func(int, int) string { return m.profileView() }
		case tabVisual:
			content, help = func(int, int) string { return m.visualView() }, "↑/↓ escolhe · ←/→ muda · j vê o modo jardim · enter salva"
		}
		nav := "a/d troca de aba · q sai"
		if m.typing() {
			nav = "tab troca de aba · ctrl+c sai"
		}
		if help != "" {
			nav = help + " · " + nav
		}
		// cabeçalho em cima, rodapé preso na última linha e o conteúdo no espaço do meio
		inner := max(m.width-4, 1)
		top, foot := m.header(), m.st.muted.Width(inner).Render(nav)
		h := max(1, m.height-2-lipgloss.Height(top)-2-lipgloss.Height(foot))
		mid := fit(content(inner, h-lipgloss.Height(m.errLine())+1)+m.errLine(), inner, h)
		body = top + "\n\n" + mid + "\n\n" + foot
	}
	v := tea.NewView(lipgloss.NewStyle().Padding(1, 2).Render(body))
	v.AltScreen = true
	return v
}

// fit deixa s com exatamente h linhas de no máximo w colunas: corta o que passa (com "…" na
// largura) e completa com linhas vazias.
func fit(s string, w, h int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	lines = lines[:h]
	for i, l := range lines {
		tail := "" // espaço sobrando no fim não é texto cortado: sem reticências
		if ansi.StringWidth(strings.TrimRight(ansi.Strip(l), " ")) > w {
			tail = "…"
		}
		lines[i] = ansi.Truncate(l, w, tail)
	}
	return strings.Join(lines, "\n")
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
		return m.st.muted.Render("sem tarefa")
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
