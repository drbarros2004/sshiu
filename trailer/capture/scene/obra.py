# Um estágio da obra para todos: minutos de hoje que deixam cada peça no estágio K (0 a 3, nos terços
# da meta de 1 h), com uma variação pequena entre as pessoas. Uso: python3 obra.py sshiu.db K
import sqlite3, sys, time
db = sqlite3.connect(sys.argv[1])
k = int(sys.argv[2])
MAYA = "SHA256:31iEWmxaNynV2QsLf3Fcky/uEUdkRavVdiC+cOZVpc8"
IDS = [MAYA, "bot:ana", "bot:bruno", "bot:caio", "bot:duda", "bot:edu"]
STAGES = [[0, 4, 2, 0, 6, 1], [22, 26, 21, 24, 30, 20], [44, 48, 41, 42, 52, 40], [61, 66, 63, 60, 70, 62]]
today = time.strftime("%Y-%m-%d")
for uid, mins in zip(IDS, STAGES[k]):
    db.execute("UPDATE users SET today=?,today_min=? WHERE id=?", (today, mins, uid))
db.commit()
