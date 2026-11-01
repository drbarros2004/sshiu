package store

import (
	"crypto/rand"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"plantao/internal/pet"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

//go:embed schema.sql
var schema string

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
	d.Exec(`ALTER TABLE users ADD COLUMN pet_color TEXT NOT NULL DEFAULT ''`)
	d.Exec(`ALTER TABLE users ADD COLUMN today TEXT NOT NULL DEFAULT ''`)
	d.Exec(`ALTER TABLE users ADD COLUMN today_min REAL NOT NULL DEFAULT 0`)
	d.Exec(`ALTER TABLE users ADD COLUMN garden_wallpaper TEXT NOT NULL DEFAULT ''`)
	d.Exec(`ALTER TABLE users ADD COLUMN garden_clock INTEGER NOT NULL DEFAULT 1`)
	d.Exec(`ALTER TABLE users ADD COLUMN garden_task INTEGER NOT NULL DEFAULT 1`)
	d.Exec(`ALTER TABLE users ADD COLUMN garden_room INTEGER NOT NULL DEFAULT 1`)
	d.Exec(`ALTER TABLE rooms ADD COLUMN max_members INTEGER NOT NULL DEFAULT 10`)
	d.Exec(`ALTER TABLE tasks ADD COLUMN room_code TEXT NOT NULL DEFAULT ''`)
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

// SetPetColor guarda a cor escolhida para o pet ("" volta para a cor do humor).
func SetPetColor(id, c string) error {
	_, err := db.Exec(`UPDATE users SET pet_color = ? WHERE id = ?`, c, id)
	return err
}

func scanUser(row *sql.Row) (nick string, p pet.Pet, ok bool, err error) {
	var last int64
	err = row.Scan(&nick, &p.Energy, &last, &p.Streak, &p.StreakDay, &p.Today, &p.TodayMin, &p.Color)
	if errors.Is(err, sql.ErrNoRows) {
		return "", pet.Pet{}, false, nil
	}
	p.LastActive = time.Unix(last, 0)
	return nick, p, err == nil, err
}

const selectUser = `SELECT nick, energy, last_active, streak, streak_day, today, today_min, pet_color FROM users WHERE id = ?`

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
	Room        string // a sala em que a tarefa começou: o foco é do usuário (ADR 0002), a sala é onde ele estava
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
	if _, err = tx.Exec(`INSERT INTO tasks (user_id, started_at, ended_at, task, declared_min, focused_sec, result, room_code)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, t.Start.Unix(), t.End.Unix(), t.Task, t.DeclaredMin, int(t.Focused.Seconds()), t.Result, t.Room); err != nil {
		return pet.Pet{}, err
	}
	if _, err = tx.Exec(`DELETE FROM open_tasks WHERE user_id = ?`, id); err != nil {
		return pet.Pet{}, err
	}
	return p, tx.Commit()
}

// OpenTask: a tarefa em andamento, gravada ao começar e ao continuar para sobreviver a um restart.
type OpenTask struct {
	User, Task, Room      string
	Start, SegStart, Ends time.Time
	Declared              int
	Counted               time.Duration
}

func SaveOpenTask(t OpenTask) error {
	_, err := db.Exec(`INSERT OR REPLACE INTO open_tasks (user_id, started_at, seg_started_at, ends_at, task, declared_min, counted_sec, room_code)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.User, t.Start.Unix(), t.SegStart.Unix(), t.Ends.Unix(), t.Task, t.Declared, int(t.Counted.Seconds()), t.Room)
	return err
}

// OpenTasks: as tarefas que estavam em andamento quando o servidor caiu.
func OpenTasks() ([]OpenTask, error) {
	rows, err := db.Query(`SELECT user_id, started_at, seg_started_at, ends_at, task, declared_min, counted_sec, room_code FROM open_tasks`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OpenTask
	for rows.Next() {
		var t OpenTask
		var start, seg, ends, sec int64
		if err := rows.Scan(&t.User, &start, &seg, &ends, &t.Task, &t.Declared, &sec, &t.Room); err != nil {
			return nil, err
		}
		t.Start, t.SegStart, t.Ends, t.Counted = time.Unix(start, 0), time.Unix(seg, 0), time.Unix(ends, 0), time.Duration(sec)*time.Second
		out = append(out, t)
	}
	return out, rows.Err()
}

// Tasks: as últimas tarefas do usuário, da mais nova para a mais velha.
func Tasks(id string) ([]TaskRow, error) {
	rows, err := db.Query(`SELECT started_at, ended_at, task, declared_min, focused_sec, result, room_code FROM tasks
		WHERE user_id = ? ORDER BY id DESC LIMIT 8`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TaskRow
	for rows.Next() {
		var r TaskRow
		var start, end, sec int64
		if err := rows.Scan(&start, &end, &r.Task, &r.DeclaredMin, &sec, &r.Result, &r.Room); err != nil {
			return nil, err
		}
		r.Start, r.End, r.Focused = time.Unix(start, 0), time.Unix(end, 0), time.Duration(sec)*time.Second
		out = append(out, r)
	}
	return out, rows.Err()
}

// --- o histórico, para o calendário e os números do perfil ---

// FocusByDay: os minutos focados em cada dia, de `from` para cá, com a chave "2006-01-02".
// O dia é o do fuso do servidor, o mesmo que o pet usa para o streak.
func FocusByDay(id string, from time.Time) (map[string]float64, error) {
	rows, err := db.Query(`SELECT date(started_at, 'unixepoch', 'localtime'), SUM(focused_sec)
		FROM tasks WHERE user_id = ? AND started_at >= ? GROUP BY 1`, id, from.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]float64{}
	for rows.Next() {
		var day string
		var sec float64
		if err := rows.Scan(&day, &sec); err != nil {
			return nil, err
		}
		out[day] = sec / 60
	}
	return out, rows.Err()
}

// Totals: o resumo de todas as tarefas encerradas de um usuário.
type Totals struct {
	Tasks, Done       int
	Focused, Declared time.Duration // somadas: a média sai dividindo por Tasks
}

func GetTotals(id string) (Totals, error) {
	var t Totals
	var sec, min int64
	err := db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(result = 'done'), 0), COALESCE(SUM(focused_sec), 0),
		COALESCE(SUM(declared_min), 0) FROM tasks WHERE user_id = ?`, id).Scan(&t.Tasks, &t.Done, &sec, &min)
	t.Focused, t.Declared = time.Duration(sec)*time.Second, time.Duration(min)*time.Minute
	return t, err
}

type RoomRow struct {
	Code, Name string
	Max        int // quantos membros a sala aceita
}

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
	rows, err := db.Query(`SELECT code, name, max_members FROM rooms`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RoomRow
	for rows.Next() {
		var r RoomRow
		if err := rows.Scan(&r.Code, &r.Name, &r.Max); err != nil {
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

func CreateRoom(owner, name string, max int) (RoomRow, error) {
	r := RoomRow{newCode(), name, max}
	tx, err := db.Begin()
	if err != nil {
		return r, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO rooms (code, name, owner, slots, max_members) VALUES (?, ?, ?, '', ?)`,
		r.Code, name, owner, max); err != nil {
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
