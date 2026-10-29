# Ideias de expansão do Plantão

> **Plantão é um nome provisório (WIP).** O nome oficial do app ainda vai ser decidido com o grupo.

Backlog de ideias, sem compromisso de ordem (exceto onde indicado). Cada item diz o que é, por que importa e o que já se sabe sobre como fazer.
Conceito e arquitetura originais: [doc "Plantão: infraestrutura e como evitar o vazio"](https://claude.ai/artifact/24PH3QwV6B6enMrD5Vv6x1). O vocabulário está em [`CONTEXT.md`](CONTEXT.md), as decisões em [`docs/adr/`](docs/adr/) e os números em [`docs/regras.md`](docs/regras.md).

## O modelo novo (implementado)

Decidido no grill de 28/10/2026, validado nos protótipos e implementado em 29/10/2026. Fica aqui como registro do que foi decidido e por quê.

- **Foco individual** ([ADR 0002](docs/adr/0002-foco-individual-sem-horario-marcado.md)): sai o sprint em horário marcado (fila, falta, abandono, check-out, intervalo, horários da sala). Entra a tarefa com tempo declarado, "feito" a qualquer momento, e continuar/trocar/parar quando o tempo esgota. Desconectar equivale a parar, com tolerância de reconexão.
- **Energia por minuto focado**, com bônus por tarefa cumprida ([regras](docs/regras.md#energia-e-humor-do-pet)).
- **Só salas privadas** ([ADR 0003](docs/adr/0003-so-salas-privadas.md)): sai a sala pública e a aba amigos. Primeiro acesso vira um menu (entrar com código ou criar sala); com `-demo`, a sala demo aparece no menu. A tela da sala lista todos os membros com o status de cada um. Membros podem sair; a sala some quando o último sai.
- **Sem chat, só reações** ([ADR 0004](docs/adr/0004-sem-chat-so-reacoes.md)).
- **Tela da sala: painel + cards**, escolhida no protótipo de 28/10/2026 (guardado no branch `prototype/sala`). Seu foco no topo (relógio, tarefa, "feito"), um card por membro (pet, status, tarefa e o contador de reações recebidas) e os eventos embaixo. As reações ficam numa coluna só, em ordem fixa (👏 🔥 💪 ☕ por enquanto): a que tem zero some e as seguintes andam para a esquerda. As reações ficam com as teclas 1-4, e as abas passam para a/d (esquerda/direita), além do tab.
- **"Lock" vira "tarefa"** na interface, mantendo a ideia de ação concreta no texto da pergunta.
- **Abas:** sala · perfil · visual (a aba "tema" vira "visual" e reúne tema, o modo jardim (wallpaper e o que aparece por cima) e, depois, a escolha do pet).
- **Bots da demo:** os mesmos 5, com os mesmos humores iniciais, seguindo o modelo novo (tarefas espalhadas, perfis diferentes) e reagindo a quem está na sala. Tarefas de cerca de 1 min.

### Modo jardim (tela de descanso durante o foco)
- **O quê:** um wallpaper ASCII animado em tela cheia, com o relógio e outras informações por cima, se o usuário quiser. Liga por tecla ou depois de 30 s sem tecla, só durante o foco. Qualquer tecla volta sem executar nada; o fim do foco desliga.
- **Wallpapers:** 6, escolhidos no protótipo de 29/10/2026 (branch `prototype/jardim`): céu noturno, chuva lenta e lâmpada de lava, escritos do zero com base em projetos MIT/Apache e creditados em `CREDITS.md` ([ADR 0005](docs/adr/0005-arte-de-terceiros-so-com-licenca-permissiva.md)); café, árvore com folhas caindo e estrada à noite, desenhos nossos. A neve ficou de fora. Tudo com pouco movimento e cores apagadas, sem clarões nem nada que se acumule na tela. Cada usuário escolhe o seu na aba visual; quem não escolheu ganha um sorteado a cada tarefa.
- **O que aparece por cima** é escolhido na aba visual, junto do wallpaper (no modo jardim, qualquer tecla continua saindo): o relógio (mm:ss, só os minutos ou nenhum), a tarefa e quem mais da sala está focando. O relógio só com os minutos existe porque os segundos mudando são o único movimento rápido da tela.
- **Caixa:** o texto fica sobre uma caixa vazia, que abre um buraco no wallpaper; sem ela, a lava, a árvore e o vapor passam por trás e atrapalham a leitura.
- **Padrões** de quem nunca mexeu na aba visual: wallpaper sorteado a cada tarefa, relógio só com os minutos, tarefa e sala ligadas (quem chega vê o que o modo jardim oferece e desliga o que achar demais).

## Próximas candidatas

### Login pelo GitHub + commits na sala
- **O quê:** vincular a conta do GitHub e, opcionalmente, mostrar na sala os commits feitos durante a tarefa ("gabriel: fix: timer da atividade").
- **Por quê:** junta as contas de quem usa mais de uma máquina ([ADR 0001](docs/adr/0001-chave-ssh-e-a-identidade.md)), dá identidade real e mostra movimento na sala sem câmera.
- **Como:** o servidor baixa `https://github.com/<usuario>.keys` e confere se a chave da conexão está na lista; isso prova a posse da conta sem senha nem OAuth. Os commits públicos vêm da API de eventos (`/users/<u>/events/public`, PushEvent).
- **Cuidados:**
  - Sem token, a API aceita 60 requisições por hora, então o servidor precisa de um token próprio.
  - Repositórios privados exigem OAuth (o fluxo de código de dispositivo funciona no terminal).
  - Mostrar commits deve ser opt-in.
  - Substitui o "hook de git" do doc original.

### Gerenciar salas
- Trocar o código, remover um membro, passar a sala para outro usuário. Fora do MVP: no MVP ninguém tem poderes especiais.
- Trocar o código dá algum trabalho: hoje o código é a chave da sala no banco.

### Histórico, stats e gráficos
- Uma aba de histórico com taxa de tarefas cumpridas, streak por semana e filtro por sala.
- **Gráficos por usuário:** streak, tempo focado por dia ou semana e outras agregações no tempo.
- **Gráficos da sala:** os mesmos agregados para a sala inteira, mais o pico de usuários focando ao mesmo tempo e quanto tempo esse pico durou.
- Exige guardar a sala e quando cada tarefa começou e parou, não só o resultado.

### Vários pets (10 a 15 no MVP)
- **O quê:** em vez de um pet só, cada usuário escolhe o seu entre vários, cada um com os próprios humores.
- **Por quê:** personalização. Pelo ritmo em que os sprites atuais saíram, 10 a 15 pets parece viável para o MVP.

### Leaderboard da sala
- **O quê:** um ranking dentro de cada sala.
- **Por quê:** dá um motivo coletivo para voltar, além do pet de cada um.

### Pôr no ar: `ssh <nome>`
- **O quê:** um servidor público com domínio, ouvindo na porta 22, para entrar só com `ssh <nome>`, como o `ssh terminal.shop`.
- **Como:** escutar em todas as interfaces (hoje só `localhost`, em `main.go`), porta 22, chave do servidor fixa e fora do git (publicar o fingerprint para quem quiser conferir), binário compilado com `GOOS=linux`. Quem entra só precisa de um cliente SSH e de uma chave.
- **A chave do servidor nunca muda:** o cliente SSH guarda a chave do servidor na primeira conexão. Se ela mudar (servidor recriado, `.ssh/` apagado, outra máquina no mesmo endereço), todo usuário recebe o aviso "REMOTE HOST IDENTIFICATION HAS CHANGED" e a conexão é recusada. Guardar a chave do servidor com backup e levá-la junto em qualquer migração.
- **Quando:** sem pressa, mas o grupo precisa conseguir entrar e se ver antes da apresentação. Depende do nome do app (o domínio).

### Cara de app de dev
- **O quê:** elementos visuais que dev reconhece de cara, em algum lugar do app. Exemplo: o estilo do `git diff`, com `+` em verde e `-` em vermelho.
- **Onde:** ainda não decidido.

### Termos de servidor no vocabulário
- **O quê:** nomes tirados de servidor e de Unix para coisas do app. Se entrarem, mudam também o [`CONTEXT.md`](CONTEXT.md).
  - **Streak → uptime:** "uptime de 12 dias" é como se diz que um servidor está no ar sem parar.
  - **Moeda → ciclos:** ciclos de CPU ganhos focando (ver [Moeda e loja](#moeda-e-loja)).
  - **Humor "murcho" → zumbi:** um processo zumbi já terminou, mas continua na lista porque ninguém veio buscar o resultado. É o pet largado.
- **Keepalive**, para a apresentação: o SSH manda sinais para a conexão não cair, e o foco de cada um mantém o pet vivo.

### Textos da demo mais engraçados
- **O quê:** trocar as tarefas dos bots da demo por piadas de dev, como "reescrevendo X em Rust".
- **Por quê:** a demo vai ser apresentada para uma plateia.
- **Onde:** as tarefas dos bots ficam em `internal/rooms/bots.go`.

### Tela "sobre"
- **O quê:** um FAQ com perguntas e respostas sérias, misturadas com piadinhas. Tom de exemplo: "a focus app for real developers (or vibe-coders)".

## Mais adiante

### Sessões marcadas
- Um horário opcional e configurável da sala para focar todo mundo junto, sem ser o único jeito de focar ([ADR 0002](docs/adr/0002-foco-individual-sem-horario-marcado.md)).

### Recompensas coletivas da sala
- **O quê:** além do objetivo de cada usuário (deixar o pet feliz), a sala tem metas e recompensas em conjunto. Ainda não se sabe quais.
- **Por quê:** a turma foca junto; a recompensa também pode ser junto.

### Moeda e loja
- **O quê:** uma moeda ganha focando, usada para comprar pets ou coisas para a sala.
- **Por quê:** amarra as recompensas e os vários pets num mesmo ciclo. Ainda é uma ideia solta.

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

### Trailer
- Um vídeo de até 30 s para a apresentação, feito com ferramentas de vídeo com IA. Pensar depois.

## Por último: app de celular

> Pensar com calma e fazer por último. **Não** moldar o desenvolvimento atual em função disso (otimização prematura).

- **Ideia:** facilitar o "te fazer voltar":
  - Notificações ("3 usuários da sua sala estão focando agora").
  - Sincronizar com o foco: ao começar uma tarefa, o celular entra em modo foco ou bloqueia apps.
- **Restrição:** não complicar. O núcleo continua sendo o terminal.
- **O que já ajuda sem esforço extra:** o servidor já concentra tudo (salas, status); um app seria mais um cliente, como o web.

## Dívida técnica conhecida

Também marcada no código com comentários `ponytail:`.

- **Migração ingênua:** o `ALTER TABLE` ignora o erro de "coluna já existe". Trocar por uma ferramenta de migração quando houver várias.
- **Foco e eventos só em memória:** a tarefa em andamento, os eventos e as reações vivem em `internal/rooms`; reiniciar o servidor perde as tarefas em andamento (sem gravar) e os eventos. Gravar a tarefa aberta no banco quando um restart virar rotina (deploy).
- **Um lock só** para todas as salas e usuários em `internal/rooms`: dá para uma turma; um lock por sala se houver muitas.
- **Contornos para servidor no Windows** (`lfWriter` e perfil de cores em `main.go`): só existem porque o Wish/Bubble Tea assumem Unix. No deploy em Linux, nem são usados.
