package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

type DealHandler struct {
	service ports.DealService
}

func NewDealHandler(service ports.DealService) *DealHandler {
	return &DealHandler{service: service}
}

func (h *DealHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/deals")
	api.Post("/", h.Create)
	api.Get("/", h.List)
	api.Get("/:id", h.Get)
	api.Put("/:id", h.Update)
	api.Delete("/:id", h.Delete)
}

func (h *DealHandler) Create(c *fiber.Ctx) error {
	var body struct {
		ContactID uint    `json:"contact_id"`
		Title     string  `json:"title"`
		Value     float64 `json:"value"`
		Currency  string  `json:"currency"`
		Stage     string  `json:"stage"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}

	if body.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": "El campo 'title' es obligatorio"},
		})
	}

	currency := body.Currency
	if currency == "" {
		currency = "EUR"
	}
	stage := body.Stage
	if stage == "" {
		stage = "prospect"
	}

	userID := c.Locals("userID").(uint)
	deal := &domain.Deal{
		UserID:    userID,
		ContactID: body.ContactID,
		Title:     body.Title,
		Value:     body.Value,
		Currency:  currency,
		Stage:     stage,
	}

	if err := h.service.CreateDeal(deal); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"deal": deal},
	})
}

func (h *DealHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if limit > 100 {
		limit = 100
	}

	deals, total, err := h.service.ListDeals(userID, offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": deals, "total": total},
	})
}

func (h *DealHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	userID := c.Locals("userID").(uint)
	deal, err := h.service.GetDeal(uint(id), userID)
	if err != nil {
		status := fiber.StatusNotFound
		code := "NOT_FOUND"
		if err.Error() == "no tienes permiso sobre este deal" {
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
		"data":    fiber.Map{"deal": deal},
	})
}

func (h *DealHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	var body struct {
		Title    string  `json:"title"`
		Value    float64 `json:"value"`
		Currency string  `json:"currency"`
		Stage    string  `json:"stage"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}

	userID := c.Locals("userID").(uint)
	deal := &domain.Deal{
		ID:       uint(id),
		Title:    body.Title,
		Value:    body.Value,
		Currency: body.Currency,
		Stage:    body.Stage,
	}

	if err := h.service.UpdateDeal(deal, userID); err != nil {
		status := fiber.StatusBadRequest
		if err.Error() == "no tienes permiso sobre este deal" {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "UPDATE_FAILED", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"deal": deal},
	})
}

func (h *DealHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	userID := c.Locals("userID").(uint)
	if err := h.service.DeleteDeal(uint(id), userID); err != nil {
		status := fiber.StatusBadRequest
		if err.Error() == "no tienes permiso sobre este deal" {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "DELETE_FAILED", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Deal eliminado correctamente"},
	})
}
