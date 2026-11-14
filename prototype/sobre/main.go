// PROTÓTIPO DESCARTÁVEL (#10): três formatos para a aba "sobre", com o rascunho do FAQ.
// Roda no terminal local, sem SSH: go run ./prototype/sobre
// 1/2/3 troca o formato, ↑/↓ escolhe ou rola, q sai. Não vai para a master.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type faq struct{ q, a, flag string }

// rascunho: metade sérias, metade piadas. É para cortar.
var faqs = []faq{
	{"o que é isso?", "um lugar quieto para focar junto, pelo terminal. você diz a próxima coisa concreta que vai fazer, foca por um tempo e a sala vê. sem chat, sem notificação: só um pet que reflete o seu foco.", "--o-que-e"},
	{"por que sshiu?", "é o ssh de quem pede silêncio. shhh.", "--nome"},
	{"por que SSH?", "porque o terminal já está aberto. e porque nenhum app de foco no navegador sobrevive à aba do lado.", "--por-que-ssh"},
	{"o sshiu vai me tornar um dev melhor?", "não podemos prometer isso por escrito. mas o seu pet acredita em você.", "--dev-melhor"},
	{"o que é uma tarefa?", "a próxima coisa concreta, com um tempo: \"fazer o exercício 1 da lista 3\", não \"estudar Cálculo\". dá para marcar feito antes da hora. ninguém fiscaliza; o pet sabe.", "--tarefa"},
	{"funciona para vibe-coders?", "funciona. a tarefa pode ser \"esperar o Claude terminar\". o pet não julga.", "--vibe-coding"},
	{"o que move o pet?", "minutos focados e tarefas cumpridas enchem a energia, e cada dia parado tira um pouco. o humor sai da energia.", "--pet"},
	{"o pet morre?", "não. ele só fica murcho e te olha. o que é pior.", "--o-pet-morre"},
	{"por que não tem chat?", "aqui é uma biblioteca. dá para reagir a quem começou ou cumpriu uma tarefa, e só.", "--chat"},
	{"e se a conexão cair?", "a tarefa espera 2 minutos você voltar. se não voltar, ela fica interrompida, o que não é castigo: é o mesmo que parar.", "--conexao"},
	{"que dados vocês guardam?", "o seu apelido, a impressão digital da sua chave SSH, as suas tarefas e os minutos. nada de senha, e-mail ou nome. a chave privada nunca sai do seu computador.", "--dados"},
	{"focar pelo terminal rende mais?", "fonte monoespaçada, fundo escuro, zero notificação. não temos dados, mas temos convicção.", "--rende-mais"},
	{"tem modo escuro?", "é um terminal.", "--modo-escuro"},
}

var (
	accent = lipgloss.Color("#ff8700")
	muted  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7a7a7a"))
	acc    = lipgloss.NewStyle().Foreground(accent)
	bold   = lipgloss.NewStyle().Bold(true)
	tabOn  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#11111b")).Background(accent).Padding(0, 1)
	tabOff = muted.Padding(0, 1)
)

var formats = []string{"acordeão #", "man page", "--help"}

type model struct{ w, h, format, sel, scroll int }

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyPressMsg:
		switch k := msg.String(); k {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1", "2", "3":
			m.format, m.scroll = int(k[0]-'1'), 0
		case "up", "w":
			m.sel, m.scroll = max(m.sel-1, 0), max(m.scroll-1, 0)
		case "down", "s":
			m.sel, m.scroll = min(m.sel+1, len(faqs)-1), m.scroll+1
		}
	}
	return m, nil
}

func wrap(s string, w int) string { return lipgloss.NewStyle().Width(w).Render(s) }

func indent(s, pre string) string {
	return pre + strings.ReplaceAll(s, "\n", "\n"+pre)
}

// acordeão: as perguntas como comentários; a escolhida abre a resposta.
func (m model) accordion(w int) []string {
	var out []string
	for i, f := range faqs {
		if i != m.sel {
			out = append(out, muted.Render("  # "+f.q))
			continue
		}
		out = append(out, acc.Render("› # ")+bold.Render(f.q))
		out = append(out, strings.Split(indent(wrap(f.a, w-6), "    "), "\n")...)
		out = append(out, "")
	}
	return out
}

// man page: SSHIU(1), com as perguntas numa seção, rolando.
func (m model) man(w int) []string {
	head := "SSHIU(1)"
	mid := "manual do sshiu"
	gap := max(w-2*len(head)-len(mid), 2)
	out := []string{head + strings.Repeat(" ", gap/2) + mid + strings.Repeat(" ", gap-gap/2) + head, ""}
	sec := func(t string, body ...string) {
		out = append(out, bold.Render(strings.ToUpper(t)))
		for _, b := range body {
			out = append(out, strings.Split(indent(wrap(b, w-8), "       "), "\n")...)
		}
		out = append(out, "")
	}
	sec("nome", "sshiu — um lugar quieto para focar junto")
	sec("sinopse", acc.Render("ssh sshiu"))
	out = append(out, bold.Render("PERGUNTAS"))
	for _, f := range faqs {
		out = append(out, "       "+acc.Render(f.q))
		out = append(out, strings.Split(indent(wrap(f.a, w-12), "           "), "\n")...)
		out = append(out, "")
	}
	sec("bugs", "o pet ainda não sabe ler. não conte para ele.")
	sec("veja também", "café(1), silêncio(7)")
	return out
}

// --help: as perguntas viram flags; a escolhida mostra a resposta inteira.
func (m model) help(w int) []string {
	out := []string{"uso: sshiu [--pergunta]", "", "um lugar quieto para focar junto, pelo terminal.", "", "perguntas:"}
	fw := 0
	for _, f := range faqs {
		fw = max(fw, len(f.flag))
	}
	for i, f := range faqs {
		flag := fmt.Sprintf("  %-*s  ", fw, f.flag)
		if i != m.sel {
			out = append(out, muted.Render(flag)+f.q)
			continue
		}
		out = append(out, acc.Render(flag)+bold.Render(f.q))
		out = append(out, strings.Split(indent(wrap(f.a, w-fw-6), strings.Repeat(" ", fw+4)), "\n")...)
	}
	return append(out, "", muted.Render("sshiu 0.1 · feito com café e pouco silêncio"))
}

func (m model) View() tea.View {
	if m.w == 0 {
		return tea.NewView("")
	}
	w := min(m.w-4, 90)
	var tabs []string
	for _, t := range []string{"sala", "perfil", "visual", "sobre"} {
		if t == "sobre" {
			tabs = append(tabs, tabOn.Render(t))
		} else {
			tabs = append(tabs, tabOff.Render(t))
		}
	}
	top := acc.Bold(true).Render("SSHIU") + "  " + strings.Join(tabs, " ")
	var body []string
	switch m.format {
	case 0:
		body = m.accordion(w)
	case 1:
		body = m.man(w)
	default:
		body = m.help(w)
	}
	room := max(m.h-6, 1)
	start := 0
	if m.format == 1 { // a man page rola; as outras seguem a seleção
		start = min(m.scroll, max(len(body)-room, 0))
	} else if len(body) > room {
		start = min(max(m.sel*1-room/2, 0), len(body)-room)
	}
	body = body[start:min(start+room, len(body))]
	nav := "↑/↓ escolhe"
	if m.format == 1 {
		nav = "↑/↓ rola"
	}
	foot := muted.Render(nav + " · tab troca de aba · ctrl+c sai")
	bar := acc.Render(fmt.Sprintf("protótipo · formato %d/3: %s · 1/2/3 troca · q sai", m.format+1, formats[m.format]))
	v := tea.NewView(lipgloss.NewStyle().Padding(0, 2).Render(
		top + "\n\n" + strings.Join(body, "\n") + strings.Repeat("\n", max(room-len(body), 0)) + "\n\n" + foot + "\n" + bar))
	v.AltScreen = true
	return v
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-dump" { // para conferir a tela sem terminal
		for f := range formats {
			v := model{w: 80, h: 32, format: f, sel: 3}.View()
			fmt.Println(v.Content + "\n" + strings.Repeat("=", 80))
		}
		return
	}
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
