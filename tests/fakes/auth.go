package fakes

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Auth levanta un JWKS local y emite tokens RS256 válidos para el middleware
// real Authenticate, de modo que el tenant de la petición sale del token igual
// que en producción.
type Auth struct {
	Handler fiber.Handler
	key     *rsa.PrivateKey
}

const testKID = "test-kid"

func NewAuth(t *testing.T) *Auth {
	t.Helper()
	return NewAuthWithStatus(t, nil)
}

// NewAuthWithStatus igual que NewAuth, pero exigiendo además que la empresa del
// token siga activa (comprobación contra un /auth/session simulado).
func NewAuthWithStatus(t *testing.T, status *middleware.TenantStatusChecker) *Auth {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa key: %v", err)
	}
	jwks := map[string]any{"keys": []map[string]string{{
		"kty": "RSA",
		"kid": testKID,
		"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
	}}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(srv.Close)
	return &Auth{
		Handler: middleware.Authenticate(middleware.AuthConfig{Keys: middleware.NewJWKSKeySet(srv.URL), TenantStatus: status}),
		key:     key,
	}
}

// Bearer devuelve el valor de la cabecera Authorization para el tenant dado
// (usuario "user-<tenant>").
func (a *Auth) Bearer(t *testing.T, tenantID string) string {
	t.Helper()
	return a.BearerFor(t, tenantID, "user-"+tenantID)
}

// BearerFor devuelve la cabecera Authorization para un usuario concreto del tenant.
func (a *Auth) BearerFor(t *testing.T, tenantID, userID string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"tid": tenantID,
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = testKID
	signed, err := tok.SignedString(a.key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return "Bearer " + signed
}
