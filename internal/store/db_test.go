package store

import (
	"plantao/internal/pet"
	"slices"
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
	p, err := RecordTask("u", TaskRow{end.Add(-20 * time.Minute), end, "lista 3", 25, 15 * time.Minute, pet.Done, "abc123"})
	if err != nil || p.Energy != 57 || p.Streak != 1 { // 15 min = 4, + bônus 3
		t.Fatalf("got %+v %v", p, err)
	}
	if _, saved, _, _ := GetUser("u"); saved.Energy != 57 || saved.TodayMin != 15 {
		t.Fatalf("saved %+v", saved)
	}
	if ts, err := Tasks("u"); err != nil || len(ts) != 1 || ts[0].Focused != 15*time.Minute ||
		ts[0].Result != pet.Done || ts[0].Room != "abc123" { // a sala volta do banco: as stats da sala dependem dela
		t.Fatalf("tasks %+v %v", ts, err)
	}

	start := end.Add(-20 * time.Minute) // o dia do gráfico é o do início da tarefa
	if days, err := FocusByDay("u", start.AddDate(0, 0, -1)); err != nil || days[start.Format("2006-01-02")] != 15 {
		t.Fatalf("por dia %v %v", days, err)
	}
	if got, err := GetTotals("u"); err != nil || got.Tasks != 1 || got.Done != 1 ||
		got.Focused != 15*time.Minute || got.Declared != 25*time.Minute {
		t.Fatalf("totais %+v %v", got, err)
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

// Os membros voltam do banco na ordem em que entraram, não na do id: é dela que sai o membro mais
// antigo, que fica com a sala quando o dono sai (ADR 0007).
func TestRoomMembersOrder(t *testing.T) {
	if err := Open(t.TempDir() + "/t.db"); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"zeca", "ana", "mia"} {
		if err := PutUser(id, id, pet.EnergyStart); err != nil {
			t.Fatal(err)
		}
	}
	r, err := CreateRoom("zeca", "turma", 6)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"mia", "ana"} {
		if err := JoinRoom(id, r.Code); err != nil {
			t.Fatal(err)
		}
	}
	if got, err := RoomMembers(r.Code); err != nil || !slices.Equal(got, []string{"zeca", "mia", "ana"}) {
		t.Fatalf("membros %v %v", got, err)
	}
}
