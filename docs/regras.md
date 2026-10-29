# Regras numéricas

Todo número que decide como o app se comporta fica listado aqui, com o porquê e o link para onde está no código. Serve para auditar as regras sem ler o código e como freio: se uma regra não cabe aqui de forma simples, ela complicou antes da hora.

Ao mudar um número no código, mude aqui também.

## Em vigor

### Energia e humor do pet

Tudo em [`internal/pet/mood.go`](../internal/pet/mood.go).

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Energia inicial | 50 (de 0 a 100) | Começa no meio: humor "normal". | [mood.go:19](../internal/pet/mood.go#L19) |
| Decaimento | −8 por dia inteiro sem focar | Sumir é o que pune. | [mood.go:20](../internal/pet/mood.go#L20) |
| Bônus de streak | +2 por dia seguido, até +10 | Recompensa a constância sem deixar o streak dominar a energia. | [mood.go:21-22](../internal/pet/mood.go#L21-L22) |
| Faixas de humor | murcho < 20 ≤ cansado < 40 ≤ normal < 60 ≤ bem < 80 ≤ radiante | Cinco faixas iguais. | [mood.go:37](../internal/pet/mood.go#L37) |

```go
var moodMin = [...]float64{Wilted: 0, Tired: 20, Okay: 40, Good: 60, Radiant: 80}
```

A energia mostrada é a guardada, menos o decaimento, mais o bônus de streak, sempre entre 0 e 100 ([mood.go:56-58](../internal/pet/mood.go#L56-L58)).

### Ritmo da tela

| Regra | Valor | Por quê | Código |
|---|---|---|---|
| Tick da tela | 500 ms | Anima o pet e o relógio e atualiza a sala. | [ui.go:107](../internal/tui/ui.go#L107) |
| Respiração do pet | 1 a 4 ticks por quadro (radiante rápido, murcho lento) | O cansaço aparece no ritmo, não só no desenho. | [render.go:89](../internal/pet/render.go#L89) |

## Em vigor, mas vai sair

Estas regras pertencem ao modelo de sprint em horário marcado, substituído pelo foco individual ([ADR 0002](adr/0002-foco-individual-sem-horario-marcado.md)).

| Regra | Valor | Código |
|---|---|---|
| Resultado do sprint | cumpriu +15, não cumpriu +4, abandono −15, falta −10 | [mood.go:25](../internal/pet/mood.go#L25) |
| Sprint | 25 min de foco + 5 de intervalo, nos horários da sala | [room.go:47](../internal/rooms/room.go#L47) |
| Sprint na demo | 1 min de foco, 40 s de intervalo, próximo 20 s depois | [room.go:45](../internal/rooms/room.go#L45) |

## Decidido, ainda não implementado

| Regra | Valor | Por quê |
|---|---|---|
| Energia por minuto focado | a definir | A energia vem do tempo focado de verdade, não do tempo declarado: declarar 60 min e marcar "feito" no minuto 1 não rende. |
| Bônus por tarefa cumprida | a definir, só a partir de 10 min focados | Terminar cedo é bom e ganha o bônus; uma tarefa de 5 min rende só os minutos, para o bônus não virar máquina de pontos. |
| Tempo esgotado sem resposta | os minutos param de contar | Estar ausente nunca rende mais do que o tempo que o próprio usuário declarou. |
| Trocar, parar, desconectar | sem penalidade | Os minutos já contaram; punir faria os usuários declararem tempos folgados ou deixarem o terminal aberto. |
| Tolerância de reconexão | 2 min | Conexão instável não interrompe a tarefa. |
| Streak | a definir quanto foco faz o dia contar | |
| Modo jardim | liga depois de 30 s sem tecla, só durante o foco | |
| Reações | umas 4, uma tecla cada | |
| Tarefa na demo | cerca de 1 min | Dá para a plateia entender o fluxo. |
