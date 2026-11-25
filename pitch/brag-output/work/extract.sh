#!/bin/bash
# corta os trechos das gravações reais em quadros JPEG (30 fps)
cd "$(dirname "$0")"
seg(){ mkdir -p frames/$1; rm -f frames/$1/*; ffmpeg -v error -ss $3 -to $4 -i $2 -vf "fps=30${5:+,$5}" -q:v 3 frames/$1/%04d.jpg; echo $1 $(ls frames/$1 | wc -l); }
seg room   take2.mov 3.4 8.8
seg task   take2.mov 8.8 14.8
seg mural  take2.mov 16.1 20.3
seg shop   take2.mov 32.8 38.8
T1="crop=2140:1204:0:119,scale=1920:1080:flags=lanczos"
seg coffee take1.mov 15.8 18.3 "$T1"
seg tree   take1.mov 18.4 21.0 "$T1"
seg road   take1.mov 21.0 23.6 "$T1"
