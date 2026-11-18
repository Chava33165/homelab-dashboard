package httpapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidSambaPassword(t *testing.T) {
	for _, tc := range []struct {
		pass string
		ok   bool
	}{
		{"corta", false},
		{"12345678", true},
		{"contraseña con ñ y espacios", true},
		{"con\nsalto de linea", false},
		{"con\ttab dentro", false},
		{string(make([]byte, 129)), false},
	} {
		if got := validSambaPassword(tc.pass) == ""; got != tc.ok {
			t.Errorf("validSambaPassword(%q) ok=%v, quiero %v", tc.pass, got, tc.ok)
		}
	}
}

func TestWriteFileAtomicReemplaza(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pass")
	for _, v := range []string{"primera123\n", "segunda456\n"} {
		if err := writeFileAtomic(path, []byte(v)); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != v {
			t.Fatalf("leido %q (%v), quiero %q", got, err, v)
		}
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("quedaron temporales: %v", entries)
	}
}
