package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/ports"
)

// publicPaths son rutas que no requieren autenticación.
var publicPaths = map[string]bool{
	"/health":                true,
	"/api/v1/auth/login":     true,
	"/api/v1/auth/refresh":   true,
}

// NewJWTMiddleware retorna un handler de Fiber que valida el Bearer token.
// Inyecta userID y userRole en el contexto para uso en handlers downstream.
func NewJWTMiddleware(authService ports.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if publicPaths[c.Path()] {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "UNAUTHORIZED", "message": "Token requerido"},
			})
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "TOKEN_EXPIRED", "message": "Token inválido o expirado"},
			})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("userRole", claims.Role)

		return c.Next()
	}
}

// RequireAdmin es un middleware que verifica que el usuario tenga rol "admin".
func RequireAdmin(c *fiber.Ctx) error {
	role, ok := c.Locals("userRole").(string)
	if !ok || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "FORBIDDEN", "message": "Acceso restringido a administradores"},
		})
	}
	return c.Next()
}
