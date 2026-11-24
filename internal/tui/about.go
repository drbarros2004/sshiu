package tui

// A aba sobre (#10): o FAQ, no formato da saída de um --help. Cada pergunta é uma flag, e a
// escolhida abre a resposta embaixo. Formato escolhido no protótipo prototype/sobre.

import (
	"fmt"
	"sshiu/internal/rooms"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// faqs: metade sérias, metade piadas. O inglês de cada campo fica no dicionário en; a resposta
// sobre a conexão leva o número da tolerância de reconexão.
var faqs = []struct{ flag, q, a string }{
	{"--o-que-e", "o que é isso?", "um lugar quieto para focar junto, pelo terminal. você diz a próxima coisa concreta que vai fazer, foca por um tempo e a sala vê. sem chat, sem notificação: só um pet que reflete o seu foco."},
	{"--nome", "por que sshiu?", "é o ssh de quem pede silêncio. shhh."},
	{"--por-que-ssh", "por que SSH?", "porque o terminal já está aberto. e porque nenhum app de foco no navegador sobrevive à aba do lado."},
	{"--dev-melhor", "o sshiu vai me tornar um dev melhor?", "não podemos prometer isso por escrito. mas o seu pet acredita em você."},
	{"--tarefa", "o que é uma tarefa?", "a próxima coisa concreta, com um tempo: \"fazer o exercício 1 da lista 3\", não \"estudar Cálculo\". dá para marcar feito antes da hora. ninguém fiscaliza; o pet sabe."},
	{"--vibe-coding", "funciona para vibe-coders?", "funciona. a tarefa pode ser \"esperar o Claude terminar\". o pet não julga."},
	{"--pet", "o que move o pet?", "minutos focados e tarefas cumpridas enchem a energia, e cada dia parado tira um pouco. o humor sai da energia."},
	{"--o-pet-morre", "o pet morre?", "não. ele só fica murcho e te olha. o que é pior."},
	{"--chat", "por que não tem chat?", "aqui é tipo uma biblioteca. dá para reagir quando alguém começa ou termina uma tarefa. mas conversar... sshhh."},
	{"--conexao", "e se a conexão cair?", "a tarefa espera %d minutos você voltar. se não voltar, ela fica interrompida, o que não é castigo: é o mesmo que parar."},
	{"--loja", "pra que servem as moedas?", "tarefas cumpridas rendem moedas pelo tempo focado. na loja, você troca moedas por personagens, cores, chapéus e seus modelos. seu foco agora rende."},
	{"--dados", "que dados vocês guardam?", "o seu apelido, a impressão digital da sua chave SSH, as suas tarefas e os minutos. nada de senha, e-mail ou nome. a chave privada nunca sai do seu computador."},
	{"--rende-mais", "focar pelo terminal rende mais?", "fonte monoespaçada, fundo escuro, zero notificação. não temos dados, mas temos convicção."},
	{"--modo-escuro", "tem modo escuro?", "é um terminal."},
}

func (m model) aboutKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.aboutSel = max(0, m.aboutSel-1)
	case "down":
		m.aboutSel = min(len(faqs)-1, m.aboutSel+1)
	}
	return m, nil
}

// aboutView: o --help em w colunas e h linhas. A resposta fica na coluna depois das flags; numa
// tela estreita ela ocupa a largura toda, embaixo da flag. Se não cabe tudo, a lista rola até a
// resposta aberta.
func (m model) aboutView(w, h int) string {
	answerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248")) // cinza xterm #a8a8a8
	invite := m.invite
	if invite == "" {
		invite = "ssh sshiu"
	}
	lines := []string{m.tf("uso: %s", invite), "", m.t("um lugar quieto para focar junto"), "", m.t("perguntas:")}
	fw := 0
	for _, f := range faqs {
		fw = max(fw, ansi.StringWidth(m.t(f.flag)))
	}
	pad := fw + 4 // a coluna da resposta
	if w-pad < 30 {
		pad = 4
	}
	end := 0 // a última linha da resposta aberta
	for i, f := range faqs {
		flag := "  " + m.t(f.flag) + strings.Repeat(" ", fw-ansi.StringWidth(m.t(f.flag))+2)
		if i != m.aboutSel {
			lines = append(lines, ansi.Truncate(m.st.muted.Render(flag)+m.t(f.q), w, "…"))
			continue
		}
		lines = append(lines, "")
		a := m.t(f.a)
		if strings.Contains(a, "%d") {
			a = fmt.Sprintf(a, int(rooms.ReconnectGrace.Minutes()))
		}
		if pad == 4 { // estreita: a pergunta inteira desce para junto da resposta
			lines = append(lines, m.st.title.Render(strings.TrimRight(flag, " ")))
			a = m.t(f.q) + " " + a
		} else {
			lines = append(lines, m.st.title.Render(flag)+bold.Render(m.t(f.q)))
		}
		for _, l := range strings.Split(ansi.Wrap(a, w-pad, ""), "\n") {
			lines = append(lines, strings.Repeat(" ", pad)+answerStyle.Render(l))
		}
		end = len(lines) - 1
		if i < len(faqs)-1 {
			lines = append(lines, "")
		}
	}
	lines = append(lines, "", m.st.muted.Render(m.t("feito com café e pouco silêncio")))
	if len(lines) > h {
		start := min(max(end-h+1, 0), len(lines)-h)
		lines = lines[start : start+h]
	}
	return strings.Join(lines, "\n")
}
