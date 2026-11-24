# moedas e loja

- `../coins/coins.json`: segundos por recompensa, moedas por recompensa, limite, saldo inicial, sprite e cor dos números.
- `../coins/coin.png`: ícone 2×2, desenhado em duas colunas e uma linha do terminal, ao lado do número.
- `shop.json`: textos em português/inglês, cores dos estados e IDs dos itens iniciais.
- `../pets/characters.json` e `../pets/hats.json`: `price` da base e `defaultVariantPrice` das cores/modelos. `price` em uma variante substitui o padrão da base. A primeira variante vem incluída por zero.

Os arquivos são embutidos no executável. Depois de editar um JSON ou PNG, compile e reabra o app. Sobrescritas de preço feitas no dev ficam no banco dev, valem para todos nesse servidor e prevalecem sobre os JSONs.

Saldo inicial e inventário inicial só se aplicam ao criar a carteira. Reconectar não repõe itens removidos no dev. A primeira migração também preserva a aparência equipada e suas opções iniciais incluídas.

## controles da loja

- `↑/↓`: categoria; `←/→`: escolher item ou humor.
- `enter`: comprar o item da categoria atual; `ctrl+b`: comprar a seleção inteira.
- `ctrl+r`: voltar a prévia à aparência equipada.
- `y/n`: confirmar/cancelar compra; depois da compra, responder separadamente se deseja equipar.
- Com `-dev`, `x`: abrir o rascunho; `ctrl+o`: alternar disponível/possuído/bloqueado; `ctrl+p`: editar preço.
- No preço, `←/→` muda 1 e `j/l` muda 10; `enter` confirma no rascunho e `esc` desfaz esse campo.
- `ctrl+s`: salvar o rascunho e continuar; `x`: salvar e sair; `esc`: cancelar o rascunho.

## saldo e tarefas no dev

Dentro do editor dev de sala/perfil, `ctrl+m` edita seu saldo: setas mudam 1, `j/l` mudam 10, `enter` confirma no rascunho. Salve com `ctrl+s` ou `x`.

Na sala dev, `ctrl+f` cria uma tarefa já cumprida. Com membro selecionado, usa esse membro; sem seleção fixa, usa sua identidade atual. Digite:

```text
testar moedas | 2026-11-24 10:00:00 | 2026-11-24 10:04:23
```

Os horários seguem a hora local do servidor e o fim não pode estar no futuro. A duração pode incluir segundos e sobrepor outras tarefas. Ao salvar, gera o recibo e as moedas uma vez, sem usar/alterar o resto de segundos das tarefas normais. Não edita nem apaga tarefas antigas.

## verificar no Ubuntu

Os testes não são executados automaticamente pela alteração dos arquivos. Na raiz do repositório:

```sh
go vet ./...
go test ./... -count=1
```

Para conferir visualmente, use um banco dev de teste, sem reaproveitar o banco normal. Verifique a loja em inglês/português e com janela estreita, compras gratuitas, confirmação/cancelamento, equipamento, limite de saldo, recibo próprio e ausência de recibo dos outros membros.
