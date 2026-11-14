package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpserver "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func perm(resource, action string) middleware.Permission {
	return middleware.Permission{Resource: resource, Action: action}
}

func TestAllows_MatchingRules(t *testing.T) {
	cases := []struct {
		name     string
		perms    []middleware.Permission
		resource string
		action   string
		want     bool
	}{
		{"exact read", []middleware.Permission{perm("invoices", "read")}, "invoices", "read", true},
		{"exact manage", []middleware.Permission{perm("invoices", "manage")}, "invoices", "manage", true},
		{"manage implies read", []middleware.Permission{perm("invoices", "manage")}, "invoices", "read", true},
		{"read does not imply manage", []middleware.Permission{perm("invoices", "read")}, "invoices", "manage", false},
		{"other resource denied", []middleware.Permission{perm("customers", "manage")}, "invoices", "read", false},
		{"resource prefix is not a match", []middleware.Permission{perm("invoice", "manage")}, "invoice-items", "read", false},
		{"wildcard resource", []middleware.Permission{perm("*", "read")}, "products", "read", true},
		{"wildcard resource keeps action", []middleware.Permission{perm("*", "read")}, "products", "manage", false},
		{"wildcard resource with manage implies read", []middleware.Permission{perm("*", "manage")}, "customers", "read", true},
		{"wildcard action", []middleware.Permission{perm("invoices", "*")}, "invoices", "manage", true},
		{"wildcard action keeps resource", []middleware.Permission{perm("invoices", "*")}, "products", "read", false},
		{"full wildcard", []middleware.Permission{perm("*", "*")}, "invoice-items", "manage", true},
		{"normalizes case and spaces", []middleware.Permission{perm(" Invoices ", "MANAGE")}, "invoices", "read", true},
		{"any of several", []middleware.Permission{perm("products", "read"), perm("invoices", "manage")}, "invoices", "manage", true},
		{"unknown action not implied", []middleware.Permission{perm("invoices", "manage")}, "invoices", "delete", false},
		{"no permissions", nil, "invoices", "read", false},
		{"empty required action", []middleware.Permission{perm("*", "*")}, "invoices", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := middleware.Allows(tc.perms, tc.resource, tc.action); got != tc.want {
				t.Fatalf("Allows(%v, %q, %q) = %v, want %v", tc.perms, tc.resource, tc.action, got, tc.want)
			}
		})
	}
}

// Sin permisos en el contexto (p. ej. sin Authenticate o sin comprobación de
// sesión) RequirePermission falla cerrado con 403 a través del manejador central.
func TestRequirePermission_NoPermissionsInContext_FailsClosed(t *testing.T) {
	app := httpserver.NewFiberApp()
	app.Get("/x", middleware.RequirePermission("invoices", "read"), func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/x", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}
