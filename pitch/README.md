# Pitch do sshiu

Pitch de 3 minutos para o Shark Tank simulado da aula de empreendedorismo.

Online: <https://drbarros2004.github.io/sshiu/> (o `deck/` publicado no GitHub Pages pelo [`pages.yml`](../.github/workflows/pages.yml) a cada mudança no master).

| Arquivo | O que é |
|---|---|
| [`deck/index.offline.html`](deck/index.offline.html) | **O deck para apresentar:** um arquivo só, com fontes, imagens e vídeo dentro. Abre sem internet. |
| [`deck/index.html`](deck/index.html) | O deck para editar (lê `demo.mp4`, `demo.jpg` e a pasta `pets/`) |
| [`deck/sshiu-pitch.pdf`](deck/sshiu-pitch.pdf) | Reserva em PDF, uma página por slide |
| [`deck/demo.mp4`](deck/demo.mp4) | A demo: 40 s, 1280×720, o trailer do sshiu (versão `trailer2-flow` de [`../trailer`](../trailer), em 8-bit): tomadas reais do app em modo demo, um título por tela e trilha gerada |
| [`brag-output/`](brag-output/) | O vídeo anterior da demo, em 1920×1080 (`brag.mp4`), pôster, plano, texto para postar e o que gera tudo (`work/render.html`, `work/music.py`, `work/capture.mjs`) |
| [`script.md`](script.md) | Roteiro cronometrado, gerado das notas do deck |
| [`dossier.md`](dossier.md) | Solução, dor, mercado, concorrência, posicionamento e perguntas da banca |
| [`source-checks.md`](source-checks.md) | De onde vem cada número e o que dá para afirmar |
| [`revenue-assumptions.md`](revenue-assumptions.md) | As três fontes de receita e os cenários |
| [`rehearsal.md`](rehearsal.md) | Papéis, checklist do dia, ensaios e plano B |
| [`pets.py`](pets.py) | Desenha os pets da capa e do fechamento a partir dos sprites do app (`assets/pets`) |
| [`soundtrack.py`](soundtrack.py) | A trilha do vídeo, composta do zero em Python (numpy + scipy) |

## Como apresentar

Abra `deck/index.offline.html` no **Google Chrome**.

| Tecla | Faz |
|---|---|
| `→` `←` espaço | Passa os slides |
| `F` | Tela cheia (no projetor) |
| `S` | Abre o modo apresentador numa segunda janela (no notebook), sincronizado |
| `T` | Liga o cronômetro: dê na primeira palavra |
| `N` | Mostra os cards de roteiro do slide |
| `V` | Pausa e retoma o vídeo |
| `J` `L` | Volta e avança 5 s no vídeo (o vídeo também tem barra de controles) |
| `0` | Volta o vídeo para o começo |
| `D` | Tema claro ou escuro |
| `P` | Exporta o PDF |

O HUD embaixo mostra o tempo, o ritmo (no ritmo, atrasado, adiantado) e quanto do orçamento do slide já foi usado. O vídeo começa sozinho no slide 4, com legendas e uma trilha de fundo: fale pouco, uma frase por cena, como está nos cards, e teste o volume da sala antes.

## Mudar o deck

1. Edite `deck/index.html` (o texto dos cards fica em `data-notes`).
2. Gere o arquivo único: `python3 ~/.claude/skills/interactive-slides/scripts/embed_assets.py deck/index.html -o deck/index.offline.html`.
3. Gere o PDF de novo com `P` no Chrome.

Valores entre colchetes (`[R$ 100 mil]`, `[10%]`) ainda precisam ser decididos: o atalho `?` mostra quantos faltam.
