#!/bin/sh
# Grava os 5 momentos do timelapse: a sala (roupas) e o mural (obra) de cada um.
cd "$(dirname "$0")"
H=../harness
for k in 0 1 2 3 4; do
  P=$(lsof -tnP -iTCP:23235 -sTCP:LISTEN); [ -n "$P" ] && kill $P && sleep 1
  rm -f sshiu.db sshiu.db-shm sshiu.db-wal
  ./sshiu-trailer -demo -port 23235 > server.log 2>&1 & sleep 2; kill $!; sleep 1
  python3 scene.py sshiu.db && python3 moment.py sshiu.db $k
  TRAILER_KEEP_TODAY=1 TRAILER_FAST_TASK="ship the trailer" nohup ./sshiu-trailer -demo -port 23235 > server.log 2>&1 &
  sleep 2
  (cd $H && rm -rf clips/tl-room-$k clips/tl-mural-$k && ./harness -user maya wait:"demo room" enter wait:"/10" sleep:2.5 \
    rec:clips/tl-room-$k sleep:2 stop key:o sleep:0.8 rec:clips/tl-mural-$k sleep:2 stop) | grep quadros
done
P=$(lsof -tnP -iTCP:23235 -sTCP:LISTEN); [ -n "$P" ] && kill $P
