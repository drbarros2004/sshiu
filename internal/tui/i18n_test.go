package tui

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sshiu/internal/pet"
	"sshiu/internal/rooms"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// translatedLiterals: os textos escritos direto em m.t("..."), l.t("..."), m.tf e l.tf no código
// do pacote, com o arquivo e a linha.
func translatedLiterals(t *testing.T) map[string]string {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	found := map[string]string{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "t" && sel.Sel.Name != "tf" {
				return true
			}
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				s, _ := strconv.Unquote(lit.Value)
				found[s] = fset.Position(lit.Pos()).String()
			}
			return true
		})
	}
	return found
}

// dynamicTexts: os textos que chegam ao m.t por variável (nomes de tema, de humor, de obra...).
func dynamicTexts() []string {
	var s []string
	for _, x := range tabNames {
		s = append(s, x)
	}
	for _, x := range themes {
		s = append(s, x.name)
	}
	for _, x := range wallpapers {
		s = append(s, x.name)
	}
	s = append(s, clockModes...)
	for _, x := range pet.MoodName {
		s = append(s, x)
	}
	for _, x := range pet.Colors {
		s = append(s, x.Name)
	}
	for _, x := range rooms.Obras[1:] {
		s = append(s, x)
	}
	for _, x := range resultLabel {
		s = append(s, x)
	}
	s = append(s, weekdays[:]...)
	s = append(s, periods...) // os períodos da tabela de foco
	s = append(s, months[:]...)
	for _, x := range bandLabels { // as faixas do tempo declarado nas stats
		s = append(s, x[0], x[1])
	}
	s = append(s, statsLayouts[false], statsLayouts[true])
	s = append(s, "começou", "cumpriu", "entrou", "saiu") // os verbos dos eventos, em rooms
	// os erros fixos de rooms, que chegam pelo errText (os com nome ou número casam com enErrors)
	s = append(s, "só o dono muda a sala", "a sala precisa de um nome", "para sair da sala, use sair",
		"essa pessoa não está na sala", "a sala solo não tem obra",
		"a sala demo não tem saída: ela some quando o servidor desliga")
	return append(s, gardenEditorSizeHint)
}

// Todo texto da interface tem inglês, e o inglês começa com minúscula como o português (#86).
func TestEveryTextTranslated(t *testing.T) {
	found := translatedLiterals(t)
	for _, s := range dynamicTexts() {
		found[s] = "(dinâmico)"
	}
	for pt, where := range found {
		e, ok := en[pt]
		if !ok {
			t.Errorf("%s: sem inglês para %q", where, pt)
			continue
		}
		if strings.Count(e, "%") != strings.Count(pt, "%") {
			t.Errorf("%q: o inglês %q tem outra quantidade de %%", pt, e)
		}
		if r := []rune(strings.TrimLeft(e, " ·(←↑/→↓‹")); len(r) > 0 && unicode.IsUpper(r[0]) &&
			!strings.HasPrefix(string(r), "SSH") {
			t.Errorf("%q começa com maiúscula", e)
		}
	}
	for pt := range en { // tradução que ninguém usa mais: a chave mudou em português
		if _, ok := found[pt]; !ok {
			t.Errorf("tradução sem uso: %q", pt)
		}
	}
}

func TestLang(t *testing.T) {
	var pt, vazio, ingles lang = "pt", "", "en"
	if pt.t("sala") != "sala" || vazio.t("sala") != "sala" || ingles.t("sala") != "room" {
		t.Errorf("sala: %q %q %q", pt.t("sala"), vazio.t("sala"), ingles.t("sala"))
	}
	if got := ingles.tf("%d de %d min focados", 3, 25); got != "3 of 25 min focused" {
		t.Errorf("tf: %q", got)
	}
	for in, want := range map[string]string{
		`nenhuma sala com o código "abc123"`: "no room with the code abc123",
		"turma já está cheia (10 membros)":   "turma is already full (10 members)",
		"a sala precisa de um nome":          "the room needs a name",
	} {
		if got := ingles.errText(errors.New(in)); got != want {
			t.Errorf("errText(%q) = %q, queria %q", in, got, want)
		}
	}
}
