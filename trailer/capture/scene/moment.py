# Um momento do timelapse: roupas de cada pet e minutos de hoje (que fazem a obra crescer).
# Uso: python3 moment.py sshiu.db K   (K de 0 a 4)
import sqlite3, sys, time
db = sqlite3.connect(sys.argv[1])
k = int(sys.argv[2])
MAYA = "SHA256:31iEWmxaNynV2QsLf3Fcky/uEUdkRavVdiC+cOZVpc8"
# Em cada momento, todo pet troca de personagem, cor e chapéu: o timelapse mostra a variedade da loja.
IDS = [MAYA, "bot:ana", "bot:bruno", "bot:caio", "bot:duda", "bot:edu"]
CHARS = ["fruit", "bunny", "claude"]
COLORS = {"fruit": ["blue", "green", "yellow", "coral", "purple", "turquoise", "orange", "pink"],
          "bunny": ["pink", "turquoise", "yellow", "purple", "green", "coral", "blue", "orange"],
          "claude": ["normal", "purple", "green", "orange", "blue", "pink", "yellow", "turquoise"]}
HATS = [("", ""), ("top_hat", "black"), ("crown", "normal"), ("cowboy", "sand"), ("cap", "red_backwards"),
        ("pirate", "normal"), ("xmas_hat", "green"), ("little_hair", "blonde"), ("firefighter_helmet", "normal"),
        ("crown", "rgb"), ("cap", "turquoise"), ("top_hat", "emerald"), ("cowboy", "black"), ("pirate", "red")]
MINUTES = [[0, 5, 3, 0, 0, 0], [14, 20, 12, 6, 10, 2], [25, 30, 24, 15, 22, 10],
           [45, 47, 41, 32, 43, 22], [64, 70, 61, 62, 66, 61]]
# Só o claude fica mais alto com chapéu (18 linhas em vez de 16) e a fileira inteira acompanha o mais
# alto. Para os cards não mudarem de tamanho entre os cortes, todo claude usa chapéu e todo momento
# tem pelo menos um claude.
LOOKS = {}
for p, uid in enumerate(IDS):
    LOOKS[uid] = []
    for m in range(len(MINUTES)):
        c = CHARS[(p + m) % 3]
        if uid == "bot:duda" and c == "claude":  # o humor dela vira "nonfunctional" no claude e alarga os cards
            c = "bunny"
        v = COLORS[c][(p * 3 + m * 5) % 8]
        h, hv = HATS[(p * 3 + m * 4 + (1 if m else 0)) % len(HATS)] if (p, m) != (0, 0) else ("", "")
        if c == "claude" and not h:
            h, hv = HATS[(p + m) % (len(HATS) - 1) + 1]
        LOOKS[uid].append((c, v, h, hv, MINUTES[m][p]))
for m in range(len(MINUTES)):
    assert any(LOOKS[u][m][0] == "claude" and LOOKS[u][m][2] for u in IDS), m
today = time.strftime("%Y-%m-%d")
for uid, looks in LOOKS.items():
    c, v, h, hv, mins = looks[k]
    db.execute("UPDATE users SET pet_character=?,pet_variant=?,pet_hat=?,pet_hat_variant=?,today=?,today_min=? WHERE id=?",
               (c, v, h, hv, today, mins, uid))
db.commit()
