#!/bin/sh
# Grava todas as cenas do trailer numa sessão só, a partir de uma cena zerada (reset.sh).
set -e
cd "$(dirname "$0")"
../demo/reset.sh 2>/dev/null
rm -rf clips && mkdir clips
./harness -user maya \
  wait:"demo room" enter wait:"6/10" sleep:2 \
  key:d sleep:0.3 key:d sleep:0.3 key:d sleep:0.5 down down down right sleep:0.3 enter sleep:0.5 key:a sleep:0.3 key:a sleep:0.3 key:a sleep:1 \
  rec:clips/01-room sleep:1.5 down sleep:0.8 key:2 sleep:3 stop \
  enter wait:concrete type:"finish problem set 3" enter wait:"how long" enter sleep:1 \
  key:j sleep:0.5 rec:clips/02-garden sleep:2 right sleep:1.8 right sleep:1.8 right sleep:1.8 right sleep:1.8 right sleep:1.8 right sleep:1.8 stop \
  esc sleep:0.5 key:d sleep:0.3 key:d sleep:0.3 key:d sleep:0.8 \
  rec:clips/03-allow sleep:0.8 down down down down down down down down sleep:0.6 right sleep:1.2 enter sleep:1.2 stop \
  key:a sleep:0.3 key:a sleep:0.3 key:a sleep:0.8 key:j sleep:0.5 until:right:"brainrot" sleep:0.3 rec:clips/04-brainrot sleep:5 stop \
  esc sleep:0.5 key:t wait:concrete type:"ship the trailer" enter wait:"how long" right right enter sleep:1.5 \
  rec:clips/05-coins sleep:1 key:f sleep:3 stop \
  key:d sleep:0.3 key:d sleep:0.8 \
  rec:clips/06-shop sleep:0.8 down down sleep:0.5 right sleep:0.6 right sleep:0.6 right sleep:0.6 right sleep:0.6 right sleep:1.2 enter sleep:1.5 key:y sleep:1.5 dump key:y sleep:1.5 stop dump \
  key:a sleep:0.3 key:a sleep:1 rec:clips/07-room-hat sleep:3 stop \
  key:d sleep:1 rec:clips/08-profile sleep:3 stop
