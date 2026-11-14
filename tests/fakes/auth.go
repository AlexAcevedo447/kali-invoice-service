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
		Handler: middleware.Authenticate(middleware.AuthConfig{Keys: middleware.NewJWKSKeySet(srv.URL)}),
		key:     key,
	}
}

// Bearer devuelve el valor de la cabecera Authorization para el tenant dado.
func (a *Auth) Bearer(t *testing.T, tenantID string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"tid": tenantID,
		"sub": "user-" + tenantID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = testKID
	signed, err := tok.SignedString(a.key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return "Bearer " + signed
}
