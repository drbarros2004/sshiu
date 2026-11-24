# moedas e loja

- `../coins/coins.json`: segundos por recompensa, moedas por recompensa, limite, saldo inicial, sprite e cor dos números.
- `../coins/coin.png`: ícone 2×2, desenhado em duas colunas e uma linha do terminal, ao lado do número.
- `shop.json`: textos em português/inglês, cores dos estados e IDs dos itens iniciais.
- `../pets/characters.json` e `../pets/hats.json`: `price` da base e `defaultVariantPrice` das cores/modelos. `price` em uma variante substitui o padrão da base. A primeira variante vem incluída por zero.

Os arquivos são embutidos no executável. Depois de editar um JSON ou PNG, compile e reabra o app. Sobrescritas de preço feitas no dev ficam no banco dev, valem para todos nesse servidor e prevalecem sobre os JSONs.

Saldo inicial e inventário inicial só se aplicam ao criar a carteira. Reconectar não repõe itens removidos no dev. A primeira migração também preserva a aparência equipada e suas opções iniciais incluídas.

A lista usa quatro colunas: categoria, seleção (`‹ item › X/Y`), estado/preço e coleção (`X/Y owned`). A prévia de humor não tem estado nem coleção. A explicação continua abaixo da categoria; as contagens não se repetem abaixo. As colunas reservam as maiores larguras do catálogo e só se comprimem quando a janela exige.

## confirmação de compra

A compra e a pergunta de equipamento usam uma caixa centralizada na área entre as abas e os controles. `confirmation` em `shop.json` controla a largura mínima (62 colunas), a proporção horizontal da janela (45%) e os preenchimentos. A altura acompanha o conteúdo, com uma linha vazia antes dos controles `y/n`, em cinza secundário. O terminal informa células; o tamanho da fonte é controlado pelo próprio terminal.

A confirmação destaca a pergunta em negrito, seguida de um divisor. Os rótulos dos saldos ficam em cinza; os valores continuam verdes. Sob `itens:`, em negrito, compras ficam recuadas duas colunas (`itemIndent`) e opções incluídas quatro (`includedIndent`). Uma linha vazia separa compras; opções incluídas ficam junto da base. Preços são vermelhos com sinal de menos; opções incluídas mostram apenas `included`. As cores `confirm_text`, `confirm_question`, `confirm_muted`, `confirm_border`, `confirm_balance` e `confirm_price` ficam no mesmo JSON, com tons da paleta xterm. Fundo e navegação seguem o tema atual.

Em janelas muito baixas, `↑/↓` rola o conteúdo do diálogo mantendo a resposta `y/n` visível. Os atalhos de navegação e saída do rodapé continuam funcionando.

## controles da loja

- `↑/↓`: categoria; `←/→`: escolher item ou humor.
- `enter`: comprar o item da categoria atual; `ctrl+b`: comprar a seleção inteira.
- `ctrl+r`: voltar a prévia à aparência equipada.
- `y/n`: confirmar/cancelar compra; depois da compra, responder separadamente se deseja equipar.
- Com `-dev`, `x`: abrir o rascunho; `ctrl+o`: alternar disponível/adquirido/bloqueado; `ctrl+p`: editar preço.
- No preço, `←/→` muda 1 e `j/l` muda 10; `enter` confirma no rascunho e `esc` desfaz esse campo.
- `ctrl+s`: salvar o rascunho e continuar; `x`: salvar e sair; `esc`: cancelar o rascunho.

## saldo e tarefas no dev

As chaves de textos exclusivos do dev começam com `dev_`. Por exemplo,
`dev_price_help` orienta a edição de preço, `dev_wallet_help` orienta o saldo,
`dev_equipped_item_error` impede retirar um item em uso e
`dev_included_variant_error` impede retirar a variante incluída de uma base adquirida.
Textos compartilhados com compras normais, como `stale`, não recebem esse prefixo.

Dentro do editor dev de sala/perfil, `ctrl+m` edita seu saldo: setas mudam 1 e `j/l` mudam 50. Sem precisar de `enter`, `x` salva e sai ou `ctrl+s` salva e continua no dev. `enter` apenas confirma o valor no rascunho; `esc` cancela o campo.

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
