package middleware_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	httpinfra "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/tests/fakes"
	"github.com/gofiber/fiber/v2"
)

type result struct {
	status   int
	body     string
	replayed bool
}

func setup(t *testing.T) (*fiber.App, *fakes.Auth, *fakes.IdempotencyRepo, map[string]int) {
	t.Helper()
	auth := fakes.NewAuth(t)
	repo := fakes.NewIdempotencyRepo()
	calls := map[string]int{}
	app := httpinfra.NewFiberApp()
	app.Post("/things", auth.Handler, middleware.Idempotency(repo), func(c *fiber.Ctx) error {
		tenant := middleware.TenantID(c)
		calls[tenant]++
		return c.Status(fiber.StatusCreated).SendString("created-by-" + tenant)
	})
	return app, auth, repo, calls
}

func post(t *testing.T, app *fiber.App, auth *fakes.Auth, tenant, key, body string) result {
	t.Helper()
	req := httptest.NewRequest("POST", "/things", strings.NewReader(body))
	req.Header.Set("Authorization", auth.Bearer(t, tenant))
	req.Header.Set("Idempotency-Key", key)
	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return result{resp.StatusCode, string(raw), resp.Header.Get("X-Idempotent-Replayed") == "true"}
}

// La misma Idempotency-Key en dos tenants no colisiona: cada tenant ejecuta su
// operación y nunca recibe la respuesta cacheada del otro.
func TestIdempotency_KeyIsScopedByTenant(t *testing.T) {
	app, auth, repo, calls := setup(t)

	a := post(t, app, auth, "tenant-a", "same-key", `{"x":1}`)
	b := post(t, app, auth, "tenant-b", "same-key", `{"x":1}`)

	if a.status != 201 || a.body != "created-by-tenant-a" || a.replayed {
		t.Errorf("tenant A: %+v", a)
	}
	if b.status != 201 || b.body != "created-by-tenant-b" || b.replayed {
		t.Errorf("tenant B must not replay tenant A's response: %+v", b)
	}
	if calls["tenant-a"] != 1 || calls["tenant-b"] != 1 {
		t.Errorf("each tenant must execute once, calls=%v", calls)
	}
	for _, k := range []string{"tenant-a|same-key", "tenant-b|same-key"} {
		if _, ok := repo.Records[k]; !ok {
			t.Errorf("expected stored key %q, have %v", k, keys(repo))
		}
	}
	if len(repo.Records) != 2 {
		t.Errorf("expected 2 records, have %v", keys(repo))
	}

	// Un payload distinto en tenant B con la misma clave tampoco choca con A.
	b2 := post(t, app, auth, "tenant-b", "other-key", `{"x":2}`)
	if b2.status != 201 || b2.replayed {
		t.Errorf("tenant B other key: %+v", b2)
	}
}

// Dentro del mismo tenant se mantiene el comportamiento idempotente.
func TestIdempotency_SameTenantReplaysAndRejectsDifferentPayload(t *testing.T) {
	app, auth, _, calls := setup(t)

	first := post(t, app, auth, "tenant-a", "k1", `{"x":1}`)
	again := post(t, app, auth, "tenant-a", "k1", `{"x":1}`)
	if first.status != 201 || again.status != 201 || !again.replayed || again.body != first.body {
		t.Errorf("expected replay, first=%+v again=%+v", first, again)
	}
	if calls["tenant-a"] != 1 {
		t.Errorf("handler must run once, calls=%v", calls)
	}

	diff := post(t, app, auth, "tenant-a", "k1", `{"x":2}`)
	if diff.status != 422 {
		t.Errorf("different payload with same key must be 422, got %+v", diff)
	}
}

func keys(r *fakes.IdempotencyRepo) []string {
	out := make([]string, 0, len(r.Records))
	for k := range r.Records {
		out = append(out, k)
	}
	return out
}
