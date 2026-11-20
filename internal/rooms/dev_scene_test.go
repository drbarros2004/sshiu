package rooms

import (
	"path/filepath"
	"sshiu/internal/pet"
	"sshiu/internal/store"
	"testing"
	"time"
)

func setupInline(t *testing.T) (string, string, string) {
	t.Helper()
	old := Dev
	Dev = true
	t.Cleanup(func() { Dev = old })
	mu.Lock()
	users = map[string]*user{}
	rooms = map[string]*room{}
	devSaved = devState{Presence: map[string]bool{}, Progress: map[string]map[string]devOffset{}}
	mu.Unlock()
	if err := store.Open(filepath.Join(t.TempDir(), "dev.db")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	a, e := DevCreate("ana")
	if e != nil {
		t.Fatal(e)
	}
	b, e := DevCreate("bia")
	if e != nil {
		t.Fatal(e)
	}
	code, e := Create(a, "teste", 5)
	if e != nil {
		t.Fatal(e)
	}
	if e = Join(b, code); e != nil {
		t.Fatal(e)
	}
	return a, b, code
}
func mustScene(t *testing.T, code string) DevScene {
	t.Helper()
	s, e := DevRead(code)
	if e != nil {
		t.Fatal(e)
	}
	return s
}

func TestInlinePreviewSaveAndRollback(t *testing.T) {
	a, b, code := setupInline(t)
	now := time.Now()
	base := mustScene(t, code)
	name := "bia-nova"
	energy := 19.0
	stage := 2
	on := true
	obra := "trem"
	changes := []DevChange{{ID: b, Nick: &name, Energy: &energy, Stage: &stage, Online: &on}}
	preview, e := DevPreview(base, changes, &obra, now)
	if e != nil {
		t.Fatal(e)
	}
	w := DevWorld(SnapshotWorld(), preview, changes, now)
	if w.Users[b].Nick != name || w.Users[b].Energy != 19 || Stage(w.Rooms[code].Members[1].TodayMin, 60) != 2 {
		t.Fatalf("prévia incorreta: %+v", w)
	}
	nick, _, _, _ := store.GetUser(b)
	if nick != "bia" || SnapshotWorld().Users[b].Energy != 50 {
		t.Fatal("prévia gravou dados")
	}
	if _, e = DevSave(base, changes, &obra, now); e != nil {
		t.Fatal(e)
	}
	nick, p, _, _ := store.GetUser(b)
	if nick != name || p.EnergyAt(now) != 19 {
		t.Fatalf("salvo: %s %+v", nick, p)
	}
	if tasks, _ := store.Tasks(b); len(tasks) != 0 {
		t.Fatal("criou histórico para atingir estágio")
	}
	// A linha de progresso soma foco real ao estágio escolhido, sem alterar a tarefa.
	Start(b, code, "ler", 25)
	mu.Lock()
	users[b].focus.Start = now.Add(-21 * time.Minute)
	users[b].focus.SegStart = now.Add(-21 * time.Minute)
	users[b].focus.Ends = now.Add(4 * time.Minute)
	mu.Unlock()
	w = SnapshotWorld()
	if Stage(w.Rooms[code].Members[1].TodayMin, 60) != 3 {
		t.Fatal("a peça não progrediu a partir do estágio escolhido")
	}
	Finish(b, pet.Done)
	_, p, _, _ = store.GetUser(b)
	if p.EnergyAt(time.Now()) <= 19 {
		t.Fatal("energia ficou congelada depois do foco")
	}
	if len(SnapshotWorld().Rooms[code].Members) != 2 || a == b {
		t.Fatal("membros alterados")
	}
}

func TestInlineStaleAndInvalidSaveAreAtomic(t *testing.T) {
	a, b, code := setupInline(t)
	base := mustScene(t, code)
	invalid := 101.0
	name := "nome-valido"
	if _, e := DevSave(base, []DevChange{{ID: a, Nick: &name}, {ID: b, Energy: &invalid}}, nil, time.Now()); e == nil {
		t.Fatal("aceitou energia inválida")
	}
	if nick, _, _, _ := store.GetUser(a); nick != "ana" {
		t.Fatal("salvou parte do rascunho")
	}
	if e := DevAdjust(b, DevEnergy, 1, time.Now()); e != nil {
		t.Fatal(e)
	}
	if _, e := DevSave(base, []DevChange{{ID: a, Nick: &name}}, nil, time.Now()); e == nil {
		t.Fatal("sobrescreveu cenário modificado")
	}
}

func TestInlineCreationMembershipAndPresence(t *testing.T) {
	_, b, code := setupInline(t)
	base := mustScene(t, code)
	name := "caio"
	yes, no := true, false
	changes := []DevChange{{ID: "dev:caio", Create: true, Nick: &name, Member: &yes, Online: &yes}, {ID: b, Member: &no}}
	if _, e := DevPreview(base, changes, nil, time.Now()); e != nil {
		t.Fatal(e)
	}
	if _, _, ok, _ := store.GetUser("dev:caio"); ok {
		t.Fatal("prévia criou usuário órfão")
	}
	if _, e := DevSave(base, changes, nil, time.Now()); e != nil {
		t.Fatal(e)
	}
	if _, _, ok, _ := store.GetUser(b); !ok {
		t.Fatal("remover membro apagou a conta")
	}
	rows, _ := store.DevUsers(time.Now())
	if len(rows) != 3 {
		t.Fatal("novo usuário não consta no catálogo")
	}
	if _, e := DevSave(mustScene(t, code), []DevChange{{ID: b, Member: &yes, Online: &no}}, nil, time.Now()); e != nil {
		t.Fatal(e)
	}
	w := SnapshotWorld()
	if len(w.Rooms[code].Members) != 3 || w.Users[b].Status != Offline {
		t.Fatal("presença confundida com participação")
	}
}

func TestInlineEnergyKeepsRunningTaskAndStreak(t *testing.T) {
	a, _, code := setupInline(t)
	now := time.Now()
	_, p, _, _ := store.GetUser(a)
	p.Streak = 3
	p.StreakDay = now.Format(time.DateOnly)
	store.SetPet(a, p)
	Start(a, code, "estudar", 25)
	base := mustScene(t, code)
	focus := SnapshotWorld().Users[a].Focus
	energy := 2.0
	if _, e := DevSave(base, []DevChange{{ID: a, Energy: &energy}}, nil, now); e != nil {
		t.Fatal(e)
	}
	if got := SnapshotWorld().Users[a].Focus; !got.Start.Equal(focus.Start) || !got.SegStart.Equal(focus.SegStart) || !got.Ends.Equal(focus.Ends) || got.Task != focus.Task || got.Declared != focus.Declared {
		t.Fatal("ajuste de energia alterou o cronômetro")
	}
	_, p, _, _ = store.GetUser(a)
	if p.Streak != 3 || p.EnergyAt(now) != 2 {
		t.Fatalf("energia/streak: %+v", p)
	}
}

func TestInlineTasksSaveOnceAndPersist(t *testing.T) {
	a, _, code := setupInline(t)
	now := time.Now()
	base := mustScene(t, code)
	start := []DevChange{{ID: a, Action: "start", Task: "ler capítulo", Minutes: 20}}
	if _, e := DevPreview(base, start, nil, now); e != nil {
		t.Fatal(e)
	}
	if SnapshotWorld().Users[a].Focus.Task != "" {
		t.Fatal("prévia iniciou tarefa")
	}
	if _, e := DevSave(base, start, nil, now); e != nil {
		t.Fatal(e)
	}
	base = mustScene(t, code)
	finish := []DevChange{{ID: a, Action: "done"}}
	if _, e := DevSave(base, finish, nil, now.Add(10*time.Minute)); e != nil {
		t.Fatal(e)
	}
	if _, e := DevSave(base, finish, nil, now.Add(10*time.Minute)); e == nil {
		t.Fatal("aceitou conclusão duplicada")
	}
	tasks, _ := store.Tasks(a)
	if len(tasks) != 1 {
		t.Fatalf("%d tarefas", len(tasks))
	}
	// Simula só a memória perdida; o banco temporário continua intacto.
	stage := 1
	if _, e := DevSave(mustScene(t, code), []DevChange{{ID: a, Stage: &stage}}, nil, now); e != nil {
		t.Fatal(e)
	}
	mu.Lock()
	rooms = map[string]*room{}
	users = map[string]*user{}
	devSaved = devState{}
	mu.Unlock()
	if e := Load(); e != nil {
		t.Fatal(e)
	}
	if e := LoadDev(); e != nil {
		t.Fatal(e)
	}
	w := SnapshotWorld()
	if w.Users[a].Status != Idle || Stage(w.Rooms[code].Members[0].TodayMin, 60) != 1 {
		t.Fatal("cenário salvo não sobreviveu à reabertura")
	}
	mu.Lock()
	offset := devSaved.Progress[code][a]
	mu.Unlock()
	if offset.Day != now.Format(time.DateOnly) {
		t.Fatal("progresso não tem limite diário")
	}
}

func TestInlineUnavailableWithoutDev(t *testing.T) {
	old := Dev
	Dev = false
	defer func() { Dev = old }()
	if _, e := DevRead(""); e == nil {
		t.Fatal("editor disponível no modo normal")
	}
	if _, e := DevSave(DevScene{}, nil, nil, time.Now()); e == nil {
		t.Fatal("modo normal salvou dados dev")
	}
}

// A validação aceita nomes finais distintos, mas o UNIQUE imediato do SQLite pode recusar
// uma troca circular. Mesmo após um UPDATE bem-sucedido, a transação desfaz tudo.
func TestInlineDatabaseFailureRollsBackEarlierWrites(t *testing.T) {
	a, b, code := setupInline(t)
	c, err := DevCreate("caio")
	if err != nil {
		t.Fatal(err)
	}
	nameA, nameB, nameC := "novo-nome", "caio", "bia"
	base := mustScene(t, code)
	changes := []DevChange{{ID: a, Nick: &nameA}, {ID: b, Nick: &nameB}, {ID: c, Nick: &nameC}}
	if _, err = DevPreview(base, changes, nil, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err = DevSave(base, changes, nil, time.Now()); err == nil {
		t.Fatal("esperava conflito UNIQUE na troca circular")
	}
	if nick, _, _, _ := store.GetUser(a); nick != "ana" {
		t.Fatal("erro de banco deixou um UPDATE anterior salvo")
	}
	if SnapshotWorld().Users[a].Nick != "ana" {
		t.Fatal("erro de banco alterou a memória")
	}
}
