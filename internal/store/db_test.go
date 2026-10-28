package store

import "testing"

func TestParseSlots(t *testing.T) {
	if got, err := ParseSlots(" 20, 9,14,9"); err != nil || FormatSlots(got) != "9,14,20" {
		t.Fatalf("got %v %v", got, err)
	}
	if _, err := ParseSlots("9,25"); err == nil {
		t.Fatal("25h deveria falhar")
	}
}
