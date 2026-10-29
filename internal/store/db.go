package store

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"plantao/internal/pet"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// As tabelas sprints e friends, do modelo antigo, ficam nos bancos que já as têm, sem uso.
const schema = `
CREATE TABLE IF NOT EXISTS users (
	id          TEXT PRIMARY KEY, -- fingerprint da chave SSH
	nick        TEXT NOT NULL UNIQUE,
	energy      REAL NOT NULL,
	last_active INTEGER NOT NULL,
	streak      INTEGER NOT NULL DEFAULT 0,
	streak_day  TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS tasks ( -- uma por tarefa encerrada
	id           INTEGER PRIMARY KEY,
	user_id      TEXT NOT NULL REFERENCES users(id),
	started_at   INTEGER NOT NULL,
	ended_at     INTEGER NOT NULL,
	task         TEXT NOT NULL,
	declared_min INTEGER NOT NULL, -- tempo declarado (com o que foi dado ao continuar)
	focused_sec  INTEGER NOT NULL, -- o que de fato contou
	result       TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS rooms (
	code  TEXT PRIMARY KEY, -- o convite: quem tem o código entra
	name  TEXT NOT NULL,
	owner TEXT NOT NULL REFERENCES users(id), -- quem criou; sem poderes no MVP (ADR 0003)
	slots TEXT NOT NULL DEFAULT ''           -- do modelo antigo, sem uso
);
CREATE TABLE IF NOT EXISTS room_members (
	code    TEXT NOT NULL REFERENCES rooms(code),
	user_id TEXT NOT NULL REFERENCES users(id),
	PRIMARY KEY (code, user_id)
);`

// Open abre o banco (criando as tabelas se preciso) e o guarda para o resto do pacote.
func Open(path string) error {
	d, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	d.SetMaxOpenConns(1) // SQLite aceita um escritor por vez; assim nunca dá SQLITE_BUSY
	if _, err = d.Exec(schema); err != nil {
		return err
	}
	// ponytail: migração ingênua (o erro "duplicate column" é esperado após a 1ª vez); goose/migrate quando houver várias.
	d.Exec(`ALTER TABLE users ADD COLUMN theme TEXT NOT NULL DEFAULT ''`)
	d.Exec(`ALTER TABLE users ADD COLUMN today TEXT NOT NULL DEFAULT ''`)
	d.Exec(`ALTER TABLE users ADD COLUMN today_min REAL NOT NULL DEFAULT 0`)
	d.Exec(`ALTER TABLE users ADD COLUMN garden_wallpaper TEXT NOT NULL DEFAULT ''`)
	d.Exec(`ALTER TABLE users ADD COLUMN garden_clock INTEGER NOT NULL DEFAULT 1`)
	d.Exec(`ALTER TABLE users ADD COLUMN garden_task INTEGER NOT NULL DEFAULT 1`)
	d.Exec(`ALTER TABLE users ADD COLUMN garden_room INTEGER NOT NULL DEFAULT 1`)
	db = d
	return nil
}

// Garden: o que o usuário escolheu para o modo jardim na aba visual.
type Garden struct {
	Wallpaper  string // "" = sorteado a cada tarefa
	Clock      int    // 0 mm:ss, 1 só os minutos, 2 nenhum
	Task, Room bool   // mostrar a tarefa e quem mais está focando
}

// DefaultGarden: os padrões de quem nunca mexeu na aba visual (IDEIAS.md).
var DefaultGarden = Garden{Clock: 1, Task: true, Room: true}

func GetGarden(id string) Garden {
	g := DefaultGarden
	db.QueryRow(`SELECT garden_wallpaper, garden_clock, garden_task, garden_room FROM users WHERE id = ?`, id).
		Scan(&g.Wallpaper, &g.Clock, &g.Task, &g.Room) // sem usuário ainda: os padrões
	return g
}

func SetGarden(id string, g Garden) error {
	_, err := db.Exec(`UPDATE users SET garden_wallpaper = ?, garden_clock = ?, garden_task = ?, garden_room = ? WHERE id = ?`,
		g.Wallpaper, g.Clock, g.Task, g.Room, id)
	return err
}

func Theme(id string) (t string) {
	db.QueryRow(`SELECT theme FROM users WHERE id = ?`, id).Scan(&t) // sem usuário ainda: tema padrão
	return t
}

func SetTheme(id, t string) error {
	_, err := db.Exec(`UPDATE users SET theme = ? WHERE id = ?`, t, id)
	return err
}

func scanUser(row *sql.Row) (nick string, p pet.Pet, ok bool, err error) {
	var last int64
	err = row.Scan(&nick, &p.Energy, &last, &p.Streak, &p.StreakDay, &p.Today, &p.TodayMin)
	if errors.Is(err, sql.ErrNoRows) {
		return "", pet.Pet{}, false, nil
	}
	p.LastActive = time.Unix(last, 0)
	return nick, p, err == nil, err
}

const selectUser = `SELECT nick, energy, last_active, streak, streak_day, today, today_min FROM users WHERE id = ?`

func GetUser(id string) (string, pet.Pet, bool, error) {
	return scanUser(db.QueryRow(selectUser, id))
}

// PutUser cria o usuário; se já existe, só redefine a energia (usado para os bots da demo).
func PutUser(id, nick string, energy float64) error {
	_, err := db.Exec(`INSERT INTO users (id, nick, energy, last_active) VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET energy = excluded.energy, last_active = excluded.last_active`,
		id, nick, energy, time.Now().Unix())
	return err
}

type TaskRow struct {
	Start, End  time.Time
	Task        string
	DeclaredMin int
	Focused     time.Duration
	Result      pet.Result
}

// RecordTask guarda uma tarefa encerrada e soma os minutos focados ao pet.
func RecordTask(id string, t TaskRow) (pet.Pet, error) {
	tx, err := db.Begin()
	if err != nil {
		return pet.Pet{}, err
	}
	defer tx.Rollback()
	_, p, _, err := scanUser(tx.QueryRow(selectUser, id))
	if err != nil {
		return pet.Pet{}, err
	}
	p.Focus(t.Focused.Minutes(), t.Result == pet.Done, t.End)
	if _, err = tx.Exec(`UPDATE users SET energy = ?, last_active = ?, streak = ?, streak_day = ?, today = ?, today_min = ? WHERE id = ?`,
		p.Energy, p.LastActive.Unix(), p.Streak, p.StreakDay, p.Today, p.TodayMin, id); err != nil {
		return pet.Pet{}, err
	}
	if _, err = tx.Exec(`INSERT INTO tasks (user_id, started_at, ended_at, task, declared_min, focused_sec, result) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, t.Start.Unix(), t.End.Unix(), t.Task, t.DeclaredMin, int(t.Focused.Seconds()), t.Result); err != nil {
		return pet.Pet{}, err
	}
	return p, tx.Commit()
}

// Tasks: as últimas tarefas do usuário, da mais nova para a mais velha.
func Tasks(id string) ([]TaskRow, error) {
	rows, err := db.Query(`SELECT started_at, ended_at, task, declared_min, focused_sec, result FROM tasks
		WHERE user_id = ? ORDER BY id DESC LIMIT 8`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TaskRow
	for rows.Next() {
		var r TaskRow
		var start, end, sec int64
		if err := rows.Scan(&start, &end, &r.Task, &r.DeclaredMin, &sec, &r.Result); err != nil {
			return nil, err
		}
		r.Start, r.End, r.Focused = time.Unix(start, 0), time.Unix(end, 0), time.Duration(sec)*time.Second
		out = append(out, r)
	}
	return out, rows.Err()
}

type RoomRow struct{ Code, Name string }

// newCode: 6 letras sem as ambíguas (0/o, 1/l/i), fácil de ditar para um colega.
func newCode() string {
	const alphabet = "abcdefghjkmnpqrstuvwxyz23456789"
	b := make([]byte, 6)
	rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

// Rooms lista todas as salas.
func Rooms() ([]RoomRow, error) {
	rows, err := db.Query(`SELECT code, name FROM rooms`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RoomRow
	for rows.Next() {
		var r RoomRow
		if err := rows.Scan(&r.Code, &r.Name); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func RoomMembers(code string) ([]string, error) {
	rows, err := db.Query(`SELECT user_id FROM room_members WHERE code = ?`, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func CreateRoom(owner, name string) (RoomRow, error) {
	r := RoomRow{newCode(), name}
	tx, err := db.Begin()
	if err != nil {
		return r, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO rooms (code, name, owner, slots) VALUES (?, ?, ?, '')`, r.Code, name, owner); err != nil {
		return r, err
	}
	if _, err := tx.Exec(`INSERT INTO room_members (code, user_id) VALUES (?, ?)`, r.Code, owner); err != nil {
		return r, err
	}
	return r, tx.Commit()
}

func JoinRoom(user, code string) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rooms WHERE code = ?`, code).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("nenhuma sala com o código %q", code)
	}
	_, err := db.Exec(`INSERT OR IGNORE INTO room_members (code, user_id) VALUES (?, ?)`, code, user)
	return err
}

// LeaveRoom tira o usuário da sala e apaga a sala se ele era o último membro.
func LeaveRoom(user, code string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM room_members WHERE code = ? AND user_id = ?`, code, user); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM rooms WHERE code = ? AND NOT EXISTS (SELECT 1 FROM room_members WHERE code = ?)`, code, code); err != nil {
		return err
	}
	return tx.Commit()
}
