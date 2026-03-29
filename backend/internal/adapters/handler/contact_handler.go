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
	api.Get("/search", h.Search)
	api.Get("/:id", h.Get)
	api.Put("/:id", h.Update)
	api.Delete("/:id", h.Delete)
}

// ─── Payload ─────────────────────────────────────────────────

type contactBody struct {
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
	Position  string  `json:"position"`
	Status    string  `json:"status"`
	CompanyID *uint   `json:"company_id"`
}

// ─── Handlers ────────────────────────────────────────────────

func (h *ContactHandler) Create(c *fiber.Ctx) error {
	var body contactBody
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
	status := body.Status
	if status == "" {
		status = "active"
	}

	contact := &domain.Contact{
		UserID:    userID,
		CompanyID: body.CompanyID,
		Name:      body.Name,
		Email:     body.Email,
		Phone:     body.Phone,
		Position:  body.Position,
		Status:    status,
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

func (h *ContactHandler) Search(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	query := c.Query("q", "")
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	contacts, total, err := h.service.SearchContacts(userID, query, offset, limit)
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
		status, code := errorStatus(err)
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

	var body contactBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}

	userID := c.Locals("userID").(uint)
	contact := &domain.Contact{
		ID:        uint(id),
		CompanyID: body.CompanyID,
		Name:      body.Name,
		Email:     body.Email,
		Phone:     body.Phone,
		Position:  body.Position,
		Status:    body.Status,
	}

	if err := h.service.UpdateContact(contact, userID); err != nil {
		status, code := errorStatus(err)
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
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Contacto eliminado correctamente"},
	})
}
