# A música "quebrando" na entrada do brainrot: o trecho da faixa que toca naquele instante desacelera até
# parar, como uma fita. Uso: .venv-earshot/bin/python tools/tapestop.py MUSICA.wav INICIO_S SAIDA.wav
# INICIO_S é o segundo da faixa em que o brainrot entra (Trailer2.tsx mostra a conta em BREAK_AT).
import sys
import numpy as np
import soundfile as sf

src, start, out = sys.argv[1], float(sys.argv[2]), sys.argv[3]
y, sr = sf.read(src, always_2d=True)
DUR = 0.9  # segundos até parar
n = int(DUR * sr)
t = np.arange(n) / n
speed = (1 - t) ** 1.6  # desacelera rápido no começo e se arrasta no fim
pos = start * sr + np.cumsum(speed)
i = np.clip(pos.astype(int), 0, len(y) - 2)
frac = (pos - i)[:, None]
z = y[i] * (1 - frac) + y[i + 1] * frac
z *= np.minimum(1, (1 - t) * 8)[:, None]  # some nos últimos quadros, sem estalo
sf.write(out, z, sr, subtype="PCM_16")
