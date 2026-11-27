# Acha andamento e batidas fortes de uma música: python tools/beats.py arquivo.mp3 [segundos]
import subprocess, sys
import numpy as np
f, dur = sys.argv[1], float(sys.argv[2]) if len(sys.argv) > 2 else 40
sr = 11025
raw = subprocess.run(["ffmpeg", "-loglevel", "error", "-t", str(dur), "-i", f, "-ac", "1", "-ar", str(sr), "-f", "f32le", "-"], capture_output=True).stdout
x = np.frombuffer(raw, np.float32)
hop, win = 256, 1024
frames = np.lib.stride_tricks.sliding_window_view(x, win)[::hop] * np.hanning(win)
spec = np.abs(np.fft.rfft(frames, axis=1))
freqs = np.fft.rfftfreq(win, 1 / sr)
low = spec[:, freqs < 150].sum(1)  # bumbo
flux = np.maximum(0, np.diff(np.log1p(spec), axis=0)).sum(1)
t = np.arange(len(flux)) * hop / sr
# andamento pela autocorrelação do fluxo
ac = np.correlate(flux - flux.mean(), flux - flux.mean(), "full")[len(flux) - 1:]
lags = np.arange(len(ac)) * hop / sr
ok = (lags > 60 / 180) & (lags < 60 / 60)
beat = lags[ok][np.argmax(ac[ok])]
print(f"andamento ~ {60 / beat:.1f} bpm (tempo {beat:.3f} s, compasso {4 * beat:.2f} s)")
# energia do grave por janela de meio segundo, para ver onde a batida entra
step = int(0.5 * sr / hop)
lowdb = 20 * np.log10(np.array([low[i:i + step].mean() for i in range(0, len(low), step)]) + 1e-9)
print("grave (dB rel.) a cada 0,5 s:", " ".join(f"{i*0.5:.1f}:{v - lowdb.max():.0f}" for i, v in enumerate(lowdb)))
# bumbos fortes: picos do grave
lo = low / low.max()
TH = float(sys.argv[3]) if len(sys.argv) > 3 else 0.45  # limiar do pico, relativo ao maior
peaks = [i for i in range(1, len(lo) - 1) if lo[i] > TH and lo[i] >= lo[i - 1] and lo[i] >= lo[i + 1]]
pt, last = [], -1
for i in peaks:
    if t[i] - last > 0.2:
        pt.append(t[i]); last = t[i]
print("bumbos:", " ".join(f"{v:.2f}" for v in pt[:60]))
