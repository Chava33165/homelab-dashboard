package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFinanzasProxy(t *testing.T) {
	var gotAuth, gotPath, gotCookie, gotBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath, gotCookie = r.Header.Get("Authorization"), r.URL.RequestURI(), r.Header.Get("Cookie")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		http.SetCookie(w, &http.Cookie{Name: "x", Value: "y"})
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()
	h := newFinanzasHandlers(upstream.URL, "secreto")

	req := httptest.NewRequest(http.MethodPost, "/api/finanzas/gastos?x=1", strings.NewReader(`{"monto":10}`))
	req.Header.Set("Cookie", "hd_session=abc")
	req.Header.Set("Authorization", "Bearer del-navegador")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != `{"ok":true}` {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if gotAuth != "Bearer secreto" || gotPath != "/api/v1/gastos?x=1" || gotCookie != "" || gotBody != `{"monto":10}` {
		t.Fatalf("auth=%q path=%q cookie=%q body=%q", gotAuth, gotPath, gotCookie, gotBody)
	}
	if rec.Header().Get("Set-Cookie") != "" {
		t.Fatal("se colo un Set-Cookie de task-hub")
	}

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/api/finanzas/../../health"},
		{http.MethodDelete, "/api/finanzas/tarjetas/nu"},
		{http.MethodGet, "/api/finanzas/docs"},
		{http.MethodPost, "/api/finanzas/resumen"},
		{http.MethodDelete, "/api/finanzas/cashy/config"},
		{http.MethodPost, "/api/finanzas/cashy/config"},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(c.method, c.path, nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s = %d, quiero 404", c.method, c.path, rec.Code)
		}
	}

	big := httptest.NewRequest(http.MethodPost, "/api/finanzas/gastos", strings.NewReader(strings.Repeat("a", maxFinanzasBody+10)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, big)
	if len(gotBody) > maxFinanzasBody {
		t.Fatalf("paso un body de %d bytes", len(gotBody))
	}
}

func TestFinanzasCaido(t *testing.T) {
	h := newFinanzasHandlers("http://127.0.0.1:1", "secreto")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/finanzas/resumen", nil))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("code = %d, quiero 502", rec.Code)
	}

	sinToken := newFinanzasHandlers("http://127.0.0.1:1", "")
	rec = httptest.NewRecorder()
	sinToken.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/finanzas/resumen", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d, quiero 503", rec.Code)
	}
}

func TestFinanzasCashyPermitido(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer upstream.Close()
	h := newFinanzasHandlers(upstream.URL, "secreto")
	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/api/finanzas/cashy"},
		{http.MethodPost, "/api/finanzas/cashy"},
		{http.MethodDelete, "/api/finanzas/cashy"},
		{http.MethodGet, "/api/finanzas/cashy/config"},
		{http.MethodPut, "/api/finanzas/cashy/config"},
		{http.MethodGet, "/api/finanzas/gastos/periodo"},
		{http.MethodGet, "/api/finanzas/hoy"},
		{http.MethodPost, "/api/finanzas/calibrar"},
		{http.MethodPost, "/api/finanzas/simulador"},
		{http.MethodPost, "/api/finanzas/nominas/recibo"},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(c.method, c.path, strings.NewReader("{}")))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s %s = %d, quiero 200", c.method, c.path, rec.Code)
		}
	}
}

func TestFinanzasReciboPasaCompleto(t *testing.T) {
	var got int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = len(b)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer upstream.Close()
	h := newFinanzasHandlers(upstream.URL, "secreto")
	pdf := strings.Repeat("a", 200<<10) // un PDF de 200 KB: mas que el limite de los formularios
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/finanzas/nominas/recibo", strings.NewReader(pdf)))
	if rec.Code != http.StatusOK || got != len(pdf) {
		t.Fatalf("code=%d llegaron %d de %d bytes", rec.Code, got, len(pdf))
	}
	for _, metodo := range []string{http.MethodGet, http.MethodDelete} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(metodo, "/api/finanzas/nominas/recibo", nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s recibo = %d, quiero 404", metodo, rec.Code)
		}
	}
}
