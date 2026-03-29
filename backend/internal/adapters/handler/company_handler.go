package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"github.com/sparkbigs/crm/internal/core/services"
)

type CompanyHandler struct {
	service ports.CompanyService
}

func NewCompanyHandler(service ports.CompanyService) *CompanyHandler {
	return &CompanyHandler{service: service}
}

func (h *CompanyHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/companies")
	api.Post("/", h.Create)
	api.Get("/", h.List)
	api.Get("/search", h.Search)
	api.Get("/:id", h.Get)
	api.Put("/:id", h.Update)
	api.Delete("/:id", h.Delete)
}

// ─── Payload compartido ──────────────────────────────────────────

type companyBody struct {
	Name              string  `json:"name"`
	Sector            string  `json:"sector"`
	Status            string  `json:"status"`
	Website           string  `json:"website"`
	Phone             string  `json:"phone"`
	Address           string  `json:"address"`
	RelationStartDate *string `json:"relation_start_date"` // ISO 8601: "2024-01-15"
}

func (b *companyBody) toModel(userID uint) (*domain.Company, error) {
	c := &domain.Company{
		UserID:  userID,
		Name:    b.Name,
		Sector:  b.Sector,
		Status:  b.Status,
		Website: b.Website,
		Phone:   b.Phone,
		Address: b.Address,
	}
	if b.RelationStartDate != nil && *b.RelationStartDate != "" {
		t, err := time.Parse("2006-01-02", *b.RelationStartDate)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "relation_start_date debe tener formato YYYY-MM-DD")
		}
		c.RelationStartDate = &t
	}
	if c.Status == "" {
		c.Status = "prospect"
	}
	return c, nil
}

// ─── Handlers ────────────────────────────────────────────────────

func (h *CompanyHandler) Create(c *fiber.Ctx) error {
	var body companyBody
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
	company, err := body.toModel(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": err.Error()},
		})
	}

	if err := h.service.CreateCompany(company); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"company": company},
	})
}

func (h *CompanyHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit > 100 {
		limit = 100
	}

	companies, total, err := h.service.ListCompanies(userID, offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": companies, "total": total},
	})
}

func (h *CompanyHandler) Search(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	query := c.Query("q", "")
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	companies, total, err := h.service.SearchCompanies(userID, query, offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"list": companies, "total": total},
	})
}

func (h *CompanyHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}
	userID := c.Locals("userID").(uint)
	company, err := h.service.GetCompany(uint(id), userID)
	if err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"company": company},
	})
}

func (h *CompanyHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}

	var body companyBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_BODY", "message": "Cuerpo de petición inválido"},
		})
	}

	userID := c.Locals("userID").(uint)
	company, err := body.toModel(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_FAILED", "message": err.Error()},
		})
	}
	company.ID = uint(id)

	if err := h.service.UpdateCompany(company, userID); err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"company": company},
	})
}

func (h *CompanyHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_PARAM", "message": "ID inválido"},
		})
	}
	userID := c.Locals("userID").(uint)
	if err := h.service.DeleteCompany(uint(id), userID); err != nil {
		status, code := errorStatus(err)
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Empresa eliminada correctamente"},
	})
}

// ─── Helper ──────────────────────────────────────────────────────

func errorStatus(err error) (int, string) {
	switch err {
	case services.ErrCompanyNotFound, services.ErrContactNotFound, services.ErrDealNotFound:
		return fiber.StatusNotFound, "NOT_FOUND"
	case services.ErrCompanyForbidden, services.ErrContactForbidden, services.ErrDealForbidden:
		return fiber.StatusForbidden, "FORBIDDEN"
	default:
		return fiber.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
