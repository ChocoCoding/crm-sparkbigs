package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/ports"
	"github.com/sparkbigs/crm/internal/core/services"
)

type LeadScraperHandler struct {
	service ports.LeadScraperService
}

func NewLeadScraperHandler(service ports.LeadScraperService) *LeadScraperHandler {
	return &LeadScraperHandler{service: service}
}

func (h *LeadScraperHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/lead-scraper")
	api.Post("/scrape", h.StartScrape)
	api.Get("/stream/:jobId", h.StreamJob)
	api.Get("/jobs", h.ListJobs)
	api.Get("/jobs/:id", h.GetJob)
	api.Get("/leads", h.ListLeads)
	api.Get("/leads/:id", h.GetLead)
	api.Patch("/leads/:id", h.UpdateLeadEstado)
	api.Delete("/leads/:id", h.DeleteLead)
	api.Get("/logs/:jobId", h.GetAPILogs)
	api.Post("/leads/:id/import", h.ImportToCRM)
}

// ─── Start Scrape ─────────────────────────────────────────────────

func (h *LeadScraperHandler) StartScrape(c *fiber.Ctx) error {
	var body struct {
		Query      string `json:"query"`
		Location   string `json:"location"`
		MaxResults int    `json:"max_results"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}
	if body.Query == "" || body.Location == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": "Se requiere query y location"},
		})
	}
	if body.MaxResults <= 0 {
		body.MaxResults = 5
	}
	if body.MaxResults > 50 {
		body.MaxResults = 50
	}

	jobID, err := h.service.StartScrape(body.Query, body.Location, body.MaxResults)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"job_id": jobID, "message": "Scrape iniciado"},
	})
}

// ─── SSE Stream ───────────────────────────────────────────────────

func (h *LeadScraperHandler) StreamJob(c *fiber.Ctx) error {
	jobID := c.Params("jobId")

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	ch, unsub := h.service.StreamJob(jobID)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer unsub()

		// Evento de conexión
		fmt.Fprintf(w, "event: connected\ndata: %s\n\n",
			mustJSON(map[string]string{"jobId": jobID}))
		w.Flush()

		for evt := range ch {
			data := mustJSON(evt.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Event, data)
			w.Flush()

			// Cerrar stream cuando el pipeline termina
			if evt.Event == "pipeline_done" || evt.Event == "pipeline_error" {
				return
			}
		}
	})

	return nil
}

// ─── Jobs ─────────────────────────────────────────────────────────

func (h *LeadScraperHandler) ListJobs(c *fiber.Ctx) error {
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit > 100 {
		limit = 100
	}

	jobs, total, err := h.service.ListJobs(offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": jobs, "total": total},
	})
}

func (h *LeadScraperHandler) GetJob(c *fiber.Ctx) error {
	jobID := c.Params("id")
	job, err := h.service.GetJob(jobID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "NOT_FOUND", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"job": job},
	})
}

// ─── Leads ────────────────────────────────────────────────────────

func (h *LeadScraperHandler) ListLeads(c *fiber.Ctx) error {
	jobID := c.Query("jobId", "")

	if jobID != "" {
		leads, err := h.service.GetLeadsByJob(jobID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
			})
		}
		return c.JSON(fiber.Map{
			"success": true,
			"data":    fiber.Map{"list": leads, "total": len(leads)},
		})
	}

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit > 100 {
		limit = 100
	}

	leads, total, err := h.service.ListLeads(offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": leads, "total": total},
	})
}

func (h *LeadScraperHandler) GetLead(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	lead, err := h.service.GetLead(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "NOT_FOUND", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"lead": lead},
	})
}

func (h *LeadScraperHandler) UpdateLeadEstado(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	var body struct {
		Estado string `json:"estado"`
	}
	if err := c.BodyParser(&body); err != nil || body.Estado == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": "El campo 'estado' es obligatorio"},
		})
	}

	if err := h.service.UpdateLeadEstado(uint(id), body.Estado); err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Estado actualizado"},
	})
}

func (h *LeadScraperHandler) DeleteLead(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	if err := h.service.DeleteLead(uint(id)); err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Lead eliminado correctamente"},
	})
}

// ─── API Logs ─────────────────────────────────────────────────────

func (h *LeadScraperHandler) GetAPILogs(c *fiber.Ctx) error {
	jobID := c.Params("jobId")
	logs, err := h.service.GetAPILogs(jobID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": logs},
	})
}

// ─── Import to CRM ───────────────────────────────────────────────

func (h *LeadScraperHandler) ImportToCRM(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	userID := c.Locals("userID").(uint)

	company, contact, err := h.service.ImportToCRM(uint(id), userID)
	if err != nil {
		if err == services.ErrLeadNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "NOT_FOUND", "message": err.Error()},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"company": company,
			"contact": contact,
			"message": "Lead importado al CRM correctamente",
		},
	})
}

// ─── Helper ───────────────────────────────────────────────────────

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
