package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmpleoProxy(t *testing.T) {
	var gotPath, gotXRW, gotCookie, gotBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotXRW, gotCookie = r.URL.Path, r.Header.Get("X-Requested-With"), r.Header.Get("Cookie")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()
	h := newEmpleoHandlers(upstream.URL)

	req := httptest.NewRequest(http.MethodPatch, "/api/empleo/vacantes/55c4de5d27", strings.NewReader(`{"estado":"aplicada"}`))
	req.Header.Set("Cookie", "hd_session=abc")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || gotPath != "/vacantes/55c4de5d27" || gotXRW != "empleo" || gotCookie != "" || gotBody != `{"estado":"aplicada"}` {
		t.Fatalf("code=%d path=%q xrw=%q cookie=%q body=%q", rec.Code, gotPath, gotXRW, gotCookie, gotBody)
	}
	for _, p := range []string{"resumen", "vacantes", "contactos", "perfil", "cvs", "bandeja"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/empleo/"+p, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s = %d", p, rec.Code)
		}
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/empleo/avisar", strings.NewReader(`{"id":"18c2f0a1b2"}`)))
	if rec.Code != http.StatusOK || gotPath != "/avisar" || gotXRW != "empleo" {
		t.Fatalf("POST avisar: code=%d path=%q xrw=%q", rec.Code, gotPath, gotXRW)
	}
	for _, c := range []struct{ method, path string }{
		{http.MethodPost, "/api/empleo/entrevista/responder"},
		{http.MethodDelete, "/api/empleo/perfil/hechos/abc"},
		{http.MethodGet, "/api/empleo/../salud"},
		{http.MethodPatch, "/api/empleo/vacantes/../../x"},
		{http.MethodGet, "/api/empleo/entrevista"},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(c.method, c.path, nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s = %d, quiero 404", c.method, c.path, rec.Code)
		}
	}
}

func TestFinanzasTarjetaNueva(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{}`)) }))
	defer upstream.Close()
	h := newFinanzasHandlers(upstream.URL, "secreto")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/finanzas/tarjetas", strings.NewReader("{}")))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST tarjetas = %d", rec.Code)
	}
}
