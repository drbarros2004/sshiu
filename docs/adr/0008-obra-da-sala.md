# A obra da sala

O pet recompensa quem foca, um por um. A sala não tinha recompensa nenhuma: ela mostra quem está focando agora e guarda os eventos, mas o dia de todo mundo não somava em lugar nenhum. A **obra** é essa soma. Cada membro tem uma **peça** que cresce com os minutos que a pessoa focou hoje, e as peças lado a lado formam um desenho só — um trem em que cada um constrói o seu vagão, uma cidade em que cada um levanta a sua casa. Sozinha, uma peça é pouca coisa; o que fica bonito é o conjunto, e é por isso que ela puxa a turma.

A alternativa era o **leaderboard** da [#16](https://github.com/drbarros2004/ssh-focus-app/issues/16), um ranking da sala. Ranking ordena as pessoas umas contra as outras, e numa turma que estuda junto isso premia quem já estuda muito e afunda quem está voltando. A obra premia o mesmo esforço sem comparar ninguém: quem focou dez minutos adiantou a peça dele, e o desenho da sala andou junto.

## As regras

- **Os minutos são da pessoa, não da sala.** Valem os minutos focados hoje, em qualquer sala, inclusive sozinho. Cada sala divide esses minutos pela **própria meta**: com 30 min focados, a peça está na metade numa sala de meta 1 h e em um quarto numa de meta 2 h.
- **A meta é da sala**, escolhida pelo dono ([ADR 0007](0007-dono-da-sala.md)). O padrão é 1 h.
- **A obra zera todo dia.** Sem zerar, ela acabaria na primeira semana e não teria mais graça; zerando, a sala tem a obra de hoje e um motivo para voltar amanhã.
- **Quatro estágios por peça**, contando o vazio, nas frações iguais da meta: numa sala de 1 h, a peça muda aos 20, aos 40 e aos 60 minutos. São quatro desenhos por tema, não quatro por pessoa.
- **A peça é pequena, e leva o nome de quem a construiu.** O pet tem 16 colunas, e dez peças desse tamanho dariam 160: não cabe em terminal nenhum. A peça fica perto de 8×4 células, com o apelido embaixo do desenho — é o nome que deixa a sala ver quem focou hoje e quem não focou. A coluna acompanha o apelido mais comprido da sala, até o teto de 16 caracteres do apelido ([`ui.go`](../../internal/tui/ui.go)).
- **O arranjo é do tema, não da tela.** O trem é uma fileira que quebra de linha quando não cabe na largura, como os cards. A cidade tem um arranjo escrito na mão para cada quantidade de gente: 2×5 com dez, 3×3 com nove, 2×4 com oito, 2-3-2 com sete, 2×3 com seis, 3-2 com cinco, 2×2 com quatro, 2-1 com três, lado a lado com duas. Nada de fórmula: a tabela é curta e é ela que faz o conjunto parecer um lugar em vez de uma lista.
- **A obra diz a meta.** Uma linha acima do desenho com o tema e quantos minutos de foco completam uma peça, porque sem ela o desenho não explica o que falta para crescer.
- **A sala solo não tem obra.** Uma peça sozinha não é um conjunto, e a tela da sala solo é outra ([ADR 0006](0006-sala-solo.md)).
- **A obra começa desligada.** O dono escolhe o tema para ligar, e escolher "nenhuma" desliga.
- **Só entram temas que têm desenho.** A lista de temas é a dos que têm os quatro estágios prontos; oferecer um tema sem arte é oferecer uma sala vazia. Hoje são o trem e a cidade, e as outras ideias esperam em [`IDEIAS.md`](../../IDEIAS.md).

## Consequências

- **A arte é por peça.** Cada tema custa quatro desenhos pequenos, não quatro por membro, então um tema novo é barato.
- **Focar sozinho constrói na sala da turma.** Os minutos são seus; a sala só mede o que você já fez. Quem passou a tarde focando numa sala solo chega na sala da turma com a peça adiantada.
- **A obra é a primeira coisa a sumir quando falta altura**, e isso custa caro: ela só aparece quando a tela inteira cabe com ela. A tela da sala encolhe cortando eventos e depois achatando os pets ([`room.go`](../../internal/tui/room.go)), mas a obra sai antes dos dois, inteira. Numa sala de seis pessoas a 120 colunas isso quer dizer uma janela de umas 50 linhas; abaixo disso a obra não aparece. É alto, e a razão é que os cards já ocupam 14 linhas: cards e obra mostram a mesma gente e disputam a mesma tela. Se a obra ficar invisível demais na prática, o caminho é deixar ela sobreviver ao encolhimento do feed e dos pets, em vez de sair primeiro.
- **A obra não é histórico.** Ela mostra o dia de hoje, e some à meia-noite. Quem quiser os números exatos tem o perfil e, um dia, as stats da sala.
- **O leaderboard fica em aberto.** A obra ocupa o lugar que ele ocuparia na tela da sala; se um ranking ainda fizer falta, ele volta a ser discutido na #16.
- Uma tarefa que começa num dia e termina no outro conta inteira no dia em que terminar, porque é o pet que guarda os minutos do dia ([`internal/pet`](../../internal/pet/mood.go)).
