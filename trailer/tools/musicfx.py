# Efeitos feitos da própria faixa, para a música conversar com o vídeo:
#   stop T   a fita parando: o trecho que começa em T desacelera até parar (entrada do brainrot)
#   start T  a fita voltando: acelera do zero e chega em velocidade normal exatamente em T (volta nas moedas)
#   ring T   o fim: o golpe que começa em T (um bumbo) ecoando até sumir (o logo do fecho)
# Uso: .venv-earshot/bin/python tools/musicfx.py stop|start|ring MUSICA.wav T SAIDA.wav
import sys
import numpy as np
import soundfile as sf

mode, src, at, out = sys.argv[1], sys.argv[2], float(sys.argv[3]), sys.argv[4]
y, sr = sf.read(src, always_2d=True)


def read(pos):  # lê a faixa em posições fracionárias (amostras), com interpolação linear
    i = np.clip(pos.astype(int), 0, len(y) - 2)
    frac = (pos - i)[:, None]
    return y[i] * (1 - frac) + y[i + 1] * frac


if mode in ("stop", "start"):
    DUR = 0.9 if mode == "stop" else 0.7
    n = int(DUR * sr)
    t = np.arange(n) / n
    if mode == "stop":
        speed = (1 - t) ** 1.6  # desacelera rápido no começo e se arrasta no fim
        z = read(at * sr + np.cumsum(speed)) * np.minimum(1, (1 - t) * 8)[:, None]
    else:
        speed = t**1.6  # o contrário: se arrasta no começo e chega em velocidade normal em T
        z = read(at * sr - np.cumsum(speed[::-1])[::-1]) * np.minimum(1, t * 8)[:, None]
elif mode == "ring":
    HIT, TAIL = 0.35, 2.2  # segundos do golpe e do eco
    seg = y[int(at * sr) : int((at + HIT) * sr)].copy()
    seg *= np.minimum(1, (HIT - np.arange(len(seg)) / sr) / 0.08)[:, None]  # sem estalo no corte
    rng = np.random.default_rng(1)
    k = np.arange(int(TAIL * sr)) / sr
    ir = rng.standard_normal(len(k)) * np.exp(-k / 0.45) * 0.05  # um eco de sala, decaindo
    ir[0] = 1.0
    n = len(seg) + len(ir) - 1
    z = np.stack([np.fft.irfft(np.fft.rfft(seg[:, c], n) * np.fft.rfft(ir, n), n) for c in range(seg.shape[1])], 1)
    z /= max(1, np.abs(z).max() / np.abs(seg).max())
    z *= np.minimum(1, (len(z) - np.arange(len(z))) / (0.3 * sr))[:, None]  # termina em zero
else:
    sys.exit("modo: stop, start ou ring")
sf.write(out, z, sr, subtype="PCM_16")
