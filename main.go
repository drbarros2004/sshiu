package main

import (
	"flag"
	"io"
	"net"
	"plantao/internal/rooms"
	"plantao/internal/store"
	"plantao/internal/tui"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/colorprofile"
	gossh "golang.org/x/crypto/ssh"
)

func main() {
	flag.BoolVar(&rooms.Demo, "demo", false, "liga a sala demo, com bots, e a tarefa de 1 min")
	port := flag.String("port", "23234", "porta SSH")
	// o padrão só aceita quem está na própria máquina: abrir para a rede é uma decisão de quem sobe
	host := flag.String("host", "localhost", "endereço de escuta (vazio = todas as interfaces)")
	flag.Parse()

	if err := store.Open("plantao.db"); err != nil {
		log.Fatal("sqlite", "err", err)
	}
	if err := rooms.Load(); err != nil {
		log.Fatal("salas", "err", err)
	}
	if rooms.Demo {
		rooms.SeedBots()
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(*host, *port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool { return true }), // a chave é a identidade
		// quem chega sem chave entra só para ler como criar uma (em vez de "Permission denied")
		wish.WithKeyboardInteractiveAuth(func(ssh.Context, gossh.KeyboardInteractiveChallenge) bool { return true }),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(programHandler),
			activeterm.Middleware(),
			requireKey,
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("wish", "err", err)
	}
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal("porta ocupada? talvez o Plantão já esteja rodando", "err", err)
	}
	for _, h := range hosts(*host) {
		log.Info("Plantão no ar: "+sshTo(h, *port), "demo", rooms.Demo)
	}
	log.Fatal(s.Serve(ln))
}

func requireKey(next ssh.Handler) ssh.Handler {
	return func(s ssh.Session) {
		if s.PublicKey() == nil {
			host, port, _ := net.SplitHostPort(s.LocalAddr().String()) // por onde esta sessão chegou
			wish.Fatalln(s, "\nO Plantão usa sua chave SSH como login, e você ainda não tem uma.\n"+
				"crie com:  ssh-keygen -t ed25519   (enter em tudo)\n"+
				"depois entre de novo:  "+sshTo(host, port))
			return
		}
		next(s)
	}
}

// sshTo: o comando de entrada neste servidor. Sem "-p" na porta 22, que o ssh usa sozinho.
func sshTo(host, port string) string {
	if port == "22" {
		return "ssh " + host
	}
	return "ssh -p " + port + " " + host
}

// hosts: os endereços por onde dá para entrar, para a mensagem de subida. Escutando em todas as
// interfaces não existe um endereço só, então valem os IPv4 da máquina na rede: quem sobe o
// servidor precisa saber qual passar para a turma, e uma máquina costuma ter mais de um.
func hosts(listen string) []string {
	if listen != "" && listen != "0.0.0.0" {
		return []string{listen}
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []string{"<ip desta máquina>"}
	}
	var out []string
	for _, a := range addrs {
		n, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		// o IPv4 de cada placa, menos o loopback e o 169.254.x de rede sem DHCP
		if ip := n.IP.To4(); ip != nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
			out = append(out, ip.String())
		}
	}
	if len(out) == 0 {
		return []string{"<ip desta máquina>"}
	}
	return out
}

func programHandler(s ssh.Session) *tea.Program {
	id := gossh.FingerprintSHA256(s.PublicKey())
	rooms.Connect(id)
	go func() { <-s.Context().Done(); rooms.Disconnect(id) }()

	opts := bubbletea.MakeOptions(s)
	if runtime.GOOS == "windows" {
		// Com o servidor no Windows, o Wish não define o perfil de cores: o Bubble Tea vê que a
		// sessão não é TTY e remove todas as sequências de escape (cores e movimentos de cursor).
		pty, _, _ := s.Pty()
		opts = append(opts,
			tea.WithColorProfile(colorprofile.Env(append(s.Environ(), "TERM="+pty.Term))),
			tea.WithOutput(lfWriter{s}))
	}
	return tea.NewProgram(tui.New(id), opts...)
}

// lfWriter: no Windows o renderer do Bubble Tea supõe que "\n" desce mantendo a coluna, mas a
// sessão SSH emulada transforma "\n" em "\r\n" e o cursor volta à coluna 0 (a tela embaralha).
// Trocamos o "\n" solto por IND (ESC D), que desce sem mexer na coluna.
// ponytail: some quando o servidor roda em Linux (o deploy de verdade), aí nem é usado.
type lfWriter struct{ io.Writer }

func (w lfWriter) Write(p []byte) (int, error) {
	b := make([]byte, 0, len(p)+8)
	for i, c := range p {
		if c == '\n' && (i == 0 || p[i-1] != '\r') {
			b = append(b, "\x1bD"...)
		} else {
			b = append(b, c)
		}
	}
	if _, err := w.Writer.Write(b); err != nil {
		return 0, err
	}
	return len(p), nil
}
