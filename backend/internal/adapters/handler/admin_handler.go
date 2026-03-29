package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"github.com/sparkbigs/crm/internal/adapters/middleware"
)

type AdminHandler struct {
	service ports.AdminService
}

func NewAdminHandler(service ports.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) RegisterRoutes(app *fiber.App) {
	admin := app.Group("/api/v1/admin", middleware.RequireAdmin)
	admin.Post("/users", h.CreateUser)
	admin.Get("/users", h.ListUsers)
	admin.Get("/users/:id", h.GetUser)
	admin.Put("/users/:id", h.UpdateUser)
	admin.Delete("/users/:id", h.DeactivateUser)
	admin.Put("/users/:id/license", h.SetLicense)
}

func (h *AdminHandler) CreateUser(c *fiber.Ctx) error {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}

	if body.Name == "" || body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": "Nombre, email y contraseña son obligatorios"},
		})
	}

	role := body.Role
	if role == "" {
		role = "user"
	}

	user := &domain.User{Name: body.Name, Email: body.Email, Role: role}

	if err := h.service.CreateUser(user, body.Password); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "CREATE_USER_FAILED", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"user": user},
	})
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if limit > 100 {
		limit = 100
	}

	users, total, err := h.service.ListUsers(offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": users, "total": total},
	})
}

func (h *AdminHandler) GetUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	user, err := h.service.GetUser(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "USER_NOT_FOUND", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"user": user},
	})
}

func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	var body struct {
		Name     string `json:"name"`
		Role     string `json:"role"`
		IsActive *bool  `json:"is_active"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}

	user := &domain.User{ID: uint(id), Name: body.Name, Role: body.Role}
	if body.IsActive != nil {
		user.IsActive = *body.IsActive
	}

	if err := h.service.UpdateUser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "UPDATE_FAILED", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"user": user},
	})
}

func (h *AdminHandler) DeactivateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	if err := h.service.DeactivateUser(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "DEACTIVATE_FAILED", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Usuario desactivado correctamente"},
	})
}

func (h *AdminHandler) SetLicense(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	var body struct {
		Plan string `json:"plan"`
	}

	if err := c.BodyParser(&body); err != nil || body.Plan == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": "El campo 'plan' es obligatorio"},
		})
	}

	if err := h.service.SetLicense(uint(id), body.Plan); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "LICENSE_FAILED", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Licencia actualizada correctamente"},
	})
}
