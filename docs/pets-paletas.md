# Personagens, cores e chapéus

Os pets usam os PNGs coloridos em [assets/pets](../assets/pets/README.md), sem recoloração por humor.
Os templates de marcadores são auxiliares de edição e não são desenhos selecionáveis.

- `characters.json`: personagens, variantes, nomes por idioma, amostras de cor e composição.
- `hats.json`: tipos/modelos de chapéu, nomes e ordem. Todos ficam disponíveis nesta versão.
- `sprite-layout.json`: tamanho dos quadros, ordem dos humores, quantidade e duração dos quadros.
  O arquivo com o mesmo nome dentro de cada personagem sobrescreve somente os campos presentes.
- `moods.json`: nomes padrão; o arquivo do personagem contém apenas exceções por idioma.

Para mudar as cores do desenho, edite os PNGs da variante e suas camadas. `displayColor` muda
somente a amostra na interface. `artworkColor` liga a variante à cor/instrumento da obra e
não muda os pixels do pet. Consulte os caminhos e exemplos no README dos assets.

## Composição e posições

O aplicativo compõe as camadas na ordem de `withoutHat` ou `withHat`. Assim, tirar a planta
quando há chapéu é uma decisão do catálogo da fruta, não uma exceção escondida no código.
`hatY` posiciona a linha inferior da caixa completa do chapéu, contando a partir do pé do
quadro. O topo pode ultrapassar o corpo: a composição reserva essa altura sem cortar o desenho.
O perfil e a sala usam a mesma composição; numa fileira, os corpos ficam alinhados pela base.

`frameDurationMs` é a duração de um quadro em milissegundos. A lista acompanha `moodOrder`.
A tela é atualizada a cada 500 ms; esse também é o mínimo aceito para a duração de um quadro.
Para acrescentar quadros, atualize `frameCounts`, a folha PNG, as camadas e os respectivos `hatY`.
Os seis humores continuam sendo os mesmos estados de energia/presença.

## Personalização

No perfil, `c` abre as opções. ↑/↓ seleciona uma linha e ←/→ altera seu valor.
Enter salva a aparência; Esc cancela. A última linha permite apenas visualizar outro humor.
Ela nunca altera energia, presença ou o humor salvo.

Durante uma edição, cada personagem lembra sua cor e cada chapéu lembra seu modelo.
A primeira visita a outra opção começa na primeira variante do catálogo. Trocar só o
personagem mantém o chapéu escolhido. Ao sair, essa memória temporária é descartada;
a aparência efetivamente salva continua selecionada na próxima abertura.

No modo dev, Enter confirma o rascunho e o salvamento final continua sendo Ctrl+S ou x.
Na sala dev, Ctrl+L abre as mesmas opções de aparência para o membro selecionado.

## Cores do terminal e dados antigos

Use cores xterm nos PNGs para reduzir aproximações em terminais de 256 cores.
Os dois casos conhecidos preservados estão no README dos assets. `displayColor` aceita HEX
ou `xterm:0` a `xterm:255`; os primeiros 16 índices dependem do tema do terminal.

Perfis antigos continuam como fruta, na cor já escolhida, sem chapéu. As novas escolhas
salvam IDs de personagem, variante, chapéu e modelo, sem alterar energia ou histórico.
A recoloração das obras permanece independente em `internal/pet/artwork_colors.go`.

## Barra de energia do perfil

Em [energy_colors.json](../internal/tui/energy_colors.json), use uma string HEX como
`"#FF8787"` ou um índice explícito como `"xterm:210"`. A barra converte HEX para a
tabela xterm de 256 cores.

- `bar` define o preenchimento e a cor do nome do humor **atual**. `bar.radiante` vale
  durante radiante. Uma cor em `bar.all` prevalece sobre os valores individuais.
- `unfilled` é a cor de fundo da parte vazia, inclusive na célula do traço.
  Vale para todos os humores. Com `null`, usa o fundo discreto do tema na tabela xterm.
- `markers` define os traços que aparecem **antes** de alcançar cada humor. Por
  exemplo, `before_radiante` colore a marca de 80 enquanto o pet está bem.
  `null` em uma chave usa a cor do humor de destino em `bar`.
- Uma cor em `markers.all` prevalece sobre todos os traços individuais. Use
  `"all": "#FF8787"` dentro de `markers` para testar vermelho em todos eles.
  Volte para `null` para usar as cores individuais ou herdadas.

Recompile a demo para ver os ajustes.
