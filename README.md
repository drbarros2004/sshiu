# Plantão

> Nome provisório (WIP): o nome oficial ainda vai ser decidido com o grupo.

Um app de foco que roda no terminal, por SSH. Você entra na fila do próximo sprint de uma sala, escreve o seu Lock (a primeira ação concreta do sprint), foca junto com a sala e, no intervalo, faz o check-out e conversa. O seu pet floresce quando você foca e murcha quando você some.

## Rodar

Precisa de Go 1.27 ou mais novo e de um cliente SSH.

1. **Crie a sua chave SSH**, se ainda não tiver uma. Ela é o seu login:
   ```bash
   ssh-keygen -t ed25519
   ```
   Pode apertar enter em todas as perguntas.

2. **Suba o servidor** na raiz do repositório:
   ```bash
   go run . -demo
   ```
   Com `-demo`, um sprint dura 1 minuto e a sala pública já vem com bots. Sem a flag, os sprints são de 25 minutos e seguem os horários de cada sala. Na primeira vez, o servidor cria a própria chave em `.ssh/` e o banco `plantao.db`.

3. **Conecte** em outro terminal:
   ```bash
   ssh -p 23234 localhost
   ```
   Na primeira conexão, o SSH pergunta se você confia no servidor: responda `yes`. Depois escolha um apelido.

Para usar outra porta, passe `-port`. Para rodar os testes, use `go test ./...`.

## Mais

- [`IDEIAS.md`](IDEIAS.md): o backlog de ideias e a dívida técnica conhecida.
