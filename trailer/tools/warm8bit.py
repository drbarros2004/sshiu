# Versão 8-bit da bed-warm: a mesma composição do earshot (seed 0, 84 bpm, fá maior, mesmos acordes,
# baixo, motivo e bateria), tocada pelos canais de chip do estilo bed/8bit.
# Uso: .venv-earshot/bin/python tools/warm8bit.py public/music/bed-warm-8bit.wav
import copy, sys
import soundfile as sf
from earshot.synth import music
from earshot.sounds import beds, registry

CHIP = {"keys": ("voice", "chip:arp"), "lead": ("voice", "chip:lead"), "bass": ("voice", "chip:bass"),
        "kick": ("piece", "chip_kick"), "snare": ("piece", "chip_snare"), "rim": ("piece", "chip_hat"),
        "shaker": ("piece", "chip_hat")}
st = copy.deepcopy(music.STYLES["warm"])
for p in st["parts"]:
    if p["name"] in CHIP:
        k, v = CHIP[p["name"]]
        p[k] = v
        p.pop("lp", None)
        p.pop("strum", None)
    if p["name"] == "crackle":
        p["level"] = -120.0  # sem chiado de vinil; a parte fica para a composição não mudar
st["room"], st["tape"] = "dry", 0.0
music.STYLES["warm"] = st
beds.KITS["warm"] = ()  # sem bateria acústica: as peças de chip são do próprio motor
y, meta = registry.render("bed/warm", dur=32, check=False)
sf.write(sys.argv[1], y, 48000, subtype="PCM_16")
print(meta.get("dur_s"), meta.get("lufs_max"), meta.get("peak_dbfs"))
