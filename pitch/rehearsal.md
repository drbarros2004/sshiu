# Plano de ensaio

> Status: plano preparado; **nenhum ensaio realizado**. Atualize este topo depois de cada passagem cronometrada.

## Papéis

| Papel | Quem | O quê |
|---|---|---|
| Apresenta | [A CONFIRMAR] | Fala o roteiro; um só apresentador cabe melhor em 3 min |
| Opera | [A CONFIRMAR] | Passa os slides no notebook, no modo apresentador (`S`) |
| Cronometra | [A CONFIRMAR] | Dá `T` na primeira palavra e avisa se o ritmo ficar vermelho |
| Responde | Todos | Cada pergunta da banca tem dono (tabela abaixo) |

| Banca | Pergunta mais provável | Quem responde |
|---|---|---|
| Investidor-anjo | valuation, receita, uso do dinheiro | [A CONFIRMAR] |
| CEO | vigiar funcionário, preço por assento | [A CONFIRMAR] |
| Founder | por que vocês, defensabilidade, por que terminal | [A CONFIRMAR] |
| Agilista | como entra no sprint, issues | [A CONFIRMAR] |

## Pré-condições (checklist do dia)

- [ ] O pedido (`[R$ 100 mil]` por `[10%]`) definido; `grep -o '\[[^]]*\]' deck/index.html` sem placeholder de texto.
- [ ] Números do slide 7 medidos de novo na véspera (`git rev-list --count master`, `gh pr list --state merged`, `grep '^func Test'`).
- [ ] Octoverse 2026 e preço do Focusmate Plus conferidos ([`source-checks.md`](source-checks.md)).
- [ ] Deck aberto pelo **`deck/index.offline.html`** no Google Chrome (o Chromium sem codec não toca o vídeo). Funciona sem internet.
- [ ] Projetor em 16:9 testado; tela cheia com `F`; modo apresentador com `S` no notebook.
- [ ] Volume testado na sala: a trilha do vídeo fica bem abaixo da voz (notebook em ~30%). O som sai pelo notebook ou pelo cabo do projetor? Confirmar.
- [ ] PDF de reserva ([`deck/sshiu-pitch.pdf`](deck/sshiu-pitch.pdf)) num pendrive ou no e-mail.

## Passagens

Duas passagens completas, em voz alta, com o deck e o cronômetro do HUD:

1. **Sozinho, no notebook:** anotar o tempo de cada bloco ao lado do previsto em [`script.md`](script.md).
2. **Com uma pessoa fazendo de banca:** 3 min de pitch + 3 perguntas sorteadas do [`dossier.md`](dossier.md).

O orçamento dos slides soma 2:57, só 3 s de folga. Passou de 3:00? Corte explicação (os bullets do slide 2 e os exemplos do slide 6 são os primeiros), nunca o vídeo nem o pedido.

| Passagem | Data | Tempo total | Estourou em | Ajuste |
|---|---|---|---|---|
| 1 | | | | |
| 2 | | | | |

## Congelamento

Depois da segunda passagem, o deck congela: só valem os números do [`source-checks.md`](source-checks.md) e o vídeo de 40 s montado em 25/11/2026 com tomadas do commit 854000a. Nada de "só mais um ajuste" no dia.

## Plano B

- **Vídeo não toca:** `V` ou clique no vídeo; se ainda assim não tocar, abra `deck/demo.mp4` direto no QuickTime.
- **Notebook falha:** o PDF tem os 8 slides (o vídeo vira um quadro); narre a demo em cima do quadro do slide 4.
- **Demo ao vivo pedida pela banca:** só se houver tempo depois do pitch. `go run . -demo` na raiz do repositório e `ssh -p 23234 localhost` em outro terminal.
