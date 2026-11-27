#!/bin/sh
# Renderiza uma composição em pedaços e junta. Numa máquina de 8 GB o compositor do Remotion às
# vezes cai no meio ("Request closed"); assim só o pedaço que caiu é refeito.
# Uso: tools/render-chunks.sh trailer2-warm out/trailer2-warm.mp4
set -e
ID=$1 OUT=$2 STEP=200
TMP=$(mktemp -d)
mkdir -p "$(dirname "$OUT")"
N=$(npx remotion compositions 2>/dev/null | awk -v id="$ID" '$1==id{print $4}')
[ -n "$N" ] || { echo "composição $ID não encontrada"; exit 1; }
i=0
while [ $i -lt "$N" ]; do
  end=$((i + STEP - 1)); [ $end -ge "$N" ] && end=$((N - 1))
  for try in 1 2 3; do
    npx remotion render "$ID" "$TMP/$i.mp4" --frames=$i-$end --muted --log=error --timeout=60000 && break
    [ $try = 3 ] && exit 1
  done
  echo "file '$TMP/$i.mp4'" >> "$TMP/list.txt"
  i=$((end + 1))
done
for try in 1 2 3; do
  npx remotion render "$ID" "$TMP/audio.wav" --codec=wav --log=error --timeout=60000 && break
  [ $try = 3 ] && exit 1
done
ffmpeg -v error -y -f concat -safe 0 -i "$TMP/list.txt" -i "$TMP/audio.wav" -c:v copy -c:a aac -b:a 192k -shortest "$OUT"
rm -r "$TMP"
echo "$OUT"
