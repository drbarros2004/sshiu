# Modo dev: outro banco, só localhost e a identidade pelo nome

Para testar o servidor, quem desenvolve precisa de várias pessoas, de salas cheias e de números arbitrários (minutos focados hoje, na semana, energia, streak). Até aqui, isso exigia uma chave SSH por pessoa e editar o `sshiu.db` com o servidor parado, porque as salas guardam os números em memória.

Decidimos por uma flag de inicialização, `-dev`, como os ambientes de desenvolvimento do Rails e do Django. Ela muda três coisas:

- o banco é o `sshiu-dev.db`, e o modo dev nunca encosta nos dados de verdade;
- o servidor só escuta em localhost, mesmo com `-host`;
- **a identidade vem do nome de usuário do SSH**, não da chave: `ssh -p 23235 ana@localhost` entra como a ana, que nasce com esse apelido. É a única exceção ao [ADR 0001](0001-chave-ssh-e-a-identidade.md), e ela só existe porque ninguém de fora alcança esse servidor.

Junto vem a aba dev, a última, com todo mundo do banco numa tabela: criar pessoas, pôr e tirar gente da sala que você está vendo, passar essa sala para alguém, e mudar com `+`/`-` a energia, os dias parado, os minutos de hoje e de antes, o streak, a cor, se a pessoa está online e a tarefa em andamento (que também se escreve e se cumpre pelo painel). Os ajustes passam pelas salas, para a memória e o banco mudarem juntos com o servidor rodando.

## Consequências

- Os minutos focados viram tarefas de mentira ("tarefa dev") no histórico, e não só o número do dia: a tabela de foco soma as tarefas na semana, no mês e no total, e o número de hoje tem de bater com elas. Os passos estão nas [regras](../regras.md#modo-dev).
- Estar online, para quem não está conectado, é uma sessão de mentira aberta pelo mesmo `Connect` das de verdade: a sala não sabe a diferença, e fechá-la no meio de uma tarefa segue a tolerância de reconexão. No ambiente dev, a presença simulada agora é gravada e restaurada ao reiniciar, independentemente das conexões SSH reais.
- Os ids do modo dev são `dev:<apelido>`, como os bots da demo são `bot:<apelido>`. Um banco de verdade nunca tem um. Com `-dev -demo`, a ana do exemplo e a ana bot querem o mesmo apelido: o bot vira `ana-bot`.
- Um subcomando (`sshiu dev ...`) que mexesse direto no banco ficou de fora: com o servidor rodando, a sala não veria a mudança.
- A aba dev fica fora da TUI de quem roda sem `-dev`, e sem `-dev` quem chega sem chave continua recusado.


## Editor na própria tela

Com `-dev`, `x` abre um rascunho na sala ou no perfil, com destaque turquesa.
O rascunho só muda a cópia usada para desenhar; o banco e as outras sessões só mudam ao salvar.
`Ctrl+S` valida e salva tudo numa transação, permanecendo no editor; `x` salva e sai.
`Enter` confirma um campo no rascunho. `Esc` cancela esse campo, volta da seleção de membro
ou, no nível externo, abandona as alterações desde o último salvamento.
Uma mudança externa no cenário invalida o salvamento: cancelar e reabrir evita sobrescrever
outra sessão ou concluir a mesma tarefa duas vezes. Os cronômetros continuam correndo.

Na sala, `←/→` escolhe um membro e `Enter` fixa a seleção. As ações são:

| Contexto | Atalho | Ação |
|---|---|---|
| Sala, fora da seleção fixa | Ctrl+B | adicionar usuário existente ou criar com Ctrl+N |
| Sala, fora da seleção fixa | Ctrl+O | escolher obra |
| Membro selecionado | Ctrl+N | editar apelido, com o validador do primeiro acesso |
| Sala/perfil | Ctrl+E | editar energia exibida, por setas ou número digitado |
| Membro selecionado | Ctrl+L | escolher cor |
| Membro selecionado | Ctrl+G | escolher estágio; -1 volta aos minutos reais |
| Membro selecionado | Ctrl+P | alternar presença simulada |
| Membro selecionado | Ctrl+D | remover da sala, preservando conta e histórico |
| Membro selecionado | Ctrl+T | iniciar tarefa (n), cumprir (f) ou parar (p) |
| Sala, membro / aba dev | Ctrl+U | assumir identidade |
| Sala/perfil no editor | Ctrl+R | voltar à identidade SSH original ao salvar |

Na aba dev, `Ctrl+U` pede `Enter` antes de assumir a identidade, descartando os rascunhos
de sala/perfil. Nas telas inline a troca fica pendente até salvar. A identidade SSH original
não muda; a identidade efetiva determina perfil, salas, tarefas e permissões.
Respostas de perfil da identidade anterior são ignoradas. As preferências visuais da sessão
permanecem para facilitar comparar os mesmos cenários.

No perfil dev, personalizar a cor e pressionar `Enter` confirma só o rascunho; cor e energia
são gravadas juntas no salvamento dev. Fora do editor, o comportamento normal permanece.

A energia editada é um novo ponto de partida, não uma faixa congelada. O marcador `dev_energy`
evita somar novamente o bônus do streak sem apagar o streak; a próxima tarefa volta ao cálculo
normal. A tarefa atual mantém seu início, duração e término.
O estágio manual usa um deslocamento de minutos por sala/usuário/dia em `dev_state`, sem
criar histórico fictício. O foco posterior avança a partir dele, até o estágio final; o
novo dia descarta o deslocamento anterior, como a obra normal.
Iniciar uma tarefa usa o fluxo de duração normal; alterar tempos de uma tarefa existente
continua exclusivo da aba dev. A conclusão usa a mesma gravação de tarefas do aplicativo,
participando da transação do rascunho.

Os dados salvos sobrevivem a Ctrl+C e reabertura; rascunhos não salvos não sobrevivem.
Com `-dev -demo`, a demo passa a ser um cenário persistente: os bots iniciais são dados para
editar, sem o piloto automático que sobrescreveria suas escolhas. Sem `-dev`, a demo normal
continua com seus bots animados. A presença dev não fecha conexões reais e as tarefas dev
não são interrompidas pela desconexão do editor.

Código: `tui/dev_inline.go` mantém os controles e rascunhos, `rooms/dev_scene.go` valida/projeta
as intenções e coordena a memória, e `store/dev_scene.go` grava a transação.
Renderização de pets, cards, energia e obras continua nas funções existentes.

A identidade assumida é lembrada por login SSH original no banco dev, inclusive ao reconectar.
Em telas pequenas da sala, `?` expande a lista de atalhos no próprio painel.
