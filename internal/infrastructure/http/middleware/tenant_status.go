package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
)

// TenantStatusChecker consulta a kali-auth-context si la empresa del token sigue
// activa (GET /api/v1/auth/session con el mismo Bearer del usuario). Así una
// empresa suspendida queda bloqueada también con tokens emitidos antes de la
// suspensión; al reactivarla, los tokens aún vigentes vuelven a funcionar.
//
// El resultado (activa o suspendida) se guarda por empresa durante TTL para no
// consultar a auth en cada petición: el bloqueo se aplica como máximo TTL
// después de suspender. TTL 0 desactiva la caché. Si auth no responde, se falla
// cerrado (503), igual que con el JWKS.
type TenantStatusChecker struct {
	url    string
	ttl    time.Duration
	client *http.Client

	mu    sync.Mutex
	cache map[string]statusEntry
	now   func() time.Time
}

type statusEntry struct {
	suspended bool
	expires   time.Time
}

func NewTenantStatusChecker(sessionURL string, ttl time.Duration) *TenantStatusChecker {
	return &TenantStatusChecker{
		url:    sessionURL,
		ttl:    ttl,
		client: &http.Client{Timeout: 3 * time.Second},
		cache:  map[string]statusEntry{},
		now:    time.Now,
	}
}

// SessionURLFromJWKS deriva la URL de sesión de auth a partir de su JWKS:
// mismo esquema y host, ruta /api/v1/auth/session.
func SessionURLFromJWKS(jwksURL string) string {
	u, err := url.Parse(jwksURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host + "/api/v1/auth/session"
}

var (
	errTenantSuspended  = fiber.NewError(fiber.StatusForbidden, "tenant is suspended")
	errStatusUnverified = fiber.NewError(fiber.StatusServiceUnavailable, "authentication service unavailable")
	errSessionRejected  = fiber.NewError(fiber.StatusUnauthorized, "invalid access token")
)

// Check devuelve nil si la empresa está activa, o el *fiber.Error a responder.
func (s *TenantStatusChecker) Check(ctx context.Context, tenantID, authorization string) error {
	if s == nil {
		return nil
	}
	if e, ok := s.cached(tenantID); ok {
		if e.suspended {
			return errTenantSuspended
		}
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return errStatusUnverified
	}
	req.Header.Set(fiber.HeaderAuthorization, authorization)
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error("tenant status check failed", logger.Fields{"error": err.Error()})
		return errStatusUnverified
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	switch {
	case resp.StatusCode == http.StatusOK:
		s.store(tenantID, false)
		return nil
	case resp.StatusCode == http.StatusForbidden && errorMessage(body) == "tenant is suspended":
		s.store(tenantID, true)
		return errTenantSuspended
	case resp.StatusCode == http.StatusUnauthorized:
		return errSessionRejected
	default:
		logger.Error("unexpected tenant status response", logger.Fields{"status": resp.StatusCode})
		return errStatusUnverified
	}
}

func (s *TenantStatusChecker) cached(tenantID string) (statusEntry, bool) {
	if s.ttl <= 0 {
		return statusEntry{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.cache[tenantID]
	if !ok || s.now().After(e.expires) {
		return statusEntry{}, false
	}
	return e, true
}

func (s *TenantStatusChecker) store(tenantID string, suspended bool) {
	if s.ttl <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[tenantID] = statusEntry{suspended: suspended, expires: s.now().Add(s.ttl)}
}

func errorMessage(body []byte) string {
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)
	return strings.TrimSpace(payload.Error)
}
