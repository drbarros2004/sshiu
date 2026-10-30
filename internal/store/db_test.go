package store

import (
	"plantao/internal/pet"
	"testing"
	"time"
)

func TestRecordTask(t *testing.T) {
	if err := Open(t.TempDir() + "/t.db"); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := PutUser("u", "ana", pet.EnergyStart); err != nil {
		t.Fatal(err)
	}
	end := time.Now()
	p, err := RecordTask("u", TaskRow{end.Add(-20 * time.Minute), end, "lista 3", 25, 15 * time.Minute, pet.Done})
	if err != nil || p.Energy != 57 || p.Streak != 1 { // 15 min = 4, + bônus 3
		t.Fatalf("got %+v %v", p, err)
	}
	if _, saved, _, _ := GetUser("u"); saved.Energy != 57 || saved.TodayMin != 15 {
		t.Fatalf("saved %+v", saved)
	}
	if ts, err := Tasks("u"); err != nil || len(ts) != 1 || ts[0].Focused != 15*time.Minute || ts[0].Result != pet.Done {
		t.Fatalf("tasks %+v %v", ts, err)
	}

	if g := GetGarden("u"); g != DefaultGarden {
		t.Fatalf("garden padrão %+v", g)
	}
	want := Garden{Wallpaper: "rain", Clock: 2, Task: false, Room: true}
	if err := SetGarden("u", want); err != nil || GetGarden("u") != want {
		t.Fatalf("garden %+v %v", GetGarden("u"), err)
	}

	r, err := CreateRoom("u", "turma", 6)
	if err != nil {
		t.Fatal(err)
	}
	if rs, _ := Rooms(); len(rs) != 1 || rs[0].Max != 6 { // o tamanho escolhido volta do banco
		t.Fatalf("rooms %+v", rs)
	}
	if err := JoinRoom("u", "nao-existe"); err == nil {
		t.Fatal("código inexistente deveria falhar")
	}
	if err := LeaveRoom("u", r.Code); err != nil {
		t.Fatal(err)
	}
	if rs, _ := Rooms(); len(rs) != 0 { // o último saiu: a sala some
		t.Fatalf("rooms %+v", rs)
	}
}
