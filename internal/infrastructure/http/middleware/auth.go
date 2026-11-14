package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	localsTenantID = "auth.tenant_id"
	localsUserID   = "auth.user_id"
	jwksMinRefresh = 10 * time.Second
)

var errJWKSUnavailable = errors.New("jwks unavailable")

// JWKSKeySet descarga y cachea las claves públicas RS256 publicadas por
// kali-auth-context en /.well-known/jwks.json.
type JWKSKeySet struct {
	url       string
	client    *http.Client
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	lastFetch time.Time
}

func NewJWKSKeySet(url string) *JWKSKeySet {
	return &JWKSKeySet{url: url, client: &http.Client{Timeout: 5 * time.Second}, keys: map[string]*rsa.PublicKey{}}
}

// Key devuelve la clave del kid indicado; si no la conoce, recarga el JWKS
// (como máximo una vez cada jwksMinRefresh) para soportar rotación de claves.
func (k *JWKSKeySet) Key(kid string) (*rsa.PublicKey, error) {
	k.mu.RLock()
	key, ok := k.keys[kid]
	recent := time.Since(k.lastFetch) < jwksMinRefresh
	k.mu.RUnlock()
	if ok {
		return key, nil
	}
	if recent {
		return nil, fmt.Errorf("unknown key id %q", kid)
	}
	if err := k.refresh(); err != nil {
		return nil, err
	}
	k.mu.RLock()
	defer k.mu.RUnlock()
	if key, ok := k.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("unknown key id %q", kid)
}

func (k *JWKSKeySet) refresh() error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if time.Since(k.lastFetch) < jwksMinRefresh && len(k.keys) > 0 {
		return nil
	}
	k.lastFetch = time.Now()

	resp, err := k.client.Get(k.url)
	if err != nil {
		return fmt.Errorf("%w: %v", errJWKSUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", errJWKSUnavailable, resp.StatusCode)
	}

	var doc struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return fmt.Errorf("%w: %v", errJWKSUnavailable, err)
	}

	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, jk := range doc.Keys {
		if jk.Kty != "RSA" {
			continue
		}
		n, errN := base64.RawURLEncoding.DecodeString(jk.N)
		e, errE := base64.RawURLEncoding.DecodeString(jk.E)
		if errN != nil || errE != nil {
			continue
		}
		keys[jk.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	}
	k.keys = keys
	return nil
}

// AuthConfig describe cómo validar los access tokens emitidos por kali-auth-context.
type AuthConfig struct {
	Keys     *JWKSKeySet
	Issuer   string
	Audience string
	// TenantStatus, si se define, exige además que la empresa del token siga
	// activa (bloquea tokens emitidos antes de una suspensión).
	TenantStatus *TenantStatusChecker
}

// Authenticate valida el Bearer token (RS256, emisor, audiencia, vigencia) y
// deja en el contexto el tenant y el usuario del token. El tenant efectivo de
// toda la petición sale exclusivamente de aquí: nunca del body, query o headers.
func Authenticate(cfg AuthConfig) fiber.Handler {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithLeeway(30 * time.Second),
		jwt.WithExpirationRequired(),
	}
	if cfg.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(cfg.Issuer))
	}
	if cfg.Audience != "" {
		opts = append(opts, jwt.WithAudience(cfg.Audience))
	}
	parser := jwt.NewParser(opts...)

	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid Authorization header")
		}

		claims := jwt.MapClaims{}
		_, err := parser.ParseWithClaims(strings.TrimSpace(parts[1]), claims, func(t *jwt.Token) (any, error) {
			kid, _ := t.Header["kid"].(string)
			return cfg.Keys.Key(kid)
		})
		if err != nil {
			if errors.Is(err, errJWKSUnavailable) {
				logger.Error("jwks unavailable", logger.Fields{"error": err.Error()})
				return fiber.NewError(fiber.StatusServiceUnavailable, "authentication service unavailable")
			}
			return fiber.NewError(fiber.StatusUnauthorized, "invalid access token")
		}

		tenantID, _ := claims["tid"].(string)
		userID, _ := claims["sub"].(string)
		if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(userID) == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "token is missing required claims")
		}

		if err := cfg.TenantStatus.Check(c.UserContext(), tenantID, header); err != nil {
			return err
		}

		c.Locals(localsTenantID, tenantID)
		c.Locals(localsUserID, userID)
		return c.Next()
	}
}

// TenantID devuelve el tenant autenticado de la petición (vacío si no pasó por Authenticate).
func TenantID(c *fiber.Ctx) string {
	v, _ := c.Locals(localsTenantID).(string)
	return v
}
