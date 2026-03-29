package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

type MeetingHandler struct {
	service ports.MeetingService
}

func NewMeetingHandler(service ports.MeetingService) *MeetingHandler {
	return &MeetingHandler{service: service}
}

func (h *MeetingHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/meetings")
	api.Post("/", h.Create)
	api.Get("/", h.List)
	api.Get("/upcoming", h.Upcoming)
	api.Get("/:id", h.Get)
	api.Put("/:id", h.Update)
	api.Delete("/:id", h.Delete)
}

// ─── Payload ─────────────────────────────────────────────────

type meetingBody struct {
	CompanyID   uint    `json:"company_id"`
	ContactID   *uint   `json:"contact_id"`
	Title       string  `json:"title"`
	StartAt     string  `json:"start_at"` // ISO 8601: "2024-06-15T10:00:00Z"
	DurationMin int     `json:"duration_min"`
	Status      string  `json:"status"`
	Notes       string  `json:"notes"`
}

func (b *meetingBody) toModel(userID uint) (*domain.Meeting, error) {
	startAt, err := time.Parse(time.RFC3339, b.StartAt)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "start_at debe tener formato ISO 8601 (ej: 2024-06-15T10:00:00Z)")
	}

	duration := b.DurationMin
	if duration <= 0 {
		duration = 60
	}

	status := b.Status
	if status == "" {
		status = "scheduled"
	}

	return &domain.Meeting{
		UserID:      userID,
		CompanyID:   b.CompanyID,
		ContactID:   b.ContactID,
		Title:       b.Title,
		StartAt:     startAt,
		DurationMin: duration,
		Status:      status,
		Notes:       b.Notes,
	}, nil
}

// ─── Handlers ────────────────────────────────────────────────

func (h *MeetingHandler) Create(c *fiber.Ctx) error {
	var body meetingBody
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
	if body.CompanyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": "El campo 'company_id' es obligatorio"},
		})
	}

	userID := c.Locals("userID").(uint)
	meeting, err := body.toModel(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": err.Error()},
		})
	}

	if err := h.service.CreateMeeting(meeting); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"meeting": meeting},
	})
}

func (h *MeetingHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit > 100 {
		limit = 100
	}

	meetings, total, err := h.service.ListMeetings(userID, offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": meetings, "total": total},
	})
}

func (h *MeetingHandler) Upcoming(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	meetings, err := h.service.UpcomingMeetings(userID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": meetings},
	})
}

func (h *MeetingHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}
	userID := c.Locals("userID").(uint)
	meeting, err := h.service.GetMeeting(uint(id), userID)
	if err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"meeting": meeting},
	})
}

func (h *MeetingHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	var body meetingBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}

	userID := c.Locals("userID").(uint)
	meeting, err := body.toModel(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": err.Error()},
		})
	}
	meeting.ID = uint(id)

	if err := h.service.UpdateMeeting(meeting, userID); err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"meeting": meeting},
	})
}

func (h *MeetingHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}
	userID := c.Locals("userID").(uint)
	if err := h.service.DeleteMeeting(uint(id), userID); err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Reunión eliminada correctamente"},
	})
}
