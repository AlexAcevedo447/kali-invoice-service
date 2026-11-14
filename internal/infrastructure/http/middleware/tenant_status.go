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

// TenantStatusChecker consulta a kali-auth-context la sesión del usuario
// (GET /api/v1/auth/session con el mismo Bearer): si la empresa sigue activa y
// cuáles son los permisos efectivos del usuario. Así una empresa suspendida queda
// bloqueada también con tokens emitidos antes de la suspensión, y un cambio de
// roles se refleja sin reemitir el token.
//
// Caché (TTL; 0 la desactiva):
//   - Respuesta 200: se guardan los permisos por (tenant, usuario). Nunca se
//     comparten entre usuarios de la misma empresa.
//   - Respuesta 403 "tenant is suspended": se marca la empresa completa como
//     suspendida, de modo que ningún usuario de ella pasa mientras dure la marca.
//
// Un cambio (suspensión o permisos) se aplica como máximo TTL después. Si auth
// no responde o la respuesta no se entiende, se falla cerrado (503).
type TenantStatusChecker struct {
	url    string
	ttl    time.Duration
	client *http.Client

	mu        sync.Mutex
	sessions  map[sessionKey]sessionEntry // por (tenant, usuario)
	suspended map[string]time.Time        // por tenant: vencimiento de la marca
	now       func() time.Time
}

type sessionKey struct {
	tenantID string
	userID   string
}

type sessionEntry struct {
	permissions []Permission
	expires     time.Time
}

// sessionLimit acota el cuerpo leído de /auth/session (la lista de permisos
// puede crecer con los roles; 64 KiB es holgado).
const sessionLimit = 64 << 10

func NewTenantStatusChecker(sessionURL string, ttl time.Duration) *TenantStatusChecker {
	return &TenantStatusChecker{
		url:       sessionURL,
		ttl:       ttl,
		client:    &http.Client{Timeout: 3 * time.Second},
		sessions:  map[sessionKey]sessionEntry{},
		suspended: map[string]time.Time{},
		now:       time.Now,
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

// Check devuelve los permisos efectivos del usuario si su empresa está activa,
// o el *fiber.Error a responder. Con un checker nil devuelve (nil, nil): sin
// permisos, y RequirePermission deniega (falla cerrado).
func (s *TenantStatusChecker) Check(ctx context.Context, tenantID, userID, authorization string) ([]Permission, error) {
	if s == nil {
		return nil, nil
	}
	key := sessionKey{tenantID: tenantID, userID: userID}
	if suspended, perms, ok := s.cached(key); ok {
		if suspended {
			return nil, errTenantSuspended
		}
		return perms, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, errStatusUnverified
	}
	req.Header.Set(fiber.HeaderAuthorization, authorization)
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error("tenant status check failed", logger.Fields{"error": err.Error()})
		return nil, errStatusUnverified
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, sessionLimit))

	switch {
	case resp.StatusCode == http.StatusOK:
		perms, err := parsePermissions(body)
		if err != nil {
			logger.Error("invalid session response", logger.Fields{"error": err.Error()})
			return nil, errStatusUnverified
		}
		s.storeSession(key, perms)
		return perms, nil
	case resp.StatusCode == http.StatusForbidden && errorMessage(body) == "tenant is suspended":
		s.storeSuspended(tenantID)
		return nil, errTenantSuspended
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, errSessionRejected
	default:
		logger.Error("unexpected tenant status response", logger.Fields{"status": resp.StatusCode})
		return nil, errStatusUnverified
	}
}

// cached devuelve (suspendida, permisos, encontrado). La marca de empresa
// suspendida tiene prioridad sobre los permisos guardados del usuario.
func (s *TenantStatusChecker) cached(key sessionKey) (bool, []Permission, bool) {
	if s.ttl <= 0 {
		return false, nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if exp, ok := s.suspended[key.tenantID]; ok {
		if now.After(exp) {
			delete(s.suspended, key.tenantID)
		} else {
			return true, nil, true
		}
	}
	e, ok := s.sessions[key]
	if !ok {
		return false, nil, false
	}
	if now.After(e.expires) {
		delete(s.sessions, key)
		return false, nil, false
	}
	return false, e.permissions, true
}

func (s *TenantStatusChecker) storeSession(key sessionKey, perms []Permission) {
	if s.ttl <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// Auth acaba de confirmar que la empresa está activa.
	delete(s.suspended, key.tenantID)
	s.sessions[key] = sessionEntry{permissions: perms, expires: s.now().Add(s.ttl)}
}

func (s *TenantStatusChecker) storeSuspended(tenantID string) {
	if s.ttl <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.suspended[tenantID] = s.now().Add(s.ttl)
}

// parsePermissions lee {"permissions":[{"resource","action"}]} de la sesión.
// Normaliza (minúsculas, sin espacios) y descarta entradas incompletas. Un
// cuerpo que no es JSON es un error (fallo cerrado).
func parsePermissions(body []byte) ([]Permission, error) {
	var payload struct {
		Permissions []Permission `json:"permissions"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	perms := make([]Permission, 0, len(payload.Permissions))
	for _, p := range payload.Permissions {
		p = p.normalized()
		if p.Resource == "" || p.Action == "" {
			continue
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func errorMessage(body []byte) string {
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)
	return strings.TrimSpace(payload.Error)
}
