package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	localsPermissions = "auth.permissions"

	// Wildcard en recurso o acción concede cualquier valor.
	Wildcard = "*"
	// ActionRead y ActionManage son las acciones del contrato de permisos.
	ActionRead   = "read"
	ActionManage = "manage"
)

// Permission es un permiso efectivo del usuario, tal como lo devuelve
// kali-auth-context en GET /api/v1/auth/session.
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

func (p Permission) normalized() Permission {
	return Permission{
		Resource: strings.ToLower(strings.TrimSpace(p.Resource)),
		Action:   strings.ToLower(strings.TrimSpace(p.Action)),
	}
}

var errInsufficientPermissions = fiber.NewError(fiber.StatusForbidden, "insufficient permissions")

// Allows indica si alguno de los permisos concede (resource, action):
//   - el recurso coincide o es "*", y
//   - la acción coincide o es "*", o se pide "read" y el permiso es "manage"
//     (manage implica read; read nunca implica manage).
func Allows(perms []Permission, resource, action string) bool {
	want := Permission{Resource: resource, Action: action}.normalized()
	if want.Resource == "" || want.Action == "" {
		return false
	}
	for _, p := range perms {
		p = p.normalized()
		if p.Resource != want.Resource && p.Resource != Wildcard {
			continue
		}
		if p.Action == want.Action || p.Action == Wildcard ||
			(want.Action == ActionRead && p.Action == ActionManage) {
			return true
		}
	}
	return false
}

// RequirePermission exige que el usuario autenticado tenga (resource, action).
// Debe ir después de Authenticate. Falla cerrado: sin permisos en el contexto
// (p. ej. sin comprobación de sesión configurada) responde
// 403 insufficient permissions.
func RequirePermission(resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !Allows(Permissions(c), resource, action) {
			return errInsufficientPermissions
		}
		return c.Next()
	}
}

// Permissions devuelve los permisos efectivos del usuario de la petición
// (vacío si no pasó por Authenticate o no hay comprobación de sesión).
func Permissions(c *fiber.Ctx) []Permission {
	v, _ := c.Locals(localsPermissions).([]Permission)
	return v
}
