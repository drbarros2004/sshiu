package store

// O que só o modo dev usa (flag -dev): listar todo mundo, mudar os números do pet direto e gravar
// tarefas de mentira, para testar o servidor sem focar de verdade.

import (
	"database/sql"
	"errors"
	"sshiu/internal/pet"
	"time"
)

// DevTask: o texto das tarefas que o painel dev grava. É por ele que o painel acha as que pode apagar.
const DevTask = "tarefa dev"

type DevRow struct {
	ID, Nick string
	Pet      pet.Pet
	Before   float64 // minutos focados em tarefas que terminaram antes de hoje (o dia é o do fim, como no pet e na tabela de foco)
}

// DevUsers: todo mundo do banco, em ordem de apelido, com os minutos de antes de hoje.
func DevUsers(now time.Time) ([]DevRow, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rows, err := db.Query(`SELECT id, nick, energy, last_active, streak, streak_day, today, today_min, pet_color,
		COALESCE((SELECT SUM(focused_sec) FROM tasks t WHERE t.user_id = users.id AND t.ended_at < ?), 0)
		FROM users ORDER BY nick`, today.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DevRow
	for rows.Next() {
		var r DevRow
		var last int64
		var sec float64
		if err := rows.Scan(&r.ID, &r.Nick, &r.Pet.Energy, &last, &r.Pet.Streak, &r.Pet.StreakDay,
			&r.Pet.Today, &r.Pet.TodayMin, &r.Pet.Color, &sec); err != nil {
			return nil, err
		}
		r.Pet.LastActive, r.Before = time.Unix(last, 0), sec/60
		out = append(out, r)
	}
	return out, rows.Err()
}

// SetPet grava o pet inteiro como está.
func SetPet(id string, p pet.Pet) error {
	_, err := db.Exec(`UPDATE users SET energy = ?, last_active = ?, streak = ?, streak_day = ?, today = ?, today_min = ?, pet_color = ? WHERE id = ?`,
		p.Energy, p.LastActive.Unix(), p.Streak, p.StreakDay, p.Today, p.TodayMin, p.Color, id)
	return err
}

// AddDevTask grava uma tarefa cumprida de min minutos, sem mexer no pet.
func AddDevTask(id string, start time.Time, min int) error {
	_, err := db.Exec(`INSERT INTO tasks (user_id, started_at, ended_at, task, declared_min, focused_sec, result)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, id, start.Unix(), start.Add(time.Duration(min)*time.Minute).Unix(),
		DevTask, min, min*60, pet.Done)
	return err
}

// DevTasks: quantas tarefas dev o usuário tem que começaram de from (inclusive) até to.
func DevTasks(id string, from, to time.Time) (n int, err error) {
	err = db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE user_id = ? AND task = ? AND started_at >= ? AND started_at < ?`,
		id, DevTask, from.Unix(), to.Unix()).Scan(&n)
	return n, err
}

// DeleteDevTask apaga a tarefa dev mais nova que começou entre from e to e devolve os minutos dela
// (0 se não havia nenhuma).
func DeleteDevTask(id string, from, to time.Time) (float64, error) {
	var tid, sec int64
	err := db.QueryRow(`SELECT id, focused_sec FROM tasks WHERE user_id = ? AND task = ? AND started_at >= ? AND started_at < ?
		ORDER BY id DESC LIMIT 1`, id, DevTask, from.Unix(), to.Unix()).Scan(&tid, &sec)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil // nenhuma: nada a apagar
	}
	if err != nil {
		return 0, err
	}
	_, err = db.Exec(`DELETE FROM tasks WHERE id = ?`, tid)
	return float64(sec) / 60, err
}
