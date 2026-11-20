package pet

import "testing"

func TestPaletteColorFormats(t *testing.T) {
	for value, want := range map[string]uint32{"#AF5F5F": 0xaf5f5f, "#af5f5f": 0xaf5f5f, "xterm:95": 0x875f5f, "xterm:131": 0xaf5f5f, "xterm:255": 0xeeeeee, "xterm:0": 0x000000} {
		c, err := ParsePaletteColor(value)
		if err != nil || pixelRGB(c) != want {
			t.Errorf("%q: %v, %v; esperado #%06X", value, c, err, want)
		}
	}
	for _, value := range []string{"", "#ABC", "#GGGGGG", "95", "xterm:-1", "xterm:256", "xterm:abc"} {
		if _, err := ParsePaletteColor(value); err == nil {
			t.Errorf("aceitou cor inválida %q", value)
		}
	}
}
