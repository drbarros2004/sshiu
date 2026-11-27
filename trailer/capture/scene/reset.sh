#!/bin/sh
# Recomeça a cena: banco novo, bots semeados, scene.py aplicado, servidor no ar na 23235.
cd "$(dirname "$0")"
P=$(lsof -tnP -iTCP:23235 -sTCP:LISTEN); [ -n "$P" ] && kill $P && sleep 1
rm -f sshiu.db sshiu.db-shm sshiu.db-wal
./sshiu-trailer -demo -port 23235 > server.log 2>&1 & sleep 2; kill $!; sleep 1
python3 scene.py sshiu.db
TRAILER_FAST_TASK="ship the trailer" nohup ./sshiu-trailer -demo -port 23235 > server.log 2>&1 &
sleep 2; tail -1 server.log
