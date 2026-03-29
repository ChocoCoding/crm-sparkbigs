package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

type SubscriptionHandler struct {
	service ports.SubscriptionService
}

func NewSubscriptionHandler(service ports.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

func (h *SubscriptionHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/subscriptions")
	api.Post("/", h.Create)
	api.Get("/", h.List)
	api.Get("/expiring", h.ExpiringSoon)
	api.Get("/:id", h.Get)
	api.Put("/:id", h.Update)
	api.Delete("/:id", h.Delete)
}

type subscriptionBody struct {
	CompanyID    uint    `json:"company_id"`
	Name         string  `json:"name"`
	PlanType     string  `json:"plan_type"`
	Status       string  `json:"status"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	BillingCycle string  `json:"billing_cycle"`
	StartDate    string  `json:"start_date"`    // "YYYY-MM-DD"
	RenewalDate  *string `json:"renewal_date"`  // nullable
	Notes        string  `json:"notes"`
}

func (b *subscriptionBody) toModel(userID uint) (*domain.Subscription, error) {
	startDate, err := time.Parse("2006-01-02", b.StartDate)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "start_date debe tener formato YYYY-MM-DD")
	}

	sub := &domain.Subscription{
		UserID:       userID,
		CompanyID:    b.CompanyID,
		Name:         b.Name,
		PlanType:     b.PlanType,
		Status:       b.Status,
		Amount:       b.Amount,
		Currency:     b.Currency,
		BillingCycle: b.BillingCycle,
		StartDate:    startDate,
		Notes:        b.Notes,
	}
	if sub.Status == "" {
		sub.Status = "active"
	}
	if sub.Currency == "" {
		sub.Currency = "EUR"
	}
	if sub.BillingCycle == "" {
		sub.BillingCycle = "monthly"
	}

	if b.RenewalDate != nil && *b.RenewalDate != "" {
		t, err := time.Parse("2006-01-02", *b.RenewalDate)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "renewal_date debe tener formato YYYY-MM-DD")
		}
		sub.RenewalDate = &t
	}
	return sub, nil
}

func (h *SubscriptionHandler) Create(c *fiber.Ctx) error {
	var body subscriptionBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo inválido"},
		})
	}
	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "VALIDATION_FAILED", "message": "El campo 'name' es obligatorio"},
		})
	}
	if body.CompanyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "VALIDATION_FAILED", "message": "El campo 'company_id' es obligatorio"},
		})
	}
	userID := c.Locals("userID").(uint)
	sub, err := body.toModel(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "VALIDATION_FAILED", "message": err.Error()},
		})
	}
	if err := h.service.CreateSubscription(sub); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": fiber.Map{"subscription": sub}})
}

func (h *SubscriptionHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	subs, total, err := h.service.ListSubscriptions(userID, offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"list": subs, "total": total}})
}

func (h *SubscriptionHandler) ExpiringSoon(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	days, _ := strconv.Atoi(c.Query("days", "30"))
	subs, err := h.service.ExpiringSoon(userID, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"list": subs}})
}

func (h *SubscriptionHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}
	userID := c.Locals("userID").(uint)
	sub, err := h.service.GetSubscription(uint(id), userID)
	if err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{"success": false, "error": fiber.Map{"code": code, "message": err.Error()}})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"subscription": sub}})
}

func (h *SubscriptionHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}
	var body subscriptionBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo inválido"},
		})
	}
	userID := c.Locals("userID").(uint)
	sub, err := body.toModel(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "VALIDATION_FAILED", "message": err.Error()},
		})
	}
	sub.ID = uint(id)
	if err := h.service.UpdateSubscription(sub, userID); err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{"success": false, "error": fiber.Map{"code": code, "message": err.Error()}})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"subscription": sub}})
}

func (h *SubscriptionHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}
	userID := c.Locals("userID").(uint)
	if err := h.service.DeleteSubscription(uint(id), userID); err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{"success": false, "error": fiber.Map{"code": code, "message": err.Error()}})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "Suscripción eliminada"}})
}
