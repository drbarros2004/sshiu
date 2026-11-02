# sshiu

> O nome tem o ssh dentro e o "shhh" de quem pede silêncio numa biblioteca. "Pet" ainda é um nome provisório.

App de foco acessado por SSH, para uma turma estudar junto: cada usuário diz a próxima coisa concreta que vai fazer, foca por um tempo e cuida de um pet que reflete o próprio foco, à vista da sala.

## Linguagem

### Gente e salas

**Usuário**:
Quem usa o app, identificado pela chave SSH com que se conecta.
_Evite_: pessoa, conta, login

**Apelido**:
O nome único e público de um usuário, escolhido no primeiro acesso.
_Evite_: nick, username

**Sala**:
Um grupo fechado de usuários que focam e se veem juntos. Toda sala é privada.
_Evite_: canal, grupo, servidor, hub

**Código**:
O convite de uma sala: quem tem o código entra, sem aprovação, enquanto houver vaga e a sala estiver aberta. Toda sala tem um; a sala solo esconde o seu enquanto for de uma pessoa.
_Evite_: senha, link

**Membro**:
Um usuário que entrou numa sala e ainda não saiu. A sala deixa de existir quando o último sai, ou quando o dono a encerra.
_Evite_: participante

**Dono da sala**:
Quem criou a sala, ou o membro para quem ela foi passada. É quem muda o nome e o tamanho, fecha a sala para novas entradas, remove um membro e encerra a sala ([ADR 0007](docs/adr/0007-dono-da-sala.md)). Nenhum outro membro tem poderes.
_Evite_: admin, moderador, host, criador

**Sala fechada**:
Uma sala que não aceita mais ninguém, mesmo com vaga e com o código na mão. O dono abre e fecha quando quiser.
_Evite_: sala trancada, privada

**Sala solo**:
Uma sala de uma pessoa, para focar sozinho. Ninguém entra nela e o código fica escondido; a tela é só o painel de foco e o modo jardim ([ADR 0006](docs/adr/0006-sala-solo.md)). Ela deixa de ser solo se o dono aumentar o tamanho.
_Evite_: sala privada, modo solo

**Sala demo**:
A sala com bots que existe só no modo de demonstração, para mostrar o app cheio.

### Foco

**Tarefa**:
A próxima coisa concreta que o usuário vai fazer, com um tempo declarado. Concreta: "fazer o exercício 1 da lista 3", não "estudar Cálculo".
_Evite_: Lock, meta, objetivo, sprint

**Tempo declarado**:
Quanto o usuário diz que a tarefa vai levar, escolhido por ele ao começar.
_Evite_: duração do sprint

**Foco**:
O período em que o usuário está com uma tarefa em andamento.
_Evite_: sessão, pomodoro

**Minutos focados**:
O tempo que de fato contou numa tarefa. Para de contar quando a tarefa acaba, é parada ou quando o tempo declarado esgota e ninguém responde.

**Feito**:
Marcar a tarefa como cumprida, a qualquer momento, inclusive antes do tempo declarado.
_Evite_: check-out, concluir

**Continuar / Trocar / Parar**:
As escolhas quando o tempo declarado esgota: dar mais tempo à mesma tarefa, encerrá-la como não cumprida e escrever outra, ou encerrar o foco.

**Resultado**:
Como uma tarefa terminou: cumprida, não cumprida (trocada ou parada) ou interrompida.

**Interrompida**:
Tarefa encerrada porque o usuário desconectou e não voltou a tempo. Não é punição: equivale a parar.
_Evite_: abandono, falta, desistência

**Modo jardim**:
A tela de descanso do foco: um wallpaper animado com, se o usuário quiser, o relógio, a tarefa e quem mais está focando por cima. Ligada por tecla ou, se o usuário deixar, depois de um tempo sem tecla.
_Evite_: screensaver, modo zen

### O pet

**Pet**:
A criatura de cada usuário, cujo humor reflete o foco recente dele.
_Evite_: Broto, bonequinho, criatura, mascote

**Energia**:
Um número de 0 a 100 que sobe com minutos focados e tarefas cumpridas e decai a cada dia parado.

**Humor**:
O estado visível do pet, derivado da energia: murcho, cansado, normal, bem ou radiante.
_Evite_: estado, mood

**Cor do pet**:
O tom que cada usuário escolhe para a sua criatura, logo depois do apelido e depois disso na aba
de perfil. A cor é da pessoa e o humor continua aparecendo no quanto ela é viva: murcho fica
apagado e radiante fica vivo, os dois no tom escolhido. Quem não escolhe fica com a primeira cor
da lista.
_Evite_: skin, tema do pet

**Streak**:
Dias seguidos em que o usuário focou.
_Evite_: sequência, ofensiva

### Na sala

**Status**:
O que um membro está fazendo agora, visto pela sala: focando (com tarefa e tempo), sem tarefa ou offline.
_Evite_: presença

**Evento**:
Algo que aconteceu na sala e aparece para todos: alguém entrou ou saiu, começou ou cumpriu uma tarefa, bateu um streak.
_Evite_: notificação, post

**Reação**:
Uma resposta rápida, de um conjunto fixo e pequeno, a um evento. Não existe chat.
_Evite_: curtida, like, mensagem

**Obra**:
O desenho que a sala constrói junto no dia: uma peça por membro, lado a lado ([ADR 0008](docs/adr/0008-obra-da-sala.md)). Zera todo dia e a sala solo não tem.
_Evite_: projeto, conquista, missão

**Peça**:
A sua parte da obra — o vagão, o vaso, a casa. Cresce com os seus minutos focados no dia.
_Evite_: item, pedaço, bloco

**Meta da sala**:
Quantos minutos focados completam uma peça. O dono escolhe; o padrão é 1 h.
_Evite_: objetivo, cota
