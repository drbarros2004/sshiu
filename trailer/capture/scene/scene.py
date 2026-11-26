# Cena do trailer: aparência dos bots, a maya com histórico, saldo e uma tarefa quase no fim.
import sqlite3, time, random, sys
db = sqlite3.connect(sys.argv[1])
MAYA = "SHA256:31iEWmxaNynV2QsLf3Fcky/uEUdkRavVdiC+cOZVpc8"
now = int(time.time())
looks = {"ana": ("bunny", "pink", "top_hat", "normal"), "bruno": ("claude", "normal", "cowboy", "normal"),
         "caio": ("fruit", "purple", "cap", "red_backwards"), "duda": ("bunny", "turquoise", "little_hair", "rainbow"),
         "edu": ("fruit", "orange", "xmas_hat", "normal")}
for nick, (c, v, h, hv) in looks.items():
    db.execute("UPDATE users SET pet_character=?,pet_variant=?,pet_hat=?,pet_hat_variant=? WHERE id=?", (c, v, h, hv, "bot:" + nick))
db.execute("DELETE FROM users WHERE id=?", (MAYA,))
db.execute("INSERT INTO users(id,nick,energy,last_active,streak,streak_day,pet_color,lang) VALUES(?,?,?,?,?,?,?,?)",
           (MAYA, "maya", 74, now, 6, time.strftime("%Y-%m-%d", time.localtime(now - 86400)), "", "en"))
db.execute("INSERT OR REPLACE INTO wallets(user_id,balance,remainder_sec) VALUES(?,?,0)", (MAYA, 240))
for item in ["character/fruit", "color/fruit/blue", "color/fruit/green", "color/fruit/yellow", "color/fruit/coral"]:
    db.execute("INSERT OR REPLACE INTO inventory(user_id,item,state) VALUES(?,?,?)", (MAYA, item, "owned"))
tasks = ["read chapter 4 of the networks book", "problem set 2 of linear algebra", "fix the flaky test", "write the lab report",
         "review lecture notes", "practice for the interview"]
random.seed(7)
for d in range(24, 0, -1):
    if random.random() < 0.25:
        continue
    for _ in range(random.randint(1, 3)):
        mins = random.choice([25, 25, 50, 15, 45])
        end = now - d * 86400 + random.randint(-4, 4) * 3600
        res = "done" if random.random() < 0.8 else "not_done"
        db.execute("INSERT INTO tasks(user_id,started_at,ended_at,task,declared_min,focused_sec,result,room_code) VALUES(?,?,?,?,?,?,?,?)",
                   (MAYA, end - mins * 60, end, random.choice(tasks), mins, mins * 60, res, "demo"))
db.commit()
