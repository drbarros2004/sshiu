package tui

// Aba "sala": o seu foco em cima, um card por membro e os eventos com as reações embaixo
// (o layout A do protótipo, ver IDEIAS.md).

import (
	"fmt"
	"slices"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var reactionGlyphs = [rooms.Reactions]string{"👏", "🔥", "💪", "☕"}

const (
	shownEvents = 6
	panelCols   = 54 // largura do painel da sala solo
	minGarden   = 36 // menos que isso não dá jardim: a tela fica só com o painel
	soloEvents  = 3  // linhas do mini log da sala solo
	clockCols   = 20 // o relógio grande: 5 glifos de 4 colunas
)

// solo: a sala é de uma pessoa (ADR 0006), então a tela não tem cards, reações nem código.
func (m model) solo() bool { return m.w.Rooms[m.room].Max == rooms.MinMembers }

// soloGardenWidth: o que sobra para o jardim ao lado do painel (0 = não cabe).
func soloGardenWidth(inner int) int {
	if w := inner - panelCols - 2; w >= minGarden {
		return w
	}
	return 0
}

// soloRows: a altura das colunas da sala solo, fora o cabeçalho, o título e a ajuda.
func soloRows(height int) int { return max(12, height-9) }

// minuteChoices: os tempos declarados que dá para escolher (na demo, 1 min para a plateia ver o fluxo).
func minuteChoices() []int {
	if rooms.Demo {
		return []int{1, 5, 10, 15, 20, 25, 30, 45, 60}
	}
	return []int{5, 10, 15, 20, 25, 30, 45, 60}
}

// events: os eventos navegáveis. A sala solo não tem nenhum (nem cursor, nem reações); o que
// você fez nela aparece no mini log.
func (m model) events() []rooms.Event {
	if m.solo() {
		return nil
	}
	ev := m.w.Rooms[m.room].Events
	return ev[:min(len(ev), shownEvents)]
}

func (m model) selIndex() int {
	return max(0, slices.IndexFunc(m.events(), func(e rooms.Event) bool { return e.ID == m.sel }))
}

// inPicker: o menu de salas está na tela, seja porque você abriu com s, seja porque você ainda
// não está em nenhuma sala (o primeiro acesso cai nele).
func (m model) inPicker() bool { return m.picking || m.room == "" }

func (m model) roomKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.inPicker() {
		return m.pickerKey(msg)
	}
	if m.cfg {
		return m.cfgKey(msg)
	}
	k, id, code := msg.String(), m.id, m.room
	m.notice = "" // o aviso de "entrou na sala" some na primeira tecla
	me := m.me()
	focusing := me.Status == rooms.Focusing
	expired := focusing && me.Focus.Expired(m.now)
	var cmd tea.Cmd
	switch {
	case m.entering:
		switch k {
		case "enter":
			if strings.TrimSpace(m.in.Value()) == "" {
				m.err = m.t("escreva a próxima coisa concreta que você vai fazer")
				return m, nil
			}
			m.entering, m.choosing, m.cont, m.err = false, true, false, ""
		case "esc":
			m.entering, m.err = false, ""
		default:
			m.in, cmd = m.in.Update(msg)
		}
		return m, cmd
	case m.choosing:
		choices := minuteChoices()
		m.mins = min(m.mins, len(choices)-1)
		switch k {
		case "left":
			m.mins = max(0, m.mins-1)
		case "right":
			m.mins = min(len(choices)-1, m.mins+1)
		case "esc":
			m.choosing = false
		case "enter":
			m.choosing = false
			task, mins := strings.TrimSpace(m.in.Value()), choices[m.mins]
			if m.cont {
				return m, act(func() { rooms.Continue(id, mins) })
			}
			return m, act(func() { rooms.Start(id, code, task, mins) })
		}
		return m, nil
	}
	switch k {
	case "left", "right": // na sala solo, trocam o wallpaper da coluna, como na tela cheia
		if m.side != nil {
			return m.swapWallpaper(m.side, map[string]int{"left": -1, "right": 1}[k])
		}
	case "up", "down":
		ev := m.events()
		if len(ev) == 0 {
			return m, nil
		}
		i := m.selIndex()
		if k == "up" {
			i = max(0, i-1)
		} else {
			i = min(len(ev)-1, i+1)
		}
		m.sel = ev[i].ID
	case "1", "2", "3", "4":
		if ev := m.events(); len(ev) > 0 {
			eid, r := ev[m.selIndex()].ID, int(k[0]-'1')
			return m, act(func() { rooms.React(id, code, eid, r) })
		}
	case "s":
		return m.openPicker()
	case "enter":
		if !focusing {
			m.entering, m.in = true, newInput(m.t("ex.: fazer o exercício 1 da lista 3"), 80)
		}
	case "f":
		if focusing { // em sequência: o loadProfile só vale depois de a tarefa estar gravada
			return m, tea.Sequence(act(func() { rooms.Finish(id, pet.Done) }), m.loadProfile())
		}
	case "t": // encerra como não cumprida, sem evento: ninguém é exposto por trocar
		if focusing {
			m.entering, m.in = true, newInput(m.t("ex.: fazer o exercício 1 da lista 3"), 80)
			return m, tea.Sequence(act(func() { rooms.Finish(id, pet.NotDone) }), m.loadProfile())
		}
	case "p":
		if focusing {
			return m, tea.Sequence(act(func() { rooms.Finish(id, pet.NotDone) }), m.loadProfile())
		}
	case "c":
		switch {
		case expired: // o painel está em "c continuar": a configuração espera a resposta
			m.choosing, m.cont = true, true
		case m.owns():
			return m.openConfig()
		}
	case "o": // a obra sozinha na tela, para quem não tem altura para ela junto com os cards
		if m.obraOf(m.w.Rooms[code]) != nil {
			m.obraFull = !m.obraFull
		}
	case "y": // o servidor não sabe se o terminal aceitou o OSC 52: o aviso diz só que mandou
		if !m.solo() && code != rooms.DemoCode { // as duas salas sem código para passar adiante
			m.notice = m.t("código enviado para a área de transferência")
			return m, tea.SetClipboard(code)
		}
	case "j": // na sala solo o jardim é a outra versão da tela, e abre a qualquer momento
		if m.solo() || focusing && !expired {
			return m.openGarden(gardenFocus)
		}
	}
	return m, nil
}

func (m model) roomHelp() string {
	me := m.me()
	focusing := me.Status == rooms.Focusing
	switch {
	case m.obraFull:
		return m.t("o volta para a sala")
	case m.cfg:
		return m.cfgHelp()
	case m.inPicker() && m.form == formNew:
		return m.t("↑/↓ escolhe a linha · ←/→ muda · enter cria · esc volta")
	case m.inPicker() && m.form == formJoin:
		return m.t("enter entra · esc volta")
	case m.inPicker() && m.form != formNone:
		return m.t("enter confirma · esc volta")
	case m.inPicker():
		return m.t("↑/↓ escolhe · enter abre") + map[bool]string{true: m.t(" · esc volta"), false: ""}[m.room != ""]
	case m.entering:
		return m.t("enter confirma · esc volta")
	case m.choosing:
		return m.t("←/→ escolhe · enter começa · esc volta")
	case focusing && me.Focus.Expired(m.now): // as escolhas já estão em destaque no painel
		return m.t("s salas") + m.reactHelp()
	case focusing:
		return m.t("f feito · t trocar · p parar · j modo jardim") + m.sideHint() + m.reactHelp() + m.obraHint() + m.cfgHint() + m.t(" · s salas")
	}
	if m.solo() { // na sala solo o jardim abre sem tarefa: é a outra versão da tela
		return m.t("enter começa uma tarefa · j modo jardim") + m.cfgHint() + m.t(" · s salas")
	}
	return m.t("enter começa uma tarefa") + m.reactHelp() + m.obraHint() + m.cfgHint() + m.t(" · s salas")
}

// cfgHint: a tecla do menu, que só o dono tem (ADR 0007). Fica de fora com o tempo esgotado,
// porque aí o c é "continuar" e o painel já diz isso.
func (m model) cfgHint() string {
	if !m.owns() {
		return ""
	}
	return m.t(" · c configura")
}

// sideHint: as setas do wallpaper, que só existem com a coluna do jardim na tela.
func (m model) sideHint() string {
	if m.side == nil {
		return ""
	}
	return m.t(" · ←/→ wallpaper")
}

// obraHint: a tecla da obra, que só existe onde há obra para ver.
func (m model) obraHint() string {
	if m.obraOf(m.w.Rooms[m.room]) == nil {
		return ""
	}
	return m.t(" · o obra")
}

// reactHelp: as teclas de evento e reação, que a sala solo não tem.
func (m model) reactHelp() string {
	if m.solo() {
		return ""
	}
	return m.t(" · ↑/↓ evento · 1-4 reage")
}

// roomView: a sala em w×h células. Quando falta altura, a lista de eventos encurta e depois os
// pets viram um ponto da cor do corpo do pet.
// ponytail: com muita gente num terminal baixo os cards ainda passam da altura e são cortados;
// rolagem nos cards se isso aparecer de verdade.
func (m model) roomView(w, h int) string {
	if m.inPicker() {
		return m.pickerView()
	}
	s := m.w.Rooms[m.room]
	if m.cfg {
		return m.cfgView()
	}
	me := m.me()
	// com o tempo esgotado o c é "continuar", e digitando a tarefa as teclas são texto
	busy := m.entering || m.choosing || me.Status == rooms.Focusing && me.Focus.Expired(m.now)
	title := bold.Render(m.roomName(s))
	if m.room != rooms.DemoCode && !m.solo() { // a sala solo não tem código para passar
		title += m.st.muted.Render(m.t(" · código ") + m.room)
		if !busy {
			title += m.st.muted.Render(m.t(" · y copia"))
		}
	}
	if !m.solo() { // quanto a sala comporta, embaixo do nome; a solo é sempre uma pessoa
		line := m.tf("%d/%d pessoas", len(s.Members), s.Max)
		if s.Closed {
			line += m.t(" · fechada")
		}
		title += "\n" + m.st.muted.Render(line)
	}
	if m.notice != "" {
		title += "\n" + m.st.ok.Render(m.notice)
	}
	if m.obraFull { // a obra sozinha, na tecla o
		return m.obraView(s, w, h)
	}
	if m.solo() { // a sala solo tem uma pessoa e nenhum card: ela já cabe
		return title + "\n\n" + m.soloView()
	}
	var out string
	for _, st := range m.fitSteps(s, w) {
		// sem eventos só depois de a obra ou os pets já terem encolhido
		for n := shownEvents; n >= 0 && (st.compact || st.ob != "" || n > 0); n -= 2 {
			out = m.roomBody(s, title, w, n, st.compact, st.ob)
			if lipgloss.Height(out) <= h {
				return out
			}
		}
	}
	return out
}

// fitStep: uma tentativa de caber na altura, da tela cheia para a mais apertada.
type fitStep struct {
	ob      string // a obra desenhada, ou "" quando ela já não cabe
	compact bool   // os pets viram um ponto da cor do corpo do pet
}

// fitSteps: a ordem em que a tela abre mão das coisas. Dentro de cada passo o feed encurta
// primeiro; depois a obra troca o arranjo do tema por uma fileira só, depois os pets viram ponto,
// e a obra sai por último — tirando ela, os pets voltam ao tamanho normal.
//
// Ela saía primeiro, e numa janela pequena isso queria dizer nunca aparecer: numa sala de seis
// pessoas a obra pedia umas 50 linhas, que é mais do que um terminal costuma ter (ADR 0008).
func (m model) fitSteps(s rooms.Snapshot, w int) []fitStep {
	steps := []fitStep{{"", false}, {"", true}}
	full := m.obraBlock(s, w, false)
	if full == "" {
		return steps
	}
	first := []fitStep{{full, false}}
	if flat := m.obraBlock(s, w, true); flat != full { // o trem já é uma fileira: achatar não muda
		first = append(first, fitStep{flat, false}, fitStep{flat, true})
	} else {
		first = append(first, fitStep{full, true})
	}
	return append(first, steps...)
}

// roomBody: a tela montada com n eventos, os pets do tamanho que couber e, se ob não for vazio,
// a obra. O trem ocupa a largura e fica embaixo dos cards; a cidade cabe ao lado do log e vai
// para o canto que hoje fica vazio, descendo para baixo dele quando a janela é estreita.
func (m model) roomBody(s rooms.Snapshot, title string, w, n int, compact bool, ob string) string {
	out := title + "\n\n" + m.focusPanel() + "\n\n" + m.cards(s, w, compact)
	alone := len(s.Members) == 1 && m.room != rooms.DemoCode // sozinho numa sala de mais
	if alone {                                               // o convite vale mais que o feed vazio, e não encolhe
		if ob != "" {
			out += "\n\n" + ob
		}
		return out + "\n\n" + m.invitePanel(s)
	}
	var log string
	if n > 0 {
		log = m.feed(n) + "\n\n" + m.legend()
	}
	switch {
	case ob == "":
	case m.obraOf(s).wide:
		out += "\n\n" + ob
	case log == "" || w-lipgloss.Width(log)-obraGap < lipgloss.Width(ob):
		out += "\n\n" + ob // não cabe ao lado: desce para baixo do log
	default:
		log = lipgloss.JoinHorizontal(lipgloss.Top, log, strings.Repeat(" ", obraGap), ob)
	}
	if log != "" {
		out += "\n\n" + log
	}
	return out
}

// invitePanel: a sala recém-criada é a tela mais vazia do app, e o que falta nela é gente.
// Enquanto você é o único membro, o convite ocupa o lugar do feed, com os dois caminhos:
// quem já usa o app entra pelo menu de salas, e quem está chegando cola o comando no terminal.
func (m model) invitePanel(s rooms.Snapshot) string {
	row := func(label, text string) string {
		return m.st.muted.Render(fmt.Sprintf("  %-15s", label)) + text
	}
	if s.Closed { // sala fechada: o código não serve para ninguém enquanto a porta não abrir
		aviso := m.st.muted.Render(m.t("esta sala não está aceitando ninguém."))
		if m.owns() {
			aviso += m.st.muted.Render(m.t(" abra as novas entradas em ")) + "c" + m.st.muted.Render(m.t(" · configurar."))
		}
		return bold.Render(m.t("por enquanto, só tem você aqui.")) + "\n\n" + aviso
	}
	return bold.Render(m.t("por enquanto, só tem você aqui.")) + m.st.muted.Render(m.t(" para a turma entrar na sala:")) + "\n\n" +
		row(m.t("no sshiu"), m.st.muted.Render(m.t("aperta "))+"s"+m.st.muted.Render(m.t(", escolhe "))+m.t("entrar com código")+
			m.st.muted.Render(m.t(" e digita o código da sala"))) + "\n" +
		row(m.t("no terminal"), m.st.title.Render(m.invite+" "+s.Code))
}

// soloView: o painel do foco à esquerda e o modo jardim à direita (ADR 0006). O jardim anda só
// durante o foco; parado, a coluna fica quieta. Em terminal estreito ele não cabe e sobra o painel,
// com o pet ao lado das informações.
func (m model) soloView() string {
	// o pet com quem você é ao lado, na altura dele: embaixo o nome ficava solto, longe do desenho
	// (o sprite é centralizado numa caixa de 16, então ele nunca encosta na margem).
	who := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(20).Render(pet.Render(pet.MoodOf(m.me().Energy), m.tick, m.pixel, m.me().Color)),
		m.soloWho())
	panel := strings.Join([]string{m.focusPanel(), "", who, "", m.soloLog()}, "\n")
	if soloGardenWidth(max(m.width-4, 40)) == 0 {
		return panel // terminal estreito: não cabe coluna nenhuma
	}
	// a direita muda com o estado: focando é o jardim, parado são os seus números.
	right := m.statsBlock(true)
	if m.side != nil {
		right = m.side.frame
	}
	// a barra separa as colunas pela altura da mais alta: o jardim e os números têm alturas diferentes
	bar := m.st.muted.Render(strings.TrimRight(strings.Repeat("│\n", max(lipgloss.Height(panel), lipgloss.Height(right))), "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(panelCols).Render(panel), bar, " ", right)
}

// soloWho: as informações ao lado do pet, em colunas, como a saída de um comando: o rótulo
// apagado e o valor na cor do texto. O streak e os minutos vêm do banco, então só aparecem
// depois da primeira tarefa encerrada no dia.
func (m model) soloWho() string {
	rows := [][2]string{{m.t("humor"), m.t(pet.MoodName[pet.MoodOf(m.me().Energy)])}}
	if p := m.pet; p.Today == m.now.Format("2006-01-02") {
		if n := p.StreakAt(m.now); n > 0 {
			rows = append(rows, [2]string{m.t("streak"), fmt.Sprintf("%d %s", n, plural(n, m.t("dia seguido"), m.t("dias seguidos")))})
		}
		rows = append(rows, [2]string{m.t("hoje"), m.tf("%d min focados", int(p.TodayMin))})
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(m.st.muted.Render(fmt.Sprintf("%-8s", r[0])) + r[1] + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func plural(n int, um, varios string) string {
	if n == 1 {
		return um
	}
	return varios
}

// soloLog: as últimas coisas que você fez nesta sala, sem reações e sem cursor.
func (m model) soloLog() string {
	ev := m.w.Rooms[m.room].Events
	if len(ev) == 0 {
		return m.st.muted.Render(m.t("a primeira tarefa que você começar aparece aqui"))
	}
	var b strings.Builder
	for _, e := range ev[:min(len(ev), soloEvents)] {
		verb := m.t(e.Verb)
		if e.Verb == "cumpriu" {
			verb = m.st.ok.Render(verb)
		}
		b.WriteString(m.st.muted.Render(e.At.Format("15:04")) + "  " + verb + " \"" + e.Task + "\"\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// focusPanel: o seu foco, conforme o momento.
func (m model) focusPanel() string {
	me := m.me()
	f := me.Focus
	choice := func(label string) string {
		return label + "  " + m.st.muted.Render("‹ ") + m.st.title.Render(fmt.Sprintf("%d min", minuteChoices()[min(m.mins, len(minuteChoices())-1)])) + m.st.muted.Render(" ›")
	}
	switch {
	case m.entering:
		return bold.Render(m.t("qual a próxima coisa concreta que você vai fazer?")) + "\n" + m.in.View()
	case m.choosing && m.cont:
		return bold.Render(f.Task) + "\n" + choice(m.t("mais quanto tempo?"))
	case m.choosing:
		return bold.Render(strings.TrimSpace(m.in.Value())) + "\n" + choice(m.t("quanto tempo?"))
	case me.Status != rooms.Focusing:
		return bold.Render(m.t("nenhuma tarefa em andamento.")) + "\n" + m.st.muted.Render(m.t("enter começa uma tarefa"))
	}
	// A tarefa e os números ficam no topo do relógio, nunca ao lado dele por inteiro: o painel da
	// sala solo tem 54 colunas, e texto comprido ao lado do relógio quebrava a tela em pedaços.
	side := max(20, m.panelWidth()-clockCols-3)
	lines := []string{bold.Render(ellipsize(f.Task, side, 2))}
	c := m.st.title
	if f.Expired(m.now) {
		c = m.st.bad
		lines = append(lines, m.st.bad.Render(m.t("tempo esgotado")))
	}
	lines = append(lines, m.st.muted.Render(m.tf("%d de %d min focados",
		int(f.Focused(m.now).Minutes()), f.Declared)))
	block := lipgloss.JoinHorizontal(lipgloss.Top, c.Render(bigText(mmss(f.Ends.Sub(m.now)))), "   ",
		strings.Join(lines, "\n"))
	if f.Expired(m.now) { // as escolhas embaixo do relógio, onde cabem em qualquer largura
		block += "\n\n" + m.st.bad.Render(m.t("c continuar")) + m.st.muted.Render(" · ") +
			m.st.bad.Render(m.t("t trocar")) + m.st.muted.Render(" · ") + m.st.bad.Render(m.t("p parar")) +
			m.st.muted.Render(" · ") + m.st.ok.Render(m.t("f feito"))
	}
	return block
}

// panelWidth: a largura útil de quem mostra o foco — o painel da sala solo ou a tela inteira.
func (m model) panelWidth() int {
	if m.solo() {
		return panelCols
	}
	return max(m.width-4, 40)
}

// received: quantas reações cada membro recebeu nos eventos da sala (o contador do card, ADR 0004).
func received(s rooms.Snapshot) map[string]int {
	n := map[string]int{}
	for _, e := range s.Events {
		for _, who := range e.Reacts {
			n[e.Who] += len(who)
			if slices.Contains(who, e.Who) { // reagir ao próprio evento não enche o próprio contador
				n[e.Who]--
			}
		}
	}
	return n
}

// orderedMembers: você primeiro, depois quem está focando, quem está sem tarefa e quem está offline.
// Numa sala cheia, quem está focando agora fica sempre nas primeiras posições.
func orderedMembers(id string, s rooms.Snapshot) []rooms.Member {
	rank := func(x rooms.Member) int {
		switch {
		case x.ID == id:
			return 0
		case x.Status == rooms.Focusing:
			return 1
		case x.Status == rooms.Idle:
			return 2
		}
		return 3
	}
	ms := slices.Clone(s.Members)
	slices.SortStableFunc(ms, func(a, b rooms.Member) int { return rank(a) - rank(b) })
	return ms
}

// cardMin e cardMax: o miolo do card, em colunas. Ele cresce até caber "apelido humor" na mesma
// linha do maior apelido da sala (nickMax mais " radiante"); antes, o nome comprido quebrava em
// duas linhas e o card ficava mais alto que os vizinhos, deslocando os pets (#84).
const (
	cardMin = 16
	cardMax = nickMax + 9 // docs/regras.md
)

// nameLine: o apelido (ou "você") e o humor, como aparecem no card.
func (m model) nameLine(x rooms.Member, mood pet.Mood) (name, moodName string) {
	name = x.Nick
	if x.ID == m.id {
		name = m.t("você")
	}
	return name, m.t(pet.MoodName[mood])
}

// cardInner: o miolo dos cards da sala, o mesmo para todos, para a grade não desalinhar.
func (m model) cardInner(ms []rooms.Member) int {
	inner := cardMin
	for _, x := range ms {
		name, mood := m.nameLine(x, m.cardMood(x))
		inner = max(inner, ansi.StringWidth(name)+1+ansi.StringWidth(mood))
	}
	return min(inner, cardMax)
}

// cardMood: o humor que o card mostra (quem fechou o terminal dorme).
func (m model) cardMood(x rooms.Member) pet.Mood {
	if x.Status == rooms.Offline {
		return pet.Sleeping
	}
	return pet.MoodOf(x.Energy)
}

// cards: quantos couberem em w por fileira; compact troca o desenho do pet por um ponto da cor do
// corpo do pet.
func (m model) cards(s rooms.Snapshot, w int, compact bool) string {
	ms := orderedMembers(m.id, s)
	got := received(s)
	inner := m.cardInner(ms)
	per := max(1, w/(inner+2))
	var rows []string
	for i := 0; i < len(ms); i += per {
		var row []string
		for _, x := range ms[i:min(i+per, len(ms))] {
			row = append(row, m.card(x, got[x.ID], compact, inner))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	}
	return strings.Join(rows, "\n")
}

// card: um membro num card de miolo inner. O apelido que ainda passa (de antes do limite) termina
// em "…" em vez de quebrar a linha.
func (m model) card(x rooms.Member, got int, compact bool, inner int) string {
	mood := m.cardMood(x)
	nick, moodName := m.nameLine(x, mood)
	nick = ansi.Truncate(nick, inner-1-ansi.StringWidth(moodName), "…")
	st, name, task := m.st.card.Width(inner+2), bold.Render(nick), ""
	if x.ID == m.id {
		st = st.BorderForeground(m.st.accent)
	}
	if x.Status == rooms.Focusing {
		task = x.Focus.Task
	}
	reacts := " "
	switch {
	case got == 1:
		reacts = m.st.muted.Render(m.t("1 reação"))
	case got > 1:
		reacts = m.st.muted.Render(m.tf("%d reações", got))
	}
	if x.Status == rooms.Offline { // quem fechou o terminal dorme: o humor volta quando a pessoa volta
		st = st.BorderForeground(m.st.off)
	}
	art := pet.Render(mood, m.tick, m.pixel, x.Color)
	if compact {
		art = lipgloss.NewStyle().Foreground(pet.Colors[pet.ColorIndex(x.Color)].Body).Render("●")
	}
	return st.Render(lipgloss.JoinVertical(lipgloss.Center,
		art,
		name+" "+m.st.muted.Render(moodName),
		m.st.task.Width(inner).Render(ellipsize(task, inner, 2)),
		m.status(x),
		reacts))
}

// feed: os n eventos mais novos, com as reações numa coluna só, em ordem fixa.
func (m model) feed(n int) string {
	ev := m.events()
	ev = ev[:min(max(n, 0), len(ev))]
	if len(ev) == 0 {
		return m.st.muted.Render(m.t("silêncio por aqui. quando alguém começar ou cumprir uma tarefa, aparece aqui"))
	}
	text := func(e rooms.Event) string {
		nick := e.Nick
		if e.Who == m.id {
			nick = m.t("você")
		}
		line := m.st.muted.Render(e.At.Format("15:04")) + " " + bold.Render(nick) + " " + m.t(e.Verb)
		if e.Task != "" { // entrar e sair não têm tarefa
			line += " \"" + e.Task + "\""
		}
		return line
	}
	col := 0
	for _, e := range ev {
		col = max(col, lipgloss.Width(text(e)))
	}
	var b strings.Builder
	sel := m.selIndex()
	for i, e := range ev {
		mark := "  "
		if i == sel {
			mark = m.st.title.Render("> ")
		}
		line := lipgloss.NewStyle().Width(col).Render(text(e))
		for k, who := range e.Reacts { // ordem fixa 1-4: quem tem zero sai e os seguintes andam para a esquerda
			if len(who) == 0 {
				continue
			}
			r := fmt.Sprintf("%s %d", reactionGlyphs[k], len(who)) // o espaço protege o número de emoji mais largos que 2 colunas
			if slices.Contains(who, m.id) {
				r = m.st.title.Render(r)
			}
			line += "  " + r
		}
		b.WriteString(mark + line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m model) legend() string {
	var ks []string
	for k, g := range reactionGlyphs {
		ks = append(ks, fmt.Sprintf("%d %s", k+1, g))
	}
	return m.st.muted.Render(m.t("reagir ao evento selecionado: ")) + strings.Join(ks, "  ")
}
