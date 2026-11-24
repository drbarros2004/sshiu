package tui

import (
	"path/filepath"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"sshiu/internal/store"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func inlineModel(t *testing.T, name string) model {
	t.Helper()
	old, oldRooms := Dev, rooms.Dev
	Dev, rooms.Dev = true, true
	t.Cleanup(func() { Dev, rooms.Dev = old, oldRooms })
	if e := store.Open(filepath.Join(t.TempDir(), "dev.db")); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { store.Close() })
	id, e := rooms.DevCreate(name)
	if e != nil {
		t.Fatal(e)
	}
	code, e := rooms.Create(id, "sala de teste", 5)
	if e != nil {
		t.Fatal(e)
	}
	m := New(id, "ssh teste", "")
	m.room = code
	m.now = time.Now()
	m.width, m.height = 120, 42
	return m
}
func pressInline(m model, code rune, ctrl bool) (model, tea.Cmd) {
	key := tea.KeyPressMsg{Code: code}
	if !ctrl && code >= 32 && code < 127 {
		key.Text = string(code)
	}
	if ctrl {
		key.Mod = tea.ModCtrl
	}
	next, cmd := m.key(key)
	return next.(model), cmd
}
func openInline(t *testing.T, m model) model {
	t.Helper()
	next, cmd := pressInline(m, 'x', false)
	if cmd == nil {
		t.Fatal("não abriu dev")
	}
	result, _ := next.Update(cmd())
	return result.(model)
}

func TestInlineCancelAndFieldPreview(t *testing.T) {
	m := openInline(t, inlineModel(t, "ui-cancel"))
	m, _ = pressInline(m, 'e', true)
	m.roomDev.number = 19
	rendered := m.previewField().inlineRender()
	if rendered.w.Users[m.id].Energy != 19 {
		t.Fatal("barra/card não usa a energia do campo")
	}
	if _, p, _, _ := store.GetUser(m.id); p.EnergyAt(time.Now()) != 50 {
		t.Fatal("prévia gravou")
	}
	next, _ := m.key(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	if len(m.roomDev.changes) != 0 {
		t.Fatal("esc do campo confirmou a energia")
	}
	next, _ = m.key(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(model)
	if m.roomDev != nil {
		t.Fatal("esc não saiu do editor")
	}
}

func TestInlineProfileColorWaitsForSave(t *testing.T) {
	m := inlineModel(t, "ui-profile")
	m.tab = tabProfile
	m = openInline(t, m)
	m.custom = true
	m.appearanceEdit = newAppearanceEditor(pet.Appearance{Character: "fruit", Variant: "yellow"})
	m, _ = pressInline(m, tea.KeyEnter, false)
	if m.profileDev.changes[0].Appearance == nil || m.profileDev.changes[0].Appearance.Variant != "yellow" {
		t.Fatal("enter não confirmou cor no rascunho")
	}
	if _, p, _, _ := store.GetUser(m.id); p.Color != "" {
		t.Fatal("enter do perfil dev salvou antes de ctrl+s")
	}
	m, cmd := pressInline(m, 's', true)
	result := cmd()
	next, _ := m.Update(result)
	m = next.(model)
	if _, p, _, _ := store.GetUser(m.id); p.Color != pet.Colors[2].ID {
		t.Fatal("ctrl+s não salvou cor")
	}
	if m.profileDev == nil {
		t.Fatal("ctrl+s saiu do dev")
	}
}

func TestInlineProfileNicknamePreviewCancelAndSave(t *testing.T) {
	m := inlineModel(t, "ui-profile-name")
	m.tab = tabProfile
	m = openInline(t, m)
	originalID, originalNick := m.id, m.nick
	m, _ = pressInline(m, 'n', true)
	if m.profileDev.field != "name" || m.in.Value() != originalNick {
		t.Fatal("perfil não abriu o apelido atual")
	}
	m.in.SetValue("novo-apelido")
	preview := m.previewField().inlineRender()
	if preview.nick != "novo-apelido" || !strings.Contains(ansi.Strip(preview.profileView(116, 32)), "novo-apelido") {
		t.Fatal("prévia não atualizou o título do perfil")
	}
	if nick, _, _, _ := store.GetUser(m.id); nick != originalNick {
		t.Fatal("prévia gravou o apelido")
	}
	m, _ = pressInline(m, tea.KeyEscape, false)
	if m.inlineRender().nick != originalNick || len(m.profileDev.changes) != 0 {
		t.Fatal("cancelar campo deixou o nome alterado")
	}
	m, _ = pressInline(m, 'n', true)
	m.in.SetValue("novo-apelido")
	m, _ = pressInline(m, tea.KeyEnter, false)
	if nick, _, _, _ := store.GetUser(m.id); nick != originalNick {
		t.Fatal("enter gravou antes do salvamento dev")
	}
	m, cmd := pressInline(m, 's', true)
	next, _ := m.Update(cmd())
	m = next.(model)
	if nick, _, _, _ := store.GetUser(m.id); nick != "novo-apelido" {
		t.Fatal("ctrl+s não gravou o apelido")
	}
	if m.nick != "novo-apelido" || m.id != originalID || m.profileDev == nil {
		t.Fatal("salvamento mudou identidade ou saiu do editor")
	}
}

func TestInlineProfileNicknameRejectsInvalidAndDuplicateNames(t *testing.T) {
	m := inlineModel(t, "ui-name-invalid")
	m.tab = tabProfile
	m = openInline(t, m)
	if _, e := rooms.DevCreate("nome-em-uso"); e != nil {
		t.Fatal(e)
	}
	// Reabre para incluir o novo usuário no cenário, como faria ao entrar no editor.
	m.profileDev = nil
	m = openInline(t, m)
	m, _ = pressInline(m, 'n', true)
	for _, name := range []string{"nome com espaço", "nome-em-uso"} {
		m.in.SetValue(name)
		m, _ = pressInline(m, tea.KeyEnter, false)
		if m.err == "" || m.profileDev.field != "name" || len(m.profileDev.changes) != 0 {
			t.Fatalf("aceitou %q", name)
		}
	}
}

func TestInlineIdentityDropsDraftsAndRejectsOldProfile(t *testing.T) {
	m := inlineModel(t, "ui-origin")
	id, e := rooms.DevCreate("ui-other")
	if e != nil {
		t.Fatal(e)
	}
	m.roomDev = &inlineDev{target: id}
	m.profileDev = &inlineDev{selected: m.id}
	next, _, ok := m.updateInline(m.assume(id)())
	if !ok {
		t.Fatal("troca não tratada")
	}
	m = next.(model)
	if m.id != id || m.original != "dev:ui-origin" || m.roomDev != nil || m.profileDev != nil {
		t.Fatal("identidade/rascunhos incorretos")
	}
	if New("dev:ui-origin", "", "").id != id {
		t.Fatal("identidade assumida não foi restaurada ao reconectar")
	}
	if e := rooms.DevAssume("dev:ui-origin", "dev:ui-origin"); e != nil {
		t.Fatal(e)
	}
	if New("dev:ui-origin", "", "").id != "dev:ui-origin" {
		t.Fatal("não voltou à identidade original")
	}
	next, _ = m.Update(profileMsg{id: "dev:ui-origin", pet: pet.Pet{Energy: 99}})
	if next.(model).pet.Energy == 99 {
		t.Fatal("perfil atrasado sobrescreveu identidade")
	}
}

func TestInlineScreenAndTypingShortcuts(t *testing.T) {
	m := openInline(t, inlineModel(t, "ui-screen"))
	m.roomDev.field = "name"
	m.in = newInput("", 16)
	m.in.SetValue("ana")
	m, _ = pressInline(m, 'x', false)
	if m.in.Value() != "anax" || m.roomDev.busy {
		t.Fatal("x digitado virou salvar")
	}
	m.roomDev.field = ""
	m.roomDev.locked = true
	screen := ansi.Strip(m.View().Content)
	if !strings.Contains(screen, "dev mode") || !strings.Contains(screen, "ctrl+e") {
		t.Fatalf("controles não visíveis: %s", screen)
	}
}

func TestInlineExactEnergyNumber(t *testing.T) {
	m := openInline(t, inlineModel(t, "ui-number"))
	m, _ = pressInline(m, 'e', true)
	for _, r := range "12.5" {
		m, _ = pressInline(m, r, false)
	}
	if m.roomDev.number != 12.5 {
		t.Fatalf("valor digitado: %v", m.roomDev.number)
	}
	if e := m.previewField().inlineRender().w.Users[m.id].Energy; e != 12.5 {
		t.Fatalf("energia da prévia: %v", e)
	}
	m, _ = pressInline(m, tea.KeyEnter, false)
	if *m.roomDev.changes[0].Energy != 12.5 {
		t.Fatal("confirmar arredondou a energia")
	}
}

func TestInlineCoinsStartFromCurrentBalance(t *testing.T) {
	m := inlineModel(t, "ui-wallet-start")
	m.tab = tabProfile
	m.economy.Balance = 87
	m.economyReady = true
	m = openInline(t, m)
	m, _ = pressInline(m, 'm', true)
	if m.profileDev.number != 87 {
		t.Fatalf("editor começou em %g, esperado o saldo 87", m.profileDev.number)
	}
	m, _ = pressInline(m, 'l', false)
	m, _ = pressInline(m, tea.KeyEnter, false)
	m, _ = pressInline(m, 'm', true)
	if m.profileDev.number != 97 {
		t.Fatalf("reabrir perdeu o rascunho: %g, esperado 97", m.profileDev.number)
	}
}

func TestInlineCoinsSaveDirectlyFromField(t *testing.T) {
	for _, tc := range []struct {
		name       string
		tab        tab
		key        rune
		ctrl, exit bool
	}{
		{"p-x", tabProfile, 'x', false, true},
		{"p-s", tabProfile, 's', true, false},
		{"r-x", tabRoom, 'x', false, true},
		{"r-s", tabRoom, 's', true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := inlineModel(t, "coin-"+tc.name)
			m.tab = tc.tab
			m = openInline(t, m)
			m, _ = pressInline(m, 'm', true)
			m, _ = pressInline(m, 'l', false)
			before, err := store.ReadEconomy(m.id)
			if err != nil || before.Balance != 0 {
				t.Fatalf("edição gravou antes de salvar: %+v %v", before, err)
			}
			m, cmd := pressInline(m, tc.key, tc.ctrl)
			if cmd == nil {
				t.Fatal("atalho não iniciou o salvamento direto do campo")
			}
			next, _ := m.Update(cmd())
			m = next.(model)
			after, err := store.ReadEconomy(m.id)
			if err != nil || after.Balance != 10 {
				t.Fatalf("saldo não foi salvo: %+v %v", after, err)
			}
			if tc.exit && m.inline() != nil {
				t.Fatal("x não saiu do editor")
			}
			if !tc.exit && (m.inline() == nil || m.inline().field != "") {
				t.Fatal("ctrl+s não permaneceu no editor com o campo confirmado")
			}
		})
	}
}

func TestInlineQuitAlsoChecksDraftsInOtherTabs(t *testing.T) {
	m := openInline(t, inlineModel(t, "ui-quit"))
	_, cmd := pressInline(m, 'c', true)
	if cmd == nil {
		t.Fatal("ctrl+c não sai do editor sem alterações")
	}
	e := 20.0
	m.roomDev.change(m.id).Energy = &e
	m.tab = tabVisual
	m, cmd = pressInline(m, 'q', false)
	if cmd != nil || m.err == "" {
		t.Fatal("não avisou sobre rascunho em outra aba")
	}
	_, cmd = pressInline(m, 'q', false)
	if cmd == nil {
		t.Fatal("repetir q não confirmou saída")
	}
}
