package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

type ContactHandler struct {
	service ports.ContactService
}

func NewContactHandler(service ports.ContactService) *ContactHandler {
	return &ContactHandler{service: service}
}

func (h *ContactHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/contacts")
	api.Post("/", h.Create)
	api.Get("/", h.List)
	api.Get("/:id", h.Get)
	api.Put("/:id", h.Update)
	api.Delete("/:id", h.Delete)
}

func (h *ContactHandler) Create(c *fiber.Ctx) error {
	var body struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Company string `json:"company"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}

	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": "El campo 'name' es obligatorio"},
		})
	}

	userID := c.Locals("userID").(uint)
	contact := &domain.Contact{
		UserID:  userID,
		Name:    body.Name,
		Email:   body.Email,
		Phone:   body.Phone,
		Company: body.Company,
	}

	if err := h.service.CreateContact(contact); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"contact": contact},
	})
}

func (h *ContactHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if limit > 100 {
		limit = 100
	}

	contacts, total, err := h.service.ListContacts(userID, offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": contacts, "total": total},
	})
}

func (h *ContactHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	userID := c.Locals("userID").(uint)
	contact, err := h.service.GetContact(uint(id), userID)
	if err != nil {
		status := fiber.StatusNotFound
		code := "NOT_FOUND"
		if err.Error() == "no tienes permiso sobre este contacto" {
			status = fiber.StatusForbidden
			code = "FORBIDDEN"
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"contact": contact},
	})
}

func (h *ContactHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	var body struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Company string `json:"company"`
		Status  string `json:"status"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}

	userID := c.Locals("userID").(uint)
	contact := &domain.Contact{
		ID:      uint(id),
		Name:    body.Name,
		Email:   body.Email,
		Phone:   body.Phone,
		Company: body.Company,
		Status:  body.Status,
	}

	if err := h.service.UpdateContact(contact, userID); err != nil {
		status := fiber.StatusBadRequest
		if err.Error() == "no tienes permiso sobre este contacto" {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "UPDATE_FAILED", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"contact": contact},
	})
}

func (h *ContactHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	userID := c.Locals("userID").(uint)
	if err := h.service.DeleteContact(uint(id), userID); err != nil {
		status := fiber.StatusBadRequest
		if err.Error() == "no tienes permiso sobre este contacto" {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "DELETE_FAILED", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Contacto eliminado correctamente"},
	})
}
