# A sala tem dono

O [ADR 0003](0003-so-salas-privadas.md) deixou a sala sem ninguém no comando: quem criava não ganhava poder nenhum, e remover um membro, passar a sala e trocar o código ficaram "para depois". O depois chegou com o uso de verdade. Uma sala tem tamanho, e vaga ocupada por quem entrou e não voltou é vaga que falta para quem ia focar; um nome digitado errado na criação ficava errado para sempre; e uma sala em que a turma já está toda dentro não tem como parar de aceitar gente. Quem criou a sala passa a ser o **dono** e é a única pessoa que muda essas coisas.

## O que o dono faz

Trocar o nome, mudar o tamanho, fechar a sala para novas entradas, remover um membro, passar a sala para outro membro e encerrar a sala.

## Consequências

- **O dono não é moderador.** Remover um membro abre a vaga, e só: quem tem o código entra de novo. O que isso resolve é lugar ocupado, não desavença — para desavença a sala é pequena e a turma se conhece.
- **A sala nunca fica órfã.** Quando o dono sai, a sala passa para o membro mais antigo que ficou. Sair é sair da sala: ficar offline não passa a sala para ninguém, porque quem fecha o terminal volta.
- **O tamanho não desce abaixo de quem já está na sala.** Ninguém é expulso por uma mudança de tamanho: para abrir espaço na lista, o caminho é remover um membro, que é uma escolha com nome e sobrenome.
- **A sala demo não tem dono**, então o menu nem aparece nela.
- **A sala solo nasce fechada para novas entradas.** Ela tem código como qualquer outra; aumentar o tamanho faz o código aparecer, e quem abre a porta é o dono, num segundo passo ([ADR 0006](0006-sala-solo.md)).
- **Trocar o código continua fora.** O código é a chave da sala no banco, e mexer nele arrasta as tarefas que guardam onde aconteceram.
- O banco já guardava quem criou cada sala (`rooms.owner`), então a mudança é de poderes, não de modelo.

## Alternativas consideradas

- **Remover com bloqueio** (o removido não volta nem com o código): resolve desavença, mas exige uma lista de bloqueados e transforma o dono em moderador de uma sala de dez pessoas que se conhecem.
- **O dono escolher o layout ou o tema da sala para todo mundo:** cada pessoa usa um terminal de tamanho diferente, e a tela já se adapta a cada um; um layout escolhido por quem tem a tela grande quebra na tela pequena dos outros. O tema continua sendo de cada usuário, na aba visual.
- **Um recado fixo da sala, escrito pelo dono:** é a porta dos fundos do chat que o [ADR 0004](0004-sem-chat-so-reacoes.md) recusou.
