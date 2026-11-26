#!/bin/sh
# Efeitos do trailer, gerados com o earshot (MIT; o som gerado é livre para qualquer uso, sem crédito).
# Os WAVs ficam em public/sfx/. Para gerar de novo: uv venv .venv-earshot --python 3.11 &&
#   uv pip install --python .venv-earshot/bin/python "git+https://github.com/Leonxlnx/earshot@7a7f7c5"
set -e
E=${EARSHOT:-.venv-earshot/bin/earshot}
D=public/sfx
rm -f $D/*.wav
for i in 1 2 3 4; do $E sfx render typing_key/laptop -o $D/key$i.wav --seed $i; done   # teclas de "ssh sshiu"
$E sfx render typing_key/thock -o $D/enter.wav --seed 1             # enter
$E sfx render pro_pop_in/round -o $D/logo.wav --seed 1              # logo aparece
$E sfx render swish/soft -o $D/swish.wav --seed 1                   # troca de wallpaper
$E sfx render hover_tick/detent -o $D/toggle.wav --seed 1           # "allow ... ? › yes"
$E sfx render glitch/burst -o $D/glitch.wav --seed 1                # entra o brainrot
$E sfx render pro_notify/ding_2045 -o $D/coin.wav --seed 1          # +43 coins
$E sfx render ui_pop/blip -o $D/dialog.wav --seed 1                 # abre a confirmação da compra
$E sfx render notify/mid_wood -o $D/confirm.wav --seed 1            # y: compra e equipa
$E sfx render whoosh_fast/tight -o $D/whoosh.wav --seed 1          # tela deslizando para a próxima
# fecho: um "shh" sintetizado de ruído rosa filtrado (o earshot não faz voz)
ffmpeg -loglevel error -y -f lavfi -i "anoisesrc=color=pink:d=1.4:seed=7:r=48000" \
  -af "highpass=f=1800,lowpass=f=7000,equalizer=f=3000:t=q:w=1.2:g=8,afade=t=in:d=0.18:curve=qsin,afade=t=out:st=0.55:d=0.85:curve=exp,loudnorm=I=-18:TP=-2" -ar 48000 $D/shh.wav
