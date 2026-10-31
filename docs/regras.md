# Regras numéricas

Todo número que decide como o app se comporta fica listado aqui, com o porquê e o link para onde está no código. Serve para auditar as regras sem ler o código e como freio: se uma regra não cabe aqui de forma simples, ela complicou antes da hora.

Ao mudar um número no código, mude aqui também.

## Energia e humor do pet

Tudo em [`internal/pet/mood.go`](../internal/pet/mood.go). A energia vem dos minutos focados de verdade, não do tempo declarado: declarar 60 min e marcar "feito" no minuto 1 não rende.

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Energia inicial | 50 (de 0 a 100) | Começa no meio: humor "normal". | [mood.go:19](../internal/pet/mood.go#L19) |
| Decaimento | −8 por dia inteiro sem focar | Sumir é o que pune. | [mood.go:20](../internal/pet/mood.go#L20) |
| Energia por minuto focado | 8 a cada 30 min (≈ 0,27 por minuto) | 30 min de foco devolvem o que um dia parado tira. | [mood.go:27](../internal/pet/mood.go#L27) |
| Bônus por tarefa cumprida | +3, só a partir de 10 min focados | Vale uns 10 min de foco: recompensa clara sem fazer várias tarefas curtas renderem mais que uma longa. Uma tarefa de 5 min rende só os minutos, para o bônus não virar máquina de pontos. | [mood.go:28-29](../internal/pet/mood.go#L28-L29) |
| Streak | o dia conta a partir de 15 min focados | Pouco o bastante para um dia corrido, mais que abrir e fechar o app. | [mood.go:30](../internal/pet/mood.go#L30) |
| Bônus de streak | +2 por dia seguido, até +10 | Recompensa a constância sem deixar o streak dominar a energia. | [mood.go:21-22](../internal/pet/mood.go#L21-L22) |
| Faixas de humor | murcho < 20 ≤ cansado < 40 ≤ normal < 60 ≤ bem < 80 ≤ radiante | Cinco faixas iguais. | [mood.go:67](../internal/pet/mood.go#L67) |

```go
var moodMin = [...]float64{Wilted: 0, Tired: 20, Okay: 40, Good: 60, Radiant: 80}
```

A energia mostrada é a guardada, menos o decaimento, mais o bônus de streak, sempre entre 0 e 100 ([mood.go:89](../internal/pet/mood.go#L89)). Cada tarefa encerrada soma os seus minutos e o bônus na hora ([mood.go:34](../internal/pet/mood.go#L34)).

## A tarefa

Tudo em [`internal/rooms/room.go`](../internal/rooms/room.go), salvo onde indicado.

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Tempos declarados | 10, 15, 25, 45 ou 60 min (padrão 25); na demo, também 1 min | A plateia da demo vê o fluxo inteiro. | [tui/room.go:42-45](../internal/tui/room.go#L42-L45) |
| Tempo esgotado sem resposta | os minutos param de contar | Estar ausente nunca rende mais do que o tempo que o próprio usuário declarou. | [room.go:61](../internal/rooms/room.go#L61) |
| Trocar, parar, desconectar | sem penalidade | Os minutos já contaram; punir faria os usuários declararem tempos folgados ou deixarem o terminal aberto. | [room.go:212](../internal/rooms/room.go#L212) |
| Tolerância de reconexão | 2 min | Conexão instável não interrompe a tarefa. Sem volta, a tarefa é interrompida com os minutos até a desconexão. | [room.go:26](../internal/rooms/room.go#L26) |

## Na sala

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Tamanho da sala | de 1 a 10, escolhido na criação (padrão 10) | Uma sala é um grupo de estudo, não um auditório: todo mundo cabe na tela e o custo de desenhar a sala cresce com o quadrado do tamanho dela. Quem cria sabe o tamanho da turma; o tamanho não muda depois. O 1 é a sala solo ([ADR 0006](adr/0006-sala-solo.md)). | [room.go:24-25](../internal/rooms/room.go#L24-L25) |
| Reações | 4 (👏 🔥 💪 ☕), teclas 1 a 4; as abas ficam em a/d e tab | Uma tecla cada, sem chat ([ADR 0004](adr/0004-sem-chat-so-reacoes.md)). | [room.go:23](../internal/rooms/room.go#L23) |
| Eventos guardados | 20 por sala, 6 na tela | O bastante para reagir ao que acabou de acontecer. | [room.go:27](../internal/rooms/room.go#L27), [tui/room.go:20](../internal/tui/room.go#L20) |
| Bots da demo | tarefas de 40 a 90 s, uns 90 s parados entre uma e outra | A sala parece viva sem empurrar os seus eventos para fora da tela. | [bots.go:78-79](../internal/rooms/bots.go#L78-L79) |

## Ritmo da tela

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Tick da tela | 500 ms | Anima o pet e o relógio e atualiza a sala. | [ui.go:103](../internal/tui/ui.go#L103) |
| Respiração do pet | 1 a 4 ticks por quadro (radiante rápido, murcho lento) | O cansaço aparece no ritmo, não só no desenho. | [render.go:89](../internal/pet/render.go#L89) |
| Modo jardim | liga depois de 30 s sem tecla, só durante o foco | | [garden.go:26](../internal/tui/garden.go#L26) |
| Jardim na sala solo | ocupa a direita a partir de 96 colunas de terminal | Menos que isso e o painel fica apertado; aí a tela é só o painel. | [room.go:21-22](../internal/tui/room.go#L21-L22) |
| Quadros do modo jardim | 8 por segundo | Movimento suave. Medido no protótipo: de 2 a 24 KB/s e de 4 a 6% de um núcleo por usuário; 5 fps gasta uns 40% menos CPU. | [garden.go:25](../internal/tui/garden.go#L25) |
| Padrões do modo jardim | wallpaper sorteado a cada tarefa, relógio só com os minutos, tarefa e sala aparecem | Quem chega vê o que o modo jardim oferece e desliga o que achar demais. | [db.go:78](../internal/store/db.go#L78) |
