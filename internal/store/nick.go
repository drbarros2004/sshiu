package store

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
	"strings"
	"unicode"
)

const NickMax = 16

// ValidateNick é compartilhado pelo primeiro acesso e pelo dev.
func ValidateNick(nick string) error {
	if nick == "" {
		return errors.New("escreva um apelido")
	}
	if strings.ContainsAny(nick, " :") {
		return errors.New("apelido sem espaços, por favor")
	}
	if ansi.StringWidth(nick) > NickMax {
		return fmt.Errorf("apelido de até %d letras, por favor (日 e 🐱 contam como 2)", NickMax)
	}
	state := -1
	for s := nick; s != ""; {
		var g string
		g, s, _, state = uniseg.FirstGraphemeClusterInString(s, state)
		if r := []rune(g); len(r) > 1 || !unicode.IsGraphic(r[0]) || ansi.StringWidth(g) == 0 {
			return errors.New("apelido só com letras, números e símbolos simples, por favor")
		}
	}
	return nil
}
