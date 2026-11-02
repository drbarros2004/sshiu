-- O schema do Plantão, embutido no binário (go:embed em db.go) e rodado a cada Open.
-- Coluna nova entra aqui e também como ALTER TABLE no Open, para os bancos que já existem.
-- As tabelas sprints e friends, do modelo antigo, ficam nos bancos que já as têm, sem uso.

CREATE TABLE IF NOT EXISTS users (
	id          TEXT PRIMARY KEY, -- fingerprint da chave SSH
	nick        TEXT NOT NULL UNIQUE,
	energy      REAL NOT NULL,
	last_active INTEGER NOT NULL,
	streak      INTEGER NOT NULL DEFAULT 0,
	streak_day  TEXT NOT NULL DEFAULT '',
	pet_color   TEXT NOT NULL DEFAULT ''   -- cor escolhida para o pet ("" = sai do humor)
);
CREATE TABLE IF NOT EXISTS tasks ( -- uma por tarefa encerrada
	id           INTEGER PRIMARY KEY,
	user_id      TEXT NOT NULL REFERENCES users(id),
	started_at   INTEGER NOT NULL,
	ended_at     INTEGER NOT NULL,
	task         TEXT NOT NULL,
	declared_min INTEGER NOT NULL, -- tempo declarado (com o que foi dado ao continuar)
	focused_sec  INTEGER NOT NULL, -- o que de fato contou
	result       TEXT NOT NULL,
	room_code    TEXT NOT NULL DEFAULT '' -- a sala em que ela começou; "" nas tarefas anteriores à coluna
);
CREATE TABLE IF NOT EXISTS rooms (
	code        TEXT PRIMARY KEY, -- o convite: quem tem o código entra
	name        TEXT NOT NULL,
	owner       TEXT NOT NULL REFERENCES users(id), -- o dono da sala (ADR 0007)
	slots       TEXT NOT NULL DEFAULT '',     -- do modelo antigo, sem uso
	max_members INTEGER NOT NULL DEFAULT 10,  -- tamanho, mudado pelo dono
	closed      INTEGER NOT NULL DEFAULT 0,   -- 1 = não aceita mais ninguém, mesmo com vaga
	obra        TEXT NOT NULL DEFAULT '',     -- o tema da obra ("" = sem obra, ADR 0008)
	obra_meta   INTEGER NOT NULL DEFAULT 60   -- minutos focados que completam uma peça
);
CREATE TABLE IF NOT EXISTS room_members (
	code    TEXT NOT NULL REFERENCES rooms(code),
	user_id TEXT NOT NULL REFERENCES users(id),
	PRIMARY KEY (code, user_id)
);
CREATE TABLE IF NOT EXISTS open_tasks ( -- a tarefa em andamento, uma por usuário; sai daqui ao virar uma linha de tasks
	user_id        TEXT PRIMARY KEY REFERENCES users(id),
	started_at     INTEGER NOT NULL,
	seg_started_at INTEGER NOT NULL, -- início do trecho que está contando (muda ao continuar)
	ends_at        INTEGER NOT NULL,
	task           TEXT NOT NULL,
	declared_min   INTEGER NOT NULL,
	counted_sec    INTEGER NOT NULL, -- o que contou antes do último continuar
	room_code      TEXT NOT NULL
);
