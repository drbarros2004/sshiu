package rooms

import (
	"sshiu/internal/store"
	"sync"
	"testing"
	"time"
)

// O painel dev muda o banco e a memória juntos, e os minutos viram tarefas que a tabela de foco soma.
func TestDevAdjust(t *testing.T) {
	if err := store.Open(":memory:"); err != nil {
		t.Fatal(err)
	}
	id, err := DevCreate("ana")
	if err != nil || id != "dev:ana" {
		t.Fatalf("criar: %q %v", id, err)
	}
	if _, err := DevCreate("a b"); err == nil {
		t.Fatal("aceitou apelido com espaço")
	}
	if _, err := DevCreate("ana"); err == nil {
		t.Fatal("criou a ana de novo")
	}
	now := time.Date(2026, 11, 10, 15, 0, 0, 0, time.Local)
	midnight := time.Date(2026, 11, 10, 0, 0, 0, 0, time.Local)
	for _, a := range []struct {
		f DevField
		d int
	}{{DevEnergy, 1}, {DevToday, 1}, {DevToday, 1}, {DevToday, -1}, {DevBefore, 1}, {DevBefore, 1}, {DevStreak, 1}, {DevColor, 1}} {
		if err := DevAdjust(id, a.f, a.d, now); err != nil {
			t.Fatal(err)
		}
	}
	_, p, _, _ := store.GetUser(id)
	if p.Energy != 55 || p.TodayMin != 15 || p.StreakAt(now) != 1 || p.Color == "" {
		t.Fatalf("pet: %+v", p)
	}
	if u := users[id]; u.todayMin != 15 || u.color != p.Color {
		t.Fatalf("a memória não acompanhou: %+v", u)
	}
	rows, _ := store.DevUsers(midnight)
	if len(rows) != 1 || rows[0].Before != 2*DevBeforeStep {
		t.Fatalf("antes: %+v", rows)
	}
	// vários "+" ao mesmo tempo, como os Cmd do Bubble Tea: nenhum se perde
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() { DevAdjust(id, DevToday, 1, now) })
	}
	wg.Wait()
	if _, p, _, _ := store.GetUser(id); p.TodayMin != 15+4*DevTodayStep {
		t.Fatalf("ajustes em paralelo: %v", p.TodayMin)
	}
	// as de antes: uma por dia para trás (ontem e anteontem)
	for i := 1; i <= 2; i++ {
		from := midnight.AddDate(0, 0, -i)
		if n, _ := store.DevTasks(id, from, from.AddDate(0, 0, 1)); n != 1 {
			t.Fatalf("%d dias atrás: %d tarefas", i, n)
		}
	}
	// a tarefa que atravessa a meia-noite é de hoje, como no pet e na tabela de foco: fica fora de "antes"
	if _, err := store.RecordTask(id, store.TaskRow{Start: midnight.Add(-10 * time.Minute), End: midnight.Add(20 * time.Minute),
		Task: "virada", Focused: 30 * time.Minute}); err != nil {
		t.Fatal(err)
	}
	if rows, _ := store.DevUsers(now); rows[0].Before != 2*DevBeforeStep {
		t.Fatalf("antes com a tarefa da virada: %v", rows[0].Before)
	}
}

// A ana do modo dev não impede a demo de subir: o bot fica com outro apelido.
func TestPutBotNickTaken(t *testing.T) {
	if err := store.Open(":memory:"); err != nil {
		t.Fatal(err)
	}
	if _, err := DevCreate("ana"); err != nil {
		t.Fatal(err)
	}
	if err := putBot(bots[0]); err != nil {
		t.Fatal(err)
	}
	if nick, _, _, _ := store.GetUser("bot:ana"); nick != "ana-bot" {
		t.Fatalf("apelido do bot: %q", nick)
	}
	if err := putBot(bots[0]); err != nil { // o bot já existe: a segunda subida também passa
		t.Fatal(err)
	}
}

// Presença e tarefa de mentira: a sala vê a pessoa como se ela estivesse conectada.
func TestDevPresence(t *testing.T) {
	if err := store.Open(":memory:"); err != nil {
		t.Fatal(err)
	}
	id, _ := DevCreate("bia")
	u := users[id]
	now := time.Now()
	if DevAdjust(id, DevOnline, 1, now); u.status() != Idle {
		t.Fatalf("online: %v", u.status())
	}
	if DevAdjust(id, DevOnline, 1, now); u.online != 1 { // ligar de novo não abre outra sessão
		t.Fatalf("sessões: %d", u.online)
	}
	DevAdjust(id, DevFocus, 1, now) // sem tarefa: começa uma
	if u.status() != Focusing || u.focus.Declared != DevFocusMin {
		t.Fatalf("tarefa: %v %+v", u.status(), u.focus)
	}
	ends := u.focus.Ends
	DevAdjust(id, DevFocus, -1, now) // 5 min passaram
	if got := u.focus.Ends; got != ends.Add(-DevFocusStep*time.Minute) || u.focus.Focused(time.Now()) < DevFocusStep*time.Minute {
		t.Fatalf("passou: %v %v", got, u.focus.Focused(time.Now()))
	}
	DevStart(id, "", "ler o cap. 4")
	if u.focus.Task != "ler o cap. 4" || u.focus.Ends != ends.Add(-DevFocusStep*time.Minute) {
		t.Fatalf("trocar o texto mexeu no tempo: %+v", u.focus)
	}
	DevAdjust(id, DevFocus, -5, now) // passou do tempo: esgotou
	if !u.focus.Expired(now) {
		t.Fatalf("não esgotou: %+v", u.focus)
	}
	DevAdjust(id, DevFocus, 1, now) // mais tempo depois de esgotar: o tempo sem resposta não conta
	if u.focus.Expired(now) || u.focus.Focused(now) != DevFocusMin*time.Minute {
		t.Fatalf("continuou: %v %+v", u.focus.Focused(now), u.focus)
	}
	if DevAdjust(id, DevOnline, -1, now); u.status() != Offline || u.focus == nil { // a tarefa espera a tolerância
		t.Fatalf("offline: %v", u.status())
	}
	DevAdjust(id, DevIdle, 2, now)
	if _, p, _, _ := store.GetUser(id); now.Sub(p.LastActive) < 47*time.Hour {
		t.Fatalf("parado: %v", p.LastActive)
	}
}
