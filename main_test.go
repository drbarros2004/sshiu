package main

import (
	"net"
	"testing"
)

func TestSSHTo(t *testing.T) {
	if got := sshTo("sshiu.exemplo", "22"); got != "ssh sshiu.exemplo" {
		t.Errorf("porta 22 não precisa de -p: %q", got)
	}
	if got := sshTo("192.168.0.33", "23234"); got != "ssh -p 23234 192.168.0.33" {
		t.Errorf("outra porta: %q", got)
	}
}

func TestHosts(t *testing.T) {
	if got := hosts("localhost"); len(got) != 1 || got[0] != "localhost" {
		t.Errorf("endereço escolhido: %v", got)
	}
	open := hosts("") // escuta aberta: os IPs que servem para a turma
	if len(open) == 1 && net.ParseIP(open[0]) == nil {
		t.Skip("máquina sem rede: sobra o aviso no lugar do IP")
	}
	for _, h := range open {
		ip := net.ParseIP(h)
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.To4() == nil {
			t.Errorf("%s não serve para entrar de outra máquina", h)
		}
	}
}
