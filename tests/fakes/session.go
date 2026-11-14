package fakes

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
)

// Session simula GET /api/v1/auth/session de kali-auth-context: lee "tid" y
// "sub" del Bearer reenviado y responde 200 con los permisos efectivos de ese
// usuario, o 403 {"error":"tenant is suspended"} si la empresa está suspendida.
// Cuenta las llamadas recibidas.
type Session struct {
	mu          sync.Mutex
	suspended   map[string]bool
	permissions map[string][]middleware.Permission // clave tenant + "|" + usuario
	Calls       atomic.Int32
	URL         string
}

func NewSession(t *testing.T) *Session {
	t.Helper()
	s := &Session{suspended: map[string]bool{}, permissions: map[string][]middleware.Permission{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Calls.Add(1)
		tenant, user := ClaimsFromBearer(r.Header.Get("Authorization"))
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.suspended[tenant] {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"tenant is suspended"}`))
			return
		}
		perms := s.permissions[tenant+"|"+user]
		if perms == nil {
			perms = []middleware.Permission{}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tenant_id": tenant, "user_id": user, "status": "ACTIVE", "permissions": perms,
		})
	}))
	t.Cleanup(srv.Close)
	s.URL = srv.URL
	return s
}

// Grant fija los permisos efectivos del usuario ("recurso/acción", p. ej. "invoices/read").
func (s *Session) Grant(tenantID, userID string, perms ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]middleware.Permission, 0, len(perms))
	for _, p := range perms {
		resource, action, _ := strings.Cut(p, "/")
		list = append(list, middleware.Permission{Resource: resource, Action: action})
	}
	s.permissions[tenantID+"|"+userID] = list
}

func (s *Session) Suspend(tenantID string, suspended bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.suspended[tenantID] = suspended
}

// ClaimsFromBearer extrae (tid, sub) del JWT sin verificarlo (solo para fakes).
func ClaimsFromBearer(header string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(header, "Bearer "), ".")
	if len(parts) != 3 {
		return "", ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ""
	}
	var claims struct {
		Tid string `json:"tid"`
		Sub string `json:"sub"`
	}
	_ = json.Unmarshal(raw, &claims)
	return claims.Tid, claims.Sub
}
