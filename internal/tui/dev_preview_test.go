package tui

import (
	"strings"
	"testing"
)

// A prévia deve produzir a própria interface normal, não apenas apagar o texto "dev".
func TestCleanDevPreviewMatchesNormalScreens(t *testing.T) {
	m := inlineModel(t, "clean-screens")
	for _, current := range []tab{tabRoom, tabProfile, tabVisual, tabAbout} {
		m.tab = current
		clean, cmd := pressInline(m, 'h', false)
		if !clean.hideDev || cmd != nil || !Dev {
			t.Fatalf("aba %d: H deve ocultar só a interface desta conexão", current)
		}
		got := clean.View().Content
		Dev = false
		want := m.View().Content
		Dev = true
		if got != want {
			t.Fatalf("aba %d: a prévia difere da tela normal", current)
		}
		if m.hideDev {
			t.Fatal("ocultar uma conexão alterou outro modelo")
		}
	}
}

func TestCleanDevPreviewNavigationAndReturn(t *testing.T) {
	m := inlineModel(t, "clean-navigation")
	m.tab = tabAbout
	m, _ = pressInline(m, 'h', false)
	m, _ = pressInline(m, 'd', false)
	if m.tab != tabRoom {
		t.Fatal("a navegação entrou na aba dev oculta")
	}
	m, _ = pressInline(m, 'a', false)
	if m.tab != tabAbout {
		t.Fatal("a navegação reversa entrou na aba dev oculta")
	}
	m.tab = tabRoom
	m, cmd := pressInline(m, 'x', false)
	if m.hideDev || m.inline() != nil || cmd != nil {
		t.Fatal("X deveria apenas restaurar os controles, sem abrir o editor")
	}
	view := strings.ReplaceAll(m.View().Content, "\u00a0", " ")
	if !strings.Contains(view, m.t(" · x modo dev")) || !strings.Contains(view, m.t(" · h oculta opções dev")) {
		t.Fatal("os atalhos não voltaram após X")
	}
	m, cmd = pressInline(m, 'x', false)
	if m.roomDev == nil || cmd == nil {
		t.Fatal("o próximo X deveria abrir o editor normalmente")
	}
	m.roomDev = nil
	m.tab = tabDev
	m, _ = pressInline(m, 'h', false)
	if !m.hideDev || m.tab != tabRoom {
		t.Fatal("ocultar na aba dev deve levar à sala")
	}
}

func TestCleanDevPreviewPreservesEditingAndNormalMode(t *testing.T) {
	m := inlineModel(t, "clean-editing")
	m = openInline(t, m)
	draft := m.roomDev
	m.tab = tabProfile
	m, _ = pressInline(m, 'h', false)
	if m.hideDev || m.roomDev != draft {
		t.Fatal("H não pode esconder nem descartar um editor dev aberto em outra aba")
	}
	m.roomDev = nil
	m.tab = tabRoom
	m.entering = true
	m.in = newInput("", 80)
	m, _ = pressInline(m, 'h', false)
	if m.hideDev || m.in.Value() != "h" {
		t.Fatal("H digitado numa tarefa deve continuar sendo texto")
	}
	m.hideDev = true
	m, _ = pressInline(m, 'x', false)
	if !m.hideDev || m.in.Value() != "hx" {
		t.Fatal("X digitado numa tarefa não deve sair da prévia")
	}
	m.hideDev, m.entering = false, false
	Dev = false
	m, _ = pressInline(m, 'h', false)
	if m.hideDev || m.tabs() != tabDev {
		t.Fatal("H não deve ativar uma função dev no modo normal")
	}
}
