# Pets em pixel art com licença livre

> **É uma sugestão, não uma decisão.** Esta pesquisa serve de ponto de partida: quem for trabalhar nessa parte deve conferir as fontes e as licenças de novo e procurar outras referências antes de escolher.

Pesquisa da issue [#21](https://github.com/drbarros2004/ssh-focus-app/issues/21), feita em 31/10/2026. Cada licença foi conferida na página do autor (OpenGameArt, itch.io, kenney.nl ou o repositório), não em agregadores. Nada foi baixado: tamanhos, contagens e animações vêm do texto das páginas. Quando a página não diz, está escrito "não informado".

Regra do [ADR 0005](../adr/0005-arte-de-terceiros-so-com-licenca-permissiva.md): só MIT, Apache-2.0 ou CC0, creditados em `CREDITS.md`.

## Resumo

- **Melhores pacotes para aproveitar (CC0, 16×16):** [Tiny Creatures](https://opengameart.org/content/tiny-creatures) (mais de 50 animais e mais de 100 monstros num estilo só, sem animação), [Ninja Adventure](https://pixel-boy.itch.io/ninja-adventure-asset-pack) (mais de 30 monstros animados, mais animais) e [GB Funky Fauna](https://opengameart.org/content/gb-funky-fauna) (8 bichos em 4 cores, fáceis de recolorir).
- **Tamanho:** 16×16 px é o teto nos cards. O card tem 18 colunas com a borda ([ui.go:50](../../internal/tui/ui.go#L50)), ou seja, 16 por dentro, e com o `▀` um sprite de 16×16 vira 16 colunas × 8 linhas: é exatamente o pet de hoje. 32×32 só cabe no perfil.
- **Custo cheio:** 5 humores × 2 quadros = 10 quadros por pet, ou de 100 a 150 quadros para 10 a 15 pets.
- **Alternativa barata:** 1 sprite por pet (2 com a respiração) e o humor como efeito comum feito no código: saturação da cor, ritmo da respiração (já existe) e um acessório sobreposto, desenhado uma vez só. Para 15 pets são de 15 a 30 quadros, mais uns 4 acessórios.

## Pacotes que dá para aproveitar (CC0)

| Pacote | Autor | Licença conferida | Tamanho | Bichos | Poses e animações |
|---|---|---|---|---|---|
| [Tiny Creatures](https://opengameart.org/content/tiny-creatures) | Clint Bellanger, com Kenney | CC0 ("giving attribution is not required") | 16×16, contorno grosso | 180 sprites: "Over 100 monsters", "Over 50 animals" | Estáticos ("don't need a lot of animations"). Combina com o Tiny Dungeon e o Tiny Town da Kenney. |
| [Ninja Adventure](https://pixel-boy.itch.io/ninja-adventure-asset-pack) | pixel-boy | CC0 ("Attribution is not required but appreciated") | tag 16×16 (a página não diz o tamanho) | "30+ Monster with animations", "9 Bosses", animais (quantos, não informado) | Animados (a página não lista os quadros). Animais desde o update #8. |
| [GB Funky Fauna](https://opengameart.org/content/gb-funky-fauna) | pebonius | CC0 ("its cc0 do whatever") | 16×16, paleta de 4 cores do Game Boy | 8: besouro, raposa, rato, gato, cachorro, sapo, pássaro, esquilo | Animação não informada. Com 4 cores, trocar a paleta por humor é trivial. |
| [Surtizens](https://opengameart.org/content/surtizens) | pebonius | CC0 ("it's cc0 go wild") | 16×16 | cachorro, sapo, pato (mais orc, goblin, minotauro e 3 humanos) | Animação não informada. |
| [Forest Animal Pack, amostra grátis](https://opengameart.org/node/185579) | inkBubi | CC0 (só a amostra; o [pacote completo](https://inkbubi.itch.io/seasons-of-forest-animal-pack) é pago) | 16×16 | 5 animais da floresta | idle e corrida em 4 direções, com um `duration.txt` de velocidades |
| [16x16 Animated Critters](https://opengameart.org/content/16x16-animated-critters) | patvanmackelberg | CC0 ("no strings attached") | 16×16 | coelho, pintinho com ovo, aranha com ovo pulsante, slime | Animados (quadros não informados). |
| [Animal Icons](https://ydo4ki.itch.io/animalicons) | Ydo4ki | CC0 ("cc0 public domain") | 16×16 | 15 | Estáticos. |
| [Cute Little Things](https://dreamgryphon.itch.io/cute-little-things) | DreamGryphon | CC0 | 8×8 e 16×16 | não informado (tag "Animals") | "Sprites (2 frames)": o mesmo formato da respiração de hoje. |
| [Tiny Farm Animals](https://tnypixels.itch.io/tiny-farm-animals-pixel-art-asset-pack) | Tnypixels | CC0 | não informado | 8 espécies (pintinho, pato, galinha, coelho, vaca, porco, ovelha, cavalo), uns 13 sprites | Animação não informada. |
| [Pet Cats Pack](https://luizmelo.itch.io/pet-cat-pack) | LuizMelo | CC0 ("Credit is not required") | de 20×14 a 22×15 | 6 gatos | 12 animações: idle (10 quadros), andar, correr, miar, deitar, coçar, dormir, sentar, lamber, espreguiçar. É mais largo que o card (veja abaixo). |
| [1-Bit Pack](https://kenney.nl/assets/1-bit-pack) e [Tiny Dungeon](https://kenney.nl/assets/tiny-dungeon) | Kenney | CC0 | 16×16 | 1078 e 130 arquivos; se há bichos, a página não diz | Estáticos. Valem como base de estilo para o Tiny Creatures. |

### Para os humores (CC0)

| Pacote | Autor | O que tem | Uso |
|---|---|---|---|
| [16x16 emotes for RPGs and digital pets](https://opengameart.org/content/16x16-emotes-for-rpgs-and-digital-pets) | Tomcat94 | 11 balões de emoção em 3 orientações, mais "!" e "?". Feito pensando em Tamagotchi. "Credit is not necessary." | Acessório de humor comum a todos os pets. |
| [Emotes Pack](https://kenney.nl/assets/emotes-pack) | Kenney | 480 arquivos de balões de emoção (se há versão em pixel art e o tamanho, a página não diz) | Idem, a conferir. |
| megupets: [#008 Loobits](https://opengameart.org/content/megupet-008-loobits-public-domain), [#012 Piddles](https://opengameart.org/content/megupet-012-piddles-public-domain), [Shuaagotchi](https://opengameart.org/content/shuaagotchi-crab-virtual-pet) | megupets | pets virtuais CC0 com sprites idle, feliz e triste (o #012 traz idle, triste e "GUAA"). Tamanho não informado; arquivos de 70 a 90 KB, provavelmente maiores que 16×16. | Referência de como um bichinho de Tamagotchi mostra o humor. |

### Grandes demais para o card (CC0)

- [little weirdos](https://opengameart.org/content/little-weirdos) (pebonius): criaturas em 32×32. Só para o perfil, ou reduzidas à mão.

## CC-BY (exige crédito; fora do ADR 0005 hoje)

- [DawnLike](https://opengameart.org/content/dawnlike-16x16-universal-rogue-like-tileset-v181) (DragonDePlatino): CC-BY 4.0, 16×16, um tileset de roguelike enorme, com centenas de monstros e bichos. A exigência de crédito é simples, mas dupla: o autor e o DawnBringer, criador da paleta ("you must accredit DawnBringer"). Usar pede uma emenda ao ADR 0005 para aceitar CC-BY com crédito em `CREDITS.md`.

## Só inspiração (licença que não serve)

As licenças "royalty free, sem redistribuição" não são MIT, Apache-2.0 nem CC0, e embutir o PNG no repositório ou no binário é redistribuir o arquivo original.

| Pacote | Por que não serve | O que ensina |
|---|---|---|
| [Pixel Art Animals 16x16](https://iiarielsl.itch.io/pixel-art-animals-16x16) (iiarielsl) | "Redistribution, resale, or modification for resale is not permitted" | 10 bichos em 16×16 com idle e andar/voar, 2 quadros cada: o nosso formato exato. |
| [Animal Asset Pack](https://deepdivegamestudio.itch.io/animalassetpack) (DeepDiveGameStudio) | proíbe "redistribute the assets in original or modified form" | 45 bichos em 16×16, cada um com idle. |
| [Idle Animals Free](https://toffeecraft.itch.io/idle-animals-free) (ToffeeCraft) | "Not redistribute or resell" | 10 bichos em idle. |
| [Animals Character Asset Pack](https://comfymattres.itch.io/animals-character-asset-pack) (ComfyMattres) | licença própria (proíbe revenda e NFT) | 4 bichos em 16×16, com idle e andar em 4 direções. |
| [Free Pixel Animal Pack](https://pixel-wanderer.itch.io/free-pixel-animal-pack-16x16-weekly-updates) (Pixel Wanderer) | a versão grátis é só para uso não comercial | 8 bichos de fazenda; a versão paga tem as animações de comer e dormir. |
| [Pixanimals](https://zeeshan1.itch.io/pixanimals-32x32-pet-animals-2d) (Zeeshan) | proíbe redistribuir os arquivos | 4 pets em 32×32 (o gato em 48×48), com idle de piscar em 2 quadros e dormir, machucado e morte: um mapa pronto de poses por humor. |
| [Mono Pets 16](https://assertchris.itch.io/mono-pets-16) (assertchris) | a página não publica licença | 35 ícones e animações para um jogo no estilo Tamagotchi. |
| [Cute cats](https://hutsul.itch.io/cute-cats-pixel-art) (Hutsul) | o campo diz CC0, mas o texto diz "You may not redistribute it": contraditório, melhor evitar | 2 gatinhos. |
| [vscode-pets](https://github.com/tonybaloney/vscode-pets) | o código é MIT, mas a arte vem de pacotes do itch.io de vários autores, cada um com a sua licença (créditos no README) | A referência mais próxima do nosso produto: uns 19 pets animados num editor. |

## Que tamanho cabe no card

O card tem `Width(18)` e borda arredondada ([ui.go:50](../../internal/tui/ui.go#L50)); no Lip Gloss v2 a largura inclui a borda, então sobram 16 colunas. Com o `▀`, 1 coluna = 1 pixel de largura e 1 linha = 2 pixels de altura, e como a célula do terminal tem mais ou menos 1:2, o pixel sai quadrado ([render.go](../../internal/pet/render.go)).

| Sprite | No terminal | No card |
|---|---|---|
| 8×8 | 8 col × 4 linhas | Cabe, mas o rosto tem 2 ou 3 pixels: o humor não se lê. |
| 12×12 | 12 × 6 | Cabe; o rosto fica apertado. |
| **16×16** | **16 × 8** | **Cabe exato. É o de hoje e o tamanho da maioria dos pacotes.** |
| 20×14 a 22×15 (gatos do LuizMelo) | 20 a 22 × 7 a 8 | Estoura: precisa cortar para 16 ou alargar o card para 24 (com 80 colunas caem de 4 para 3 cards por linha, por [room.go:220](../../internal/tui/room.go#L220)). |
| 32×32 | 32 × 16 | Só no perfil, onde o `RenderBig` já mostra o pet em 32×16. |

Cuidados de legibilidade:

- **Contorno.** O Tiny Creatures e a Kenney usam contorno escuro grosso; num tema escuro ele se confunde com o fundo do terminal e a silhueta some. O pet de hoje contorna com um tom mais escuro da própria cor, que funciona nos dois temas. Vale recolorir o contorno ou conferir nos dois temas.
- **256 cores.** Paletas sutis (DawnBringer de 32 cores, degradês) perdem tons quando o terminal tem só 256 cores. Prefira poucas cores bem saturadas e confira rasterizando a tela num PNG, como pede o `AGENTS.md`.
- **Pose de frente.** Os pacotes de RPG visto de cima trazem o bicho de frente, que lê bem num card. Os de vista lateral (gatos do LuizMelo) também servem.
- **Plano B em texto.** Cada pet novo também precisaria de uma arte ASCII de 5 × 2 quadros (`art` em [render.go](../../internal/pet/render.go)). O mais barato é manter a bolinha atual como plano B de todos os pets.

## Quanto desenho cada pet exige

| Abordagem | Quadros por pet | 10 pets | 15 pets | Como o humor aparece |
|---|---|---|---|---|
| **Cheia** (a de hoje) | 5 humores × 2 = 10 | 100 | 150 | Desenho próprio por humor, como no `pet.png`. |
| **Meio-termo** | 3 rostos (triste, neutro, feliz) × 2 = 6 | 60 | 90 | O rosto dá a direção; cor e acessório separam os 5 humores. |
| **Barata** | 1 sprite + 1 quadro de respiração = 2 | 20 + 4 acessórios | 30 + 4 acessórios | Tudo por efeito comum (abaixo). |
| **Mínima** | 1 | 10 + 4 | 15 + 4 | Respiração feita no código: o sprite sobe e desce 1 pixel. |

Com um pacote CC0 como base, o sprite base sai de graça, e o trabalho vira recolorir, ajustar o contorno e, na abordagem barata, desenhar só o quadro de respiração. O Tiny Creatures sozinho dá 10 a 15 bichos num estilo coerente; misturar pacotes pede uma paleta em comum.

## Alternativas mais baratas: o humor como efeito comum

Tudo isso roda uma vez no `init` de `internal/pet`, com a stdlib (`image`, `image/color`), como o `halfBlocks` já faz:

1. **Respiração.** Já existe: de 1 a 4 ticks por quadro conforme o humor ([regras](../regras.md#ritmo-da-tela), `moodPeriod` em [render.go](../../internal/pet/render.go)). Na abordagem mínima, o quadro 2 é o sprite 1 pixel mais baixo, feito no código.
2. **Cor, sem trocar o matiz.** Hoje cada humor tem uma cor (o pet murcho é cinza, o radiante é amarelo). Com bichos de cores próprias, trocar o matiz apaga a identidade do pet; melhor mexer na saturação e no brilho: radiante e bem com a cor cheia, normal um pouco menos, cansado pela metade e murcho quase cinza. A cor do humor continua no texto do card (`MoodColor`). Num pacote de 4 cores, como o GB Funky Fauna, dá para trocar a paleta inteira por humor.
3. **Acessório sobreposto.** Uns 4 desenhos de 16×16 com fundo transparente, desenhados uma vez para todos: brilhos no radiante (o pet atual já tem `*` e pontos brancos), "z" no cansado (também já existe), uma gota ou lágrima no murcho e nada no normal. Os balões do [Tomcat94](https://opengameart.org/content/16x16-emotes-for-rpgs-and-digital-pets) (CC0) servem de pronto. Isso pede que o sprite base deixe uns 3 pixels livres nos cantos de cima, ou o acessório cobre a cabeça.
4. **Postura.** No murcho, o sprite desce 2 pixels (o bicho "deitado"). Combina com a ideia do zumbi para o murcho no `IDEIAS.md`.

Recomendação: começar pela abordagem barata com o Tiny Creatures (ou o GB Funky Fauna) como base, num protótipo com 3 bichos e os 5 humores por efeito, e ver no terminal se os humores se distinguem antes de desenhar rostos.

## Fontes

Todas lidas em 31/10/2026.

- OpenGameArt: [Tiny Creatures](https://opengameart.org/content/tiny-creatures), [16x16 Animated Critters](https://opengameart.org/content/16x16-animated-critters), [Forest Animal Pack, amostra](https://opengameart.org/node/185579), [GB Funky Fauna](https://opengameart.org/content/gb-funky-fauna), [Surtizens](https://opengameart.org/content/surtizens), [little weirdos](https://opengameart.org/content/little-weirdos), [DawnLike](https://opengameart.org/content/dawnlike-16x16-universal-rogue-like-tileset-v181), [16x16 emotes](https://opengameart.org/content/16x16-emotes-for-rpgs-and-digital-pets), [megupet #008](https://opengameart.org/content/megupet-008-loobits-public-domain), [megupet #012](https://opengameart.org/content/megupet-012-piddles-public-domain), [Shuaagotchi](https://opengameart.org/content/shuaagotchi-crab-virtual-pet)
- itch.io: [Ninja Adventure](https://pixel-boy.itch.io/ninja-adventure-asset-pack), [Animal Icons](https://ydo4ki.itch.io/animalicons), [Cute Little Things](https://dreamgryphon.itch.io/cute-little-things), [Tiny Farm Animals](https://tnypixels.itch.io/tiny-farm-animals-pixel-art-asset-pack), [Pet Cats Pack](https://luizmelo.itch.io/pet-cat-pack), e os da tabela "Só inspiração"
- Kenney: [1-Bit Pack](https://kenney.nl/assets/1-bit-pack), [Tiny Dungeon](https://kenney.nl/assets/tiny-dungeon), [Tiny Town](https://kenney.nl/assets/tiny-town), [Emotes Pack](https://kenney.nl/assets/emotes-pack)
- GitHub: [vscode-pets](https://github.com/tonybaloney/vscode-pets) (README, seção Credits)
- Lip Gloss v2.0.6, `Style.Width` em `set.go`: a largura inclui a borda e o padding.
