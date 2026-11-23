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

- PNGs coloridos e camadas em [assets/pets](../assets/pets/README.md). A energia não muda as cores.
- Layout padrão: quadros 16×16, seis linhas, [2,2,2,2,2,4] quadros ativos. Tamanhos, quadros e
  duração vêm de sprite-layout.json, com sobrescrita por personagem no arquivo de mesmo nome.
- Chapéus: caixa 16×6 definida em hats.json; hatY de 1 até frameHeight. A composição reserva
  o maior espaço necessário acima do corpo; os cartões da mesma fileira alinham suas bases.
- IDs e ordem das oito cores das obras são preservados no catálogo de obras. Claude normal usa coral.
- O terminal pode aproximar cores fora da tabela xterm; os PNGs não são recoloridos pelo aplicativo.
- Cache de desenho: até 512 quadros, renovado ao atingir o teto, para limitar a memória usada pelas combinações de aparência.
- Perfil com menos de 26 linhas usa o desenho pequeno e uma disposição lateral; o tamanho do
  desenho não é reduzido por aproximação. O espaçamento é editável nas constantes de profile.go.

Veja [como editar os personagens e chapéus](pets-paletas.md).

## A tarefa

Tudo em [`internal/rooms/room.go`](../internal/rooms/room.go), salvo onde indicado.

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Tempos declarados | 5, 10, 15, 20, 25, 30, 45 ou 60 min; na demo, também 1 min | O 5 min é para a tarefa curta que também é concreta ("responder o e-mail da monitoria"); sem ele, quem ia fazer algo rápido declarava 10 e parava no meio. 20 e 30 min permitem estimativas intermediárias sem exigir digitação. A plateia da demo vê o fluxo inteiro com 1 min. | [tui/room.go:48](../internal/tui/room.go#L48) |
| Outro tempo | o último item da lista, "outro", abre um campo de 5 a 120 min | O mínimo é o menor tempo da lista. O teto existe porque, com o tempo esgotado e sem resposta, os minutos só param no tempo declarado: 4 h declaradas seriam 4 h de foco com o terminal aberto e ninguém na frente ([#45](https://github.com/drbarros2004/sshiu/issues/45)). | [tui/room.go:57](../internal/tui/room.go#L57) |
| Último tempo usado | o seletor abre nele, marcado com ↺; se ele veio do "outro", entra na lista no lugar dele. Quem nunca declarou começa em 25 min | Quem sempre faz 50 min não percorre a lista toda vez, e a lista não cresce com um item por tempo já usado. Vale para começar e para continuar. | [tui/room.go:61](../internal/tui/room.go#L61) |
| Tempo esgotado sem resposta | os minutos param de contar | Estar ausente nunca rende mais do que o tempo que o próprio usuário declarou. | [room.go:62](../internal/rooms/room.go#L62) |
| Trocar, parar, desconectar | sem penalidade | Os minutos já contaram; punir faria os usuários declararem tempos folgados ou deixarem o terminal aberto. | [room.go:247](../internal/rooms/room.go#L247) |
| Tolerância de reconexão | 2 min | Conexão instável não interrompe a tarefa. Sem volta, a tarefa é interrompida com os minutos até a desconexão. | [room.go:30](../internal/rooms/room.go#L30) |
| Quem fechou o terminal | offline na hora, com o pet dormindo | A tolerância é da tarefa, não do que a sala mostra: ninguém "foca" com o terminal fechado, e o relógio parar de correr no card é o que avisa a sala. | [room.go:79](../internal/rooms/room.go#L79) |
| Restart do servidor | desconexão de todo mundo na hora em que o servidor volta, com a mesma tolerância de 2 min | Uma atualização não derruba o foco da turma: quem reconecta a tempo vê o relógio onde estava. Sem volta, a tarefa é interrompida com os minutos até o restart. A tarefa em andamento fica no banco; eventos e reações se perdem. | [room.go:121](../internal/rooms/room.go#L121) |

## Na sala

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Apelido | até 16 colunas, sem espaço nem ":"; caracteres largos (日, 🐱) valem 2; nada de acento combinante, emoji composto ou caractere invisível | Cabe no card da sala ao lado do humor e na linha do modo jardim. Cada terminal conta a largura dos caracteres compostos de um jeito, e a tela desalinhava ([#84](https://github.com/drbarros2004/sshiu/issues/84), [#85](https://github.com/drbarros2004/sshiu/issues/85)). | [ui.go:469](../internal/tui/ui.go#L469) |
| Card de cada pessoa na sala | miolo de 20 a 29 colunas, o mesmo para todos: "apelido humor" do maior apelido da sala + 4 colunas de respiro; três tamanhos: cheio (18 linhas, com respiro em volta do pet), médio (13 linhas: sem o respiro, a tarefa numa linha e as reações no status quando cabem) e ponto; as fileiras se dividem por igual (9 em fileiras de 8: 5 e 4) | A grade fica alinhada e o pet não sai do lugar com apelido largo; o apelido que passa (de antes do limite) termina em "…" ([#84](https://github.com/drbarros2004/sshiu/issues/84)). O médio guarda o rosto quando o cheio não cabe ([ADR 0012](adr/0012-arranjo-da-sala.md)). | [cardExtraWidth, cardPetPaddingTop, cardPetPaddingBottom e cards](../internal/tui/room.go) |
| Tamanho da sala | de 1 a 10, escolhido na criação (padrão 10) e mudado depois pelo dono, nunca abaixo do número de membros | Uma sala é um grupo de estudo, não um auditório: todo mundo cabe na tela e o custo de desenhar a sala cresce com o quadrado do tamanho dela. Quem cria sabe o tamanho da turma, e o dono corrige se ela mudar ([ADR 0007](adr/0007-dono-da-sala.md)); o chão é a sala como ela está, para ninguém ser expulso por uma mudança de número. O 1 é a sala solo ([ADR 0006](adr/0006-sala-solo.md)). | [room.go:24-25](../internal/rooms/room.go#L24-L25) |
| Reações | 4 (👏 🔥 💪 ☕), teclas 1 a 4; as abas ficam em a/d e tab | Uma tecla cada, sem chat ([ADR 0004](adr/0004-sem-chat-so-reacoes.md)). | [room.go:23](../internal/rooms/room.go#L23) |
| Meta da obra | 30 min, 1 h, 1 h 30, 2 h ou 3 h por pessoa (padrão 1 h), escolhida pelo dono | O dia de estudo de uma turma: perto o bastante para a peça andar numa tarde, longe o bastante para a obra não acabar no primeiro intervalo. | [room.go:28](../internal/rooms/room.go#L28) |
| Estágios da peça | 4, contando o vazio (muda nos terços da meta) | Com 1 h de meta, a peça muda aos 20, 40 e 60 min: mudança que dá para ver numa sessão de foco, sem virar animação. São 4 desenhos por tema ([ADR 0008](adr/0008-obra-da-sala.md)). | [room.go:29](../internal/rooms/room.go#L29) |
| Arranjo da obra | até 3 peças numa fileira; daí em diante duas, com a sobra em cima (9: 5 e 4); no acampamento, a de baixo leva um número par de barracas, com a fogueira no meio; sem largura para duas fileiras, a fileira que quebra | Baixa e larga, a obra cabe na faixa ao lado da tabela de foco ([ADR 0012](adr/0012-arranjo-da-sala.md)). | [twoRows e campRows](../internal/tui/obra.go) |
| Eventos guardados | 20 por sala, 6 na tela | O bastante para reagir ao que acabou de acontecer. | [room.go:27](../internal/rooms/room.go#L27), [tui/room.go:20](../internal/tui/room.go#L20) |
| Reinício diário da obra | meia-noite no horário local do servidor; contador HH:MM:SS no cabeçalho | Usa o tick existente da tela e arredonda o tempo restante para cima. Na virada, minutos de tarefas encerradas ontem saem do progresso; a tarefa em andamento continua contando, conforme a regra atual. | [obra.go](../internal/tui/obra.go), [room.go](../internal/rooms/room.go) |
| Tabela de foco | fechada (só o dono vê) e em tabela, até o dono mudar; no ranking, empate divide a posição e quem tem 0 min fica sem número | O ranking compara as pessoas, então alguém escolhe ligá-lo ([ADR 0010](adr/0010-tabela-de-foco.md)). Hoje são os minutos da obra e zeram junto com ela. | [board.go](../internal/tui/board.go) |
| Períodos da tabela de foco | hoje, esta semana (de segunda), este mês ou o total, escolhido por quem olha; a tarefa conta no dia em que termina; o banco é relido ao abrir, ao trocar o período e na virada do dia | Os dias anteriores só mudam à meia-noite; hoje vem ao vivo no retrato. | [board.go](../internal/tui/board.go) |
| Faixa da obra e da tabela na sala | embaixo dos cards, a obra e a tabela inteira lado a lado; os eventos entram nela, à esquerda, quando sobram 50 colunas, ocupando até 80; apertada, a tabela fica na altura da obra (no mínimo 3 pessoas) com "+N no o"; fora da tela, uma linha aponta a tecla o; com ela na tela, pelo menos 2 eventos | A faixa encosta na esquerda e a sobra fica à direita. A ordem de encolher é do [ADR 0012](adr/0012-arranjo-da-sala.md): a faixa encolhe com os cards cheios e com os médios, sai inteira, e só então os pets viram ponto. | [feedMin, feedMax, fitSteps e roomBand](../internal/tui/room.go), [board.go](../internal/tui/board.go) |
| Bots da demo | tarefas de 40 a 90 s, uns 90 s parados entre uma e outra; se o apelido já é de alguém, o bot vira `<apelido>-bot` | A sala parece viva sem empurrar os seus eventos para fora da tela. Sem o sufixo, uma pessoa chamada ana impedia a demo de subir. | [bots.go:102-103](../internal/rooms/bots.go#L102-L103), [bots.go:80](../internal/rooms/bots.go#L80) |

## As stats

O que o perfil e a sala solo parada mostram sobre o seu foco ([#45](https://github.com/drbarros2004/sshiu/issues/45)). Tudo em [`internal/tui/stats.go`](../internal/tui/stats.go), salvo onde indicado.

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Dia de uma tarefa | o dia em que ela termina, no fuso do servidor | O mesmo dia do pet e da tabela de foco. Antes o calendário contava pelo início, e uma tarefa das 23h50 à 0h20 caía em dias diferentes no calendário e na tabela. | [db.go:279](../internal/store/db.go#L279) |
| Meses do calendário | de janeiro de 2026 até o mês de agora, com ←/→ no perfil; a sala solo mostra só o mês de agora | O histórico do sshiu começa em 2026. Na sala solo, ←/→ já trocam o wallpaper. O perfil sempre abre no mês de agora. | [stats.go:78](../internal/tui/stats.go#L78) |
| Cor do dia | 4 níveis: sem foco, até 25 min, de 25 min a 1 h, 1 h ou mais; os dias que ainda não chegaram ficam apagados | Os mesmos cortes de antes, agora em blocos do fundo apagado do tema até a cor de destaque, com a legenda "menos ▪▪▪▪ mais". | [stats.go:89](../internal/tui/stats.go#L89) |
| Mês passado | no mês de agora, a comparação vai até o mesmo dia do mês anterior | Meio mês contra o anterior inteiro dava sempre "menos", e enganava. | [stats.go:153](../internal/tui/stats.go#L153) |
| Recorde de streak | a maior sequência de dias com 15 min ou mais desde 2026, nunca abaixo do streak de agora | A mesma regra do dia que conta para o streak do pet. | [stats.go:208](../internal/tui/stats.go#L208) |
| Faixas do tempo declarado | curtas até 15 min, médias até 30, longas acima de 30 | O tempo declarado inclui o que foi somado ao continuar e, com o "outro", pode ser qualquer número: as faixas não se espalham em minutos soltos. | [db.go:335](../internal/store/db.go#L335) |
| Horário em que você foca | o foco de cada tarefa se espalha por igual entre o início e o fim; o pico são as 2 horas seguidas com mais foco; manhã é das 5h às 12h, tarde das 12h às 18h e noite o resto | Duas horas dizem mais que uma hora solta, e três partes do dia cabem numa linha. | [db.go:300](../internal/store/db.go#L300), [stats.go:275](../internal/tui/stats.go#L275) |
| Layout | painéis por padrão; lista na aba visual | Provisório: os dois convivem até a turma escolher um, e aí o outro sai. | [profile.go:237](../internal/tui/profile.go#L237) |
| Terminal baixo | as seções saem de baixo para cima (últimas tarefas, faixas, horário, números do mês), sempre inteiras; o pet, o estado e o calendário ficam | O que identifica a pessoa e o mês vem primeiro. Seção pela metade, ou o título das tarefas sem nenhuma, não aparece. | [stats.go:384](../internal/tui/stats.go#L384) |

## Ritmo da tela

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Pet menor no perfil | quando falta altura para o pet grande (umas 30 linhas de terminal), também na prévia dos humores | Em 80×24 cabem inteiros o pet, o estado (com o aviso da prévia) e o calendário, nos dois layouts das stats. | [profile.go](../internal/tui/profile.go) |
| Tick da tela | 500 ms | Anima o pet e o relógio e atualiza a sala. | [ui.go:103](../internal/tui/ui.go#L103) |
| Avisos na tela | somem depois de 8 s, ou na primeira tecla da sala | "salvo", "código enviado" e "a sala agora é de…" dizem o que acabou de acontecer e logo deixam de valer. Ficam no fim de uma linha que já existe, para a tela não pular quando somem. | [ui.go:186](../internal/tui/ui.go#L186) |
| Enter depois de chegar na sala | ignorado por 600 ms depois de criar a sala ou entrar por código | Quem aperta enter de novo enquanto a sala nasce caía no campo da tarefa da sala nova, e o enter seguinte mostrava um erro. | [room.go:264](../internal/tui/room.go#L264) |
| Animação dos personagens | frameDurationMs: 2000, 1500, 1000, 1000, 500, 1500 ms por quadro; frameCounts: 2, 2, 2, 2, 2, 4 | Valores padrão na ordem dos humores. Cada personagem pode sobrescrever os mesmos campos no seu sprite-layout.json. A tela amostra a animação a cada tick de 500 ms. | [sprite-layout.json](../assets/pets/sprite-layout.json), [appearance.go](../internal/pet/appearance.go) |
| Movimento da obra | 2 quadros: trem e fogueira 1 tick por quadro, orquestra e passarinhos 2, porto 3 | Mexe no ritmo do pet: o trem anda e o fogo crepita mais depressa, a água balança devagar. Pouco movimento, porque a obra fica no canto do olho de quem foca. | [obra.go:106-110](../internal/tui/obra.go#L106-L110) |
| Modo jardim | liga depois de 30 s sem tecla, só durante o foco, se "liga sozinho" estiver em "depois de 30 s" na aba visual | Quem nunca abriu o jardim de propósito pode desligar e usar só o `j` ([#35](https://github.com/drbarros2004/sshiu/issues/35)). | [garden.go:27](../internal/tui/garden.go#L27) |
| Jardim na sala solo | ocupa a direita a partir de 96 colunas de terminal | Menos que isso e o painel fica apertado; aí a tela é só o painel. | [room.go:21-22](../internal/tui/room.go#L21-L22) |
| Editar no jardim | `v` abre o painel a partir de 64×20 células. A lista completa pode aparecer a partir de 120×36 | Abaixo desse tamanho ou quando a lista encobre a cena ou as informações de foco, mostra só a opção selecionada. Espaço oculta o painel, outra tecla o traz de volta e Esc volta para a aba visual. | [garden_settings.go](../internal/tui/garden_settings.go) |
| Nome do wallpaper no modo jardim | 5 s, ao abrir e a cada troca (←/→), no rodapé | Quem troca vê o nome do novo e descobre as setas, sem texto fixo na tela de descanso. No rodapé, ele some sem mexer no relógio. | [garden.go:28](../internal/tui/garden.go#L28) |
| Tarefa no modo jardim | até 40 caracteres numa linha; mais que isso, duas, quebrando no espaço a até 12 do meio (ou com hífen no meio de uma palavra) | O bloco fica estreito e organizado em cima do desenho. | [garden.go:255](../internal/tui/garden.go#L255) |
| Quem está focando, no modo jardim | até 4 apelidos; com mais gente, 3 e "mais N". Se a linha não cabe na tela, saem nomes até caber, e o último encurta com "…" | A frase nunca passa da largura da tela, com nome largo ou terminal estreito ([#85](https://github.com/drbarros2004/sshiu/issues/85)). | [garden.go:204](../internal/tui/garden.go#L204) |
| Idioma da interface | inglês por padrão, para todo mundo (inclusive quem já usava o app); português na aba visual, com prévia ao vivo como o tema | O padrão pedido na [#86](https://github.com/drbarros2004/sshiu/issues/86); quem prefere português troca uma vez e a escolha fica salva. | [db.go](../internal/store/db.go), [i18n.go](../internal/tui/i18n.go) |
| Quadros do modo jardim | 8 por segundo | Movimento suave. Medido no protótipo: de 2 a 42 KB/s (em pixel art: árvore 4, café 21, estrada 31 KB/s com o painel e 29 sem, farol 42, gato e cachorro dormindo 3) e de 4 a 6% de um núcleo por usuário; 5 fps gasta uns 40% menos CPU. | [garden.go:26](../internal/tui/garden.go#L26) |
| Facho e mar do farol | o facho dá uma volta a cada 20 s, andando 4 vezes por segundo; as cristas das ondas andam 1 vez por segundo | Cada passo reescreve as linhas por onde o facho e as ondas passam. Andando a cada quadro, o farol mandava 83 KB/s em 160×50; assim, 42 KB/s, abaixo da chuva ([#20](https://github.com/drbarros2004/sshiu/issues/20)). | [garden_pixels.go:802](../internal/tui/garden_pixels.go#L802), [garden_pixels.go:866](../internal/tui/garden_pixels.go#L866) |
| Padrões do modo jardim | wallpaper sorteado a cada tarefa, relógio só com os minutos, tarefa e sala aparecem, liga sozinho depois de 30 s, wallpapers para devs de verdade não | Quem chega vê o que o modo jardim oferece e desliga o que achar demais. O brainrot é de piada e gasta ~8× a banda da chuva: só entra na lista, nas setas e no sorteio de quem pede ([#36](https://github.com/drbarros2004/sshiu/issues/36)). | [db.go:67](../internal/store/db.go#L67) |
| Brainrot | dois vídeos de 1 min, a 8 quadros por segundo, em 54 × 48 células no tamanho cheio (ou 75%, 50%, 35% do que couber) | Vertical, como no celular. Os quadros em ANSI (~70 MB) são montados na primeira vez que alguém abre o brainrot, em ~0,5 s, e servem para todo o servidor. | [brainrot.go](../internal/tui/brainrot.go), [brainrot_gen.go](../internal/tui/brainrot_gen.go) |
| Brinquedos do brainrot | a prensa esmaga um objeto a cada 4 s (pato, latinha, melancia, ursinho de goma); o sabão perde uma fatia a cada 1,2 s e tem 8 fatias; 3 palavras de aura sobem em cada vão dos lados, cada uma em 3 a 5 s; o vão do meio tem até 32 colunas, a prensa até 14 linhas e o sabão até 10 | Ocupam o vão vazio entre os vídeos, em cima e embaixo do relógio ([#79](https://github.com/drbarros2004/sshiu/issues/79)). Desenhados em Go, sem vídeo de terceiros; numa tela baixa o sabão sai primeiro e depois a prensa. | [brainrot_toys.go](../internal/tui/brainrot_toys.go) |

## Modo dev

Só com a flag `-dev` ([ADR 0011](adr/0011-modo-dev.md)). Os passos da aba dev estão em [`internal/rooms/dev.go`](../internal/rooms/dev.go).

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Energia | de 5 em 5, na energia guardada | Chega a qualquer faixa de humor em poucas teclas. Os dias parados continuam decaindo por cima. | [dev.go:19](../internal/rooms/dev.go#L19) |
| Dias parado | um dia por passo na última atividade, nunca depois de agora | Mostra o decaimento de 8 por dia sem esperar dias de verdade. Encerrar uma tarefa zera, como sempre. | [dev.go:111](../internal/rooms/dev.go#L111) |
| Minutos de hoje | tarefas de 15 min, começando agora | O passo do streak (15 min): um `+` já faz o dia contar. O `-` apaga a última tarefa dev de hoje. | [dev.go:20](../internal/rooms/dev.go#L20) |
| Minutos de antes de hoje | tarefas de 60 min, uma por dia para trás (ontem, anteontem...) | Cada `+` anda um dia, então a semana, o mês e o total da tabela de foco se separam sozinhos. Uma tabela já aberta na tecla `o` só mostra o ajuste depois de reaberta ou de trocar o período. | [dev.go:21](../internal/rooms/dev.go#L21), [dev.go:141](../internal/rooms/dev.go#L141) |
| Tarefa em andamento | o `+` sem tarefa começa uma de 25 min; com tarefa, `+` dá mais 5 min e `-` faz 5 min passarem (até esgotar) | 25 min é o tempo padrão de quem foca de verdade. O `-` leva a tarefa ao "tempo esgotado" sem esperar, e os minutos passados contam como focados. Depois de esgotar, o `+` funciona como o continuar: o tempo sem resposta não conta. | [dev.go:22-23](../internal/rooms/dev.go#L22-L23), [dev.go:197](../internal/rooms/dev.go#L197) |
| Online no ambiente dev | presença simulada persistente | Independente da conexão SSH e da participação na sala; ficar offline não fecha conexões reais nem interrompe a tarefa do cenário dev. | [dev.go](../internal/rooms/dev.go), [dev_scene.go](../internal/rooms/dev_scene.go) |
| Banco e endereço | `sshiu-dev.db`, só localhost | O modo dev não encosta nos dados de verdade, e ninguém de fora entra sem chave. O WAL vem do `store.Open`, como no banco de verdade: com o journal padrão, cada ajuste da aba dev levava ~0,4 s no Windows. | [main.go:34](../main.go#L34) |
| Tamanho da aba dev | cabe em 80 colunas: apelido em 10, cada coluna em 6 e a cor como um bloco pintado; a lista rola na altura do terminal, com o cursor sempre à vista | É o menor terminal que o app suporta. O nome da cor ("turquesa") não cabia em 6. | [tui/dev.go](../internal/tui/dev.go) |


### Editor dev inline

- Energia exibida: 0 a 100, passo 1 por seta; número digitado aceita o valor exato. Preserva o streak e o cronômetro, sem bônus oculto na energia escolhida.
- Estágio da obra: 0 a 3, ou -1 para usar novamente o progresso real. O estágio salvo é o novo ponto de partida, com deslocamento diário de minutos; tarefas não são inventadas.
- Tarefa nova: os mesmos limites do fluxo normal, 5 a 120 min (1 min também na demo); setas ajustam a duração em passos de 5 min. Tempos de tarefas existentes só mudam na aba dev.
- Conta removida da sala permanece disponível para adicionar novamente; tamanho máximo e nomes válidos continuam seguindo as regras normais.
- Salvamento inline é único para todas as alterações; Enter em um campo não grava o banco. Ver [ADR 0011](adr/0011-modo-dev.md#editor-na-própria-tela).
