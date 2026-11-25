# Premissas de receita

> Status: hipóteses para o pitch; nada foi vendido nem testado com usuários pagantes. Os cenários são exploração, não previsão.

Valores em reais, nominais, por mês, **sem impostos, sem taxas de pagamento ou de loja, sem custos e sem cancelamentos**. "Pessoas ativas" são as que focaram pelo menos uma vez no mês.

## Entradas

| Entrada | Conservador | Base | Otimista | Tipo | De onde vem |
|---|---|---|---|---|---|
| Pessoas ativas no mês 12 | 500 | 2.000 | 10.000 | H | Piloto com turmas da Poli e depois boca a boca entre devs |
| Assinam o sshiu+ | 2% | 4% | 5% | H | Faixa comum de conversão freemium; não medida aqui |
| Preço do sshiu+ | R$ 9,90 | R$ 9,90 | R$ 9,90 | H | Abaixo do Focusmate Plus (US$ 8–12) e do Finch Plus (US$ 5,99–9,99) |
| Compram skins no mês | 3% | 5% | 6% | H | — |
| Gasto médio de quem compra skin | R$ 8 | R$ 10 | R$ 12 | H | Itens a R$ 4,90–14,90 |
| Empresas parceiras | 1 | 5 | 20 | H | — |
| Assentos por empresa | 15 | 20 | 25 | H | — |
| Preço por assento | R$ 15 | R$ 15 | R$ 15 | H | O Focusmate Team começa em US$ 50/mês por time |

## Fórmula

```
sshiu+    = pessoas × % assinantes × 9,90
skins     = pessoas × % compradores × gasto médio
empresas  = empresas × assentos × 15
mês       = sshiu+ + skins + empresas
ano       = mês × 12   (o mês 12 repetido; o ano real seria menor, porque a base cresce ao longo do ano)
```

## Resultado (estimativa calculada)

| Cenário | sshiu+ | Skins | Empresas | Por mês | Por ano |
|---|---|---|---|---|---|
| Conservador | R$ 99 | R$ 120 | R$ 225 | R$ 444 | R$ 5.328 |
| Base | R$ 792 | R$ 1.000 | R$ 1.500 | R$ 3.292 | R$ 39.504 |
| Otimista | R$ 4.950 | R$ 7.200 | R$ 7.500 | R$ 19.650 | R$ 235.800 |

O que pesa: no cenário base, as empresas já são 46% da receita com só 5 clientes. É por isso que o pitch fecha no B2B e na integração com issues.

## Fora da conta

- Impostos, taxas do meio de pagamento e cancelamentos.
- Custo do servidor: a sala manda ~90 KB/s por pessoa (medido em 30/10/2026), então a banda cresce com quem está conectado ao mesmo tempo, não com o total de contas.
- Arte das skins: personagens pagos precisam de arte própria ou licenciada ([ADR 0005](../docs/adr/0005-arte-de-terceiros-so-com-licenca-permissiva.md)). O personagem "claude" da loja não deve ser vendido sem autorização da marca.
- Receita de parcerias além do assento (implantação, integração sob medida).
