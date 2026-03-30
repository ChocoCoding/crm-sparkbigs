package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/ports"
)

type DashboardHandler struct {
	service ports.DashboardService
}

func NewDashboardHandler(service ports.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/v1/dashboard", h.Stats)
}

func (h *DashboardHandler) Stats(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	stats, err := h.service.GetStats(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"stats": stats},
	})
}
