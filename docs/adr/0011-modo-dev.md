# Modo dev: outro banco, só localhost e a identidade pelo nome

Para testar o servidor, quem desenvolve precisa de várias pessoas, de salas cheias e de números arbitrários (minutos focados hoje, na semana, energia, streak). Até aqui, isso exigia uma chave SSH por pessoa e editar o `sshiu.db` com o servidor parado, porque as salas guardam os números em memória.

Decidimos por uma flag de inicialização, `-dev`, como os ambientes de desenvolvimento do Rails e do Django. Ela muda três coisas:

- o banco é o `sshiu-dev.db`, e o modo dev nunca encosta nos dados de verdade;
- o servidor só escuta em localhost, mesmo com `-host`;
- **a identidade vem do nome de usuário do SSH**, não da chave: `ssh -p 23235 ana@localhost` entra como a ana, que nasce com esse apelido. É a única exceção ao [ADR 0001](0001-chave-ssh-e-a-identidade.md), e ela só existe porque ninguém de fora alcança esse servidor.

Junto vem a aba dev, a última, com todo mundo do banco numa tabela: criar pessoas, pôr e tirar gente da sala que você está vendo, passar essa sala para alguém, e mudar com `+`/`-` a energia, os dias parado, os minutos de hoje e de antes, o streak, a cor, se a pessoa está online e a tarefa em andamento (que também se escreve e se cumpre pelo painel). Os ajustes passam pelas salas, para a memória e o banco mudarem juntos com o servidor rodando.

## Consequências

- Os minutos focados viram tarefas de mentira ("tarefa dev") no histórico, e não só o número do dia: a tabela de foco soma as tarefas na semana, no mês e no total, e o número de hoje tem de bater com elas. Os passos estão nas [regras](../regras.md#modo-dev).
- Estar online, para quem não está conectado, é uma sessão de mentira aberta pelo mesmo `Connect` das de verdade: a sala não sabe a diferença, e fechá-la no meio de uma tarefa segue a tolerância de reconexão. Ela some quando o servidor reinicia.
- Os ids do modo dev são `dev:<apelido>`, como os bots da demo são `bot:<apelido>`. Um banco de verdade nunca tem um. Com `-dev -demo`, a ana do exemplo e a ana bot querem o mesmo apelido: o bot vira `ana-bot`.
- O banco dev roda em WAL e sem fsync a cada commit, porque perder o último ajuste numa queda não importa num banco de teste. O banco de verdade continua como estava.
- Um subcomando (`sshiu dev ...`) que mexesse direto no banco ficou de fora: com o servidor rodando, a sala não veria a mudança.
- A aba dev fica fora da TUI de quem roda sem `-dev`, e sem `-dev` quem chega sem chave continua recusado.
