# Ideias de expansão do sshiu

> **O nome é sshiu**, decidido em 02/11/2026 ([#28](https://github.com/drbarros2004/ssh-focus-app/issues/28)): tem o ssh dentro e o "shhh" de quem pede silêncio numa biblioteca. "Pet" continua provisório.

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
- **O quê:** um wallpaper ASCII animado em tela cheia, com o relógio e outras informações por cima, se o usuário quiser. Liga por tecla ou depois de 30 s sem tecla, só durante o foco; quem preferir só a tecla escolhe "nunca" em "liga sozinho", na aba visual ([#35](https://github.com/drbarros2004/sshiu/issues/35)). ←/→ trocam o wallpaper (o nome dele aparece por 5 s, ao abrir e a cada troca); qualquer outra tecla volta sem executar nada; o fim do foco desliga.
- **Wallpapers:** 6, escolhidos no protótipo de 29/10/2026 (branch `prototype/jardim`): céu noturno, chuva lenta e lâmpada de lava, escritos do zero com base em projetos MIT/Apache e creditados em `CREDITS.md` ([ADR 0005](docs/adr/0005-arte-de-terceiros-so-com-licenca-permissiva.md)); café, árvore com folhas caindo e estrada à noite, desenhos nossos em pixel art (meio bloco, com o cenário parado e o movimento junto, para a banda ficar perto da chuva). A neve ficou de fora. Tudo com pouco movimento e cores apagadas, sem clarões nem nada que se acumule na tela. Cada usuário escolhe o seu na aba visual; quem não escolheu ganha um sorteado a cada tarefa. A troca com ←/→ dentro do modo jardim vale só para aquela tarefa ([#27](https://github.com/drbarros2004/ssh-focus-app/issues/27)); na prévia (`j` na aba visual), as setas mudam a escolha da aba.
- **O que aparece por cima** é escolhido na aba visual, junto do wallpaper (no modo jardim, só o wallpaper se troca): o relógio (mm:ss, só os minutos ou nenhum), a tarefa e quem mais da sala está focando. O relógio só com os minutos existe porque os segundos mudando são o único movimento rápido da tela.
- **Caixa:** o texto fica sobre uma caixa vazia, que abre um buraco no wallpaper; sem ela, a lava e a chuva passam por trás e atrapalham a leitura. Os de pixel art escolhem o próprio lugar e as cores do texto, sem caixa, num quadro que é parte do desenho: a lousa de giz na parede do café, o painel do carro na estrada (a tela com o texto, velocímetro, conta-giros, gasolina, temperatura e hodômetro da viagem) e uma placa de madeira fincada no campo na árvore. O nome do wallpaper fica no rodapé, fora do bloco, para sumir sem mexer no resto.
- **Padrões** de quem nunca mexeu na aba visual: wallpaper sorteado a cada tarefa, relógio só com os minutos, tarefa e sala ligadas, liga sozinho depois de 30 s (quem chega vê o que o modo jardim oferece e desliga o que achar demais).

### Sala solo (implementada em 30/10/2026)
- **O quê:** uma sala de uma pessoa, sem código, para focar sozinho ([ADR 0006](docs/adr/0006-sala-solo.md)). Criar com tamanho 1 é a única diferença.
- **Por quê:** antes, quem criava uma sala sozinho via uma grade com um card, um feed vazio e um código para passar para ninguém. Focar sozinho é o caso de quem acabou de chegar e de quem estuda fora do horário da turma.
- **Tela:** o layout A, escolhido entre três candidatos e validado no protótipo `prototype/solo`: painel à esquerda (relógio, tarefa, o pet com humor, streak e minutos de hoje, e um mini log de 3 linhas). Sem reações, sem cursor de eventos e sem cards. O `j` abre o jardim em tela cheia a qualquer momento, não só durante o foco.
- **A coluna da direita muda com o estado:** focando é o modo jardim; parado, são os seus números (calendário do mês, barras da semana, totais). Decidido em 31/10/2026, depois de testar e descartar o jardim a 2 fps (parecia travado) — assim começar a tarefa muda a tela, e quem está parado não paga os 4 a 6% de um núcleo que o jardim custa a 8 fps. Abaixo de 96 colunas não cabe coluna nenhuma e sobra só o painel.

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
- **Feito:** a sala tem dono ([ADR 0007](docs/adr/0007-dono-da-sala.md)), que troca o nome e o tamanho, fecha para novas entradas, remove um membro, passa a sala e a encerra. Sair sendo dono passa a sala para o membro mais antigo.
- **A tela** (`c` na sala, só para o dono; com o tempo esgotado o `c` volta a ser "continuar", que é o que o painel está pedindo, e a dica some do cabeçalho): nome, tamanho e novas entradas mudam na própria página, com ←/→ na linha escolhida, e as setas só aparecem na linha em que o cursor está — a que não anda mais some, em vez de explicar o limite. O tamanho é `membros / ‹ teto ›`, então o chão fica à vista. Passar a sala, remover alguém e encerrar não têm valor à direita: abrem uma página própria. O cabeçalho da sala ganha `· fechada` e `· c configura`.
- **Quem perde a sala** (removido ou sala encerrada) cai no menu de salas com o aviso "você não está mais em X", que o menu passou a mostrar.
- **Falta:** trocar o código, que dá algum trabalho — hoje o código é a chave da sala no banco, e as tarefas guardam em que sala aconteceram.

### Histórico, stats e gráficos
- **Prioridade:** essencial. É a próxima feature grande, antes dos vários pets.
- **Primeira parte, feita em 31/10/2026** (na aba perfil e na coluna da sala solo parada, em `internal/tui/stats.go`): calendário do mês com o dia mais aceso quanto mais tempo focado, barras de minutos por dia da semana, e três números — focado no mês e no total, tarefas cumpridas de quantas, e a média declarada contra a focada. Sai de duas consultas (`store.FocusByDay` e `store.GetTotals`).
- **Segunda parte, feita em 14/11/2026** ([#45](https://github.com/drbarros2004/sshiu/issues/45)): o calendário em blocos de calor, com os meses desde janeiro de 2026 (←/→ no perfil); os números do mês (focado, dias com foco, melhor dia e a comparação com o mês passado até o mesmo dia); o recorde de streak; o horário em que a pessoa foca (pico e manhã, tarde e noite); e o quanto ela cumpre em tarefas curtas, médias e longas. Em painéis ou em lista, escolhido na aba visual, até a turma decidir um dos dois.
- **Ideias que ficaram da #45:** `ssh sshiu stats`, que imprime o resumo e sai (o layout em lista já é quase isso); uma retrospectiva "o seu 2026" na virada do ano; "focou junto", os minutos em que alguém da mesma sala também focava; "no que o tempo foi", as tarefas agrupadas pelo texto; e uma linha com o resumo da semana passada na primeira conexão de segunda.
- **Gráficos da sala:** os mesmos agregados para a sala inteira, mais o pico de usuários focando ao mesmo tempo e quanto tempo esse pico durou. O pico sai dos intervalos de início e fim das tarefas dos membros. A tarefa encerrada guarda a sala em que começou (`tasks.room_code`, desde 31/10/2026), então o agregado por sala é uma consulta.
- **Aba de histórico:** com filtro por sala (a coluna já existe), se os números do perfil não bastarem.

### Vários pets (10 a 15 no MVP)
- **Prioridade:** essencial, mas depois dos gráficos.
- **O quê:** em vez de um pet só, cada usuário escolhe o seu entre vários, cada um com os próprios humores.
- **Por quê:** personalização. Pelo ritmo em que os sprites atuais saíram, 10 a 15 pets parece viável para o MVP.
- **O nome "pet":** continua provisório. Ficou de fora da decisão do nome do app ([#28](https://github.com/drbarros2004/ssh-focus-app/issues/28)).

### Tabela de foco (implementada em 10/11/2026)
- **O que entrou:** os minutos de cada membro hoje, na semana, no mês ou no total, no menu do dono, em tabela ou ranking, fechada ou aberta para a sala na tecla `o` ([ADR 0010](docs/adr/0010-tabela-de-foco.md), [#16](https://github.com/drbarros2004/sshiu/issues/16)).
- **Falta da [#83](https://github.com/drbarros2004/sshiu/issues/83):** o log de todas as tarefas de todos, com filtro por pessoa. Sai da tabela `tasks`, que já guarda a sala de cada tarefa (`room_code`).
- **Administrador de instituição:** uma empresa parceira com um espaço mais privado e alguém acima das salas acompanhando tudo. É um modelo novo de gente (o dono hoje é de uma sala só, [ADR 0007](docs/adr/0007-dono-da-sala.md)), então pede um ADR próprio antes de código.

### Pôr no ar: `ssh sshiu`
- **O quê:** um servidor público com domínio, ouvindo na porta 22, para entrar só com `ssh sshiu` (ou `ssh sshiu.<domínio>`), como o `ssh terminal.shop`. O convite de uma sala vira uma linha para colar no grupo da turma: `ssh -t sshiu <código>`. O `-t` está lá porque, com o código no fim, o ssh não pede terminal; ler o código do usuário (`ssh <código>@<nome>`) dispensaria o `-t` e deixaria o convite mais curto.
- **Como:** escutar em todas as interfaces (já dá, com `-host 0.0.0.0`), porta 22, chave do servidor fixa e fora do git (publicar o fingerprint para quem quiser conferir), binário compilado com `GOOS=linux`. Quem entra só precisa de um cliente SSH e de uma chave.
- **A chave do servidor nunca muda:** o cliente SSH guarda a chave do servidor na primeira conexão. Se ela mudar (servidor recriado, `.ssh/` apagado, outra máquina no mesmo endereço), todo usuário recebe o aviso "REMOTE HOST IDENTIFICATION HAS CHANGED" e a conexão é recusada. Guardar a chave do servidor com backup e levá-la junto em qualquer migração.
- **Quando:** sem pressa, mas o grupo precisa conseguir entrar e se ver antes da apresentação. O nome já é sshiu; falta escolher o domínio ([#24](https://github.com/drbarros2004/ssh-focus-app/issues/24), [#25](https://github.com/drbarros2004/ssh-focus-app/issues/25)).
- **O repositório:** ele ainda se chama `ssh-focus-app` no GitHub. Renomear para `sshiu` quando o app for ao ar; o GitHub redireciona os links antigos.

### Telas de tamanhos diferentes
- **O quê:** testar o app em tamanhos de terminal diferentes e ver que adaptações cabem. Já foi testado em tela cheia num notebook e em parte da tela de um monitor maior.
- **Por quê:** alguém pode querer deixar o app numa janela menor, num canto da tela (1/4 dela, por exemplo), enquanto trabalha em outra coisa.
- **Cuidado:** a princípio, sem malabarismos por causa disso; o primeiro passo é ver o que quebra.
- **Referência:** [Learning Go by Building a TUI - Full Screen & Responsive Layout](https://www.youtube.com/watch?v=Db7YmM-ZnoA) (Proomptly), layout responsivo com Bubble Tea.

### Cara de app de dev
- **O quê:** elementos visuais que dev reconhece de cara, em algum lugar do app. Exemplo: o estilo do `git diff`, com `+` em verde e `-` em vermelho.
- **Onde:** ainda não decidido.

### Termos de servidor no vocabulário
- **O quê:** nomes tirados de servidor e de Unix para coisas do app. Se entrarem, mudam também o [`CONTEXT.md`](CONTEXT.md).
  - **Streak → uptime:** "uptime de 12 dias" é como se diz que um servidor está no ar sem parar.
  - **Moeda → ciclos:** ciclos de CPU ganhos focando (ver [Moeda e loja](#moeda-e-loja)).
  - **Humor "murcho" → zumbi:** um processo zumbi já terminou, mas continua na lista porque ninguém veio buscar o resultado. É o pet largado.
- **Keepalive**, para a apresentação: o SSH manda sinais para a conexão não cair, e o foco de cada um mantém o pet vivo.
- **Mais termos** (brainstorming de 14/11/2026):
  - **Interrompida → SIGHUP:** é o sinal que o processo recebe quando o SSH cai, ao pé da letra.
  - **Feito / não cumprida → `exit 0` / `exit 1`:** o resultado da tarefa como código de saída.
  - **Status como estado de processo:** focando → running, sem tarefa → idle, offline → sleep.
  - **Dono da sala → root**, **sala solo → localhost**, **feed → `tail -f` da sala**.
  - **Tempo declarado → timeout**, **continuar → retry**.
  - **Humor como status HTTP:** radiante → 200 OK, murcho → 503.
- **Jeitos de usar, além de renomear:**
  - **Piada pontual:** os termos aparecem só no FAQ, nos bots e em mensagens raras.
  - **Legenda:** o termo do app fica, e o de dev aparece apagado ao lado ("streak 12 dias · uptime").
  - **Idioma `sysadmin`:** um terceiro idioma na aba visual que troca os termos. É caro: o `i18n.go` diz que um terceiro idioma pede trocar a chave por um id.
  - **Easter eggs:** a tarefa digitada como comando (`sudo …`, `rm -rf`, `exit`) ganha uma resposta seca.
  - **Forma visual:** o resultado no feed no estilo `git diff` (cumpriu `+` verde, parou `-` vermelho), sem trocar palavra nenhuma.
- **Decidido no grill de 14/11/2026 ([#13](https://github.com/drbarros2004/ssh-focus-app/issues/13)):** fica tudo para depois. Nenhum termo renomeado, nenhum toque visível; as ideias ficam só anotadas aqui.

### Textos da demo mais engraçados
- **O quê:** trocar as tarefas dos bots da demo por piadas de dev, como "reescrevendo X em Rust".
- **Por quê:** a demo vai ser apresentada para uma plateia.
- **Onde:** as tarefas dos bots ficam em `internal/rooms/bots.go`.
- **Decidido no grill de 14/11/2026 ([#14](https://github.com/drbarros2004/ssh-focus-app/issues/14)):** cada bot ganha um personagem, com tarefas que misturam faculdade de verdade e piada de dev. As frases são traduzidas na hora de mostrar, como o nome da sala demo. Esforço pequeno: a sala demo não é mais o foco, e melhorá-la de verdade fica para uma sessão própria.

### Wallpaper de piada: "para desenvolvedores de verdade"
- **O quê:** um wallpaper do modo jardim no estilo "brainrot": a tela dividida com um vídeo de Subway Surfers ou de corte de sabão, como nos cortes de vídeo que prendem a atenção. Só de piada, pensado para a demo.
- **Como dá para fazer:** sem nada especial (nem Lua, nem protocolo de imagem do terminal). O mesmo truque do pet em pixel art: cada caractere `▀` mostra 2 pixels, um na cor da frente e outro no fundo. Os quadros do vídeo seriam convertidos antes (ffmpeg, em baixa resolução), embutidos no binário com `go:embed` e tocados no ritmo do modo jardim.
- **Cuidados:**
  - **Licença:** resolvida. Os dois vídeos (Subway Surfers e Minecraft parkour) estão em `assets/videos/`, com fonte e licença no `CREDITS.md` e a exceção no [ADR 0005](docs/adr/0005-arte-de-terceiros-so-com-licenca-permissiva.md). Os quadros saem deles com `go generate ./internal/tui` e ficam embutidos no binário; o crédito do CC BY fica embaixo do vídeo do Minecraft.
  - **Só para quem pede:** "permitir wallpapers para devs de verdade?" na aba visual, com "não" por padrão ([#36](https://github.com/drbarros2004/sshiu/issues/36)).
  - **Peso:** medido no protótipo ([PR #4](https://github.com/drbarros2004/ssh-focus-app/pull/4)): ~380 KB/s por usuário, 8× a chuva. Fica como está até a [passada de performance](#passada-de-performance).

### Tela "sobre"
- **O quê:** um FAQ com perguntas e respostas sérias, misturadas com piadinhas. Tom de exemplo: "a focus app for real developers (or vibe-coders)".
- **O nome:** é o lugar natural para explicar o sshiu, com as respostas que ficaram do brainstorming do nome: "por que sshiu? é o ssh de quem pede silêncio" e "por que não tem chat? aqui é uma biblioteca" ([ADR 0004](docs/adr/0004-sem-chat-so-reacoes.md)).
- **Créditos:** a tela leva o crédito do Minecraft parkour do wallpaper brainrot (canal Orbital, CC BY): a licença pede crédito para quem vê o vídeo, e o `CREDITS.md` só chega a quem lê o repo ([ADR 0005](docs/adr/0005-arte-de-terceiros-so-com-licenca-permissiva.md)). Enquanto a tela não existe, o crédito fica embaixo do vídeo, no próprio wallpaper.
- **Decidido no grill de 14/11/2026 ([#10](https://github.com/drbarros2004/ssh-focus-app/issues/10)):** vira uma 4ª aba, "sobre", depois de sala/perfil/visual. O formato saiu do protótipo `prototype/sobre` (acordeão `#`, man page e `--help`): ficou o `--help`, com cada pergunta como uma flag e a resposta da escolhida aberta embaixo. A linha de uso mostra o `ssh` de verdade do servidor, para não prometer um comando que não existe. O conteúdo começa com um rascunho de umas 10 perguntas, metade sérias e metade piadas, nos dois idiomas, para o usuário cortar. O crédito do Minecraft fica fora deste trabalho. #10, #13 e #14 vão num PR só.

### O silêncio do sshiu
O nome pede silêncio, e a ideia aparece no app sem forçar. Do brainstorming de 02/11/2026 entraram só o feed vazio ("silêncio por aqui") e a linha embaixo do logo na primeira tela ("um lugar quieto para focar junto"). O resto ficou aqui, das mais discretas às mais ousadas:
- **"reagir em silêncio":** a dica das reações. É verdade ao pé da letra: as reações estão lá para ninguém precisar falar ([ADR 0004](docs/adr/0004-sem-chat-so-reacoes.md)).
- **"saiu na ponta dos pés":** o evento de saída no feed. Aparece muitas vezes por dia, então pode cansar.
- **Plaquinha de silêncio:** "🤫 3 focando" no cabeçalho da sala, só quando duas ou mais pessoas estão focando, como a placa de uma biblioteca.
- **Despedida ao sair:** "shhh… até a próxima." no terminal, depois do `q`. Seria o único lugar com o "shhh" escrito.
- **O pet faz "shh":** um quadro do pet com o dedo na boca quando alguém reage enquanto você foca.
- **Modo biblioteca:** durante o foco, as reações recebidas ficam guardadas e só aparecem quando a tarefa acaba.
- **Cuidado:** o "shhh" é sutil. Uma ou duas dessas por vez, e só as que já são verdade no app.

### Modo dev: o que ficou de fora
- **O quê:** o modo dev ([ADR 0011](docs/adr/0011-modo-dev.md)) cobre pessoas, salas e números. Ficaram de fora: um seed que já sobe uma sala cheia, as outras configurações da sala (tamanho, obra, meta) pelo painel, parar uma tarefa sem cumprir e um convite com `<apelido>@` no lugar do endereço puro.
- **Por quê:** só entram quando um teste pedir; hoje dá para fazer tudo isso entrando com outro nome.

## Mais adiante

### Passada de performance
- **O quê:** uma rodada só de performance, perto do fim do projeto. Até aqui ninguém pensou nisso.
- **Medido em 30/10/2026** (servidor de teste, 160×50 e 200×50):

  | | banda por usuário | CPU por usuário (5 juntos) |
  |---|---|---|
  | brainrot | ~380 KB/s (~1,4 GB por hora de foco) | ~6% de um núcleo |
  | chuva lenta | ~48 KB/s | ~3,5% de um núcleo |
  | sala, sem modo jardim | ~90 KB/s | — |

  Os quadros do brainrot ocupam ~60 MB de memória, fixos, não por usuário. O gargalo é a banda: 20 pessoas no brainrot mandam uns 60 Mbit/s.
- **Por onde começar:** o brainrot a 4 fps (metade da banda); por que dois vídeos custam 3× um só (117 KB/s), e não 2×; por que a sala parada manda 90 KB/s; e o custo de desenhar o modo jardim em si, que já é a maior parte da CPU da chuva.
- **Como medir:** o harness em duas rodadas com durações diferentes (5 s e 35 s), diferença dos bytes recebidos e do `TotalProcessorTime` do servidor ÷ 30 s; para CPU, vários usuários ao mesmo tempo, porque um só fica no ruído.

### Tempo declarado personalizado (implementado em 14/11/2026)

O "outro", no fim da lista, abre um campo de 5 a 120 min ([`docs/regras.md`](docs/regras.md)). O
seletor abre no último tempo usado, marcado com ↺, e esse tempo entra na lista no lugar dele, sem
virar mais uma opção fixa para percorrer com as setas ([#45](https://github.com/drbarros2004/sshiu/issues/45)).

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

## Temas de obra que faltam desenhar

Cada tema custa quatro desenhos de 12×12 pixels ([ADR 0008](docs/adr/0008-obra-da-sala.md)). O gerador
dos que existem é um mapa de caracteres, uma linha por linha de pixel. Horta, estante, feira,
acampamento, orquestra, passarinhos, porto e foguetes já entraram; a ponte foi desenhada e
descartada em 01/11/2026.

- **Bolo de casamento:** cada pessoa faz um andar, e os andares se compõem numa escada. Quem não
  focou deixa a escada desigual. Precisa de um arranjo vertical, que nenhum tema de hoje usa, e de
  um teto de altura: com dez andares não cabe em terminal nenhum.
- **Varal de bandeirinhas:** corda vazia, bandeirinha pequena, bandeirinha grande, bandeirinha com
  lampião. Usa o engate do trem como a corda, então a sala inteira vira um varal de festa junina.
- **Estação espacial:** anel de acoplamento, módulo, módulo com painel solar, módulo com antena.
  Também engata, como o trem, mas pode quebrar em fileiras sem parecer errado.
- **Muralha:** uma torre por pessoa, do alicerce à torre com bandeira; a muralha entre elas é o
  engate, e quem não focou deixa a torre baixa no meio do muro.
- **Recife:** pedra, coral pequeno, coral ramificado, coral com peixe. Areia no chão.
- **Quebra-cabeça:** cada pessoa é uma peça de uma imagem só, e a imagem só fecha com todo mundo.
  O único tema em que o conjunto é literalmente um desenho, e por isso o mais caro: uma arte
  grande para cada tamanho de sala, cortada em peças.
- **Constelação, vitral:** anotadas na conversa, sem desenho.

## Paletas do pet: revisar aproximações em 256 cores

As oito paletas já seguem os PNGs desenhados para as partes do pet.
Os miolos das flores de verde e turquesa e as estrelinhas do rosa usam HEX que são aproximados
em 256 cores; decidir se as cores do editor devem ser ajustadas à tabela xterm.
Os valores originais e convertidos estão no [guia](docs/pets-paletas.md).

## Dívida técnica conhecida

Também marcada no código com comentários `ponytail:`.

- **Migração ingênua:** o `ALTER TABLE` ignora o erro de "coluna já existe". Trocar por uma ferramenta de migração quando houver várias.
- **Um lock só** para todas as salas e usuários em `internal/rooms`: dá para uma turma; um lock por sala se houver muitas.
- **Sala sem rolagem** (`roomView` em `internal/tui/room.go`): quando falta altura, a lista de eventos encurta e os pets viram um ponto da cor do corpo do pet, mas com muita gente num terminal baixo (80×24 já com 6 membros) os cards passam da altura e são cortados. Rolagem nos cards se isso aparecer de verdade.
- **O card não encolhe de verdade** (`card` em `internal/tui/room.go`): no modo compacto o pet vira
  um ponto, mas a caixa continua com as duas linhas fixas da tarefa, então cada fileira de cards
  custa 8 linhas mesmo vazia. É o que decide hoje se a obra cabe: em 120 colunas os cards cabem
  numa fileira e a obra entra com 32 linhas; em 100 eles quebram em duas e ela só entra com 40.
- **Contornos para servidor no Windows** (`lfWriter` e perfil de cores em `main.go`): só existem porque o Wish/Bubble Tea assumem Unix. No deploy em Linux, nem são usados.
- **Tradução com o português como chave** (`internal/tui/i18n.go`): `m.t("texto em português")` busca o inglês num mapa. Mudar uma frase em português pede mudar a chave no mapa também, e o `TestEveryTextTranslated` avisa quando isso fica para trás. Os erros de `rooms` chegam em português e são traduzidos na tela, os que têm nome ou número por padrão (`enErrors`). Um terceiro idioma pede trocar a chave por um id.
- **A demo fala português** (`internal/rooms/bots.go`): as tarefas dos bots são conteúdo, não interface, e aparecem em português mesmo para quem usa o app em inglês.
- **Pixel art pela fonte** (`pet.Pixels`, `solid` em `internal/tui/garden.go`): os meios blocos saem como "▄" invertido e sublinhado, para o fundo e o sublinhado cobrirem a célula no Terminal.app (issue #56). Medido só no Terminal.app com a fonte padrão: lá sobram emendas verticais bem fracas, da suavização da fonte. Falta ver o terminal do VS Code; se o sublinhado aparecer como linha em algum terminal, a saída é tirar o sublinhado e aceitar 1 px na base da meia célula.
- **Dois layouts das stats** (`statsLayouts` em `internal/tui/profile.go`): painéis e lista convivem, com a escolha na aba visual e a coluna `users.stats_list`, só até a turma escolher um ([#45](https://github.com/drbarros2004/sshiu/issues/45)). Escolhido, sai o outro ramo de `profileView` e de `soloStats`, a linha da aba visual e a coluna.
