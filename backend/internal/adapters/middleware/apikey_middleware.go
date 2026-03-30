package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/ports"
	"github.com/sparkbigs/crm/internal/core/services"
)

// NewAPIKeyMiddleware valida el header X-API-Key para rutas de webhooks.
// Si la clave es válida, inyecta userID en el contexto exactamente igual
// que el middleware JWT, por lo que los handlers son intercambiables.
func NewAPIKeyMiddleware(apiKeySvc ports.APIKeyService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawKey := c.Get("X-API-Key")
		if rawKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "MISSING_API_KEY", "message": "Header X-API-Key requerido"},
			})
		}

		key, err := apiKeySvc.VerifyKey(rawKey)
		if err != nil {
			// Mismo código de error para clave inválida y expirada (no revelar diferencia)
			code := "INVALID_API_KEY"
			if err == services.ErrAPIKeyExpired {
				code = "API_KEY_EXPIRED"
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": code, "message": "API Key inválida, revocada o expirada"},
			})
		}

		// Inyectar userID igual que el middleware JWT para reutilizar handlers
		c.Locals("userID", key.UserID)
		c.Locals("userRole", "webhook") // rol especial, no puede acceder a rutas JWT

		return c.Next()
	}
}
