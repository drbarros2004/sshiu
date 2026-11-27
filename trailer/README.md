# trailer

Trailer de 33 a 38 s do sshiu, em inglês, 1920×1080 a 30 fps, sem voz, montado em Remotion. Duas versões, cada uma com duas músicas:

- `trailer` (`src/Trailer.tsx`): as telas do app numa janela grande do Terminal, com legenda embaixo.
- `trailer2` (`src/Trailer2.tsx`, 37,7 s): um título em cima de cada tela, sem numerar. Começa com a maya digitando a tarefa "ship the trailer", que ela termina na cena das moedas e que aparece no perfil. A câmera (`cam`) só anda na tarefa; nas moedas o feed fica parado ao lado de um contador grande (`counter`, só do trailer), e o perfil ganha molduras laranja (`rings`) em vez de zoom. As transições variam (`enter`: corte seco, janela abrindo, deslizar e fade no fecho). Ainda em ajuste.

As cenas, os trechos de cada clipe, os textos e os efeitos sonoros de cada versão ficam no `SCENES` do arquivo dela. A abertura, o fecho e a música são os mesmos (`Title`, `Outro`, `Opening`, no `Trailer.tsx`).

```
npm install
npm run studio
npm run render   # out/trailer-{warm,8bit}.mp4 e out/trailer2-{warm,8bit}.mp4
```

Uma versão só: `tools/render-chunks.sh trailer2-warm out/trailer2-warm.mp4`. Ele renderiza em pedaços de 200 quadros e refaz o pedaço que falhar, porque numa máquina de 8 GB o compositor do Remotion às vezes cai no meio ("Request closed"). Se o `npm install` falhar com EACCES no cache, use `--cache` apontando para outra pasta.

## O que tem aqui

- `public/clips/`: os clipes gravados do app de verdade (modo demo, sem dev mode). Os do trailer2 vêm do `capture/scene/record-v2.sh`: `00-task`, `05-coins-v2`, `08-profile-v2` e `obra-<tema>-<0..3>` (trem, acampamento e orquestra, cada estágio da obra). `garden-0..3` são os trechos de 12,3, 16,3, 20,3 e 28,3 s do `02-garden-long`, cortados à parte para o grid do trailer2: `ffmpeg -ss <t> -i 02-garden-long.mp4 -t 5 -vf scale=1304:734 -c:v libx264 -crf 16 -g 15 -pix_fmt yuv420p -an garden-<i>.mp4`.
- `public/sfx/`: efeitos gerados com o earshot pelo `sfx.sh` (livres, sem crédito). `shh-voice.wav` é o "shush" do começo de `sfx-shh/elevenlabs-shush-source.mp3` (ElevenLabs), de 0,32 s a 0,85 s.
- `public/music/`: `bed-warm.wav` vem do `earshot make --sound bed/warm --dur 40 --seed 0`. `bed-warm-8bit.wav` é a mesma partitura tocada pelos canais de chip do earshot (`tools/warm8bit.py`). As duas são livres, sem crédito, e entram no enter da abertura, num tempo forte. `tools/beats.py` acha o andamento e os tempos fortes de uma faixa.
- `capture/`: como os clipes foram gravados (abaixo).

Fora do git, porque se regeneram: `node_modules/`, `out/`, `.venv-earshot/`, as faixas brutas do earshot (`music/`) e as prévias em m4a (`listen/`).

O earshot, para os efeitos e a música:

```
uv venv .venv-earshot --python 3.11
uv pip install --python .venv-earshot/bin/python "git+https://github.com/Leonxlnx/earshot@7a7f7c5"
```

## Gravar os clipes de novo

1. Exporte o sshiu (`git archive HEAD`) para uma pasta, aplique `capture/scene/trailer.patch` e compile um binário `sshiu-trailer`. O patch deixa os bots previsíveis, faz a tarefa "ship the trailer" chegar quase no fim (`TRAILER_FAST_TASK`) e escolhe a obra da sala demo (`TRAILER_OBRA`). Rode tudo numa pasta de rascunho e na porta 23235, nunca na raiz do repositório (o servidor lê o `sshiu.db` do diretório atual).
2. Compile o harness (`cd capture/harness && go build -o harness .`). Os emojis já estão em `emoji/`; outros saem do `swift emoji.swift 👏 🔥 …` (só no Mac).
3. Ajuste os caminhos do `record.sh` e do `reset.sh` e rode o `record.sh`. Ele zera o banco, monta a cena (`scene.py`) e grava as cenas. Cada pasta de quadros vira um mp4 com `ffmpeg -framerate 30 -i %05d.png -c:v libx264 -crf 16 -pix_fmt yuv420p`.
4. Os clipes do trailer2: `capture/scene/record-v2.sh task`, `coins` (moedas e perfil) e `obras trem acampamento orquestra` (com `STAGES="0 1 2 3"`, o padrão). As obras usam o `TRAILER_OBRA` do patch e o `obra.py`, que põe todo mundo no mesmo estágio. Copie o `record-v2.sh`, o `obra.py` e o `scene.py` para a pasta do servidor, ao lado do `sshiu-trailer`, com o harness em `../harness`.
5. O timelapse (roupas e mural em cinco momentos) vem do `capture/scene/timelapse.sh` + `moment.py`. Todo claude usa chapéu e a duda nunca é claude, para os cards não mudarem de tamanho. O jardim longo vem do `record-garden.sh`, com 4 s em cada wallpaper.

A maya é identificada pela chave de teste em `capture/harness/keys/maya.seed`, que só serve para a sala demo. O `scene.py` espera o fingerprint dela: se a chave mudar, atualize o `MAYA`.
