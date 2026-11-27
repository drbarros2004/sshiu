#!/bin/sh
# Clipes novos do trailer2, cada grupo numa cena zerada na porta 23235:
#   task: a maya digita "ship the trailer", escolhe o tempo e começa a focar (sem a tarefa rápida)
#   coins + profile: termina a mesma tarefa (quase no fim, pelo TRAILER_FAST_TASK) e abre o perfil
#   obras: cada obra em cada estágio (obra.py), 2 s de cada
# Uso: ./record-v2.sh task|coins|obras [obra...]
set -e
cd "$(dirname "$0")"
H=../harness

serve() { # serve "<script extra>" VAR=valor...
  P=$(lsof -tnP -iTCP:23235 -sTCP:LISTEN || true); [ -n "$P" ] && kill $P && sleep 1
  rm -f sshiu.db sshiu.db-shm sshiu.db-wal
  ./sshiu-trailer -demo -port 23235 > server.log 2>&1 & sleep 2; kill $!; sleep 1
  python3 scene.py sshiu.db
  [ -n "$1" ] && python3 $1
  shift
  env "$@" nohup ./sshiu-trailer -demo -port 23235 > server.log 2>&1 &
  sleep 2
}
stop() { P=$(lsof -tnP -iTCP:23235 -sTCP:LISTEN || true); [ -n "$P" ] && kill $P; true; }

case $1 in
task)
  serve "" X=1
  (cd $H && rm -rf clips/00-task && ./harness -user maya wait:"demo room" enter wait:"/10" sleep:2.5 \
    rec:clips/00-task sleep:1.2 enter wait:concrete sleep:0.8 type:"ship the trailer" sleep:0.9 enter wait:"how long" \
    sleep:1.2 enter sleep:6 stop) ;;
coins)
  serve "" TRAILER_FAST_TASK="ship the trailer"
  (cd $H && rm -rf clips/05-coins clips/08-profile && ./harness -user maya wait:"demo room" enter wait:"/10" sleep:2 \
    enter wait:concrete type:"ship the trailer" enter wait:"how long" right right enter sleep:1.5 \
    rec:clips/05-coins sleep:1 key:f sleep:6 stop \
    key:d sleep:1.5 rec:clips/08-profile sleep:6.5 stop) ;;
obras)
  shift
  for o in "$@"; do for k in ${STAGES:-0 1 2 3}; do
    serve "obra.py sshiu.db $k" TRAILER_KEEP_TODAY=1 TRAILER_OBRA=$o
    (cd $H && rm -rf clips/obra-$o-$k && ./harness -user maya wait:"demo room" enter wait:"/10" sleep:2 \
      key:o sleep:0.8 rec:clips/obra-$o-$k sleep:2 stop) | grep quadros || true
  done; done ;;
esac
stop
