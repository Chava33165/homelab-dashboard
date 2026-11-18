package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKorraStats(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	cases := []struct {
		name string
		file string
		code int
	}{
		{"valido", write("ok.json", `{"version":1}`), http.StatusOK},
		{"faltante", filepath.Join(dir, "nada.json"), http.StatusNotFound},
		{"invalido", write("roto.json", `{"version":`), http.StatusBadGateway},
		{"enorme", write("enorme.json", `"`+strings.Repeat("a", maxKorraStats)+`"`), http.StatusBadGateway},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := &korraHandlers{statsFile: c.file}
			rec := httptest.NewRecorder()
			h.stats(rec, httptest.NewRequest(http.MethodGet, "/api/korra/stats", nil))
			if rec.Code != c.code {
				t.Fatalf("code = %d, quiero %d (%s)", rec.Code, c.code, rec.Body.String())
			}
			if c.code == http.StatusOK && rec.Body.String() != `{"version":1}` {
				t.Fatalf("body = %q", rec.Body.String())
			}
		})
	}
}
