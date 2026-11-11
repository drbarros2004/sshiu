package tui

// A casca da TUI, no padrão Elm do Bubble Tea: tudo que vem de fora (salas, banco) chega como
// mensagem no Update, efeitos saem como tea.Cmd, e o View só lê o modelo.
// Cada aba tem seu arquivo: room.go (+ picker.go), profile.go (perfil e visual); o modo jardim em garden.go.

import (
	"fmt"
	"image/color"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
	"strings"
	"time"
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

// Temas: só as cores de destaque; o texto normal usa a cor do próprio terminal (funciona em fundo claro e escuro).
// off é a borda de quem está offline: a mesma cor da borda puxada para longe do olho (mais escura
// nos temas escuros, mais clara no claro), para a caixa de quem não está ali ficar apagada.
type theme struct{ name, accent, muted, border, off, good, bad string }

var themes = []theme{
	{"sshiu", "#ff8700", "#7a7a7a", "#4a4a4a", "#212121", "#5fd787", "#ff5f5f"},
	{"catppuccin", "#cba6f7", "#7f849c", "#45475a", "#1f1f28", "#a6e3a1", "#f38ba8"},
	{"dracula", "#bd93f9", "#6272a4", "#44475a", "#1e1f28", "#50fa7b", "#ff5555"},
	{"gruvbox", "#fe8019", "#928374", "#665c54", "#2d2925", "#b8bb26", "#fb4934"},
	{"tokyo night", "#7aa2f7", "#565f89", "#3b4261", "#1a1d2b", "#9ece6a", "#f7768e"},
	{"nord", "#88c0d0", "#616e88", "#4c566a", "#22262f", "#a3be8c", "#bf616a"},
	{"latte (claro)", "#8839ef", "#8c8fa1", "#bcc0cc", "#ebecf0", "#40a02b", "#d20f39"},
}

type styles struct {
	accent, good, off                                     color.Color
	title, muted, bad, ok, card, task, sel, tabOn, tabOff lipgloss.Style
}

func newStyles(t theme) styles {
	c := lipgloss.Color
	return styles{
		accent: c(t.accent), good: c(t.good), off: c(t.off),
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
	tabDev // só com a flag -dev (dev.go)
)

var tabNames = [...]string{"sala", "perfil", "visual", "dev"}

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
	id, nick        string // nick vazio = primeiro acesso
	invite          string // o "ssh ..." com que um colega chega a este servidor
	pending         string // código que veio no comando do ssh, a entrar na primeira chance
	w               rooms.World
	room            string    // código da sala que você está vendo ("" = nenhuma: o menu de salas)
	left            string    // a sala de que você saiu ou que encerrou: um retrato de antes não a traz de volta
	first           bool      // primeiro acesso: o apelido já foi, falta a cor do pet
	arrived         time.Time // quando a sala atual foi criada ou aberta por código: segura o enter repetido
	now             time.Time
	tab             tab
	tick, width     int
	height          int
	pixel           bool // terminal com 256+ cores: pet em pixel art
	err, notice     string
	noticeSeen      string    // o aviso como estava no último tick...
	noticeAt        time.Time // ...e desde quando: um aviso velho sai da tela sozinho
	st              styles
	theme, saved    int          // tema em prévia e tema salvo
	lang, langSaved lang         // idioma em prévia e idioma salvo (#86)
	pcolor          int          // cor do pet em prévia (índice em pet.Colors)
	pcSaved         int          // a cor do pet que está no banco
	custom          bool         // aba de perfil no modo personalizar
	previewMood     pet.Mood     // só o desenho da personalização; nunca vai para o banco ou para a sala
	vis, visSaved   store.Garden // modo jardim em prévia e salvo
	vrow            int          // linha selecionada na aba visual
	ncur, nobra     int          // linha e obra escolhidas na tela de criar sala
	obraFull        bool         // a obra sozinha na tela (tecla o)
	garden          *garden      // modo jardim em tela cheia (nil = desligado)
	dev             devPanel     // a aba dev (só com -dev)
	gardenGen       int          // distingue os ticks de cada abertura
	side            *garden      // o jardim da coluna da sala solo (nil = não tem)
	lastKey         time.Time    // a última tecla: 30 s sem nenhuma liga o modo jardim, se a aba visual deixar
	wall            int          // o wallpaper trocado no jardim, que vale para a tarefa...
	wallFor         time.Time    // ...que começou neste instante

	in       textinput.Model    // o campo ativo: apelido, tarefa, nome ou código de sala
	entering bool               // escrevendo a tarefa
	choosing bool               // escolhendo o tempo declarado
	cont     bool               // ...para continuar a tarefa esgotada
	mins     int                // índice em minuteChoices()
	sel      int                // evento selecionado (ID)
	picking  bool               // menu de salas aberto
	cfg      bool               // menu do dono da sala aberto (tecla c)
	cpage    cfgPage            // página dentro dele: o menu, passar, remover ou encerrar
	ccur     int                // linha escolhida no menu
	mcur     int                // membro escolhido em passar/remover
	cname    bool               // digitando o nome da sala na própria linha
	cyes     bool               // "sim" escolhido na confirmação de passar ou de encerrar
	cwho     string             // quem vai receber a sala, na confirmação de passar
	brow     int                // linha escolhida na página da tabela de foco: quem vê, formato ou período
	bper     int                // período da tabela de foco que você está vendo (índice em periods)
	bfor     int                // o período a que bmins se refere...
	bday     string             // ...lido neste dia
	bmins    map[string]float64 // minutos do período até ontem, por usuário
	size     int                // tamanho escolhido para a sala nova
	pcur     int                // item selecionado no menu
	form     form               // formulário de sala em andamento
	joining  string             // código da sala na prévia, esperando o enter
	pet      pet.Pet            // perfil, carregado ao abrir a aba
	tasks    []store.TaskRow
	days     map[string]float64 // minutos focados por dia (uns 40 dias)
	totals   store.Totals
}

// me: você como a sala te vê (status e tarefa).
func (m model) me() rooms.Member { return m.w.Users[m.id] }

// noticeTTL: quanto tempo um aviso fica na tela sem nenhuma tecla (docs/regras.md).
const noticeTTL = 8 * time.Second

// setWorld guarda o retrato e garante que a sala vista é uma das suas.
func (m *model) setWorld(w rooms.World) {
	was := ""
	if s, ok := m.w.Rooms[m.room]; ok {
		was = s.Name
	}
	m.w = w
	// a sala de que você saiu não volta: um tick tirado antes da saída ainda a tem
	mine := slices.DeleteFunc(w.MyRooms(m.id), func(c string) bool { return c == m.left })
	if !slices.Contains(mine, m.room) {
		m.room = ""
		if len(mine) > 0 {
			m.room = mine[0]
		}
		if was != "" { // encerrada pelo dono, ou a sua vaga foi removida (ADR 0007)
			m.notice, m.cfg, m.cname = m.tf("você não está mais em %s", was), false, false
		}
	}
	m.pcur = max(0, min(m.pcur, len(m.pickerItems())-1)) // a lista encolhe quando uma sala some
}

// noticeStyle: verde para o que deu certo; a perda de uma sala sai na cor de alerta.
func (m model) noticeStyle() lipgloss.Style {
	if strings.HasPrefix(m.notice, m.tf("você não está mais em %s", "")) { // a perda não é uma boa notícia
		return m.st.bad
	}
	return m.st.ok
}

func New(id, invite, code string) model {
	m := model{id: id, invite: invite, pending: code, width: 80, height: 24, now: time.Now(),
		lastKey: time.Now(), mins: slices.Index(minuteChoices(), 25)}
	m.setWorld(rooms.SnapshotWorld())
	nick, p, ok, err := store.GetUser(id)
	if err != nil {
		log.Error("getUser", "err", err)
	}
	m.nick = nick
	m.pcolor = pet.ColorIndex(p.Color)
	m.pcSaved = m.pcolor
	m.theme = themeIndex(store.Theme(id))
	m.saved, m.st = m.theme, newStyles(themes[m.theme])
	m.lang = lang(store.Lang(id))
	m.langSaved = m.lang
	m.vis = store.GetGarden(id)
	m.visSaved = m.vis
	if !ok {
		m.in = newInput(m.t("seu apelido"), 16)
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
	if m.pending != "" && m.nick != "" { // "ssh -t <servidor> <código>": entra sem passar pelo menu
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
			bare: true, w: w, h: h, wall: i}
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
		if m.garden != nil && m.garden.mode >= gardenEditor && !m.gardenEditorFits() {
			m.garden, m.notice = nil, gardenEditorSizeHint
		}
	case tea.ColorProfileMsg:
		m.pixel = msg.Profile >= colorprofile.ANSI256
	case tickMsg:
		m.tick++
		m.now = msg.now
		m.setWorld(msg.w)
		if m.notice != m.noticeSeen {
			m.noticeSeen, m.noticeAt = m.notice, m.now
		} else if m.notice != "" && m.now.Sub(m.noticeAt) >= noticeTTL {
			m.notice = ""
		}
		active := m.active()
		switch {
		case m.garden != nil && !m.garden.keep && !active: // o foco acabou ou esgotou: volta para responder
			m.garden = nil
		case m.garden == nil && active && m.nick != "" && m.vis.Auto && m.now.Sub(m.lastKey) >= gardenIdle:
			var cmd tea.Cmd
			m, cmd = m.openGarden(gardenFocus)
			return m, tea.Batch(doTick(), cmd)
		}
		m = m.syncSide()
		next := doTick()
		if m.boardStale() { // virou o dia: ontem saiu do retrato e entra pelo banco
			m.bday = dayKey(m.now) // uma leitura só, até a resposta chegar
			next = tea.Batch(next, m.loadBoard())
		}
		if m.side != nil && !m.side.ticking && m.garden == nil { // põe os 8 fps para andar
			m.side.ticking = true
			return m, tea.Batch(next, sideTick(m.side.gen))
		}
		if m.tab == tabDev { // os números mudam também por quem foca: a tabela acompanha o tick
			return m, tea.Batch(next, devDo(nil))
		}
		return m, next
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
		if msg.code == m.left { // entrou de novo na sala de que tinha saído
			m.left = ""
		}
		m.setWorld(msg.w)
		m.room, m.picking, m.form, m.err, m.arrived = msg.code, false, formNone, "", time.Now()
		m.notice = m.tf("você entrou em %s", m.roomName(m.w.Rooms[msg.code]))
		if msg.created {
			m.notice = "" // o código está no título e o convite embaixo explica o que fazer
			if m.solo() {
				m.notice = m.t("sala solo criada: só você entra aqui")
			}
		}
		return m, m.loadProfile()
	case nickMsg:
		switch {
		case msg.err != nil && strings.Contains(msg.err.Error(), "UNIQUE"):
			m.err = m.t("esse apelido já está em uso")
		case msg.err != nil:
			m.err = m.lang.errText(msg.err)
		default:
			m.nick, m.err = msg.nick, ""
			// escolher a cor do pet é o passo seguinte ao apelido: o perfil abre já no modo
			// personalizar, com o pet grande do lado mudando junto com a escolha
			m.tab, m.custom, m.first = tabProfile, true, true
			m.previewMood = pet.MoodOf(pet.EnergyStart)
			cmds := []tea.Cmd{m.loadProfile()}
			if m.pending != "" { // o código do comando esperava o apelido do primeiro acesso
				code := m.pending
				m.pending = ""
				cmds = append(cmds, m.join(code))
			}
			return m, tea.Batch(cmds...)
		}
	case devMsg:
		if msg.at.After(m.dev.at) {
			m.dev.rows, m.dev.at, m.dev.row = msg.rows, msg.at, min(m.dev.row, max(len(msg.rows)-1, 0))
		}
		if msg.err != nil {
			m.err = msg.err.Error()
		}
	case profileMsg:
		m.pet, m.tasks, m.days, m.totals = msg.pet, msg.tasks, msg.days, msg.totals
	case boardMsg:
		switch {
		case msg.err != nil: // sem marcar o dia como lido: o próximo tick tenta de novo
			m.err, m.bday = msg.err.Error(), ""
		case msg.period == m.bper: // a resposta de um período que você já deixou não vale
			m.bmins, m.bfor, m.bday = msg.mins, msg.period, msg.day
		}
	case errMsg:
		m.err = m.lang.errText(msg.error)
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
		if m.garden.mode >= gardenEditor {
			return m.gardenEditorKey(msg)
		}
		switch k {
		case "left":
			return m.swapWallpaper(m.garden, -1)
		case "right":
			return m.swapWallpaper(m.garden, 1)
		}
		m.garden = nil
		return m, nil
	}
	n := tabs()
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
	case tabProfile:
		return m.profileKey(msg)
	case tabVisual:
		return m.visualKey(msg)
	case tabDev:
		return m.devKey(msg)
	}
	return m, nil
}

func (m model) switchTab(t tab) (tea.Model, tea.Cmd) {
	if m.tab == tabVisual && t != tabVisual { // saiu sem salvar: desfaz a prévia
		m.theme, m.st, m.vis, m.lang = m.saved, newStyles(themes[m.saved]), m.visSaved, m.langSaved
	}
	if m.tab == tabProfile && t != tabProfile { // idem para a cor do pet
		m.pcolor, m.custom = m.pcSaved, false
	}
	m.obraFull = false
	m.tab, m.err, m.notice, m.entering, m.choosing, m.picking, m.form = t, "", "", false, false, false, formNone
	if t == tabProfile {
		return m, m.loadProfile()
	}
	if t == tabDev {
		m.dev.input = ""
		return m, devDo(nil)
	}
	return m, nil
}

// typing: algum campo de texto está recebendo as teclas (então "q", "a" e "d" são texto).
// Os formulários sem campo (o tamanho da sala, a confirmação de saída) não entram: neles as
// teclas de sempre valem, inclusive o q.
func (m model) typing() bool {
	return m.entering || m.cname || m.form == formNew || m.form == formCode || m.tab == tabDev && m.dev.input != ""
}

func (m model) nickKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() != "enter" {
		var cmd tea.Cmd
		m.in, cmd = m.in.Update(msg)
		m.err = "" // o erro era do apelido de antes
		return m, cmd
	}
	nick := strings.TrimSpace(m.in.Value())
	if m.err = nickProblem(m.lang, nick); m.err != "" {
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

// nickMax: o apelido mais largo, em colunas do terminal. Caracteres largos (日, 🐱) valem 2.
const nickMax = 16 // docs/regras.md

// nickProblem: o que impede o apelido, ou "" se ele serve. Cada caractere ocupa uma ou duas
// colunas sozinho: acento combinante, emoji composto e caractere invisível ficam de fora, porque
// cada terminal conta a largura deles de um jeito e a sala e o jardim desalinhavam (#84, #85).
func nickProblem(l lang, nick string) string {
	if nick == "" {
		return l.t("escreva um apelido")
	}
	if strings.ContainsAny(nick, " :") {
		return l.t("apelido sem espaços, por favor")
	}
	if ansi.StringWidth(nick) > nickMax {
		return l.tf("apelido de até %d letras, por favor (日 e 🐱 contam como 2)", nickMax)
	}
	state := -1
	for s := nick; s != ""; {
		var g string
		g, s, _, state = uniseg.FirstGraphemeClusterInString(s, state)
		if r := []rune(g); len(r) > 1 || !unicode.IsGraphic(r[0]) || ansi.StringWidth(g) == 0 {
			return l.t("apelido só com letras, números e símbolos simples, por favor")
		}
	}
	return ""
}

func (m model) View() tea.View {
	if m.garden != nil {
		v := tea.NewView(m.garden.frame)
		v.AltScreen = true
		return v
	}
	var body string
	if m.nick == "" {
		body = m.st.title.Render("sshiu") + "\n" + m.st.muted.Render(m.t("um lugar quieto para focar junto")) +
			"\n\n" + m.t("primeira vez por aqui. qual vai ser o seu apelido?") + "\n\n" + m.in.View() +
			m.errLine() + "\n\n" + m.st.muted.Render(m.t("enter confirma · sua chave SSH é o seu login"))
	} else {
		var help string
		content := m.roomView // o conteúdo recebe o espaço que sobra (largura, altura)
		switch m.tab {
		case tabRoom:
			help = m.roomHelp()
		case tabProfile:
			content, help = m.profileView, m.profileHelp()
		case tabVisual:
			content, help = func(int, int) string { return m.visualView() }, m.t("↑/↓ escolhe · ←/→ muda · j prévia · v editar no jardim · enter salva")
		case tabDev:
			content, help = func(int, int) string { return m.devView() }, m.devHelp()
		}
		inner := max(m.width-4, 1)
		footer := func(help string) string {
			nav := m.t("a/d troca de aba · q sai")
			if m.typing() {
				nav = m.t("tab troca de aba · ctrl+c sai")
			}
			if help != "" {
				nav = help + " · " + nav
			}
			// o rodapé quebra entre um atalho e outro, nunca no meio de um ("q" numa linha, "sai" na outra)
			nbsp := string(rune(0xa0)) // o espaço que não quebra
			nav = strings.ReplaceAll(strings.ReplaceAll(nav, " ", nbsp), nbsp+"·"+nbsp, nbsp+"· ")
			return m.st.muted.Width(inner).Render(nav)
		}
		// cabeçalho em cima, rodapé preso na última linha e o conteúdo no espaço do meio
		top, foot := m.header(), footer(help)
		h := max(1, m.height-2-lipgloss.Height(top)-2-lipgloss.Height(foot))
		c := content(inner, h-lipgloss.Height(m.errLine())+1)
		// as teclas de reagir só valem com o feed na tela, e é a altura que decide se ele cabe:
		// sem a legenda do feed no conteúdo, elas saem do rodapé
		keys := m.t(" · ↑/↓ evento · 1-4 reage")
		if m.tab == tabRoom && !strings.Contains(c, m.t("reagir ao evento selecionado: ")) && strings.Contains(help, keys) {
			foot = footer(strings.Replace(help, keys, "", 1))
			h = max(1, m.height-2-lipgloss.Height(top)-2-lipgloss.Height(foot))
		}
		mid := fit(c+m.errLine(), inner, h)
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

// ellipsize: s quebrada nas palavras em até n linhas de w colunas. Se o texto não cabe, a última linha
// termina em "…": antes a tarefa comprida sumia sem aviso, e parecia que ela acabava ali.
func ellipsize(s string, w, n int) string {
	lines := strings.Split(ansi.Wrap(s, w, ""), "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	lines = lines[:n]
	last := strings.TrimRight(lines[n-1], " ")
	if ansi.StringWidth(last) >= w { // a linha cheia abre espaço para a reticência
		last = ansi.Truncate(last, w-1, "")
	}
	lines[n-1] = last + "…"
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
	parts := []string{m.st.title.Render("sshiu") + " "}
	for i, name := range tabNames[:tabs()] {
		if tab(i) == m.tab {
			parts = append(parts, m.st.tabOn.Render(m.t(name)))
		} else {
			parts = append(parts, m.st.tabOff.Render(m.t(name)))
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
		return m.st.bad.Render(m.t("tempo esgotado"))
	case x.Status == rooms.Focusing:
		return m.tf("focando %s", mmss(x.Focus.Ends.Sub(m.now)))
	case x.Status == rooms.Idle:
		return m.st.muted.Render(m.t("sem tarefa"))
	}
	return m.st.muted.Render(m.t("offline"))
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

// paintBlocks pinta o texto de bigText com st, com "█" e "▀" invertidos (espaço e "▄" sublinhado):
// a cor vai para o fundo, que cobre a célula toda no Terminal.app (ver solid e pet.Pixels).
func paintBlocks(s string, st lipgloss.Style) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == ' ' {
			b.WriteRune(r)
			continue
		}
		if r2, rev := solid(r); rev {
			b.WriteString(st.Reverse(true).Underline(r2 == '▄').Render(string(r2)))
		} else {
			b.WriteString(st.Render(string(r)))
		}
	}
	return b.String()
}
