// PROTÓTIPO DESCARTÁVEL: a tela da sala no modelo novo (ADR 0002 a 0004), para decidir o layout
// vendo rodar por SSH. Sala falsa em memória, uma por sessão, com os 5 bots da demo; nada toca no
// banco nem em internal/tui. Não vai para o app: o layout escolhido é reescrito lá.
//
//	go build -o <rascunho>/proto-sala.exe ./prototype/sala   (e rode dentro de <rascunho>)
//	ssh -p 23235 localhost
//
// v troca o layout (A painel + cards, B lista, C lado a lado) · e troca emoji/ASCII nas reações
// ] pula para 5 s do fim do tempo declarado, para ver o continuar/trocar/parar sem esperar
package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"plantao/internal/pet"
	"runtime"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"github.com/charmbracelet/colorprofile"
)

var (
	accent = lipgloss.Color("#ff8700")
	title  = lipgloss.NewStyle().Bold(true).Foreground(accent)
	muted  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7a7a7a"))
	good   = lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd787"))
	bad    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f5f"))
	bold   = lipgloss.NewStyle().Bold(true)
	card   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4a4a4a")).Width(18).Align(lipgloss.Center)
	italic = lipgloss.NewStyle().Italic(true).Width(16).Height(2).MaxHeight(2).Align(lipgloss.Center)
	bar    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#11111b")).Background(accent).Padding(0, 1)
)

var (
	layouts    = []string{"A painel + cards", "B lista compacta", "C lado a lado"}
	emoji      = [4]string{"👏", "🔥", "💪", "☕"}
	ascii      = [4]string{`\o/`, "+1", "<3", "^^"}
	minutes    = []int{10, 15, 25, 45, 60}
	botTasks   = map[string][]string{"ana": {"terminar a lista 3 de Cálculo", "revisar a prova de Álgebra"}, "bruno": {"revisar o PR do login", "escrever os testes do parser"}, "caio": {"ler o cap. 4 de Sistemas Operacionais", "resumir o artigo de IHC"}, "duda": {"começar o relatório de Física", "configurar o Docker do projeto"}, "edu": {"instalar o ambiente do TCC", "abrir o projeto do TCC"}}
	statusVerb = [...]string{offline: "offline", idle: "parado", focusing: "focando"}
)

type status int

const (
	offline status = iota
	idle
	focusing
)

type member struct {
	nick   string
	energy float64
	st     status
	task   string
	ends   time.Time
	got    int // reações recebidas: no card só aparece o contador (ADR 0004)
}

type event struct {
	id         int
	who        *member
	verb, task string
	at         time.Time
	n          [4]int
	mine       [4]bool
}

type reaction struct {
	at time.Time
	ev *event
	k  int
}

type phase int

const (
	pIdle phase = iota
	pWriting
	pPicking
	pFocus
	pExpired // tempo esgotado: os minutos param de contar até você responder
)

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type model struct {
	now           time.Time
	tick, width   int
	pixel, ascii  bool
	layout        int
	ph            phase
	in            textinput.Model
	task          string
	mins          int  // índice em minutes
	cont          bool // escolhendo tempo para continuar a mesma tarefa
	started, ends time.Time
	focused       time.Duration // minutos focados hoje, sem o trecho em andamento
	ms            []*member     // ms[0] é você
	evs           []*event      // mais novo primeiro
	sel, nextID   int
	pending       []reaction
}

func newModel() *model {
	now := time.Now()
	m := &model{now: now, width: 100, mins: 2}
	m.ms = []*member{
		{nick: "você", energy: 55, st: idle},
		{nick: "ana", energy: 92, st: focusing, task: botTasks["ana"][0], ends: now.Add(50 * time.Second)},
		{nick: "bruno", energy: 68, st: focusing, task: botTasks["bruno"][0], ends: now.Add(75 * time.Second)},
		{nick: "caio", energy: 48, st: idle},
		{nick: "duda", energy: 27, st: focusing, task: botTasks["duda"][1], ends: now.Add(20 * time.Second)},
		{nick: "edu", energy: 8, st: offline},
	}
	m.add(m.ms[1], "cumpriu", botTasks["ana"][1], now.Add(-3*time.Minute)).n = [4]int{2, 1, 0, 0}
	m.ms[1].got = 3
	m.add(m.ms[4], "começou", botTasks["duda"][1], now.Add(-40*time.Second))
	m.add(m.ms[1], "começou", botTasks["ana"][0], now.Add(-30*time.Second))
	m.add(m.ms[2], "começou", botTasks["bruno"][0], now.Add(-15*time.Second))
	return m
}

func (m *model) add(who *member, verb, task string, at time.Time) *event {
	m.nextID++
	e := &event{id: m.nextID, who: who, verb: verb, task: task, at: at}
	m.evs = append([]*event{e}, m.evs...)
	if len(m.evs) > 20 {
		m.evs = m.evs[:20]
	}
	// os bots reagem: sempre a você, às vezes aos outros
	n := 0
	if who == m.ms[0] {
		n = 1 + rand.N(3)
	} else if rand.Float64() < 0.3 {
		n = 1
	}
	for range n {
		m.pending = append(m.pending, reaction{m.now.Add(2*time.Second + rand.N(6*time.Second)), e, rand.N(4)})
	}
	return e
}

func (m *model) Init() tea.Cmd { return doTick() }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.ColorProfileMsg:
		m.pixel = msg.Profile >= colorprofile.ANSI256
	case tickMsg:
		m.tick++
		m.now = time.Time(msg)
		m.step()
		return m, doTick()
	case tea.KeyPressMsg:
		return m.key(msg.String(), msg)
	}
	return m, nil
}

// step avança a sala falsa: seu tempo esgotando, bots começando e cumprindo, reações chegando.
func (m *model) step() {
	if m.ph == pFocus && !m.now.Before(m.ends) {
		m.ph, m.focused = pExpired, m.focused+m.ends.Sub(m.started)
	}
	for _, b := range m.ms[1:] {
		switch {
		case b.st == focusing && m.now.After(b.ends):
			b.st = idle
			if rand.Float64() < 0.75 {
				m.add(b, "cumpriu", b.task, m.now)
			}
		case b.st == idle && rand.N(40) == 0:
			b.st, b.task, b.ends = focusing, botTasks[b.nick][rand.N(2)], m.now.Add(40*time.Second+rand.N(50*time.Second))
			m.add(b, "começou", b.task, m.now)
		case b.st == idle && rand.N(300) == 0, b.st == offline && rand.N(120) == 0:
			b.st = 1 - b.st // parado <-> offline
		}
	}
	rest := m.pending[:0]
	for _, r := range m.pending {
		if m.now.Before(r.at) {
			rest = append(rest, r)
			continue
		}
		r.ev.n[r.k]++
		r.ev.who.got++
	}
	m.pending = rest
}

func (m *model) key(k string, msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if k == "ctrl+c" {
		return m, tea.Quit
	}
	if m.ph == pWriting {
		switch k {
		case "enter":
			if t := strings.TrimSpace(m.in.Value()); t != "" {
				m.task, m.cont, m.ph = t, false, pPicking
			}
		case "esc":
			m.ph = pIdle
		default:
			var cmd tea.Cmd
			m.in, cmd = m.in.Update(msg)
			return m, cmd
		}
		return m, nil
	}
	switch k {
	case "q":
		return m, tea.Quit
	case "v":
		m.layout = (m.layout + 1) % len(layouts)
	case "e":
		m.ascii = !m.ascii
	case "up", "down":
		i := m.selIndex()
		if k == "up" {
			i--
		} else {
			i++
		}
		m.sel = m.evs[max(0, min(i, m.shown()-1))].id
	case "1", "2", "3", "4":
		if e := m.evs[m.selIndex()]; e.who != m.ms[0] {
			r := int(k[0] - '1')
			d := 1
			if e.mine[r] {
				d = -1
			}
			e.mine[r], e.n[r], e.who.got = !e.mine[r], e.n[r]+d, e.who.got+d
		}
	case "]":
		if m.ph == pFocus {
			d := m.ends.Sub(m.now) - 5*time.Second
			m.started, m.ends = m.started.Add(-d), m.ends.Add(-d)
		}
	}
	switch m.ph {
	case pIdle:
		if k == "enter" {
			m.write()
		}
	case pPicking:
		switch k {
		case "left":
			m.mins = max(0, m.mins-1)
		case "right":
			m.mins = min(len(minutes)-1, m.mins+1)
		case "enter":
			if !m.cont {
				m.add(m.ms[0], "começou", m.task, m.now)
			}
			m.ph, m.started, m.ends = pFocus, m.now, m.now.Add(time.Duration(minutes[m.mins])*time.Minute)
		case "esc":
			m.ph = map[bool]phase{true: pExpired, false: pIdle}[m.cont]
		}
	case pFocus, pExpired:
		if m.ph == pFocus && (k == "f" || k == "t" || k == "p") {
			m.focused += m.now.Sub(m.started)
		}
		switch k {
		case "f":
			m.add(m.ms[0], "cumpriu", m.task, m.now)
			m.ph = pIdle
		case "t": // encerra como não cumprida, sem evento: ninguém é exposto por trocar
			m.write()
		case "p":
			m.ph = pIdle
		case "c":
			if m.ph == pExpired {
				m.ph, m.cont = pPicking, true
			}
		}
	}
	return m, nil
}

func (m *model) write() {
	m.ph, m.in = pWriting, textinput.New()
	m.in.Placeholder = "ex.: fazer o exercício 1 da lista 3"
	m.in.CharLimit = 80
	m.in.SetWidth(50)
	m.in.Focus()
}

func (m *model) shown() int { return min(len(m.evs), 6) }

func (m *model) selIndex() int {
	for i, e := range m.evs[:m.shown()] {
		if e.id == m.sel {
			return i
		}
	}
	return 0
}

// --- pedaços comuns aos três layouts ---

func (m *model) glyph(k int) string {
	if m.ascii {
		return ascii[k]
	}
	return emoji[k]
}

func (m *model) focusedMin() int {
	d := m.focused
	if m.ph == pFocus {
		d += m.now.Sub(m.started)
	}
	return int(d.Minutes())
}

// statusOf: o que a sala vê de um membro, com o tempo que falta.
func (m *model) statusOf(x *member) (string, string) {
	if x == m.ms[0] {
		switch m.ph {
		case pFocus:
			return "focando " + mmss(m.ends.Sub(m.now)), m.task
		case pExpired:
			return bad.Render("tempo esgotado"), m.task
		}
		return "parado", ""
	}
	if x.st == focusing {
		return "focando " + mmss(x.ends.Sub(m.now)), x.task
	}
	return muted.Render(statusVerb[x.st]), ""
}

// you: seu foco, conforme a fase.
func (m *model) you(clock bool) string {
	var head, body string
	switch m.ph {
	case pIdle:
		head = bold.Render("Você está parado.")
		body = muted.Render(fmt.Sprintf("%d min focados hoje", m.focusedMin()))
	case pWriting:
		head = bold.Render("Qual a próxima coisa concreta que você vai fazer?")
		body = m.in.View()
	case pPicking:
		head = bold.Render(m.task)
		body = "Quanto tempo?  " + muted.Render("‹ ") + title.Render(fmt.Sprintf("%d min", minutes[m.mins])) + muted.Render(" ›")
		if m.cont {
			body = "Mais quanto tempo?  " + muted.Render("‹ ") + title.Render(fmt.Sprintf("%d min", minutes[m.mins])) + muted.Render(" ›")
		}
	case pFocus:
		head = bold.Render(m.task)
		body = muted.Render(fmt.Sprintf("de %d min · %d min focados hoje", minutes[m.mins], m.focusedMin()))
	case pExpired:
		head = bad.Render("Tempo esgotado: ") + bold.Render(m.task)
		body = "c continuar · t trocar · p parar · f feito\n" + muted.Render(fmt.Sprintf("os minutos pararam de contar · %d min focados hoje", m.focusedMin()))
	}
	text := head + "\n" + body
	if !clock || (m.ph != pFocus && m.ph != pExpired) {
		return text
	}
	c := title
	if m.ph == pExpired {
		c = bad
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, c.Render(bigClock(m.ends.Sub(m.now))), "   ", text)
}

func (m *model) feed() string {
	var b strings.Builder
	sel, evs := m.selIndex(), m.evs[:m.shown()]
	text := func(e *event) string {
		return muted.Render(e.at.Format("15:04")) + " " + bold.Render(e.who.nick) + " " + e.verb + " \"" + e.task + "\""
	}
	col := 0 // as reações começam numa coluna só, depois do texto mais longo
	for _, e := range evs {
		col = max(col, lipgloss.Width(text(e)))
	}
	for i, e := range evs {
		mark := "  "
		if i == sel {
			mark = title.Render("> ")
		}
		line := lipgloss.NewStyle().Width(col).Render(text(e))
		for k, n := range e.n { // ordem fixa 1-4: quem tem zero sai e os seguintes andam para a esquerda
			if n == 0 {
				continue
			}
			r := fmt.Sprintf("%s %d", m.glyph(k), n) // o espaço protege o número de emoji mais largos que 2 colunas
			if e.mine[k] {
				r = title.Render(r)
			}
			line += "  " + r
		}
		b.WriteString(mark + line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m *model) legend() string {
	var ks []string
	for k := range 4 {
		ks = append(ks, fmt.Sprintf("%d %s", k+1, m.glyph(k)))
	}
	return muted.Render("reagir ao evento selecionado: ") + strings.Join(ks, "  ")
}

func (m *model) help() string {
	h := map[phase]string{pIdle: "enter começa uma tarefa", pWriting: "enter confirma · esc volta",
		pPicking: "←/→ escolhe · enter começa · esc volta", pFocus: "f feito · t trocar · p parar",
		pExpired: "c continuar · t trocar · p parar · f feito"}[m.ph]
	if m.ph != pWriting {
		h += " · ↑/↓ evento · 1-4 reage · q sai"
	}
	return muted.Render(h)
}

func (m *model) switcher() string {
	r := "emoji"
	if m.ascii {
		r = "ASCII"
	}
	extra := ""
	if m.ph == pFocus {
		extra = " · ] pula p/ o fim"
	}
	return bar.Render("PROTÓTIPO · v layout " + layouts[m.layout] + " · e reações: " + r + extra)
}

func (m *model) header() string {
	return title.Render("PLANTÃO") + "  " + bold.Render("sala demo") + muted.Render(" · código K7Q2")
}

// --- os três layouts ---

func (m *model) View() tea.View {
	var body string
	switch m.layout {
	case 0:
		body = m.viewA()
	case 1:
		body = m.viewB()
	default:
		body = m.viewC()
	}
	v := tea.NewView(lipgloss.NewStyle().Padding(1, 2).Render(body + "\n\n" + m.help() + "\n" + m.switcher()))
	v.AltScreen = true
	return v
}

// A: seu foco em cima, um card por membro, eventos embaixo.
func (m *model) viewA() string {
	per := max(1, (m.width-4)/18)
	var rows []string
	for i := 0; i < len(m.ms); i += per {
		var row []string
		for _, x := range m.ms[i:min(i+per, len(m.ms))] {
			st, task := m.statusOf(x)
			c := card
			if x == m.ms[0] {
				c = c.BorderForeground(accent)
			}
			got := " "
			if x.got > 0 {
				got = muted.Render(map[bool]string{true: "1 reação", false: fmt.Sprintf("%d reações", x.got)}[x.got == 1])
			}
			mood := pet.MoodOf(x.energy)
			row = append(row, c.Render(lipgloss.JoinVertical(lipgloss.Center,
				pet.Render(mood, m.tick, m.pixel), bold.Render(x.nick)+" "+muted.Render(pet.MoodName[mood]),
				italic.Render(task), st, got)))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	}
	return m.header() + "\n\n" + m.you(true) + "\n\n" + strings.Join(rows, "\n") + "\n\n" + m.feed() + "\n" + m.legend()
}

// memberLines: uma linha por membro (layouts B e C).
func (m *model) memberLines(taskWidth int) string {
	var b strings.Builder
	for _, x := range m.ms {
		mood := pet.MoodOf(x.energy)
		dot := lipgloss.NewStyle().Foreground(lipgloss.Color(pet.MoodColor[mood])).Render("●")
		if x.st == offline {
			dot = muted.Render("○")
		}
		st, task := m.statusOf(x)
		got := ""
		if x.got > 0 {
			got = muted.Render(fmt.Sprintf(" ·%d", x.got))
		}
		name := lipgloss.NewStyle().Width(7).Render(x.nick)
		if x == m.ms[0] {
			name = title.Width(7).Render(x.nick)
		}
		line := dot + " " + name + lipgloss.NewStyle().Width(15).Render(st)
		if taskWidth > 0 {
			line += lipgloss.NewStyle().Italic(true).MaxWidth(taskWidth).Render(task)
		}
		b.WriteString(line + got + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// B: uma linha para você, uma por membro, eventos e a legenda das reações.
func (m *model) viewB() string {
	return m.header() + "\n\n" + m.you(false) + "\n\n" + m.memberLines(max(10, m.width-40)) + "\n\n" + m.feed() + "\n\n" + m.legend()
}

// C: você à esquerda (pet grande e relógio), a sala à direita.
func (m *model) viewC() string {
	left := lipgloss.NewStyle().Width(36).Render(lipgloss.JoinVertical(lipgloss.Left,
		pet.RenderBig(pet.MoodOf(m.ms[0].energy), m.tick, m.pixel), "", m.youC()))
	rw := max(30, m.width-4-36-3)
	right := lipgloss.NewStyle().Width(rw).BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).
		BorderForeground(lipgloss.Color("#4a4a4a")).PaddingLeft(1).
		Render(m.memberLines(max(0, rw-30)) + "\n" + muted.Render(strings.Repeat("─", rw-2)) + "\n" + m.feed() + "\n\n" + m.legend())
	return m.header() + "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

func (m *model) youC() string {
	if m.ph != pFocus && m.ph != pExpired {
		return lipgloss.NewStyle().Width(34).Render(m.you(false))
	}
	c := title
	if m.ph == pExpired {
		c = bad
	}
	return c.Render(bigClock(m.ends.Sub(m.now))) + "\n\n" + lipgloss.NewStyle().Width(34).Render(m.you(false))
}

// --- relógio (cópia de internal/tui) ---

func mmss(d time.Duration) string {
	d = max(d, 0).Round(time.Second)
	return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}

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

// --- servidor (cópia enxuta de main.go, com os contornos do Windows) ---

func main() {
	port := flag.String("port", "23235", "porta SSH")
	flag.Parse()
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort("localhost", *port)),
		wish.WithHostKeyPath(".ssh/proto_host_key"),
		wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool { return true }),
		wish.WithMiddleware(bubbletea.MiddlewareWithProgramHandler(programHandler), activeterm.Middleware()),
	)
	if err != nil {
		log.Fatal("wish", "err", err)
	}
	log.Info("protótipo da sala: ssh localhost -p " + *port)
	log.Fatal(s.ListenAndServe())
}

func programHandler(s ssh.Session) *tea.Program {
	opts := bubbletea.MakeOptions(s)
	if runtime.GOOS == "windows" {
		pty, _, _ := s.Pty()
		opts = append(opts,
			tea.WithColorProfile(colorprofile.Env(append(s.Environ(), "TERM="+pty.Term))),
			tea.WithOutput(lfWriter{s}))
	}
	return tea.NewProgram(newModel(), opts...)
}

type lfWriter struct{ io.Writer }

func (w lfWriter) Write(p []byte) (int, error) {
	b := make([]byte, 0, len(p)+8)
	for i, c := range p {
		if c == '\n' && (i == 0 || p[i-1] != '\r') {
			b = append(b, "\x1bD"...)
		} else {
			b = append(b, c)
		}
	}
	if _, err := w.Writer.Write(b); err != nil {
		return 0, err
	}
	return len(p), nil
}
