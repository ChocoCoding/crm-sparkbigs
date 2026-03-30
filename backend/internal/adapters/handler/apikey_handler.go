package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/ports"
)

// APIKeyHandler gestiona la creación y revocación de API Keys.
// Todas sus rutas requieren JWT válido (usuario autenticado).
type APIKeyHandler struct {
	service ports.APIKeyService
}

func NewAPIKeyHandler(service ports.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

func (h *APIKeyHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/apikeys")
	api.Post("/", h.Create)
	api.Get("/", h.List)
	api.Delete("/:id", h.Revoke)
}

// Create genera una nueva API Key para el usuario autenticado.
// ⚠️  La clave en texto plano se devuelve UNA SOLA VEZ — no hay forma de recuperarla.
func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	var body struct {
		Name   string `json:"name"`
		Scopes string `json:"scopes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo inválido"},
		})
	}
	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": "El campo 'name' es obligatorio"},
		})
	}
	if body.Scopes == "" {
		body.Scopes = "webhooks"
	}

	userID := c.Locals("userID").(uint)

	plaintext, key, err := h.service.GenerateKey(userID, body.Name, body.Scopes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			// ⚠️ plaintext_key solo aparece en esta respuesta — guárdala ahora
			"plaintext_key": plaintext,
			"key": fiber.Map{
				"id":         key.ID,
				"name":       key.Name,
				"key_prefix": key.KeyPrefix,
				"scopes":     key.Scopes,
				"is_active":  key.IsActive,
				"created_at": key.CreatedAt,
			},
		},
	})
}

// List devuelve las API Keys del usuario. Nunca expone el hash ni el texto plano.
func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	keys, err := h.service.ListKeys(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"list": keys}})
}

// Revoke desactiva permanentemente una API Key.
func (h *APIKeyHandler) Revoke(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}
	userID := c.Locals("userID").(uint)
	if err := h.service.RevokeKey(uint(id), userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "REVOKE_FAILED", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "API Key revocada"}})
}
