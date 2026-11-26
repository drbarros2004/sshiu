#!/bin/sh
# Só o modo jardim, 4 s em cada wallpaper, para o grid 2×2 do trailer.
set -e
cd "$(dirname "$0")"
../demo/reset.sh 2>/dev/null
rm -rf clips/02-garden-long
./harness -user maya \
  wait:"demo room" enter wait:"/10" sleep:2 \
  key:d sleep:0.3 key:d sleep:0.3 key:d sleep:0.5 down down down right sleep:0.3 enter sleep:0.5 key:a sleep:0.3 key:a sleep:0.3 key:a sleep:1 \
  enter wait:concrete type:"finish problem set 3" enter wait:"how long" enter sleep:1 \
  key:j sleep:0.5 rec:clips/02-garden-long sleep:4 right sleep:3.85 right sleep:3.85 right sleep:3.85 right sleep:3.85 right sleep:3.85 right sleep:3.85 right sleep:3.85 stop
P=$(lsof -tnP -iTCP:23235 -sTCP:LISTEN); [ -n "$P" ] && kill $P
