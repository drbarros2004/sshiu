# Regras numéricas

Todo número que decide como o app se comporta fica listado aqui, com o porquê e o link para onde está no código. Serve para auditar as regras sem ler o código e como freio: se uma regra não cabe aqui de forma simples, ela complicou antes da hora.

Ao mudar um número no código, mude aqui também.

## Energia e humor do pet

As regras da energia estão em [`internal/pet/mood.go`](../internal/pet/mood.go). A energia vem dos minutos focados de verdade, não do tempo declarado: declarar 60 min e marcar "feito" no minuto 1 não rende.

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Energia inicial | 50 (de 0 a 100) | Começa no meio: humor "normal". | [mood.go:19](../internal/pet/mood.go#L19) |
| Decaimento | −8 por dia inteiro sem focar | Sumir é o que pune. | [mood.go:20](../internal/pet/mood.go#L20) |
| Energia por minuto focado | 8 a cada 30 min (≈ 0,27 por minuto) | 30 min de foco devolvem o que um dia parado tira. | [mood.go:27](../internal/pet/mood.go#L27) |
| Bônus por tarefa cumprida | +3, só a partir de 10 min focados | Vale uns 10 min de foco: recompensa clara sem fazer várias tarefas curtas renderem mais que uma longa. Uma tarefa de 5 min rende só os minutos, para o bônus não virar máquina de pontos. | [mood.go:28-29](../internal/pet/mood.go#L28-L29) |
| Streak | o dia conta a partir de 15 min focados | Pouco o bastante para um dia corrido, mais que abrir e fechar o app. | [mood.go:30](../internal/pet/mood.go#L30) |
| Bônus de streak | +2 por dia seguido, até +10 | Recompensa a constância sem deixar o streak dominar a energia. | [mood.go:21-22](../internal/pet/mood.go#L21-L22) |
| Faixas de humor | murcho < 20 ≤ cansado < 40 ≤ normal < 60 ≤ bem < 80 ≤ radiante | Cinco faixas iguais. | [mood.go:67](../internal/pet/mood.go#L67) |
| Barra de energia | 20 células, 5 pontos por célula | Exibe os pontos inteiros e marca a borda direita da célula que completa o próximo limite. | [profile_energy.go](../internal/tui/profile_energy.go) |

```go
var moodMin = [...]float64{Wilted: 0, Tired: 20, Okay: 40, Good: 60, Radiant: 80}
```

A energia mostrada é a guardada, menos o decaimento, mais o bônus de streak, sempre entre 0 e 100 ([mood.go:89](../internal/pet/mood.go#L89)). Cada tarefa encerrada soma os seus minutos e o bônus na hora ([mood.go:34](../internal/pet/mood.go#L34)).

## Cores e folha do pet

- Cada paleta tem uma cor por parte, fixa nos 6 humores visuais (incluindo dormindo).
  Energia e humor não alteram as cores. Os nomes e os valores ficam em
  [`palettes.json`](../internal/pet/palettes.json).
- O template mede 64×96 pixels: 6 linhas de quadros 16×16; 2 quadros por humor e 4 no sono.
  As cores de marcador identificam partes e não são as cores mostradas no terminal.
- Os 8 IDs existentes são mantidos, para preservar escolhas salvas e os instrumentos da obra.
  As 8 paletas reproduzem os PNGs de referência.
- HEX pode ser convertido pelo terminal de 256 cores; há 3 aproximações nos PNGs atuais,
  descritas no [guia de paletas](pets-paletas.md). `xterm:0` a `xterm:255` também são aceitos;
  os primeiros 16 índices dependem do tema do terminal.

Veja [como editar cores e adicionar grupos](pets-paletas.md).

## A tarefa

Tudo em [`internal/rooms/room.go`](../internal/rooms/room.go), salvo onde indicado.

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Tempos declarados | 5, 10, 15, 20, 25, 30, 45 ou 60 min (padrão 25); na demo, também 1 min | O 5 min é para a tarefa curta que também é concreta ("responder o e-mail da monitoria"); sem ele, quem ia fazer algo rápido declarava 10 e parava no meio. 20 e 30 min permitem estimativas intermediárias sem exigir digitação. A plateia da demo vê o fluxo inteiro com 1 min. | [tui/room.go:42-46](../internal/tui/room.go#L42-L46) |
| Tempo esgotado sem resposta | os minutos param de contar | Estar ausente nunca rende mais do que o tempo que o próprio usuário declarou. | [room.go:62](../internal/rooms/room.go#L62) |
| Trocar, parar, desconectar | sem penalidade | Os minutos já contaram; punir faria os usuários declararem tempos folgados ou deixarem o terminal aberto. | [room.go:247](../internal/rooms/room.go#L247) |
| Tolerância de reconexão | 2 min | Conexão instável não interrompe a tarefa. Sem volta, a tarefa é interrompida com os minutos até a desconexão. | [room.go:30](../internal/rooms/room.go#L30) |
| Quem fechou o terminal | offline na hora, com o pet dormindo | A tolerância é da tarefa, não do que a sala mostra: ninguém "foca" com o terminal fechado, e o relógio parar de correr no card é o que avisa a sala. | [room.go:79](../internal/rooms/room.go#L79) |
| Restart do servidor | desconexão de todo mundo na hora em que o servidor volta, com a mesma tolerância de 2 min | Uma atualização não derruba o foco da turma: quem reconecta a tempo vê o relógio onde estava. Sem volta, a tarefa é interrompida com os minutos até o restart. A tarefa em andamento fica no banco; eventos e reações se perdem. | [room.go:121](../internal/rooms/room.go#L121) |

## Na sala

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Apelido | até 16 colunas, sem espaço nem ":"; caracteres largos (日, 🐱) valem 2; nada de acento combinante, emoji composto ou caractere invisível | Cabe no card da sala ao lado do humor e na linha do modo jardim. Cada terminal conta a largura dos caracteres compostos de um jeito, e a tela desalinhava ([#84](https://github.com/drbarros2004/sshiu/issues/84), [#85](https://github.com/drbarros2004/sshiu/issues/85)). | [ui.go:466-471](../internal/tui/ui.go#L466-L471) |
| Card de cada pessoa na sala | miolo de 16 a 25 colunas, o mesmo para todos: o que couber "apelido humor" do maior apelido da sala | A grade fica alinhada e o pet não sai do lugar com apelido largo; o apelido que passa (de antes do limite) termina em "…" ([#84](https://github.com/drbarros2004/sshiu/issues/84)). | [tui/room.go:523-524](../internal/tui/room.go#L523-L524) |
| Tamanho da sala | de 1 a 10, escolhido na criação (padrão 10) e mudado depois pelo dono, nunca abaixo do número de membros | Uma sala é um grupo de estudo, não um auditório: todo mundo cabe na tela e o custo de desenhar a sala cresce com o quadrado do tamanho dela. Quem cria sabe o tamanho da turma, e o dono corrige se ela mudar ([ADR 0007](adr/0007-dono-da-sala.md)); o chão é a sala como ela está, para ninguém ser expulso por uma mudança de número. O 1 é a sala solo ([ADR 0006](adr/0006-sala-solo.md)). | [room.go:24-25](../internal/rooms/room.go#L24-L25) |
| Reações | 4 (👏 🔥 💪 ☕), teclas 1 a 4; as abas ficam em a/d e tab | Uma tecla cada, sem chat ([ADR 0004](adr/0004-sem-chat-so-reacoes.md)). | [room.go:23](../internal/rooms/room.go#L23) |
| Meta da obra | 30 min, 1 h, 1 h 30, 2 h ou 3 h por pessoa (padrão 1 h), escolhida pelo dono | O dia de estudo de uma turma: perto o bastante para a peça andar numa tarde, longe o bastante para a obra não acabar no primeiro intervalo. | [room.go:28](../internal/rooms/room.go#L28) |
| Estágios da peça | 4, contando o vazio (muda nos terços da meta) | Com 1 h de meta, a peça muda aos 20, 40 e 60 min: mudança que dá para ver numa sessão de foco, sem virar animação. São 4 desenhos por tema ([ADR 0008](adr/0008-obra-da-sala.md)). | [room.go:29](../internal/rooms/room.go#L29) |
| Eventos guardados | 20 por sala, 6 na tela | O bastante para reagir ao que acabou de acontecer. | [room.go:27](../internal/rooms/room.go#L27), [tui/room.go:20](../internal/tui/room.go#L20) |
| Reinício diário da obra | meia-noite no horário local do servidor; contador HH:MM:SS no cabeçalho | Usa o tick existente da tela e arredonda o tempo restante para cima. Na virada, minutos de tarefas encerradas ontem saem do progresso; a tarefa em andamento continua contando, conforme a regra atual. | [obra.go](../internal/tui/obra.go), [room.go](../internal/rooms/room.go) |
| Bots da demo | tarefas de 40 a 90 s, uns 90 s parados entre uma e outra | A sala parece viva sem empurrar os seus eventos para fora da tela. | [bots.go:78-79](../internal/rooms/bots.go#L78-L79) |

## Ritmo da tela

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Prévia dos humores no perfil | pet menor em terminais com menos de 30 linhas | Mantém o estado da prévia e o aviso de que humor e energia não mudam visíveis em 80×24. | [profile.go](../internal/tui/profile.go) |
| Tick da tela | 500 ms | Anima o pet e o relógio e atualiza a sala. | [ui.go:103](../internal/tui/ui.go#L103) |
| Respiração do pet | 1 a 4 ticks por quadro (radiante rápido, murcho lento), 2 quadros; dormindo são 4, a 3 ticks | O cansaço aparece no ritmo, não só no desenho. No sono os dois quadros a mais são o z subindo da cabeça, num ciclo de 6 s. | [render.go:89](../internal/pet/render.go#L89) |
| Movimento da obra | 2 quadros: trem e fogueira 1 tick por quadro, orquestra e passarinhos 2, porto 3 | Mexe no ritmo do pet: o trem anda e o fogo crepita mais depressa, a água balança devagar. Pouco movimento, porque a obra fica no canto do olho de quem foca. | [obra.go:106-110](../internal/tui/obra.go#L106-L110) |
| Modo jardim | liga depois de 30 s sem tecla, só durante o foco, se "liga sozinho" estiver em "depois de 30 s" na aba visual | Quem nunca abriu o jardim de propósito pode desligar e usar só o `j` ([#35](https://github.com/drbarros2004/sshiu/issues/35)). | [garden.go:27](../internal/tui/garden.go#L27) |
| Jardim na sala solo | ocupa a direita a partir de 96 colunas de terminal | Menos que isso e o painel fica apertado; aí a tela é só o painel. | [room.go:21-22](../internal/tui/room.go#L21-L22) |
| Editar no jardim | `v` abre o painel a partir de 64×20 células. A lista completa pode aparecer a partir de 120×36 | Abaixo desse tamanho ou quando a lista encobre a cena ou as informações de foco, mostra só a opção selecionada. Espaço oculta o painel, outra tecla o traz de volta e Esc volta para a aba visual. | [garden_settings.go](../internal/tui/garden_settings.go) |
| Nome do wallpaper no modo jardim | 5 s, ao abrir e a cada troca (←/→), no rodapé | Quem troca vê o nome do novo e descobre as setas, sem texto fixo na tela de descanso. No rodapé, ele some sem mexer no relógio. | [garden.go:28](../internal/tui/garden.go#L28) |
| Tarefa no modo jardim | até 40 caracteres numa linha; mais que isso, duas, quebrando no espaço a até 12 do meio (ou com hífen no meio de uma palavra) | O bloco fica estreito e organizado em cima do desenho. | [garden.go:255](../internal/tui/garden.go#L255) |
| Quem está focando, no modo jardim | até 4 apelidos; com mais gente, 3 e "mais N". Se a linha não cabe na tela, saem nomes até caber, e o último encurta com "…" | A frase nunca passa da largura da tela, com nome largo ou terminal estreito ([#85](https://github.com/drbarros2004/sshiu/issues/85)). | [garden.go:204](../internal/tui/garden.go#L204) |
| Quadros do modo jardim | 8 por segundo | Movimento suave. Medido no protótipo: de 2 a 24 KB/s (em pixel art: árvore 4, café 21 e estrada 31 KB/s com o painel e 29 sem) e de 4 a 6% de um núcleo por usuário; 5 fps gasta uns 40% menos CPU. | [garden.go:26](../internal/tui/garden.go#L26) |
| Padrões do modo jardim | wallpaper sorteado a cada tarefa, relógio só com os minutos, tarefa e sala aparecem, liga sozinho depois de 30 s, wallpapers para devs de verdade não | Quem chega vê o que o modo jardim oferece e desliga o que achar demais. O brainrot é de piada e gasta ~8× a banda da chuva: só entra na lista, nas setas e no sorteio de quem pede ([#36](https://github.com/drbarros2004/sshiu/issues/36)). | [db.go:67](../internal/store/db.go#L67) |
| Brainrot | dois vídeos de 1 min, a 8 quadros por segundo, em 54 × 48 células no tamanho cheio (ou 75%, 50%, 35% do que couber) | Vertical, como no celular. Os quadros em ANSI (~70 MB) são montados na primeira vez que alguém abre o brainrot, em ~0,5 s, e servem para todo o servidor. | [brainrot.go](../internal/tui/brainrot.go), [brainrot_gen.go](../internal/tui/brainrot_gen.go) |
