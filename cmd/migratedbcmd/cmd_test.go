package migratedbcmd

import (
	"strings"
	"testing"
)

func TestConfirmed(t *testing.T) {
	for input, want := range map[string]bool{
		"yes\n": true,
		"YES\n": true,
		"no\n":  false,
		"y\n":   false,
		"":      false,
	} {
		if got := confirmed(strings.NewReader(input)); got != want {
			t.Errorf("confirmed(%q) = %t, want %t", input, got, want)
		}
	}
}
