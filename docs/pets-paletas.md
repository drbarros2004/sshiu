# Paletas do pet

O desenho e as cores ficam separados:

- [`internal/pet/pet.png`](../internal/pet/pet.png): template com um marcador diferente por parte.
- [`internal/pet/palettes.json`](../internal/pet/palettes.json): nomes dos marcadores e cores finais.
- [`internal/pet/palettes.go`](../internal/pet/palettes.go): leitura e validação dos dados.
- [`internal/pet/render.go`](../internal/pet/render.go): aplica a paleta e monta os quadros.

Uma paleta vale para todos os humores. Só o desenho e a animação mudam com o humor.
As oito paletas vêm dos PNGs de referência em `internal/pet/testdata/`.

## Partes

| Nome | O que representa |
|---|---|
| `eyes_mouth` | olhos e boca |
| `body` | corpo |
| `outline` | contorno |
| `glow` | brilho do corpo |
| `tear` | lágrima |
| `stem` | haste do broto/flor |
| `leaves` | folhas do broto |
| `cheeks` | bochechas |
| `petals` | pétalas |
| `flower_center` | miolo da flor |
| `tongue` | língua |
| `particle` | partículas do cansaço/sono |
| `sparkles` | estrelinhas do radiante |

Os marcadores precisam ser diferentes entre si. As cores finais de partes diferentes podem
ser iguais: `particle` e `sparkles`, por exemplo, continuam independentes mesmo com a mesma cor.
Marcadores não precisam pertencer à tabela xterm: eles nunca aparecem na tela.

## Mudar uma cor

Em `palettes.json`, `defaults` define uma cor padrão para cada parte. Começa com as cores
do azul. Cada entrada de `palettes` lista em `colors` somente as diferenças em relação a esses
padrões; a paleta azul pode ficar com `colors: {}`.

Para trocar uma cor em todas as paletas que herdam o padrão, edite `defaults`.
Para trocar em apenas uma paleta, procure seu `name` em `palettes` e acrescente ou altere
a parte em `colors`. Para voltar ao padrão, remova essa parte de `colors`.
Por exemplo, o contorno do coral pode ser `"outline": "#AF5F5F"` ou
`"outline": "xterm:131"`; ambos indicam a mesma cor. HEX é o formato principal.
Os índices 0–15 dependem do tema do terminal; 16–255 têm a tabela xterm fixa.
Um HEX fora dessa tabela é aceito, mas um terminal de 256 cores pode aproximá-lo.
Os PNGs atuais contêm três aproximações conhecidas, preservadas no JSON:

| Paleta e parte | HEX do PNG | Resultado em 256 cores |
|---|---|---|
| verde, `flower_center` | `#FFE066` | `#FFD75F` |
| turquesa, `flower_center` | `#800000` | `#870000` |
| rosa, `sparkles` | `#F7F7F7` | `#FFFFFF` |

Essas conversões são verificadas em `TestPalette256Colors`. Uma nova cor que exija aproximação
precisa ser revisada e registrada no teste; o teste não muda a cor armazenada no JSON.

Não altere `id` para trocar uma cor. Embora os IDs antigos pareçam HEX, eles agora identificam
as escolhas já salvas e preservam a relação com os instrumentos da obra. Mantenha também a
ordem existente enquanto a orquestra usar essa ordem.

O PNG e o JSON são embutidos no programa: reinicie com uma compilação nova para ver mudanças.

## Adicionar um grupo

1. No template `pet.png`, pinte os pixels da nova parte com um HEX ainda não usado no template.
   Use pixels totalmente opacos ou transparentes, sem suavização nas bordas.
2. Em `parts` do JSON, cadastre um nome e o marcador, por exemplo `"hat": "#010203"`.
3. Em `defaults`, acrescente `hat` com a cor padrão. Nas paletas que precisarem de outra cor,
   acrescente a exceção em `colors`.
4. Rode `go vet ./...` e `go test ./...`, recompile e confira no app.

Não é preciso mudar funções em Go. Se faltar o cadastro, o app para ao iniciar com a cor e a
posição do pixel desconhecido. Se faltar a cor padrão, o erro informa a parte que falta em `defaults`.
Isso evita mostrar as cores fortes do template por acidente.

## Alterar desenhos e referências

A folha continua com 64×96 pixels: cada quadro tem 16×16, e as linhas são murcho, cansado,
normal, bem, radiante e dormindo. As primeiras cinco linhas usam dois quadros; o sono usa quatro.
Um desenho novo deve conservar esse formato; mudar tamanho ou quantidade de quadros exige
ajustar `render.go` também.

Os testes comparam todos os pixels e quadros das oito paletas com os PNGs em `testdata`.
Se a mudança visual for intencional, exporte as novas referências do editor e atualize esses
PNGs junto. Não regenere referências só para esconder uma diferença inesperada.

## Obra da sala

A obra conserva o sistema anterior de recoloração por matiz, isolado em `artwork_colors.go`.
As novas paletas por parte são usadas apenas pelo pet, pelas amostras de cor e pelo desenho ASCII.
