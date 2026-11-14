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
	{regexp.MustCompile(`^nenhuma sala com o código "?([^"]*)"?$`), "no room with the code %s"},
	{regexp.MustCompile(`^(.*) já está cheia \((\d+) membros\)$`), "%s is already full (%s members)"},
	{regexp.MustCompile(`^(.*) não está aceitando mais ninguém$`), "%s isn't taking anyone else"},
	{regexp.MustCompile(`^não existe obra "?([^"]*)"?$`), "there's no build called %s"},
	{regexp.MustCompile(`^meta fora da lista: (\d+) min$`), "goal not in the list: %s min"},
}

// en: português → inglês. Em minúsculas, como todo texto da interface.
var en = map[string]string{
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
	"a sua primeira tarefa cumprida aparece aqui":                   "your first completed task shows up here",
	" · a tela de descanso durante o foco":                          " · the resting screen while you focus",
	"a/d troca de aba · q sai":                                      "a/d switch tab · q quit",
	"abertas":                                                       "open",
	" abra as novas entradas em ":                                   " open new entries in ",
	"abril":                                                         "april",
	"acampamento":                                                   "campsite",
	"agosto":                                                        "august",
	"amarelo":                                                       "yellow",
	"aparece":                                                       "shown",
	"apelido de até %d letras, por favor (日 e 🐱 contam como 2)":    "nickname up to %d letters, please (日 and 🐱 count as 2)",
	"apelido sem espaços, por favor":                               "nickname without spaces, please",
	"apelido só com letras, números e símbolos simples, por favor": "nickname with only letters, numbers and simple symbols, please",
	"apenas para visualização":                                     "just a preview",
	"aperta ":                                                      "press ",
	"até 25 min":                                                   "up to 25 min",
	"aumente o terminal para editar no jardim":                     "make the terminal bigger to edit in the garden",
	"azul":                            "blue",
	"bem":                             "good",
	"brainrot (para devs de verdade)": "brainrot (for real devs)",
	" · c configura":                  " · c settings",
	"c continuar":                     "c continue",
	"c personaliza o pet":             "c customize the pet",
	"café":                            "coffee",
	"cansado":                         "tired",
	"carregando…":                     "loading…",
	"catppuccin":                      "catppuccin",
	"chuva lenta":                     "slow rain",
	"cidade":                          "city",
	"cole com ctrl+v (cmd+v no mac): o código ou o comando ssh inteiro": "paste with ctrl+v (cmd+v on mac): the code or the whole ssh command",
	"colegas focando": "others focusing",
	"começou":         "started",
	"configurar":      "settings",
	" · configurar.":  " · settings.",
	"cor":             "color",
	"←/→ cor · ↑/↓ prévia do humor · enter salva a cor · esc volta": "←/→ color · ↑/↓ preview mood · enter saves the color · esc back",
	"coral":       "coral",
	"criar sala":  "create room",
	"✓ cumprida":  "✓ done",
	"cumpriu":     "finished",
	"céu noturno": "night sky",
	" · código ":  " · code ",
	"código %s":   "code %s",
	"código enviado para a área de transferência":        "code sent to the clipboard",
	"%d de %d cumpridas":                                 "%d of %d done",
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
	"remover %s de %s?": "remove %s from %s?",
	"←/→ cor · ↑/↓ prévia do humor · enter salva e abre a sala": "←/→ color · ↑/↓ preview mood · enter saves and opens the room",
	"sala solo":                  "solo room",
	"sala solo · focando":        "solo room · focusing",
	"%d/%d membros":              "%d/%d members",
	" · dados de exemplo":        " · sample data",
	"declara %d min, foca %s":    "declares %d min, focuses %d",
	"depois de %.0f s":           "after %.0f s",
	"dezembro":                   "december",
	"dia seguido":                "day in a row",
	"dias seguidos":              "days in a row",
	"digite o código da sala":    "type the room code",
	"dom":                        "sun",
	"dormindo":                   "asleep",
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
	"laranja":                         "orange",
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
	"murcho":                          "wilted",
	"nenhum":                          "none",
	"nenhuma":                         "none",
	"nenhuma ainda":                   "none yet",
	"nenhuma sala com o código %s":    "no room with the code %s",
	"nenhuma tarefa em andamento.":    "no task in progress.",
	"ninguém mais está focando agora": "no one else is focusing right now",
	" no mês · ":                      " this month · ",
	"no sshiu":                        "in sshiu",
	"no terminal":                     "in a terminal",
	" no total":                       " in total",
	"nome":                            "name",
	"nome da sala":                    "room name",
	"nord":                            "nord",
	"normal":                          "okay",
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
	"por tarefa": "per task",
	"porto":      "harbor",
	"primeira vez por aqui. qual vai ser o seu apelido?": "first time here. what will your nickname be?",
	"prévia do pet":               "pet preview",
	"prévia: ":                    "preview: ",
	"próximo humor em %d · %s %d": "next mood at %d · %[3]d %[2]s",
	"qua":                         "wed",
	"qual a próxima coisa concreta que você vai fazer?": "what's the next concrete thing you'll do?",
	"quantos membros":                "how many members",
	"quanto tempo?":                  "how long?",
	"quem mais está focando":         "who else is focusing",
	"qui":                            "thu",
	"radiante":                       "radiant",
	"reagir ao evento selecionado: ": "react to the selected event: ",
	"relógio":                        "clock",
	"remover alguém":                 "remove someone",
	"remover quem da sala?":          "remove whom from the room?",
	"revisar a lista 3":              "review problem set 3",
	"rosa":                           "pink",
	"roxo":                           "purple",
	" · s salas":                     " · s rooms",
	"s salas":                        "s rooms",
	"%s · meta %s · a peça muda a cada %s · reinicia em %s": "%s · goal %s · the piece changes every %s · resets in %s",
	"sair de %s":                           "leave %s",
	"sair de %s?":                          "leave %s?",
	"saiu":                                 "left",
	"sala":                                 "room",
	"sala demo":                            "demo room",
	"sala solo criada: só você entra aqui": "solo room created: only you come in here",
	"sala solo: para focar sem companhia, com o seu pet e o modo jardim": "solo room: to focus on your own, with your pet and garden mode",
	"salas":       "rooms",
	"salvo":       "saved",
	"seg":         "mon",
	"sem foco":    "no focus",
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
	"tarefas":                          "tasks",
	"tema":                             "theme",
	"tempo esgotado":                   "time's up",
	"ter":                              "tue",
	"tokyo night":                      "tokyo night",
	"trem":                             "train",
	" · ← → troca":                     " · ← → switch",
	"← → troca · qualquer tecla volta": "← → switch · any key goes back",
	"turquesa":                         "turquoise",
	"um lugar quieto para focar junto": "a quiet place to focus together",
	"viagem":                           "trip",
	"vídeo de orbital (CC BY)":         "video by orbital (CC BY)",
	"verde":                            "green",
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
	"período":                          "period",
	"quem vê":                          "who sees",
	"formato":                          "format",
	"tabela":                           "table",
	"ranking":                          "ranking",
	"a obra não coube junto da tabela: aumente a janela para ver as duas": "the mural didn't fit next to the table: enlarge the window to see both",
	" · o tabela": " · o table",
	"←/→ período · o volta para a sala":  "←/→ period · o back to the room",
	"↑/↓ escolhe · ←/→ muda · esc volta": "↑/↓ choose · ←/→ change · esc back",
}
