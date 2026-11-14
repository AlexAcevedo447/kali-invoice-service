package middleware_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	httpserver "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/tests/fakes"
	"github.com/gofiber/fiber/v2"
)

// fakeSession simula GET /api/v1/auth/session de kali-auth-context: lee el claim
// "tid" del Bearer reenviado y responde 200 si la empresa está activa o
// 403 {"error":"tenant is suspended"} si no. Cuenta las llamadas.
type fakeSession struct {
	mu        sync.Mutex
	suspended map[string]bool
	status    int // si != 0, responde siempre este código (p. ej. caída o 401)
	calls     atomic.Int32
	srv       *httptest.Server
}

func tenantFromBearer(header string) string {
	parts := strings.Split(strings.TrimPrefix(header, "Bearer "), ".")
	if len(parts) != 3 {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Tid string `json:"tid"`
	}
	_ = json.Unmarshal(raw, &claims)
	return claims.Tid
}

func newFakeSession(t *testing.T) *fakeSession {
	t.Helper()
	f := &fakeSession{suspended: map[string]bool{}}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.calls.Add(1)
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.status != 0 {
			w.WriteHeader(f.status)
			return
		}
		if f.suspended[tenantFromBearer(r.Header.Get("Authorization"))] {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"tenant is suspended"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ACTIVE"}`))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeSession) set(tenant string, suspended bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.suspended[tenant] = suspended
}

func (f *fakeSession) respond(status int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status = status
}

type statusEnv struct {
	app     *fiber.App
	auth    *fakes.Auth
	session *fakeSession
}

func newStatusEnv(t *testing.T, ttl time.Duration) *statusEnv {
	t.Helper()
	env := &statusEnv{session: newFakeSession(t)}
	env.auth = fakes.NewAuthWithStatus(t, middleware.NewTenantStatusChecker(env.session.srv.URL, ttl))
	env.app = httpserver.NewFiberApp() // manejador central de errores real (JSON)
	env.app.Get("/invoices", env.auth.Handler, func(c *fiber.Ctx) error { return c.SendString(middleware.TenantID(c)) })
	return env
}

func (e *statusEnv) get(t *testing.T, bearer string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/invoices", nil)
	req.Header.Set("Authorization", bearer)
	resp, err := e.app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body.Error
}

func TestTenantStatus_TokenIssuedBeforeSuspension_IsBlocked_AndRestoredOnReactivation(t *testing.T) {
	env := newStatusEnv(t, 0)                // sin caché: el efecto es inmediato
	bearer := env.auth.Bearer(t, "tenant-a") // emitido con la empresa activa

	if status, _ := env.get(t, bearer); status != http.StatusOK {
		t.Fatalf("active tenant: expected 200, got %d", status)
	}

	env.session.set("tenant-a", true)
	if status, msg := env.get(t, bearer); status != http.StatusForbidden || msg != "tenant is suspended" {
		t.Fatalf("suspended tenant: expected 403 tenant is suspended, got %d %q", status, msg)
	}

	// Otra empresa no se ve afectada.
	if status, _ := env.get(t, env.auth.Bearer(t, "tenant-b")); status != http.StatusOK {
		t.Fatalf("other tenant: expected 200, got %d", status)
	}

	env.session.set("tenant-a", false)
	if status, _ := env.get(t, bearer); status != http.StatusOK {
		t.Fatalf("reactivated tenant: same token expected 200, got %d", status)
	}
}

func TestTenantStatus_CacheLimitsCallsAndExpires(t *testing.T) {
	env := newStatusEnv(t, 300*time.Millisecond)
	bearer := env.auth.Bearer(t, "tenant-a")

	for i := 0; i < 3; i++ {
		if status, _ := env.get(t, bearer); status != http.StatusOK {
			t.Fatalf("expected 200, got %d", status)
		}
	}
	if calls := env.session.calls.Load(); calls != 1 {
		t.Fatalf("expected 1 call to auth within TTL, got %d", calls)
	}

	// Tras suspender, el bloqueo llega como máximo al vencer el TTL.
	env.session.set("tenant-a", true)
	time.Sleep(350 * time.Millisecond)
	if status, _ := env.get(t, bearer); status != http.StatusForbidden {
		t.Fatalf("after TTL: expected 403, got %d", status)
	}
}

func TestTenantStatus_AuthUnavailable_FailsClosed(t *testing.T) {
	env := newStatusEnv(t, 0)
	env.session.respond(http.StatusBadGateway)
	if status, msg := env.get(t, env.auth.Bearer(t, "tenant-a")); status != http.StatusServiceUnavailable || msg != "authentication service unavailable" {
		t.Fatalf("expected 503 authentication service unavailable, got %d %q", status, msg)
	}
}

func TestTenantStatus_AuthRejectsSession_Unauthorized(t *testing.T) {
	env := newStatusEnv(t, 0)
	env.session.respond(http.StatusUnauthorized)
	if status, msg := env.get(t, env.auth.Bearer(t, "tenant-a")); status != http.StatusUnauthorized || msg != "invalid access token" {
		t.Fatalf("expected 401 invalid access token, got %d %q", status, msg)
	}
}

func TestSessionURLFromJWKS(t *testing.T) {
	got := middleware.SessionURLFromJWKS("http://host.docker.internal:18080/.well-known/jwks.json")
	if got != "http://host.docker.internal:18080/api/v1/auth/session" {
		t.Fatalf("unexpected session URL %q", got)
	}
}
