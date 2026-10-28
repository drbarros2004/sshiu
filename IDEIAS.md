# Ideias de expansão do Plantão

> **Plantão é um nome provisório (WIP).** O nome oficial do app ainda vai ser decidido com o grupo.

Backlog de ideias, sem compromisso de ordem (exceto onde indicado). Cada item diz o que é, por que importa e o que já se sabe sobre como fazer.
Conceito e arquitetura originais: [doc "Plantão: infraestrutura e como evitar o vazio"](https://claude.ai/artifact/24PH3QwV6B6enMrD5Vv6x1).

## Próximas candidatas

### Login pelo GitHub + commits na sala
- **O quê:** vincular a conta do GitHub e, opcionalmente, mostrar na sala os commits feitos durante o sprint ("gabriel: fix: timer da atividade").
- **Por quê:** facilita o onboarding (a chave SSH já está no GitHub), dá identidade real e mostra movimento na sala sem câmera.
- **Como:** o servidor baixa `https://github.com/<usuario>.keys` e confere se a chave da conexão está na lista; isso prova a posse da conta sem senha nem OAuth. Os commits públicos vêm da API de eventos (`/users/<u>/events/public`, PushEvent).
- **Cuidados:**
  - Sem token, a API aceita 60 requisições por hora, então o servidor precisa de um token próprio.
  - Repositórios privados exigem OAuth (o fluxo de código de dispositivo funciona no terminal).
  - Mostrar commits deve ser opt-in.
  - Substitui o "hook de git" do doc original.

### Modo jardim (tela de descanso durante o foco)
- **O quê:** depois de alguns segundos sem tecla durante o foco, a sala vira uma cena calma em tela cheia, com os pets da sala, o relógio grande e uma animação lenta. Qualquer tecla volta.
- **Por quê:** reduz a distração e reforça o ritual. É o "wallpaper" do foco.

### Gerenciar salas privadas
- Sair de uma sala, apagar (só o dono), trocar o código, remover um membro.
- **Hoje:** quem tem o código entra e ninguém sai. Precisa existir antes de abrir para gente de fora da turma.

### Menu e histórico mais ricos
- Uma aba de histórico com taxa de Locks cumpridos, streak por semana e filtro por sala.
- A tabela `sprints` ainda não guarda a sala: adicionar a coluna quando isso for feito.

## Mais adiante

### Cliente web
- A mesma lógica do servidor exposta por WebSocket, para quem não usa terminal (e como vitrine). Já estava no doc original.
- Reduz a barreira da chave SSH, que hoje é o maior atrito para quem não é dev.

### Atualização em tempo real
- **Hoje** cada sessão lê um retrato de todas as salas a cada 500 ms (polling).
- **Upgrade:** o padrão do exemplo `realtime` do Bubble Tea, com um `Cmd` esperando num canal. Só vale quando o polling ficar lento ou caro.

### Cores completas (truecolor)
- O ssh não repassa `COLORTERM`, então o servidor reduz as cores para 256.
- **Opções:**
  - Aceitar `ssh -o SetEnv=COLORTERM=truecolor` (não testado).
  - Perguntar ao terminal via `tea.RequestCapability`.

## Por último: app de celular

> Pensar com calma e fazer por último. **Não** moldar o desenvolvimento atual em função disso (otimização prematura).

- **Ideia:** facilitar o "te fazer voltar":
  - Notificações ("sprint das 14h em 5 min, 3 amigos na fila").
  - Sincronizar com a sessão de foco: ao começar o sprint, o celular entra em modo foco ou bloqueia apps, como algo conjunto da sala.
- **Restrição:** não complicar. O núcleo continua sendo o terminal.
- **O que já ajuda sem esforço extra:** o servidor já concentra tudo (salas, presença, horários); um app seria mais um cliente, como o web.

## Dívida técnica conhecida

Também marcada no código com comentários `ponytail:`.

- **Migração ingênua:** o `ALTER TABLE` ignora o erro de "coluna já existe". Trocar por uma ferramenta de migração quando houver várias.
- **Contornos para servidor no Windows** (`lfWriter` e perfil de cores em `main.go`): só existem porque o Wish/Bubble Tea assumem Unix. No deploy em Linux, nem são usados.
- **Desconectar conta como falta ou abandono:** é duro para conexões instáveis. Considerar uma tolerância de reconexão.
