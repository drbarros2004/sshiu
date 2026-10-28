package store

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"plantao/internal/pet"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id          TEXT PRIMARY KEY, -- fingerprint da chave SSH
	nick        TEXT NOT NULL UNIQUE,
	energy      REAL NOT NULL,
	last_active INTEGER NOT NULL,
	streak      INTEGER NOT NULL DEFAULT 0,
	streak_day  TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS sprints (
	id         INTEGER PRIMARY KEY,
	user_id    TEXT NOT NULL REFERENCES users(id),
	started_at INTEGER NOT NULL,
	lock       TEXT NOT NULL,
	result     TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS rooms (
	code  TEXT PRIMARY KEY, -- o convite: quem tem o código entra
	name  TEXT NOT NULL,
	owner TEXT NOT NULL REFERENCES users(id),
	slots TEXT NOT NULL     -- horas dos sprints, ex. "9,14,20"
);
CREATE TABLE IF NOT EXISTS room_members (
	code    TEXT NOT NULL REFERENCES rooms(code),
	user_id TEXT NOT NULL REFERENCES users(id),
	PRIMARY KEY (code, user_id)
);
CREATE TABLE IF NOT EXISTS friends ( -- "seguir", mão única, como no GitHub
	user_id   TEXT NOT NULL REFERENCES users(id),
	friend_id TEXT NOT NULL REFERENCES users(id),
	PRIMARY KEY (user_id, friend_id)
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
	db = d
	return nil
}

// Rooms lista todas as salas privadas.
func Rooms() ([]RoomRow, error) {
	return queryRooms(`SELECT code, name, owner, slots FROM rooms`)
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
	err = row.Scan(&nick, &p.Energy, &last, &p.Streak, &p.StreakDay)
	if errors.Is(err, sql.ErrNoRows) {
		return "", pet.Pet{}, false, nil
	}
	p.LastActive = time.Unix(last, 0)
	return nick, p, err == nil, err
}

const selectUser = `SELECT nick, energy, last_active, streak, streak_day FROM users WHERE id = ?`

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

// Record guarda o resultado do sprint e atualiza o humor do pet.
func Record(id string, start time.Time, lock string, r pet.Result) (pet.Pet, error) {
	tx, err := db.Begin()
	if err != nil {
		return pet.Pet{}, err
	}
	defer tx.Rollback()
	_, p, _, err := scanUser(tx.QueryRow(selectUser, id))
	if err != nil {
		return pet.Pet{}, err
	}
	now := time.Now()
	p.Apply(r, now)
	if _, err = tx.Exec(`UPDATE users SET energy = ?, last_active = ?, streak = ?, streak_day = ? WHERE id = ?`,
		p.Energy, now.Unix(), p.Streak, p.StreakDay, id); err != nil {
		return pet.Pet{}, err
	}
	if _, err = tx.Exec(`INSERT INTO sprints (user_id, started_at, lock, result) VALUES (?, ?, ?, ?)`,
		id, start.Unix(), lock, r); err != nil {
		return pet.Pet{}, err
	}
	return p, tx.Commit()
}

type SprintRow struct {
	At     time.Time
	Lock   string
	Result pet.Result
}

func History(id string) ([]SprintRow, error) {
	rows, err := db.Query(`SELECT started_at, lock, result FROM sprints WHERE user_id = ? ORDER BY id DESC LIMIT 8`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SprintRow
	for rows.Next() {
		var r SprintRow
		var at int64
		if err := rows.Scan(&at, &r.Lock, &r.Result); err != nil {
			return nil, err
		}
		r.At = time.Unix(at, 0)
		out = append(out, r)
	}
	return out, rows.Err()
}

type Friend struct {
	ID, Nick string
	Pet      pet.Pet
}

func FriendsOf(id string) ([]Friend, error) {
	rows, err := db.Query(`SELECT u.id, u.nick, u.energy, u.last_active, u.streak, u.streak_day
		FROM friends f JOIN users u ON u.id = f.friend_id WHERE f.user_id = ? ORDER BY u.nick`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Friend
	for rows.Next() {
		var f Friend
		var last int64
		if err := rows.Scan(&f.ID, &f.Nick, &f.Pet.Energy, &last, &f.Pet.Streak, &f.Pet.StreakDay); err != nil {
			return nil, err
		}
		f.Pet.LastActive = time.Unix(last, 0)
		out = append(out, f)
	}
	return out, rows.Err()
}

func AddFriend(id, nick string) error {
	var fid string
	err := db.QueryRow(`SELECT id FROM users WHERE nick = ?`, nick).Scan(&fid)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("ninguém com o apelido %q", nick)
	case err != nil:
		return err
	case fid == id:
		return errors.New("você já é seu próprio amigo :)")
	}
	_, err = db.Exec(`INSERT OR IGNORE INTO friends (user_id, friend_id) VALUES (?, ?)`, id, fid)
	return err
}

func RemoveFriend(id, fid string) error {
	_, err := db.Exec(`DELETE FROM friends WHERE user_id = ? AND friend_id = ?`, id, fid)
	return err
}

type RoomRow struct {
	Code, Name, Owner string
	Slots             []int
}

// newCode: 6 letras sem as ambíguas (0/o, 1/l/i), fácil de ditar para um amigo.
func newCode() string {
	const alphabet = "abcdefghjkmnpqrstuvwxyz23456789"
	b := make([]byte, 6)
	rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

func CreateRoom(owner, name string, slots []int) (RoomRow, error) {
	r := RoomRow{newCode(), name, owner, slots}
	tx, err := db.Begin()
	if err != nil {
		return r, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO rooms (code, name, owner, slots) VALUES (?, ?, ?, ?)`, r.Code, name, owner, FormatSlots(slots)); err != nil {
		return r, err
	}
	if _, err := tx.Exec(`INSERT INTO room_members (code, user_id) VALUES (?, ?)`, r.Code, owner); err != nil {
		return r, err
	}
	return r, tx.Commit()
}

func JoinRoom(user, code string) (RoomRow, error) {
	rs, err := queryRooms(`SELECT code, name, owner, slots FROM rooms WHERE code = ?`, code)
	if err != nil {
		return RoomRow{}, err
	}
	if len(rs) == 0 {
		return RoomRow{}, fmt.Errorf("nenhuma sala com o código %q", code)
	}
	_, err = db.Exec(`INSERT OR IGNORE INTO room_members (code, user_id) VALUES (?, ?)`, code, user)
	return rs[0], err
}

func MyRooms(user string) ([]RoomRow, error) {
	return queryRooms(`SELECT r.code, r.name, r.owner, r.slots FROM room_members m JOIN rooms r ON r.code = m.code
		WHERE m.user_id = ? ORDER BY r.name`, user)
}

func queryRooms(q string, args ...any) ([]RoomRow, error) {
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RoomRow
	for rows.Next() {
		var r RoomRow
		var slots string
		if err := rows.Scan(&r.Code, &r.Name, &r.Owner, &slots); err != nil {
			return nil, err
		}
		if r.Slots, err = ParseSlots(slots); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ParseSlots lê "9, 14,20" como [9 14 20].
func ParseSlots(s string) ([]int, error) {
	var out []int
	for _, f := range strings.Split(s, ",") {
		h, err := strconv.Atoi(strings.TrimSpace(f))
		if err != nil || h < 0 || h > 23 {
			return nil, fmt.Errorf("horário inválido %q: use horas de 0 a 23 separadas por vírgula", strings.TrimSpace(f))
		}
		if !slices.Contains(out, h) {
			out = append(out, h)
		}
	}
	sort.Ints(out)
	return out, nil
}

func FormatSlots(slots []int) string {
	var s []string
	for _, h := range slots {
		s = append(s, strconv.Itoa(h))
	}
	return strings.Join(s, ",")
}
