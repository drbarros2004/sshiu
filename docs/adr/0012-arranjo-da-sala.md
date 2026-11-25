# O arranjo da sala: os cards em cima, a obra e a tabela numa faixa embaixo

A tela da sala tem quatro coisas disputando a altura: o painel de foco, os cards, a obra ([ADR 0008](0008-obra-da-sala.md)) e a tabela de foco ([ADR 0010](0010-tabela-de-foco.md)), mais os eventos. Um card tem 18 linhas, e numa sala de 8 pessoas em 120 a 160 colunas os cards ficam em duas fileiras: só eles e o painel pedem umas 46 linhas. A escada antiga sacrificava os pets primeiro (viravam um ponto da cor do pet) para a obra caber, então com obra os cards cheios só apareciam a partir de umas 60 linhas. E a tabela ficava compacta no alto, à direita do painel, com no máximo três pessoas e "+N no o": parecia sempre cortada ([#117](https://github.com/drbarros2004/sshiu/issues/117), [#121](https://github.com/drbarros2004/sshiu/issues/121)).

Decidimos manter tudo na mesma tela, com outro arranjo e outra ordem de encolher, escolhidos num protótipo com a sala demo em 120×40, 160×48, 200×50 e 330×60.

## As regras

- **A obra e a tabela ficam numa faixa embaixo dos cards**, lado a lado e alinhadas pelo topo. A tabela aparece inteira, com todo mundo, sem o que cada um faz agora (os cards já dizem). Se as duas não cabem na largura, a tabela desce para baixo da obra.
- **Com largura, os eventos entram na mesma faixa**, à esquerda: eventos, obra e tabela, nessa ordem. Os eventos ocupam só a largura que usam (até 80 colunas), e a faixa encosta na esquerda; a sobra fica à direita, em vez de um vão no meio numa tela larga.
- **Entre o card cheio e o ponto existe o card médio**: o mesmo pet, sem o respiro em volta, com a tarefa numa linha e as reações na linha do status quando cabem. São 13 linhas em vez de 18. O pet não encolhe: o desenho só aumenta de escala, e um pet menor pediria arte nova.
- **As fileiras de cards se dividem por igual.** Nove pessoas numa tela de oito cards viram 5 e 4, não 8 e 1; dez numa tela de quatro viram 4, 3 e 3, não 4, 4 e 2.
- **A obra vem em duas fileiras**, em todos os temas: até três peças numa fileira, e daí em diante duas, com a sobra em cima (nove viram 5 e 4). Duas fileiras deixam a obra baixa e larga, do jeito que cabe ao lado da tabela; as tabelas escritas à mão da cidade (3×3, 2-3-2) e do acampamento saíram. No acampamento, a fogueira fica no meio da fileira de baixo, que leva um número par de barracas para ela ficar mesmo no meio (sete viram 3 e 4). Quando nem duas fileiras cabem na largura, a obra volta à fileira que quebra.
- **A ordem de encolher** (dentro de cada passo, os eventos encurtam primeiro):
  1. cards cheios, com a obra no arranjo do tema e a tabela inteira;
  2. a obra vira uma fileira;
  3. a tabela é cortada na altura da obra, com o resto no título ("+N no o");
  4. os cards viram médios, e a faixa passa de novo pelos três tamanhos;
  5. a faixa sai inteira, e fica uma linha no lugar: "obra: 2 de 8 peças completas · tabela de foco · na tecla o"; sem ela, os cards voltam a cheios se couberem, e depois viram médios de novo;
  6. os pets viram ponto;
  7. as fileiras que não cabem viram "+N" ([#119](https://github.com/drbarros2004/sshiu/issues/119)).
- **Com a faixa na tela, ficam pelo menos dois eventos**, como a tabela já garantia ([ADR 0010](0010-tabela-de-foco.md)). Os eventos só somem com a faixa fora e os pets já encolhidos.

## Alternativas consideradas

- **Uma aba só para a obra e a tabela** ([#117](https://github.com/drbarros2004/sshiu/issues/117)): resolvia o espaço escondendo a obra de todo mundo, em qualquer tamanho de tela. A obra é a recompensa da turma e funciona no canto do olho ([ADR 0008](0008-obra-da-sala.md)); e a tela própria já existe, na tecla `o`.
- **A faixa sai antes de os cards virarem médios**: deixa os cards cheios por mais tempo, mas a obra some antes, e os rostos do card médio já dizem quem é quem.
- **A escada antiga** (os pets viram ponto antes de a obra sair): mantinha a obra numa janela média, mas a sala virava uma fileira de pontos, que é a parte que a turma mais olha.

## Consequências

- **Numa janela média com sala cheia, a obra fica na tecla `o`.** Com 8 pessoas em 120 a 160 colunas, a tela mostra os cards médios e a linha que aponta a tecla `o`; a partir de umas 200 colunas, ou com menos gente, a faixa volta. A linha diz quantas peças estão completas, para a obra não sumir sem notícia.
- Substitui a ordem de encolher do [ADR 0008](0008-obra-da-sala.md) ("a obra sai por último") e a tabela compacta ao lado do painel do [ADR 0010](0010-tabela-de-foco.md). A tabela continua vindo junto da obra: as duas saem juntas, na faixa.
- A tecla `o` não muda: a obra e a tabela inteiras, em qualquer tamanho.
