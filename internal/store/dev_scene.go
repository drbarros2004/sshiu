package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sshiu/internal/pet"
)

// DevWrite: o estado validado de um usuário, salvo junto das demais mudanças do editor.
type DevWrite struct {
	ID, Nick string
	Pet      pet.Pet
	New      bool
	Finish   *TaskRow
	History  *TaskRow
	Coins    *int
	Open     *OpenTask
}

func DevState(key string, value any) error {
	var data string
	err := db.QueryRow(`SELECT value FROM dev_state WHERE key = ?`, key).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), value)
}

// SaveDevScene é atômico: erro de nome, tarefa ou banco não deixa metade do rascunho salva.
func SaveDevScene(writes []DevWrite, room *RoomRow, members []string, state any) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, w := range writes {
		if w.New {
			if _, err = tx.Exec(`INSERT INTO users (id,nick,energy,last_active) VALUES (?,?,?,?)`, w.ID, w.Nick, w.Pet.Energy, w.Pet.LastActive.Unix()); err != nil {
				return err
			}
		}
		p := w.Pet
		if p.Appearance.Character != "" {
			if err = initializeWallet(tx, w.ID); err != nil {
				return err
			}
			for _, key := range pet.AppearanceItems(p.Look()) {
				if err = grantItem(tx, w.ID, key); err != nil {
					return err
				}
			}
		}
		if _, err = tx.Exec(`UPDATE users SET nick=?,energy=?,last_active=?,streak=?,streak_day=?,today=?,today_min=?,pet_color=?,dev_energy=?,pet_character=?,pet_variant=?,pet_hat=?,pet_hat_variant=? WHERE id=?`, w.Nick, p.Energy, p.LastActive.Unix(), p.Streak, p.StreakDay, p.Today, p.TodayMin, p.Color, p.DirectEnergy, p.Appearance.Character, p.Appearance.Variant, p.Appearance.Hat, p.Appearance.HatVariant, w.ID); err != nil {
			return err
		}
		if w.Coins != nil {
			if err = initializeWallet(tx, w.ID); err != nil {
				return err
			}
			if *w.Coins < 0 || *w.Coins > pet.Coins.Maximum {
				return errors.New("saldo inválido")
			}
			if _, err = tx.Exec(`UPDATE wallets SET balance=? WHERE user_id=?`, *w.Coins, w.ID); err != nil {
				return err
			}
		}
		if w.History != nil {
			t := w.History
			var recorded int
			if err = tx.QueryRow(`SELECT COUNT(*) FROM tasks t JOIN coin_receipts r ON r.task_id=t.id WHERE t.user_id=? AND t.started_at=? AND t.ended_at=? AND t.task=?`, w.ID, t.Start.Unix(), t.End.Unix(), t.Task).Scan(&recorded); err != nil {
				return err
			}
			if recorded == 0 {
				result, e := tx.Exec(`INSERT INTO tasks(user_id,started_at,ended_at,task,declared_min,focused_sec,result,room_code) VALUES (?,?,?,?,?,?,?,?)`, w.ID, t.Start.Unix(), t.End.Unix(), t.Task, t.DeclaredMin, int(t.Focused.Seconds()), pet.Done, t.Room)
				if e != nil {
					return e
				}
				taskID, e := result.LastInsertId()
				if e != nil {
					return e
				}
				if _, e = settleCoins(tx, w.ID, taskID, int(t.Focused.Seconds()), true); e != nil {
					return e
				}
			}
		}
		if w.Finish != nil {
			if _, err = recordTask(tx, w.ID, *w.Finish); err != nil {
				return err
			}
		}
		if w.Open != nil {
			f := w.Open
			_, err = tx.Exec(`INSERT OR REPLACE INTO open_tasks (user_id,started_at,seg_started_at,ends_at,task,declared_min,counted_sec,room_code) VALUES (?,?,?,?,?,?,?,?)`, w.ID, f.Start.Unix(), f.SegStart.Unix(), f.Ends.Unix(), f.Task, f.Declared, int(f.Counted.Seconds()), f.Room)
		} else {
			_, err = tx.Exec(`DELETE FROM open_tasks WHERE user_id=?`, w.ID)
		}
		if err != nil {
			return err
		}
	}
	if room != nil {
		r := room
		_, err = tx.Exec(`INSERT INTO rooms(code,name,owner,max_members,closed,obra,obra_meta,board_open,board_rank) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT(code) DO UPDATE SET name=excluded.name,owner=excluded.owner,max_members=excluded.max_members,closed=excluded.closed,obra=excluded.obra,obra_meta=excluded.obra_meta,board_open=excluded.board_open,board_rank=excluded.board_rank`, r.Code, r.Name, r.Owner, r.Max, r.Closed, r.Obra, r.ObraMeta, r.BoardOpen, r.BoardRank)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(`DELETE FROM room_members WHERE code=?`, r.Code); err != nil {
			return err
		}
		for _, id := range members {
			if _, err = tx.Exec(`INSERT INTO room_members(code,user_id) VALUES (?,?)`, r.Code, id); err != nil {
				return err
			}
		}
		if len(members) == 0 {
			if _, err = tx.Exec(`DELETE FROM rooms WHERE code=?`, r.Code); err != nil {
				return err
			}
		}
	}
	if _, err = tx.Exec(`INSERT OR REPLACE INTO dev_state(key,value) VALUES ('inline',?)`, string(data)); err != nil {
		return err
	}
	return tx.Commit()
}
