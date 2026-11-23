# A tabela de foco, do dono

O [ADR 0008](0008-obra-da-sala.md) trocou o leaderboard da [#16](https://github.com/drbarros2004/sshiu/issues/16) pela obra, porque um ranking ordena as pessoas umas contra as outras e numa turma isso premia quem já estuda muito e afunda quem está voltando. O argumento continua de pé para uma turma. O que mudou foi o uso: a ideia de um colega é usar o sshiu numa empresa ou num grupo para marcar o tempo de tarefas pequenas, e aí quem coordena precisa ver quanto cada um trabalhou sem contar peça por peça na obra ([#83](https://github.com/drbarros2004/sshiu/issues/83)). A **tabela de foco** é isso: os minutos focados de cada membro, hoje, na semana, no mês ou no total, no menu do dono.

## As regras

- **É do dono.** A tabela fica no menu `c`, numa página própria, e na tecla `o` do dono, junto da obra. Os outros membros não a veem até o dono abrir.
- **Abrir é escolha do dono.** "Só você" ou "a sala toda", numa página do menu `c` junto com o formato e o período. Aberta, ela aparece para todo mundo em dois lugares: na tela da sala, numa faixa embaixo dos cards ao lado da obra, e inteira na tecla `o`, em cima da obra ou sozinha. Ela ficava compacta à direita do painel de foco, com até três pessoas e "+N no o", e parecia sempre cortada; a faixa é do [ADR 0012](0012-arranjo-da-sala.md). Fechada, ela fica só no menu e na tecla `o` do dono.
- **A tabela vem junto da obra.** Na sala, as duas dividem a faixa e saem juntas quando falta altura ([ADR 0012](0012-arranjo-da-sala.md)); antes, a obra saía com a tabela ainda na tela. Os eventos nunca somem por causa da tabela (com ela, ficam pelo menos dois). Na tecla `o`, se as duas não cabem, a tabela fica e a obra dá lugar a um aviso.
- **Quem não vê a tabela não ouve falar dela.** A legenda da tecla `o` diz "obra", ou "tabela" numa sala sem obra, e some para quem não tem o que ver ali.
- **Tabela ou ranking, também do dono.** A tabela mostra o tempo exato (h:mm:ss) e o que cada um está fazendo agora, na ordem em que os membros entraram, para as linhas não pularem a cada tick. O ranking ordena do mais focado para o menos, com uma barra e a posição. O formato vale para o dono e para quem vê.
- **O período é de quem olha.** Hoje, esta semana (de segunda), este mês ou o total, com ←/→. Cada pessoa escolhe o seu, e a escolha não vai para o banco: é como rolar uma página, não uma configuração da sala. Começa em hoje.
- **Começa fechada e em tabela.** A sala que não mexe nisso fica como era antes: a obra sem comparação ([ADR 0008](0008-obra-da-sala.md)). O ranking existe, mas alguém escolhe ligá-lo.
- **No ranking, quem não focou não tem posição.** Empate divide a posição (1º, 1º, 3º); quem ainda está em zero aparece sem número, para o ranking não pôr ninguém em último por estar começando.
- **Os minutos são os da pessoa, em qualquer sala**, e contam no dia em que a tarefa termina, como o pet: os dias anteriores vêm do banco (`store.FocusBetween`) e hoje vem do retrato, com a tarefa em andamento. O banco só é relido ao abrir, ao trocar o período e na virada do dia.
- **A sala solo não tem.** Uma linha só não é tabela, e o perfil já diz quanto você focou.

## Consequências

- O dono ganha dois poderes novos ([ADR 0007](0007-dono-da-sala.md)): abrir a tabela para a sala e trocar a tabela pelo ranking. As duas escolhas ficam no banco (`rooms.board_open`, `rooms.board_rank`).
- **Fica para depois o resto da #83:** o log de todas as tarefas de todos, com filtro por pessoa, e o administrador de uma instituição acima das salas. O primeiro sai da tabela `tasks`, que guarda a sala de cada tarefa; o segundo é um modelo novo de gente, e entra em [`IDEIAS.md`](../../IDEIAS.md).

## Alternativas consideradas

- **Só o ranking, sempre aberto** (a #16 como ela foi escrita): é o que o ADR 0008 recusou, e com motivo. Aqui ele só aparece quando o dono escolhe.
- **Os minutos só desta sala** (as tarefas que começaram nela): parece mais justo para quem coordena, mas desmente a obra, que conta os minutos da pessoa em qualquer sala, e duas telas da mesma sala com números diferentes viram uma pergunta que ninguém sabe responder.
- **O período escolhido pelo dono, para todos:** quem olha a tabela quer comparar hoje com a semana sem pedir para ninguém; o que é do dono é o que expõe as pessoas (abrir e ranquear), não a janela de tempo.
