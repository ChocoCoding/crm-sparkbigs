package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

// WebhookHandler expone endpoints para integraciones externas (n8n, Zapier, etc.).
// Autenticación: header X-API-Key (no JWT).
// Reutiliza los mismos servicios de negocio del resto de la API.
type WebhookHandler struct {
	companySvc      ports.CompanyService
	contactSvc      ports.ContactService
	meetingSvc      ports.MeetingService
	subscriptionSvc ports.SubscriptionService
}

func NewWebhookHandler(
	companySvc ports.CompanyService,
	contactSvc ports.ContactService,
	meetingSvc ports.MeetingService,
	subscriptionSvc ports.SubscriptionService,
) *WebhookHandler {
	return &WebhookHandler{
		companySvc:      companySvc,
		contactSvc:      contactSvc,
		meetingSvc:      meetingSvc,
		subscriptionSvc: subscriptionSvc,
	}
}

// RegisterRoutes registra las rutas bajo /webhooks/v1/.
// El middleware de API Key se aplica al grupo completo en main.go.
func (h *WebhookHandler) RegisterRoutes(app *fiber.App, middlewares ...fiber.Handler) {
	wh := app.Group("/webhooks/v1", middlewares...)

	// Empresas — CRUD completo
	wh.Get("/companies", h.ListCompanies)
	wh.Get("/companies/:id", h.GetCompany)
	wh.Post("/companies", h.CreateCompany)
	wh.Put("/companies/:id", h.UpdateCompany)
	wh.Delete("/companies/:id", h.DeleteCompany)

	// Contactos — CRUD completo
	wh.Get("/contacts", h.ListContacts)
	wh.Get("/contacts/:id", h.GetContact)
	wh.Post("/contacts", h.CreateContact)
	wh.Put("/contacts/:id", h.UpdateContact)
	wh.Delete("/contacts/:id", h.DeleteContact)

	// Reuniones — CRUD completo
	wh.Get("/meetings", h.ListMeetings)
	wh.Get("/meetings/:id", h.GetMeeting)
	wh.Post("/meetings", h.CreateMeeting)
	wh.Put("/meetings/:id", h.UpdateMeeting)
	wh.Delete("/meetings/:id", h.DeleteMeeting)

	// Suscripciones — CRUD completo
	wh.Get("/subscriptions", h.ListSubscriptions)
	wh.Get("/subscriptions/:id", h.GetSubscription)
	wh.Post("/subscriptions", h.CreateSubscription)
	wh.Put("/subscriptions/:id", h.UpdateSubscription)
	wh.Delete("/subscriptions/:id", h.DeleteSubscription)
}

// ─── Empresas ────────────────────────────────────────────────────────────────

func (h *WebhookHandler) CreateCompany(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	var body struct {
		Name              string `json:"name"`
		Sector            string `json:"sector"`
		Status            string `json:"status"`
		Website           string `json:"website"`
		Phone             string `json:"phone"`
		Address           string `json:"address"`
		RelationStartDate string `json:"relation_start_date"` // YYYY-MM-DD opcional
	}
	if err := c.BodyParser(&body); err != nil {
		return webhookError(c, 400, "INVALID_BODY", "Cuerpo inválido")
	}
	if body.Name == "" {
		return webhookError(c, 400, "VALIDATION_FAILED", "El campo 'name' es obligatorio")
	}
	if body.Status == "" {
		body.Status = "prospect"
	}

	company := &domain.Company{
		UserID:  userID,
		Name:    body.Name,
		Sector:  body.Sector,
		Status:  body.Status,
		Website: body.Website,
		Phone:   body.Phone,
		Address: body.Address,
	}
	if body.RelationStartDate != "" {
		if t, err := time.Parse("2006-01-02", body.RelationStartDate); err == nil {
			company.RelationStartDate = &t
		}
	}

	if err := h.companySvc.CreateCompany(company); err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": fiber.Map{"company": company}})
}

func (h *WebhookHandler) UpdateCompany(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}

	existing, err := h.companySvc.GetCompany(id, userID)
	if err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Empresa no encontrada")
	}

	var body struct {
		Name    string `json:"name"`
		Sector  string `json:"sector"`
		Status  string `json:"status"`
		Website string `json:"website"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}
	if err := c.BodyParser(&body); err != nil {
		return webhookError(c, 400, "INVALID_BODY", "Cuerpo inválido")
	}

	if body.Name != "" {
		existing.Name = body.Name
	}
	if body.Sector != "" {
		existing.Sector = body.Sector
	}
	if body.Status != "" {
		existing.Status = body.Status
	}
	if body.Website != "" {
		existing.Website = body.Website
	}
	if body.Phone != "" {
		existing.Phone = body.Phone
	}
	if body.Address != "" {
		existing.Address = body.Address
	}

	if err := h.companySvc.UpdateCompany(existing, userID); err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"company": existing}})
}

// ─── Contactos ───────────────────────────────────────────────────────────────

func (h *WebhookHandler) CreateContact(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	var body struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Position  string `json:"position"`
		Status    string `json:"status"`
		CompanyID *uint  `json:"company_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return webhookError(c, 400, "INVALID_BODY", "Cuerpo inválido")
	}
	if body.Name == "" {
		return webhookError(c, 400, "VALIDATION_FAILED", "El campo 'name' es obligatorio")
	}
	if body.Status == "" {
		body.Status = "active"
	}

	contact := &domain.Contact{
		UserID:    userID,
		Name:      body.Name,
		Email:     body.Email,
		Phone:     body.Phone,
		Position:  body.Position,
		Status:    body.Status,
		CompanyID: body.CompanyID,
	}
	if err := h.contactSvc.CreateContact(contact); err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": fiber.Map{"contact": contact}})
}

func (h *WebhookHandler) UpdateContact(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}

	existing, err := h.contactSvc.GetContact(id, userID)
	if err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Contacto no encontrado")
	}

	var body struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Position  string `json:"position"`
		Status    string `json:"status"`
		CompanyID *uint  `json:"company_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return webhookError(c, 400, "INVALID_BODY", "Cuerpo inválido")
	}

	if body.Name != "" {
		existing.Name = body.Name
	}
	if body.Email != "" {
		existing.Email = body.Email
	}
	if body.Phone != "" {
		existing.Phone = body.Phone
	}
	if body.Position != "" {
		existing.Position = body.Position
	}
	if body.Status != "" {
		existing.Status = body.Status
	}
	if body.CompanyID != nil {
		existing.CompanyID = body.CompanyID
	}

	if err := h.contactSvc.UpdateContact(existing, userID); err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"contact": existing}})
}

// ─── Reuniones ───────────────────────────────────────────────────────────────

func (h *WebhookHandler) CreateMeeting(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	var body struct {
		Title       string `json:"title"`
		CompanyID   uint   `json:"company_id"`
		ContactID   *uint  `json:"contact_id"`
		StartAt     string `json:"start_at"` // RFC3339
		DurationMin int    `json:"duration_min"`
		Notes       string `json:"notes"`
		Summary     string `json:"summary"`
		Status      string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return webhookError(c, 400, "INVALID_BODY", "Cuerpo inválido")
	}
	if body.Title == "" || body.CompanyID == 0 || body.StartAt == "" {
		return webhookError(c, 400, "VALIDATION_FAILED", "title, company_id y start_at son obligatorios")
	}

	startAt, err := time.Parse(time.RFC3339, body.StartAt)
	if err != nil {
		return webhookError(c, 400, "INVALID_DATE", "start_at debe ser RFC3339 (ej: 2026-04-01T10:00:00Z)")
	}
	if body.DurationMin == 0 {
		body.DurationMin = 60
	}
	if body.Status == "" {
		body.Status = "scheduled"
	}

	meeting := &domain.Meeting{
		UserID:      userID,
		CompanyID:   body.CompanyID,
		ContactID:   body.ContactID,
		Title:       body.Title,
		StartAt:     startAt,
		DurationMin: body.DurationMin,
		Notes:       body.Notes,
		Summary:     body.Summary,
		Status:      body.Status,
	}
	if err := h.meetingSvc.CreateMeeting(meeting); err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": fiber.Map{"meeting": meeting}})
}

func (h *WebhookHandler) UpdateMeeting(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}

	existing, err := h.meetingSvc.GetMeeting(id, userID)
	if err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Reunión no encontrada")
	}

	var body struct {
		Title       string `json:"title"`
		StartAt     string `json:"start_at"`
		DurationMin int    `json:"duration_min"`
		Notes       string `json:"notes"`
		Summary     string `json:"summary"`
		Status      string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return webhookError(c, 400, "INVALID_BODY", "Cuerpo inválido")
	}
	if body.Title != "" {
		existing.Title = body.Title
	}
	if body.StartAt != "" {
		if t, err := time.Parse(time.RFC3339, body.StartAt); err == nil {
			existing.StartAt = t
		}
	}
	if body.DurationMin > 0 {
		existing.DurationMin = body.DurationMin
	}
	if body.Notes != "" {
		existing.Notes = body.Notes
	}
	if body.Summary != "" {
		existing.Summary = body.Summary
	}
	if body.Status != "" {
		existing.Status = body.Status
	}

	if err := h.meetingSvc.UpdateMeeting(existing, userID); err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"meeting": existing}})
}

// ─── Suscripciones ───────────────────────────────────────────────────────────

func (h *WebhookHandler) CreateSubscription(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	var body struct {
		Name         string  `json:"name"`
		CompanyID    uint    `json:"company_id"`
		PlanType     string  `json:"plan_type"`
		Status       string  `json:"status"`
		Amount       float64 `json:"amount"`
		Currency     string  `json:"currency"`
		BillingCycle string  `json:"billing_cycle"`
		StartDate    string  `json:"start_date"`    // YYYY-MM-DD
		RenewalDate  string  `json:"renewal_date"`  // YYYY-MM-DD opcional
		Notes        string  `json:"notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return webhookError(c, 400, "INVALID_BODY", "Cuerpo inválido")
	}
	if body.Name == "" || body.CompanyID == 0 || body.StartDate == "" {
		return webhookError(c, 400, "VALIDATION_FAILED", "name, company_id y start_date son obligatorios")
	}

	startDate, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		return webhookError(c, 400, "INVALID_DATE", "start_date debe ser YYYY-MM-DD")
	}
	if body.Status == "" {
		body.Status = "active"
	}
	if body.Currency == "" {
		body.Currency = "EUR"
	}
	if body.BillingCycle == "" {
		body.BillingCycle = "monthly"
	}

	sub := &domain.Subscription{
		UserID:       userID,
		CompanyID:    body.CompanyID,
		Name:         body.Name,
		PlanType:     body.PlanType,
		Status:       body.Status,
		Amount:       body.Amount,
		Currency:     body.Currency,
		BillingCycle: body.BillingCycle,
		StartDate:    startDate,
		Notes:        body.Notes,
	}
	if body.RenewalDate != "" {
		if t, err := time.Parse("2006-01-02", body.RenewalDate); err == nil {
			sub.RenewalDate = &t
		}
	}

	if err := h.subscriptionSvc.CreateSubscription(sub); err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": fiber.Map{"subscription": sub}})
}

func (h *WebhookHandler) UpdateSubscription(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}

	existing, err := h.subscriptionSvc.GetSubscription(id, userID)
	if err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Suscripción no encontrada")
	}

	var body struct {
		Name         string  `json:"name"`
		PlanType     string  `json:"plan_type"`
		Status       string  `json:"status"`
		Amount       float64 `json:"amount"`
		Currency     string  `json:"currency"`
		BillingCycle string  `json:"billing_cycle"`
		RenewalDate  string  `json:"renewal_date"`
		Notes        string  `json:"notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return webhookError(c, 400, "INVALID_BODY", "Cuerpo inválido")
	}
	if body.Name != "" {
		existing.Name = body.Name
	}
	if body.PlanType != "" {
		existing.PlanType = body.PlanType
	}
	if body.Status != "" {
		existing.Status = body.Status
	}
	if body.Amount > 0 {
		existing.Amount = body.Amount
	}
	if body.Currency != "" {
		existing.Currency = body.Currency
	}
	if body.BillingCycle != "" {
		existing.BillingCycle = body.BillingCycle
	}
	if body.RenewalDate != "" {
		if t, err := time.Parse("2006-01-02", body.RenewalDate); err == nil {
			existing.RenewalDate = &t
		}
	}
	if body.Notes != "" {
		existing.Notes = body.Notes
	}

	if err := h.subscriptionSvc.UpdateSubscription(existing, userID); err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"subscription": existing}})
}

// ─── List / Get / Delete — Empresas ─────────────────────────────────────────

func (h *WebhookHandler) ListCompanies(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	offset := c.QueryInt("offset", 0)
	limit := c.QueryInt("limit", 20)

	list, total, err := h.companySvc.ListCompanies(userID, offset, limit)
	if err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": list, "total": total, "offset": offset, "limit": limit},
	})
}

func (h *WebhookHandler) GetCompany(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}
	company, err := h.companySvc.GetCompany(id, userID)
	if err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Empresa no encontrada")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"company": company}})
}

func (h *WebhookHandler) DeleteCompany(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}
	if err := h.companySvc.DeleteCompany(id, userID); err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Empresa no encontrada")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"deleted": true}})
}

// ─── List / Get / Delete — Contactos ─────────────────────────────────────────

func (h *WebhookHandler) ListContacts(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	offset := c.QueryInt("offset", 0)
	limit := c.QueryInt("limit", 20)

	list, total, err := h.contactSvc.ListContacts(userID, offset, limit)
	if err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": list, "total": total, "offset": offset, "limit": limit},
	})
}

func (h *WebhookHandler) GetContact(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}
	contact, err := h.contactSvc.GetContact(id, userID)
	if err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Contacto no encontrado")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"contact": contact}})
}

func (h *WebhookHandler) DeleteContact(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}
	if err := h.contactSvc.DeleteContact(id, userID); err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Contacto no encontrado")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"deleted": true}})
}

// ─── List / Get / Delete — Reuniones ─────────────────────────────────────────

func (h *WebhookHandler) ListMeetings(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	offset := c.QueryInt("offset", 0)
	limit := c.QueryInt("limit", 20)

	list, total, err := h.meetingSvc.ListMeetings(userID, offset, limit)
	if err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": list, "total": total, "offset": offset, "limit": limit},
	})
}

func (h *WebhookHandler) GetMeeting(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}
	meeting, err := h.meetingSvc.GetMeeting(id, userID)
	if err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Reunión no encontrada")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"meeting": meeting}})
}

func (h *WebhookHandler) DeleteMeeting(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}
	if err := h.meetingSvc.DeleteMeeting(id, userID); err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Reunión no encontrada")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"deleted": true}})
}

// ─── List / Get / Delete — Suscripciones ─────────────────────────────────────

func (h *WebhookHandler) ListSubscriptions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	offset := c.QueryInt("offset", 0)
	limit := c.QueryInt("limit", 20)

	list, total, err := h.subscriptionSvc.ListSubscriptions(userID, offset, limit)
	if err != nil {
		return webhookError(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": list, "total": total, "offset": offset, "limit": limit},
	})
}

func (h *WebhookHandler) GetSubscription(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}
	sub, err := h.subscriptionSvc.GetSubscription(id, userID)
	if err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Suscripción no encontrada")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"subscription": sub}})
}

func (h *WebhookHandler) DeleteSubscription(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := parseID(c, "id")
	if err != nil {
		return webhookError(c, 400, "INVALID_PARAM", "ID inválido")
	}
	if err := h.subscriptionSvc.DeleteSubscription(id, userID); err != nil {
		return webhookError(c, 404, "NOT_FOUND", "Suscripción no encontrada")
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"deleted": true}})
}

// ─── Helpers internos ────────────────────────────────────────────────────────

func webhookError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   fiber.Map{"code": code, "message": message},
	})
}

func parseID(c *fiber.Ctx, param string) (uint, error) {
	id, err := c.ParamsInt(param)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
