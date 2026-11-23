package tui

// Idioma da interface (#86): o texto em português fica no código, como chave, e o inglês sai do
// dicionário en. O padrão é inglês; a aba visual troca. O teste TestEveryTextTranslated lê o
// código e falha se algum m.t("...") não tiver tradução.
//
// ponytail: um dicionário com o português como chave, sem pacote de i18n. Mudar uma frase em
// português pede mudar a chave aqui também (o teste avisa). Um terceiro idioma pede trocar a
// chave por um id.

import (
	"fmt"
	"regexp"
	"sshiu/internal/rooms"
	"strings"
)

// lang: "en" (o padrão de quem entra) ou "pt". O vazio também é português: é o idioma do código
// e o dos modelos montados nos testes.
type lang string

var langs = []struct {
	id   lang
	name string // o nome no próprio idioma, para quem não entende o outro achar o seu
}{{"en", "english"}, {"pt", "português"}}

// langIndex: a posição de l em langs. O vazio é o português, como em t.
func langIndex(l lang) int {
	if l != "en" {
		l = "pt"
	}
	for i, x := range langs {
		if x.id == l {
			return i
		}
	}
	return 0
}

// taskText: a tarefa de quem na tela. A de um bot da sala demo é texto do servidor, em
// português, e cada um a vê no próprio idioma; a de gente fica como foi escrita.
func (m model) taskText(who, task string) string {
	if strings.HasPrefix(who, "bot:") {
		return m.t(task)
	}
	return task
}

// t: o texto no idioma l.
func (l lang) t(pt string) string {
	if l != "en" {
		return pt
	}
	if s, ok := en[pt]; ok {
		return s
	}
	return pt
}

// tf: o texto com números no idioma l. Os dois idiomas recebem os valores na mesma ordem.
func (l lang) tf(pt string, a ...any) string { return fmt.Sprintf(l.t(pt), a...) }

// roomName: o nome da sala na tela. A sala demo é criada pelo servidor com nome em português, e
// cada um a vê no próprio idioma.
func (m model) roomName(s rooms.Snapshot) string {
	if s.Code == rooms.DemoCode {
		return m.t("sala demo")
	}
	return s.Name
}

func (m model) t(pt string) string            { return m.lang.t(pt) }
func (m model) tf(pt string, a ...any) string { return m.lang.tf(pt, a...) }

// errText: o erro vindo das salas, que fala português (o pacote rooms não sabe do idioma de quem
// lê). Os erros com nome ou número dentro casam com um padrão.
func (l lang) errText(err error) string {
	s := err.Error()
	if l != "en" {
		return s
	}
	if e, ok := en[s]; ok {
		return e
	}
	for _, p := range enErrors {
		if m := p.re.FindStringSubmatch(s); m != nil {
			a := make([]any, len(m)-1)
			for i, x := range m[1:] {
				a[i] = x
			}
			return fmt.Sprintf(p.en, a...)
		}
	}
	return s
}

var enErrors = []struct {
	re *regexp.Regexp
	en string
}{
	{regexp.MustCompile(`^apelido de até (\d+) letras, por favor \(日 e 🐱 contam como 2\)$`), "nickname up to %s letters, please (日 and 🐱 count as 2)"},
	{regexp.MustCompile(`^nenhuma sala com o código "?([^"]*)"?$`), "no room with the code %s"},
	{regexp.MustCompile(`^(.*) já está cheia \((\d+) membros\)$`), "%s is already full (%s members)"},
	{regexp.MustCompile(`^(.*) não está aceitando mais ninguém$`), "%s isn't taking anyone else"},
	{regexp.MustCompile(`^não existe obra "?([^"]*)"?$`), "there's no build called %s"},
	{regexp.MustCompile(`^meta fora da lista: (\d+) min$`), "goal not in the list: %s min"},
}

// en: português → inglês. Em minúsculas, como todo texto da interface.
var en = map[string]string{
	"personagem":        "character",
	"cor do personagem": "character color",
	"chapéu":            "hat",
	"sem chapéu":        "no hat",
	"modelo do chapéu":  "hat model",
	"prévia do humor":   "mood preview",
	"↑/↓ opção · ←/→ altera · enter salva · esc cancela": "↑/↓ option · ←/→ change · enter saves · esc cancels",
	"1 reação":          "1 reaction",
	"6 letras":          "6 letters",
	" · a obra de hoje": " · today's mural",
	"a primeira tarefa que você começar aparece aqui":               "the first task you start shows up here",
	"a sala agora é de %s":                                          "the room now belongs to %s",
	"a sala demo não tem saída: ela some quando o servidor desliga": "the demo room has no exit: it disappears when the server shuts down",
	"a sala precisa de um nome":                                     "the room needs a name",
	"a sala solo não tem obra":                                      "the solo room has no mural",
	"a sala some para os %d membros. o histórico de cada um fica.":  "the room disappears for its %d members. everyone keeps their history.",
	"a sala some. o que você focou nela fica no seu histórico.":     "the room disappears. what you focused on in it stays in your history.",
	" · a tela de descanso durante o foco":                          " · the resting screen while you focus",
	"a/d troca de aba · q sai":                                      "a/d switch tab · q quit",
	"abertas":                                                       "open",
	" abra as novas entradas em ":                                   " open new entries in ",
	"abril":                                                         "april",
	"acampamento":                                                   "campsite",
	"agosto":                                                        "august",
	"aparece":                                                       "shown",
	"apelido de até %d letras, por favor (日 e 🐱 contam como 2)":    "nickname up to %d letters, please (日 and 🐱 count as 2)",
	"apelido sem espaços, por favor":                               "nickname without spaces, please",
	"apelido só com letras, números e símbolos simples, por favor": "nickname with only letters, numbers and simple symbols, please",
	"apenas para visualização":                                     "just a preview",
	"aperta ":                                                      "press ",
	"aumente o terminal para editar no jardim":                     "make the terminal bigger to edit in the garden",
	"brainrot (para devs de verdade)":                              "brainrot (for real devs)",
	" · c configura":                                               " · c settings",
	"c continuar":                                                  "c continue",
	"café":                                                         "coffee",
	"carregando…":                                                  "loading…",
	"catppuccin":                                                   "catppuccin",
	"chuva lenta":                                                  "slow rain",
	"cidade":                                                       "city",
	"cole com ctrl+v (cmd+v no mac): o código ou o comando ssh inteiro": "paste with ctrl+v (cmd+v on mac): the code or the whole ssh command",
	"colegas focando": "others focusing",
	"começou":         "started",
	"configurar":      "settings",
	" · configurar.":  " · settings.",
	"cor":             "color",
	"criar sala":      "create room",
	"✓ cumprida":      "✓ done",
	"cumpriu":         "finished",
	"céu noturno":     "night sky",
	" · código ":      " · code ",
	"código %s":       "code %s",
	"código enviado para a área de transferência":        "code sent to the clipboard",
	"%d de %d min focados":                               "%d of %d min focused",
	"%d min focados":                                     "%d min focused",
	"%d pessoas também estão focando":                    "%d people are also focusing",
	"%d reações":                                         "%d reactions",
	"%d/%d membros · %d focando":                         "%d/%d members · %d focusing",
	"a sala está cheia.":                                 "the room is full.",
	"você já está nesta sala.":                           "you're already in this room.",
	"último passo: a cor do pet":                         "last step: your pet's color",
	"a sala não está aceitando ninguém agora.":           "the room isn't letting anyone in right now.",
	"a sala passa para %s, que está nela há mais tempo.": "the room goes to %s, who has been in it the longest.",
	"enter salva e abre a aba sala":                      "enter saves and opens the room tab",
	"escreva um apelido":                                 "type a nickname",
	"%s saiu da sala e a vaga abriu":                     "%s left the room and the spot is open",
	"a vaga abre. não é bloqueio: com o código, dá para entrar de novo.": "the spot opens up. it's not a ban: with the code, they can come back.",
	"remover %s de %s?":          "remove %s from %s?",
	"sala solo":                  "solo room",
	"sala solo · focando":        "solo room · focusing",
	"%d/%d membros":              "%d/%d members",
	" · dados de exemplo":        " · sample data",
	"depois de %.0f s":           "after %.0f s",
	"dezembro":                   "december",
	"dia seguido":                "day in a row",
	"dias seguidos":              "days in a row",
	"digite o código da sala":    "type the room code",
	"dom":                        "sun",
	"dracula":                    "dracula",
	"dê um nome para a sala":     "give the room a name",
	" e ":                        " and ",
	" e digita o código da sala": " and type the room's code",
	"encerrar %s?":               "close %s for good?",
	"encerrar a sala":            "close the room",
	"energia":                    "energy",
	"enter começa uma tarefa":    "enter starts a task",
	"enter começa uma tarefa · j modo jardim":      "enter starts a task · j garden mode",
	"enter confirma · esc cancela":                 "enter confirms · esc cancels",
	"enter confirma · esc volta":                   "enter confirms · esc back",
	"enter confirma · sua chave SSH é o seu login": "enter confirms · your SSH key is your login",
	"enter entra · esc volta":                      "enter joins · esc back",
	"entrar com código":                            "join with code",
	"entrar em %s?":                                "join %s?",
	"entrou":                                       "joined",
	" · esc volta":                                 " · esc back",
	"esc volta":                                    "esc back",
	", escolhe ":                                   ", pick ",
	"↑/↓ escolhe a linha · ←/→ muda · enter cria · esc volta":              "↑/↓ choose the line · ←/→ change · enter creates · esc back",
	"↑/↓ escolhe · enter abre":                                             "↑/↓ choose · enter opens",
	"←/→ escolhe · enter começa · esc volta":                               "←/→ choose · enter starts · esc back",
	"←/→ escolhe · enter confirma · esc volta":                             "←/→ choose · enter confirms · esc back",
	"↑/↓ escolhe · enter passa a sala · esc volta":                         "↑/↓ choose · enter hands over the room · esc back",
	"↑/↓ escolhe · enter remove · esc volta":                               "↑/↓ choose · enter removes · esc back",
	"↑/↓ escolhe · ←/→ muda · enter abre · esc volta":                      "↑/↓ choose · ←/→ change · enter opens · esc back",
	"↑/↓ escolhe · ←/→ muda · j prévia · v editar no jardim · enter salva": "↑/↓ choose · ←/→ change · j preview · v edit in the garden · enter saves",
	"escreva a próxima coisa concreta que você vai fazer":                  "write the next concrete thing you'll do",
	"espaço oculta · esc volta":                                            "space hides · esc back",
	"espaço · esc volta":                                                   "space · esc back",
	"essa pessoa não está na sala":                                         "that person isn't in the room",
	"esse apelido já está em uso":                                          "that nickname is taken",
	"esta sala não está aceitando ninguém.":                                "this room isn't taking anyone.",
	"esta semana":                         "this week",
	"estante":                             "bookshelf",
	"estrada à noite":                     "night road",
	" · ↑/↓ evento · 1-4 reage":           " · ↑/↓ event · 1-4 react",
	"ex.: fazer o exercício 1 da lista 3": "e.g. do exercise 1 of problem set 3",
	"ex.: turma de cálculo":               "e.g. calculus class",
	"f feito":                             "f done",
	"f feito · t trocar · p parar · j modo jardim": "f done · t switch · p stop · j garden mode",
	"falta":                           "left",
	"farol":                           "lighthouse",
	"gato dormindo":                   "sleeping cat",
	"cachorro dormindo":               "sleeping dog",
	"faltam":                          "left",
	" · fechada":                      " · closed",
	"fechadas":                        "closed",
	"feira":                           "market",
	"fevereiro":                       "february",
	"focado":                          "focused",
	"focando %s":                      "focusing %s",
	"foguetes":                        "rockets",
	"gruvbox":                         "gruvbox",
	"hoje":                            "today",
	"horta":                           "vegetable garden",
	"humor":                           "mood",
	"humor máximo":                    "highest mood",
	"idioma":                          "language",
	"interrompida":                    "interrupted",
	"janeiro":                         "january",
	"julho":                           "july",
	"junho":                           "june",
	"latte (claro)":                   "latte (light)",
	"liga sozinho":                    "turns on by itself",
	"lâmpada de lava":                 "lava lamp",
	"maio":                            "may",
	"mais %d":                         "%d more",
	"mais alguém está focando":        "someone else is focusing",
	"mais quanto tempo?":              "how much more time?",
	"março":                           "march",
	"meta da obra":                    "mural goal",
	"mm:ss":                           "mm:ss",
	"modo jardim":                     "garden mode",
	"nenhum":                          "none",
	"nenhuma":                         "none",
	"nenhuma ainda":                   "none yet",
	"nenhuma sala com o código %s":    "no room with the code %s",
	"nenhuma tarefa em andamento.":    "no task in progress.",
	"ninguém mais está focando agora": "no one else is focusing right now",
	"no sshiu":                        "in sshiu",
	"no terminal":                     "in a terminal",
	"nome":                            "name",
	"nome da sala":                    "room name",
	"nord":                            "nord",
	"novas entradas":                  "new entries",
	"novembro":                        "november",
	"nunca":                           "never",
	"não":                             "no",
	"(não altera humor ou energia)":   "(doesn't change mood or energy)",
	"não aparece":                     "hidden",
	"não cumprida":                    "not done",
	"não há mais ninguém na sala.":    "there's no one else in the room.",
	"o menu de configurar passa a ser de %s. você continua na sala, como os outros.": "the settings menu becomes %s's. you stay in the room, like everyone else.",
	" · o obra":                          " · o mural",
	"depois, o dono muda tudo isso em c": "later, the owner changes all of this with c",
	"o volta para a sala":                "o back to the room",
	"obra":                               "mural",
	"offline":                            "offline",
	"↑/↓ opção · ←/→ muda":               "↑/↓ option · ←/→ change",
	"↑/↓ opção · ←/→ muda · espaço oculta · esc volta": "↑/↓ option · ←/→ change · space hides · esc back",
	"↑/↓ opção ←/→ muda":                               "↑/↓ option ←/→ change",
	"orquestra":                                        "orchestra",
	"outubro":                                          "october",
	"p parar":                                          "p stop",
	" para a turma entrar na sala:":                    " to bring your group into the room:",
	"para sair da sala, use sair":                      "to leave the room, use leave",
	"para voltar, você vai precisar do código de novo. se não sobrar ninguém, a sala some.": "to come back, you'll need the code again. if no one is left, the room disappears.",
	"passar %s para %s?":       "hand %s over to %s?",
	"passar a sala":            "hand over the room",
	"passar a sala para quem?": "hand the room over to whom?",
	"passarinhos":              "birds",
	"perfil":                   "profile",
	"permitir wallpapers para devs de verdade?": "allow wallpapers for real devs?",
	"porto": "harbor",
	"primeira vez por aqui. qual vai ser o seu apelido?": "first time here. what will your nickname be?",
	"prévia: ":                    "preview: ",
	"próximo humor em %d · %s %d": "next mood at %d · %[3]d %[2]s",
	"qua":                         "wed",
	"qual a próxima coisa concreta que você vai fazer?": "what's the next concrete thing you'll do?",
	"quantos membros":                "how many members",
	"quanto tempo?":                  "how long?",
	"quem mais está focando":         "who else is focusing",
	"qui":                            "thu",
	"reagir ao evento selecionado: ": "react to the selected event: ",
	"relógio":                        "clock",
	"remover alguém":                 "remove someone",
	"remover quem da sala?":          "remove whom from the room?",
	"revisar a lista 3":              "review problem set 3",
	" · s salas":                     " · s rooms",
	"s salas":                        "s rooms",
	"%s · meta %s · a peça muda a cada %s · reinicia em %s": "%s · goal %s · the piece changes every %s · resets in %s",
	"sair de %s":  "leave %s",
	"sair de %s?": "leave %s?",
	"saiu":        "left",
	"sala":        "room",
	"sala demo":   "demo room",
	// as tarefas dos bots (rooms.BotTasks)
	"reescrever a lista 3 de Cálculo em Rust":        "rewrite calculus problem set 3 in Rust",
	"resolver a lista antes do professor publicar":   "solve the problem set before the professor posts it",
	"revisar o PR que eu mesmo abri":                 "review the PR I opened myself",
	"resolver o conflito de merge do grupo":          "fix the group project's merge conflict",
	"trocar x, y e aux2 por nomes de verdade":        "rename x, y and aux2 to real names",
	"ler a documentação até o fim (de verdade)":      "read the docs to the end (for real)",
	"descobrir por que só funciona na minha máquina": "find out why it only works on my machine",
	"centralizar uma div":                            "center a div",
	"descobrir quem deu push --force na main":        "find out who force-pushed to main",
	"sair do vim":                          "exit vim",
	"instalar o ambiente do TCC (dia 3)":   "set up the thesis environment (day 3)",
	"sala solo criada: só você entra aqui": "solo room created: only you come in here",
	"sala solo: para focar sem companhia, com o seu pet e o modo jardim": "solo room: to focus on your own, with your pet and garden mode",
	"salas":       "rooms",
	"salvo":       "saved",
	"seg":         "mon",
	"sem tarefa":  "no task",
	"setembro":    "september",
	"seu apelido": "your nickname",
	"sex":         "fri",
	"silêncio por aqui. quando alguém começar ou cumprir uma tarefa, aparece aqui": "quiet in here. when someone starts or finishes a task, it shows up here",
	"sim":                              "yes",
	"sorteado":                         "random",
	"sorteado a cada tarefa":           "random for each task",
	"sorteado a cada tarefa · agora, ": "random for each task · now, ",
	"sshiu":                            "sshiu",
	"streak":                           "streak",
	"sáb":                              "sat",
	"só o dono muda a sala":            "only the owner changes the room",
	"só os minutos":                    "minutes only",
	"t trocar":                         "t switch",
	"tab troca de aba · ctrl+c sai":    "tab switch tab · ctrl+c quit",
	"tamanho":                          "size",
	" também está focando":             " is also focusing",
	" também estão focando":            " are also focusing",
	"tarefa":                           "task",
	"tema":                             "theme",
	"tempo esgotado":                   "time's up",
	"ter":                              "tue",
	"tokyo night":                      "tokyo night",
	"trem":                             "train",
	" · ← → troca":                     " · ← → switch",
	"← → troca · qualquer tecla volta": "← → switch · any key goes back",
	"um lugar quieto para focar junto": "a quiet place to focus together",
	"viagem":                           "trip",
	"+1000 de aura":                    "+1000 aura",
	"-500 de aura":                     "-500 aura",
	"vídeo de orbital (CC BY)":         "video by orbital (CC BY)",
	"visual":                           "look",
	"você":                             "you",
	"você ainda não está em nenhuma sala. crie uma e passe o código para quem vai focar com você, ou entre com o código de alguém.": "you're not in any room yet. create one and give the code to whoever will focus with you, or join with someone's code.",
	"você entrou em %s":               "you joined %s",
	"  ← você está aqui":              "  ← you are here",
	"você não está mais em %s":        "you're no longer in %s",
	"você é a única pessoa na sala.":  "you're the only person in the room.",
	"por enquanto, só tem você aqui.": "so far it's just you here.",
	" · ←/→ wallpaper":                " · ←/→ wallpaper",
	"wallpaper":                       "wallpaper",
	"wallpapers de piada":             "joke wallpapers",
	" · y copia":                      " · y copies",
	"árvore":                          "tree",
	"últimas tarefas":                 "latest tasks",
	// a tabela de foco (ADR 0010)
	"tabela de foco":                   "focus table",
	" · só você vê":                    " · only you see it",
	" · a sala toda vê":                " · the whole room sees it",
	"só você":                          "only you",
	"a sala toda":                      "the whole room",
	"minutos focados em qualquer sala": "minutes focused in any room",
	" · reinicia em %s":                " · resets in %s",
	"este mês":                         "this month",
	"total":                            "total",
	"apelido":                          "nickname",
	"tempo":                            "time",
	"agora":                            "now",
	"%dº":                              "#%d",
	" · +%d no o":                      " · +%d on o",
	"+%d: aumente a janela para ver todo mundo": "+%d: make the window bigger to see everyone",
	"período": "period",
	"quem vê": "who sees",
	"formato": "format",
	"tabela":  "table",
	"ranking": "ranking",
	"a obra não coube junto da tabela: aumente a janela para ver as duas": "the mural didn't fit next to the table: enlarge the window to see both",
	" · o tabela": " · o table",
	"←/→ período · o volta para a sala":  "←/→ period · o back to the room",
	"↑/↓ escolhe · ←/→ muda · esc volta": "↑/↓ choose · ←/→ change · esc back",
	// a aba sobre (about.go)
	"sobre":                                "about",
	"↑/↓ escolhe a pergunta":               "↑/↓ choose a question",
	"uso: %s":                              "usage: %s",
	"perguntas:":                           "questions:",
	"feito com café e pouco silêncio":      "made with coffee and not much silence",
	"--o-que-e":                            "--what-is-this",
	"--nome":                               "--name",
	"--por-que-ssh":                        "--why-ssh",
	"--dev-melhor":                         "--better-dev",
	"--tarefa":                             "--task",
	"--vibe-coding":                        "--vibe-coding",
	"--pet":                                "--pet",
	"--o-pet-morre":                        "--does-the-pet-die",
	"--chat":                               "--chat",
	"--conexao":                            "--connection",
	"--dados":                              "--data",
	"--rende-mais":                         "--more-productive",
	"--modo-escuro":                        "--dark-mode",
	"o que é isso?":                        "what is this?",
	"por que sshiu?":                       "why sshiu?",
	"por que SSH?":                         "why SSH?",
	"o sshiu vai me tornar um dev melhor?": "will sshiu make me a better dev?",
	"o que é uma tarefa?":                  "what is a task?",
	"funciona para vibe-coders?":           "does it work for vibe-coders?",
	"o que move o pet?":                    "what drives the pet?",
	"o pet morre?":                         "does the pet die?",
	"por que não tem chat?":                "why is there no chat?",
	"e se a conexão cair?":                 "what if my connection drops?",
	"que dados vocês guardam?":             "what data do you keep?",
	"focar pelo terminal rende mais?":      "is focusing in the terminal more productive?",
	"tem modo escuro?":                     "is there a dark mode?",
	"um lugar quieto para focar junto, pelo terminal. você diz a próxima coisa concreta que vai fazer, foca por um tempo e a sala vê. sem chat, sem notificação: só um pet que reflete o seu foco.": "a quiet place to focus together, from the terminal. you say the next concrete thing you'll do, focus for a while and the room sees it. no chat, no notifications: just a pet that reflects your focus.",
	"é o ssh de quem pede silêncio. shhh.": "it's ssh with the \"shhh\" of a library inside. shhh.",
	"porque o terminal já está aberto. e porque nenhum app de foco no navegador sobrevive à aba do lado.":                                                                     "because the terminal is already open. and because no focus app in the browser survives the tab next to it.",
	"não podemos prometer isso por escrito. mas o seu pet acredita em você.":                                                                                                  "we can't promise that in writing. but your pet believes in you.",
	"a próxima coisa concreta, com um tempo: \"fazer o exercício 1 da lista 3\", não \"estudar Cálculo\". dá para marcar feito antes da hora. ninguém fiscaliza; o pet sabe.": "the next concrete thing, with a time: \"do exercise 1 of problem set 3\", not \"study calculus\". you can mark it done early. nobody checks; the pet knows.",
	"funciona. a tarefa pode ser \"esperar o Claude terminar\". o pet não julga.":                                                                                             "it does. the task can be \"wait for Claude to finish\". the pet doesn't judge.",
	"minutos focados e tarefas cumpridas enchem a energia, e cada dia parado tira um pouco. o humor sai da energia.":                                                          "focused minutes and done tasks fill its energy, and each idle day takes a little away. its mood comes from the energy.",
	"não. ele só fica murcho e te olha. o que é pior.":                                                                                                                        "no. it just wilts and stares at you. which is worse.",
	"aqui é uma biblioteca. dá para reagir a quem começou ou cumpriu uma tarefa, e só.":                                                                                       "this is a library. you can react to whoever started or finished a task, and that's it.",
	"a tarefa espera %d minutos você voltar. se não voltar, ela fica interrompida, o que não é castigo: é o mesmo que parar.":                                                 "the task waits %d minutes for you to come back. if you don't, it's interrupted, which isn't a punishment: it's the same as stopping.",
	"o seu apelido, a impressão digital da sua chave SSH, as suas tarefas e os minutos. nada de senha, e-mail ou nome. a chave privada nunca sai do seu computador.":          "your nickname, your SSH key's fingerprint, your tasks and minutes. no password, email or name. your private key never leaves your computer.",
	"fonte monoespaçada, fundo escuro, zero notificação. não temos dados, mas temos convicção.":                                                                               "monospace font, dark background, zero notifications. we have no data, but we have conviction.",
	"é um terminal.":                "it's a terminal.",
	"no total ":                     "in total ",
	"constância":                    "consistency",
	"mês":                           "month",
	"quando você foca":              "when you focus",
	"enter começa · esc volta":      "enter starts · esc back",
	"tarde":                         "afternoon",
	"recorde":                       "record",
	"menos":                         "less",
	"em média declara ":             "on average you declare ",
	", em %d tarefas":               ", over %d tasks",
	" · recorde ":                   " · record ",
	"noite":                         "night",
	"melhor dia":                    "best day",
	"dias com foco":                 "days with focus",
	" e foca ":                      " and focus ",
	"de %d a %d":                    "%d to %d",
	"%dh às %dh":                    "%dh to %dh",
	"←/→ mês · c personaliza o pet": "←/→ month · c customize the pet",
	"pico ":                         "peak ",
	"estatísticas":                  "stats",
	"outro":                         "other",
	"mais":                          "more",
	"o que você cumpre":             "what you complete",
	" · pelo tempo declarado":       " · by declared time",
	"o mês":                         "the month",
	"digite os minutos":             "type the minutes",
	"%d de %d":                      "%d of %d",
	"semana":                        "week",
	"o mínimo é %d min":             "the minimum is %d min",
	" até hoje":                     " so far",
	" até o dia %d":                 " up to day %d",
	"em %s":                         "in %s",
	"mês passado":                   "last month",
	"quantos minutos?":              "how many minutes?",
	"manhã":                         "morning",
	"o máximo é %d min":             "the maximum is %d min",
	"curtas":                        "short",
	"médias":                        "medium",
	"longas":                        "long",
	"até 15 min":                    "up to 15 min",
	"até 30 min":                    "up to 30 min",
	"mais de 30":                    "over 30",
	"painéis":                       "panels",
	"lista":                         "list",
	// a aba dev (ADR 0011)
	"dev": "dev",
	// a aba dev (ADR 0011)
	"enter confirma · o nome do SSH já é o apelido de outra pessoa": "enter confirms · the SSH name is already someone else's nickname",
	// a aba dev (ADR 0011)
	"energ.": "energy",
	"parado": "idle",
	"antes":  "before",
	"online": "online",
	"abra uma sala na aba sala para pôr gente nela":    "open a room in the room tab to put people in it",
	"abra uma sala na aba sala para passá-la a alguém": "open a room in the room tab to hand it to someone",
	"apelido da pessoa nova":                           "the new person's nickname",
	"enter confirma · esc desiste":                     "enter confirms · esc gives up",
	"setas escolhem · +/- muda · espaço põe ou tira da sala · o passa a sala · t tarefa · x cumpre · n pessoa nova": "arrows choose · +/- change · space puts in or takes out of the room · o hands over the room · t task · x completes · n new person",
	"espaço põe ou tira de ":  "space: in/out of ",
	" · * a sala é da pessoa": " · * owns the room",
	" · %d a %d de %d":        " · %d to %d of %d",
	" (você)":                 " (you)",
	"pessoa nova: ":           "new person: ",
	"tarefa de %s: ":          "%s's task: ",
	"hoje: tarefas de %d min · antes: tarefas de %d min, uma por dia para trás": "today: %d min tasks · before: %d min tasks, one per day going back",
	"energ.: de %d em %d · parado: dias sem focar · online: sessão de mentira":  "energy: %d by %d · idle: days without focus · online: fake session",
	"tarefa: + começa uma de %d min ou dá mais %d, - passa %d min":              "task: + starts one of %d min or adds %d, - skips %d min",
	"outra pessoa: ssh -p <porta> <apelido>@localhost":                          "someone else: ssh -p <port> <nickname>@localhost",
	"ninguém com esse id": "nobody with that id",

	// Editor dev na própria tela.
	"modo dev · prévia local":  "dev mode · local preview",
	" · alterações por salvar": " · unsaved changes",
	" · assumir: %s":           " · assume: %s",
	"seleção: %s":              "selected: %s",
	"energia: %g · ←/→ ou ↑/↓ ajusta · digite o valor":   "energy: %g · arrows adjust · type the value",
	"duração: %.0f min · ←/→ ajusta · digite o valor":    "duration: %.0f min · ←/→ adjust · type the value",
	"↑/↓ escolhe · enter adiciona · ctrl+n cria usuário": "↑/↓ choose · enter adds · ctrl+n creates a user",
	"estágio da obra: %d · -1 usa o progresso real":      "mural stage: %d · -1 uses real progress",
	"n inicia tarefa · f marca feito · p para":           "n starts a task · f marks done · p stops",
	"salvando…": "saving…",
	"enter confirma prévia · esc cancela campo":                      "enter confirms preview · esc cancels field",
	"ctrl+s salva · x salva e sai · esc volta · ctrl+r volta a você": "ctrl+s saves · x saves and exits · esc back · ctrl+r returns to you",
	"ctrl+e energia": "ctrl+e energy",
	"ctrl+n nome · ctrl+e energia · ctrl+l aparência · ctrl+o estágio da obra · ctrl+p presença · ctrl+d remove · ctrl+t tarefa · ctrl+u assume": "ctrl+n name · ctrl+e energy · ctrl+l appearance · ctrl+o mural stage · ctrl+p presence · ctrl+d removes · ctrl+t task · ctrl+u assumes",
	"←/→ membro · enter edita · ctrl+b adiciona · ctrl+o obra · ctrl+u assume":                                                                   "←/→ member · enter edits · ctrl+b adds · ctrl+o mural · ctrl+u assumes",
	"identidade alterada; rascunhos anteriores descartados":                                                                                      "identity changed; previous drafts discarded",
	"há alterações por salvar; repita para sair sem salvar":                                                                                      "unsaved changes; press again to quit without saving",
	" · x modo dev":          " · x dev mode",
	" · h oculta opções dev": " · h hides dev options",
	"enter assume identidade e descarta rascunhos · esc cancela": "enter assumes identity and discards drafts · esc cancels",
	"ctrl+u assume usuário":                            "ctrl+u assumes user",
	"energia de 0 a 100":                               "energy from 0 to 100",
	"modo dev desativado":                              "dev mode is disabled",
	"obra inválida":                                    "invalid mural",
	"cor inválida":                                     "invalid color",
	"estágio inválido":                                 "invalid stage",
	"já existe uma tarefa em andamento":                "a task is already running",
	"nenhuma tarefa em andamento":                      "no task is running",
	"ação inválida":                                    "invalid action",
	"a sala já está cheia":                             "the room is full",
	"o cenário mudou; cancele e abra o editor de novo": "the scene changed; cancel and reopen the editor",
	"tarefa de 1 a 80 caracteres":                      "task from 1 to 80 characters",
	"tempo de 5 a 120 min":                             "time from 5 to 120 min",
	"assumir identidade: %s":                           "assume identity: %s",
	"enter edita · ←/→ membro · ? atalhos · ctrl+s salva · x salva e sai · esc volta": "enter edits · ←/→ member · ? shortcuts · ctrl+s saves · x saves/exits · esc back",
	"salve a alteração de tarefa antes de outra ação":                                 "save the task change before another action",
	"a duração precisa ser um número inteiro":                                         "the duration must be a whole number",
	"↑/↓ opção · ←/→ altera · enter confirma prévia · esc volta":                      "↑/↓ option · ←/→ change · enter confirms preview · esc back",
	"ctrl+n nome": "ctrl+n name",
}
