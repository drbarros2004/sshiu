# Sala solo: uma pessoa, sem código, com tela própria

Dá para criar uma sala de uma pessoa só, para focar sozinho. Antes dela, quem não tinha com quem focar criava uma sala e via uma grade com um card, um feed vazio e um código para passar para ninguém — a tela mais vazia do app, logo depois de escolher o apelido. Focar sozinho é o caso de quem acabou de chegar e de quem estuda fora do horário da turma, e o pet precisa dos minutos focados do mesmo jeito.

A sala solo é uma sala de verdade: mesmo banco, mesma tarefa, mesma energia, mesmo histórico. O que muda é que ela não aceita ninguém, então não tem código, e que a tela não gasta espaço com o que uma pessoa não usa. O [ADR 0003](0003-so-salas-privadas.md) continua valendo para todas as outras: o app é para uma turma, e é a sala que mostra quem está focando.

## Consequências

- O tamanho 1 entra na escolha da criação, de 1 a 10 ([regras](../regras.md#na-sala)). Escolher 1 é a única coisa que faz uma sala solo.
- A sala solo não mostra código em lugar nenhum, e ninguém entra nela: o teto de membros já recusa, porque o primeiro membro a enche.
- A tela dela é o layout A, escolhido em 30/10/2026 entre três candidatos: painel à esquerda (relógio, tarefa, pet, identidade e um mini log de três linhas) e, à direita, o modo jardim enquanto a tarefa corre e os seus números (calendário do mês, semana, totais) quando não há tarefa. Começar a tarefa muda a coluna, e quem está parado não paga o custo do jardim.
- Saem dela as reações, o cursor de eventos, o contador de reações recebidas e a grade de cards: nada disso tem sentido com uma pessoa.
- O modo jardim em tela cheia continua na tecla `j`: na sala solo ele é a segunda versão da mesma tela, não outra tela.
- Abaixo de umas 100 colunas o jardim sai e sobra só o painel, que em qualquer largura tem o pet com o nome ao lado (ver "Telas de tamanhos diferentes" no [`IDEIAS.md`](../../IDEIAS.md)).
- Focar sozinho vale igual: mesmos minutos, mesmo bônus por tarefa cumprida, mesmo streak. Sem penalidade e sem bônus por estar só.
