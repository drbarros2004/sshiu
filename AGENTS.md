# Plantão

App de foco por SSH em Go (Wish v2 + Bubble Tea v2 + Lip Gloss v2, SQLite via `modernc.org/sqlite`) para uma turma. "Plantão" e "pet" são nomes provisórios.

- **Vocabulário**: use os termos de [`CONTEXT.md`](CONTEXT.md) na conversa, nos comentários e na interface.
- **Decisões**: antes de mexer em identidade, no modelo de foco, nas salas, em chat/reações ou em arte de terceiros, leia o ADR correspondente em [`docs/adr/`](docs/adr/).
- **Números**: toda regra numérica (energia, tempos, limites) fica em [`docs/regras.md`](docs/regras.md), com link para o código. Mudou um número, mude lá também; regra que não cabe lá de forma simples complicou cedo demais.
- **Backlog**: [`IDEIAS.md`](IDEIAS.md). Ideia nova vai para lá, e a dívida técnica conhecida também.

## Convenções

- Responda em português. Código (identificadores, arquivos, pastas) em inglês; comentários e texto da interface em português.
- Textos da interface não tratam ninguém no masculino: reestruture a frase em vez de usar "o(a)", "@" ou "x" (por exemplo, "qual vai ser o seu apelido?", não "Como quer ser chamado?").
- Textos da interface começam com letra minúscula, inclusive títulos e a frase depois de um ponto. Siglas (SSH) e o logo (PLANTÃO) ficam como estão.
- Pacotes em `internal/`, com dependência numa direção só: `tui` → `rooms` → `store` → `pet`. `main.go` sobe o servidor.
- Commits seguem o [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`), com a descrição em inglês. Commit e push só com o OK do usuário.

## Rodar e verificar

O Go fica em `C:\Program Files\Go\bin`, que pode não estar no PATH do shell.

```
go vet ./... && go test ./...
go run . -demo -port 23235
```

Toda mudança termina com `go vet ./...` e `go test ./...` passando, e toda feature precisa ser validada visualmente rodando como o app de verdade: interface nova nasce como um protótipo que roda no terminal, que o usuário vê antes de ela entrar no app.

## Ver a TUI de verdade

O usuário costuma deixar o próprio servidor rodando na porta 23234, com um binário também chamado `plantao.exe`.

- Suba o seu servidor na **porta 23235**, a partir de uma pasta de rascunho: `plantao.db` e `.ssh/` são lidos do diretório atual, então rodar na raiz mexe no banco do usuário. Ex.: `go build -o <rascunho>/plantao-teste.exe .` e rode dentro de `<rascunho>`.
- Pare o seu servidor **pelo PID ou pelo caminho do executável**. Matar pelo nome `plantao.exe` já derrubou o servidor do usuário.
- O painel de terminal não dirige a TUI. Use um harness em Go na pasta de rascunho: cliente `golang.org/x/crypto/ssh` alimentando um emulador `github.com/charmbracelet/x/vt`, com passos como digitar, enter e esperar um texto.
- Confira a tela **sem forçar redesenho** (`WindowChange`): o redesenho esconde bugs do renderer incremental. Para ver cores e pixel art, rasterize a tela num PNG.
- No Windows, o OpenSSH recusa chave privada com permissões abertas ("UNPROTECTED PRIVATE KEY FILE"): `icacls <chave> /inheritance:r /grant:r "%USERNAME%:F"`.

## Estilo

- **Enxuto (ponytail)**: a menor mudança que resolve, sem abstração nem configuração "para o futuro"; stdlib e dependências já instaladas antes de dependência nova. Um atalho consciente com teto conhecido leva um comentário `ponytail:` com o teto e o caminho de upgrade, e entra na dívida técnica de `IDEIAS.md`.
- **Padrão Elm**: tudo que vem de fora (salas, banco) chega como mensagem no `Update`; efeitos saem como `tea.Cmd`; o `View` só lê o modelo. A interface lê as salas pelo `rooms.World` do tick e só toca numa sala dentro de um `Cmd`.
- Teste o que tem lógica (regras, parsing, status) com um teste pequeno ao lado do código.
