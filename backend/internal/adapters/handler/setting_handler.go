package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

type SettingHandler struct {
	service ports.SettingService
}

func NewSettingHandler(service ports.SettingService) *SettingHandler {
	return &SettingHandler{service: service}
}

func (h *SettingHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/settings")
	api.Post("/seed", h.Seed)
	api.Get("/", h.List)
	api.Get("/:category", h.ListByCategory)
	api.Put("/", h.Upsert)
	api.Delete("/:id", h.Delete)
}

type settingBody struct {
	Category  string `json:"category"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	InputType string `json:"input_type"`
}

func (h *SettingHandler) Seed(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	if err := h.service.SeedDefaults(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "Ajustes por defecto creados"}})
}

func (h *SettingHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	settings, err := h.service.GetSettings(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"list": settings}})
}

func (h *SettingHandler) ListByCategory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	category := c.Params("category")
	settings, err := h.service.GetCategory(userID, category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"list": settings}})
}

func (h *SettingHandler) Upsert(c *fiber.Ctx) error {
	var body settingBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo inválido"},
		})
	}
	if body.Key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "VALIDATION_FAILED", "message": "El campo 'key' es obligatorio"},
		})
	}
	userID := c.Locals("userID").(uint)
	setting := &domain.Setting{
		UserID:    userID,
		Category:  body.Category,
		Key:       body.Key,
		Value:     body.Value,
		Label:     body.Label,
		InputType: body.InputType,
	}
	if setting.InputType == "" {
		setting.InputType = "text"
	}
	if err := h.service.Upsert(setting); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"setting": setting}})
}

func (h *SettingHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}
	userID := c.Locals("userID").(uint)
	if err := h.service.DeleteSetting(uint(id), userID); err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{"success": false, "error": fiber.Map{"code": code, "message": err.Error()}})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "Ajuste eliminado"}})
}
