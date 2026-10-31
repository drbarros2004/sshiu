# Wallpapers ASCII: referências e técnicas

> **É uma sugestão, não uma decisão.** Esta pesquisa serve de ponto de partida: quem for trabalhar nessa parte deve conferir as fontes e as licenças de novo e procurar outras referências antes de escolher.

Pesquisa da issue [#19](https://github.com/drbarros2004/ssh-focus-app/issues/19), feita em 31/10/2026. A pergunta: que referências de arte ASCII animada existem com licença permissiva, e que técnicas deixam um wallpaper do modo jardim bonito no terminal, em 256 cores?

Regra de licença ([ADR 0005](../adr/0005-arte-de-terceiros-so-com-licenca-permissiva.md)): só MIT, Apache-2.0 ou CC0 servem para aproveitar código ou arte. GPL, LGPL e projetos sem licença entram só como **inspiração**. Toda licença abaixo foi conferida no arquivo LICENSE (ou COPYING/COPYRIGHT) do próprio repositório, ou na página oficial do projeto.

## Resumo

- **A técnica recomendada é a do Ghostty:** uma rampa de caracteres por brilho, com **2 a 4 cores fixas por wallpaper**, em que o brilho fica no caractere e não na cor. O fundo é o do terminal. Fica com cara de ASCII, as cores ficam apagadas sem esforço e é a opção mais barata em banda.
- **O meio bloco `▀`** (o do pet e do brainrot) fica para sprites pequenos e parados, como o carro da estrada. Ele é a técnica mais cara: cada célula pode levar duas cores.
- **Braille** só serve para traços finos, como chuva fina, contornos e estrelas. Cada célula tem uma cor só e os pontos estão acesos ou apagados, então ele não serve para imagens.
- **O fantasma do ghostty.org** é um vídeo convertido por um script de shell (MIT) que separa o azul do branco por distância de cor e, dentro de cada cor, troca o brilho por um de 10 caracteres. O mesmo cabe no nosso caminho de quadros PNG (`internal/tui/brainrot.go`), trocando `halfBlockRows` por uma função de rampa. O fantasma em si é só inspiração.
- **O asciify-engine** (MIT, sugerido pelo usuário) usa a mesma rampa por brilho, com dithering Bayer 4×4 e fundos procedurais "de material" (um campo de brilho por célula, uma cor só). O algoritmo cabe em umas 40 linhas de Go, e os fundos dele são boas bases para wallpapers novos.
- **Candidatos:** mar à noite, dunas sob a lua, aurora, aquário calmo e um objeto renderizado por nós no estilo Ghostty.

## Referências

| Referência | Licença (onde foi conferida) | O que aproveitar |
|---|---|---|
| [ghostty.org](https://ghostty.org/), código em [ghostty-org/website](https://github.com/ghostty-org/website) | MIT ([LICENSE](https://github.com/ghostty-org/website/blob/main/LICENSE), "Copyright (c) 2024 Ghostty") | A técnica e o script de conversão (MIT). O fantasma é o símbolo do Ghostty: **só inspiração**, como a issue já previa. |
| [asciify-engine](https://github.com/ayangabryl/asciify-engine) | MIT ([LICENSE](https://github.com/ayangabryl/asciify-engine/blob/main/LICENSE), "Copyright (c) 2026 ayangabryl"). Dependências com licença própria em [THIRD_PARTY.md](https://github.com/ayangabryl/asciify-engine/blob/main/THIRD_PARTY.md), entre elas Mediabunny (MPL-2.0), que não nos interessa. | O algoritmo de conversão (rampa, curva de contraste, normalização, Bayer 4×4), as rampas medidas por tinta e os fundos procedurais (dune, aurora, wave, stars...) como base de wallpapers. |
| [ascii-splash](https://github.com/reowens/ascii-splash) | MIT ([LICENSE](https://github.com/reowens/ascii-splash/blob/main/LICENSE)) | Já inspira o céu noturno. Tem cenas prontas que viram candidatos: Ocean Beach, Aquarium, Night Sky com aurora. |
| [play.core](https://github.com/ertdfgcvb/play.core) (Andreas Gysin) | Apache-2.0 ([LICENSE](https://github.com/ertdfgcvb/play.core/blob/master/LICENSE)) | Dezenas de programas curtos do tipo "função por célula → caractere": [plasma](https://github.com/ertdfgcvb/play.core/blob/master/src/programs/demos/plasma.js), campos de distância ([sdf/](https://github.com/ertdfgcvb/play.core/tree/master/src/programs/sdf)), ondas. É o mesmo modelo das nossas cenas. |
| [notcurses](https://github.com/dankamongmen/notcurses) | Apache-2.0 ([COPYRIGHT](https://github.com/dankamongmen/notcurses/blob/master/COPYRIGHT)) | Não é arte: é a melhor fonte primária sobre os "blitters" (meio bloco, quadrantes, sextantes, braille) e o que cada um perde. Ver [`notcurses_visual(3)`](https://github.com/dankamongmen/notcurses/blob/master/doc/man/man3/notcurses_visual.3.md). |
| [ascii-image-converter](https://github.com/TheZoraiz/ascii-image-converter) | Apache-2.0 ([LICENSE.txt](https://github.com/TheZoraiz/ascii-image-converter/blob/master/LICENSE.txt)) | Conversor de imagem para ASCII em Go, com rampa simples e complexa, braille com limiar e dithering, cor e cinza. Serve de referência de código em Go, mas a nossa conversão é pequena demais para valer a dependência. |
| [Paul Bourke, "Character representation of grey scale images"](https://paulbourke.net/dataformats/asciiart/) | Página de referência (só fatos: as rampas) | As rampas clássicas: ` .:-=+*#%@` (10 níveis) e a de 70 caracteres. Avisa que a densidade varia com a fonte e que a célula é retangular. |
| [Terminal Rain](https://github.com/rmaake1/terminal-rain-lightning), [lavat](https://github.com/AngelJumbo/lavat), [bruh.animate](https://github.com/ethanlchristensen/bruh.animate) | MIT, MIT, Apache-2.0 (LICENSE de cada um) | Já são créditos da chuva, da lava e do céu ([`CREDITS.md`](../../CREDITS.md)). |
| [pipes.sh](https://github.com/pipeseroni/pipes.sh) | MIT ([LICENSE](https://github.com/pipeseroni/pipes.sh/blob/master/LICENSE)) | Licença boa, mas os canos enchem a tela até ela ser limpa, e acumular é justamente o que não queremos. Ficou de fora. |
| Kenney ([kenney.nl/support](https://kenney.nl/support)) | CC0 ("all game assets on the asset pages are public domain licensed") | Modelos 3D e sprites para renderizar e converter (carro, xícara, árvores low-poly). |
| Poly Haven ([polyhaven.com/license](https://polyhaven.com/license)) | CC0 (HDRIs, texturas e modelos; logos e renders de usuários ficam de fora) | Céus HDRI e modelos para um render nosso no estilo Ghostty. |
| Blender ([blender.org/about/license](https://www.blender.org/about/license/)) | O programa é GPL, mas "What you create with Blender is your sole property" | O render que fizermos é nosso. Dá para usar no app sem problema de licença. |
| [OpenGameArt, "The Complete Guide to Cows"](https://opengameart.org/content/the-complete-guide-to-cows) (coletânea de 1991 citada no ADR 0005) | CC0, segundo a busca. **Não conferida nesta pesquisa:** o site respondeu 403. | Peças estáticas. Conferir a página antes de usar. |
| [chafa](https://github.com/hpjansson/chafa) | LGPL-3.0 ([COPYING](https://github.com/hpjansson/chafa/blob/master/COPYING)) | **Só inspiração.** É o melhor conversor de imagem para terminal que existe (escolhe o símbolo que melhor cobre cada célula). |
| [cbonsai](https://gitlab.com/jallbrit/cbonsai) | GPL-3.0 ([LICENSE](https://gitlab.com/jallbrit/cbonsai/-/raw/master/LICENSE)) | **Só inspiração** (ADR 0005), para a árvore. |
| asciiquarium | GPL, segundo o ADR 0005. O espelho no GitHub não tem arquivo de licença que a API reconheça. | **Só inspiração**, para o aquário. |
| asciiart.eu, Joan Stark (jgs) | Sem licença publicada | **Só inspiração** (ADR 0005). |

### Como o fantasma do ghostty.org é feito

O código do site é público e MIT. O fantasma é um **vídeo convertido em quadros de texto antes do deploy**, não um render feito no navegador:

1. **O vídeo de origem** está no repositório: [`bin/video-to-terminal/video.mp4`](https://github.com/ghostty-org/website/tree/main/bin/video-to-terminal) (1,4 MB). Ele é convertido por [`video-to-terminal.sh`](https://github.com/ghostty-org/website/blob/main/bin/video-to-terminal/video-to-terminal.sh):
   - o `ffmpeg` extrai os quadros em PNG com 100 colunas, a 24 fps;
   - o ImageMagick achata a altura para 0,44 da original (`FONT_RATIO`), para compensar a célula retangular;
   - cada pixel é classificado pela **distância de cor** a duas cores-alvo, um azul (`0,0,230`) e um branco (`215,215,215`), cada uma com a sua tolerância. Pixel longe das duas vira espaço: é assim que o fundo some;
   - dentro de cada cor, a luminância relativa (`0.2126 R + 0.7152 G + 0.0722 B`) é reescalada numa **janela própria** (o azul de 10 a 21, o branco de 165 a 255) para 10 níveis;
   - cada nível vira um caractere de `·~ox+=*%$@`, e os trechos azuis ganham `<span class="b">`.
2. **O resultado** são [235 arquivos `.txt`](https://github.com/ghostty-org/website/tree/main/terminals/home/animation_frames) de cerca de 6 KB cada, lidos no build por [`terminal-data.tsx`](https://github.com/ghostty-org/website/blob/main/src/app/terminal-data.tsx).
3. **O componente** [`animated-terminal`](https://github.com/ghostty-org/website/blob/main/src/components/animated-terminal/index.tsx) troca o quadro com `requestAnimationFrame`, a 31 ms por quadro ([`HomeContent.tsx`](https://github.com/ghostty-org/website/blob/main/src/app/HomeContent.tsx)), numa grade de 100×41. O laço tem cerca de 7 s. A animação pausa quando a janela perde o foco e respeita `prefers-reduced-motion`. A classe `.b` pinta o azul com a cor da marca; o resto usa a cor padrão do texto.

O que faz ficar bonito: **só duas cores**, cada uma com a rampa inteira dentro da faixa de brilho que importa para ela. É isso que dá o contorno azul nítido e o corpo branco com volume. A sensação de "render" vem do vídeo de origem, que é um 3D bem iluminado.

### O que o asciify-engine faz

É uma biblioteca TypeScript para a web, desenhada em `<canvas>`/WebGL2, então nada roda direto no terminal. A técnica, porém, é a mesma que queremos:

- **Conversão** ([`src/core/renderer.ts`](https://github.com/ayangabryl/asciify-engine/blob/main/src/core/renderer.ts), [`src/core/utils.ts`](https://github.com/ayangabryl/asciify-engine/blob/main/src/core/utils.ts)):
  - luma Rec.601 (`0.299 R + 0.587 G + 0.114 B`);
  - brilho e contraste numa curva clássica (`factor = 259(c+255) / (255(259−c))`);
  - `normalize` opcional, que estica o brilho do quadro para 0–255;
  - **dithering ordenado Bayer 4×4** opcional;
  - índice `floor(lum/255 · (n−1))` na rampa.

  Há também chroma key (tira um fundo verde ou azul, como o Ghostty faz por distância de cor) e os modos de cor `grayscale`, `fullcolor`, `matrix` e `accent`.
- **Rampas** ([`src/studio/characters.ts`](https://github.com/ayangabryl/asciify-engine/blob/main/src/studio/characters.ts)), "ordered by measured glyph ink":
  - `standard` ` .:-=+*#%@` (a de Bourke);
  - `detailed` ` .,:;-~=+*xX#%@`;
  - `blocks` ` ░▒▓█`;
  - `braille` ` ⠁⠃⠇⡇⣇⣧⣷⣿`;
  - e outras.
- **Fundos "de material"** ([`src/kit/backgrounds/material.ts`](https://github.com/ayangabryl/asciify-engine/blob/main/src/kit/backgrounds/material.ts)): cada fundo é uma função `sample(u, v, t) → brilho`, mapeada na rampa com **uma cor só**. Nas palavras do código, "character density carries the material, not hue". O [`dune.ts`](https://github.com/ayangabryl/asciify-engine/blob/main/src/kit/backgrounds/dune.ts), por exemplo, é uma fórmula de 4 linhas: cristas de areia com a luz andando a ~0,09 rad/s. O [`aurora.ts`](https://github.com/ayangabryl/asciify-engine/blob/main/src/kit/backgrounds/aurora.ts) usa a rampa ` .:¦|`.

**Dá para aproveitar no nosso pipeline?**

- **A técnica, sim.** O algoritmo de conversão cabe em ~40 linhas de Go e substitui `halfBlockRows` para quem quiser o visual de ASCII. Os fundos de material têm o mesmo formato das nossas cenas (`draw(c, t, dt)`) e podem ser reescritos em Go com crédito no `CREDITS.md`, como fizemos com o lavat.
- **O modo de cor**: o `fullcolor` sozinho dá cores demais para 256. Melhor seguir o `accent`/material: uma ou duas cores fixas.
- **A arte, não há o que aproveitar**: a biblioteca não traz imagens de cena, só fórmulas e componentes de interface.
- **O Studio exporta GIF/MP4** de ASCII já desenhado em pixels. Isso não ajuda: queremos converter em Go, a partir do render original.

## As técnicas no terminal, em 256 cores

Base comum: o ssh não repassa `COLORTERM`, então tudo sai na paleta de 256 cores do xterm. No [`256colres.pl`](https://github.com/ThomasDickey/xterm-snapshots/blob/master/256colres.pl) do xterm, essa paleta tem um **cubo 6×6×6** com níveis 0, 95, 135, 175, 215 e 255 por canal (índices 16–231) e uma **rampa de 24 cinzas** de 8 a 238, de 10 em 10 (índices 232–255).

| | Rampa por brilho | Meio bloco `▀` | Braille |
|---|---|---|---|
| Pixels por célula | 1 (o brilho vira o caractere) | 2 (em cima e embaixo), quadrados | 8 (2×4), pontos acesos ou apagados |
| Cores por célula | 1 (frente); o fundo é o do terminal | 2 (frente e fundo) | 1 (frente) |
| Cara | ASCII de verdade, como o Ghostty. Lê bem de longe e combina com o resto da interface. | Pixel art ou vídeo: bonito para sprites, mas parece imagem, não texto | Traço fino e delicado. O notcurses avisa que "doesn't tend to work out very well for images". |
| Cores apagadas | Natural: com 2 a 4 cores escolhidas à mão, o brilho fica no caractere | Difícil: cada pixel pede uma cor, e o cubo de 256 cores satura | Natural (uma cor), mas o tom depende de dithering |
| Tremor em animação | Pouco: o caractere só muda quando o brilho cruza um degrau | Muito em vídeo: a quantização no cubo pisca entre cores vizinhas | Muito se o dithering for por difusão de erro; o Bayer, fixo por posição, é estável |
| Depende da fonte | Um pouco (a densidade de cada caractere varia, segundo Bourke) | Quase nada (`▀` existe em toda fonte de terminal) | Bastante: há fontes sem braille ou com pontos desalinhados |
| Banda por célula que muda | ~1 byte de caractere; a sequência de cor (~11 bytes) só aparece quando a cor muda, que é raro com 2 a 4 cores | 3 bytes de `▀` + até ~19 de cor (frente e fundo mudam juntos) | 3 bytes + cor quando muda |
| Bom para | Wallpapers inteiros, vídeo ou render convertido, campos procedurais | Sprites pequenos e parados (pet, carro) | Chuva fina, contornos, estrelas, gráficos |

Ficaram de fora quadrantes, sextantes e octantes: mais resolução que o meio bloco, mas os sextantes e octantes vêm de "Symbols for Legacy Computing" e pouca fonte de terminal tem esses símbolos. Num app acessado por qualquer cliente ssh, não vale o risco.

**Custo de banda.** O que custa é quantas células mudam por segundo, vezes os bytes de cada uma. Pela passada de performance (`IDEIAS.md`):

- o brainrot muda ~5.200 células a 8 fps e manda ~380 KB/s, cerca de **9 bytes por célula**;
- a chuva manda ~48 KB/s, o que dá umas 700 células mudando por quadro.

Daí sai uma regra de bolso (estimativa, não medida): **um wallpaper fica no custo da chuva se mudar até ~700 células por quadro a 8 fps**, ou ~10% de uma tela de 160×50. Isso favorece:

- **fundo parado e movimento localizado** (o vapor do café, folhas, um objeto que balança);
- **campos lentos com poucos degraus**: com uma rampa de 10 caracteres e luz andando a 0,1 rad/s, quase nenhuma célula cruza um degrau a cada quadro;
- **evitar vídeo com câmera andando**, que muda a tela inteira a cada quadro: é o caso do brainrot.

**Um ajuste na conversão para 256 cores.** O `xterm256` de `brainrot.go` só usa o cubo. Em cena escura, isso pesa: o primeiro nível depois do preto é 95, e todo cinza-escuro vira preto ou um cinza já claro. Para wallpapers apagados, há dois caminhos: escolher a cor mais próxima também na rampa de cinzas (232–255, degraus de 10), ou fazer como o Ghostty e fixar 2 a 4 cores por wallpaper, deixando o brilho só na rampa. O segundo é mais simples e mais bonito.

## Candidatos a wallpaper novo

Todos pensados para pouco movimento, cores apagadas, nada que pisque ou se acumule, e custo de banda perto da chuva.

1. **Mar à noite.**
   - Ondas como um campo de brilho lento (soma de senos, como o `wave`/`swell` do asciify e o Ocean Beach do ascii-splash, ambos MIT) na rampa ` .:-=+*`.
   - Duas cores: azul apagado (`24`, `67`) e o reflexo da lua num cinza (`250`), parado no meio.
   - Barato: só as cristas trocam de degrau.
2. **Dunas sob a lua.**
   - Adaptação do `dune.ts` (MIT): cristas de areia com a luz passando bem devagar.
   - Uma cor só, um areia apagado (`180` ou `137`), e o céu com as estrelas fixas do céu noturno.
   - Quase nada muda por quadro: é o mais barato da lista.
3. **Aurora.**
   - Cortinas verticais em três profundidades, como o `aurora.ts` do asciify (MIT, rampa ` .:¦|`) e o Night Sky do ascii-splash.
   - Verdes e azuis apagados (`22`, `29`, `30`, `66`).
   - O cuidado é a velocidade: as cortinas dobram em dezenas de segundos, nunca piscam.
4. **Aquário calmo.**
   - Poucos peixes atravessando devagar, bolhas que sobem e somem, algas balançando.
   - Peixes desenhados por nós (ou sprites CC0 da Kenney em meio bloco). O asciiquarium é GPL, então entra só como ideia; o Aquarium do ascii-splash é MIT.
   - Cuidado: nada de cardume grande, que vira movimento demais.
5. **Um objeto renderizado no estilo Ghostty.**
   - Um render nosso no Blender (modelos e céu CC0 da Kenney/Poly Haven), em laço curto: uma xícara girando devagar, uma planta, ou o próprio pet em 3D.
   - Convertido como o fantasma: duas cores por distância de cor, rampa `·~ox+=*%$@` dentro de cada uma, quadros PNG embutidos pelo mesmo caminho do brainrot.
   - Com fundo vazio e objeto parado no lugar, só o contorno muda: bem mais barato que o brainrot.

## Como refazer café, árvore e estrada

A ideia comum: **o que fica parado vira uma imagem convertida pela rampa** (de um render ou de um desenho nosso em PNG), desenhada uma vez por tamanho de tela. **O que se mexe continua procedural**, como hoje. Assim a beleza vem da imagem e a banda continua a de agora.

- **Café.**
  - A xícara de hoje (linhas `|`, `_`, `\`) dá lugar a um render de uma xícara numa mesa, de lado e com luz de janela. A conversão segue o Ghostty: a cerâmica num marrom apagado (`137`/`180`), a mesa num cinza escuro (`238`–`240`) e o fundo vazio.
  - O vapor procedural continua como está; ele já é bom e é barato.
  - Variante sem render: um PNG pequeno pintado à mão, convertido pela mesma rampa.
- **Árvore.**
  - A copa de hoje é uma elipse de `&%@#` com ruído fixo. No lugar dela entra uma copa com volume: um render ou um PNG nosso, ou um campo de ruído com luz vinda de um lado (claro em cima e à esquerda, escuro embaixo). Rampa `.:-=+*#%` em dois verdes (`22`, `64`), tronco em `94`.
  - A copa não balança: balançar muda centenas de células por quadro.
  - As folhas caindo continuam procedurais, como hoje (somem depois de 5 s no chão).
- **Estrada à noite.**
  - Os morros viram um degradê de rampa (` .:-` em `235`–`238`) em duas camadas de paralaxe, em vez do bloco sólido de `░`.
  - Os postes ganham um cone de luz na rampa (`·` a `+` em `179`/`137`) que passa com eles.
  - O carro vira um sprite em meio bloco `▀`, como o pet, com faróis. É o único lugar em que o meio bloco vale a pena: pequeno, parado na tela, e com silhueta melhor do que em ASCII.
  - Cuidado: os morros e os postes andam, então cada coluna muda. Manter a velocidade dos morros baixa, como hoje (1,5 col/s), e medir a banda no harness.

## Próximo passo sugerido

Um protótipo com duas cenas para comparar no harness, com print em PNG e medida de banda:

- as **dunas** (campo procedural com rampa, a mais barata);
- o **café refeito** (uma imagem convertida pela rampa + o vapor de hoje).

A função de rampa (luma → janela de brilho → índice; Bayer opcional) fica ao lado de `halfBlockRows`, com um teste pequeno. Créditos novos em `CREDITS.md`: asciify-engine (MIT) e, se a conversão seguir o script, ghostty-org/website (MIT).
