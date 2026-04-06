package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

var (
	ErrJobNotFound  = errors.New("job de scraping no encontrado")
	ErrLeadNotFound = errors.New("lead scrapeado no encontrado")
)

type leadScraperService struct {
	jobRepo     ports.ScrapeJobRepository
	leadRepo    ports.ScrapedLeadRepository
	logRepo     ports.ScrapeAPILogRepository
	companyRepo ports.CompanyRepository
	contactRepo ports.ContactRepository

	// SSE: mapa de jobID → lista de canales de suscripción
	mu          sync.RWMutex
	subscribers map[string][]chan domain.LeadScraperEvent
}

func NewLeadScraperService(
	jobRepo ports.ScrapeJobRepository,
	leadRepo ports.ScrapedLeadRepository,
	logRepo ports.ScrapeAPILogRepository,
	companyRepo ports.CompanyRepository,
	contactRepo ports.ContactRepository,
) ports.LeadScraperService {
	return &leadScraperService{
		jobRepo:     jobRepo,
		leadRepo:    leadRepo,
		logRepo:     logRepo,
		companyRepo: companyRepo,
		contactRepo: contactRepo,
		subscribers: make(map[string][]chan domain.LeadScraperEvent),
	}
}

// ═══════════════════════════════════════════════════════════
// Pipeline
// ═══════════════════════════════════════════════════════════

func (s *leadScraperService) StartScrape(query, location string, maxResults int) (string, error) {
	jobID := uuid.New().String()

	job := &domain.ScrapeJob{
		ID:          jobID,
		SearchQuery: query,
		Location:    location,
		MaxResults:  maxResults,
		Status:      "running",
	}
	if err := s.jobRepo.Create(job); err != nil {
		return "", err
	}

	// Ejecutar pipeline en goroutine (background)
	go s.runPipeline(jobID, query, location, maxResults)

	return jobID, nil
}

func (s *leadScraperService) StreamJob(jobID string) (<-chan domain.LeadScraperEvent, func()) {
	ch := make(chan domain.LeadScraperEvent, 100) // Buffer para evitar bloqueos

	s.mu.Lock()
	s.subscribers[jobID] = append(s.subscribers[jobID], ch)
	s.mu.Unlock()

	// Función de cleanup
	unsub := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		subs := s.subscribers[jobID]
		for i, sub := range subs {
			if sub == ch {
				s.subscribers[jobID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}

	return ch, unsub
}

// broadcast envía un evento a todos los suscriptores SSE de un job.
func (s *leadScraperService) broadcast(jobID, event string, data interface{}) {
	s.mu.RLock()
	subs := s.subscribers[jobID]
	s.mu.RUnlock()

	evt := domain.LeadScraperEvent{Event: event, Data: data}
	for _, ch := range subs {
		select {
		case ch <- evt:
		default:
			// Canal lleno — suscriptor lento, saltar
		}
	}

	// Loguear a DB si es api_request/api_response/step_error
	if event == "api_request" || event == "api_response" || event == "step_error" {
		dataMap, ok := data.(map[string]interface{})
		if ok {
			companyName := ""
			if cn, exists := dataMap["companyName"]; exists {
				companyName, _ = cn.(string)
			}
			stepName := event
			if sn, exists := dataMap["step"]; exists {
				stepName, _ = sn.(string)
			}
			durationMs := 0
			if d, exists := dataMap["duration"]; exists {
				switch v := d.(type) {
				case int:
					durationMs = v
				case float64:
					durationMs = int(v)
				}
			}
			status := "success"
			errMsg := ""
			if event == "step_error" {
				status = "error"
				if e, exists := dataMap["error"]; exists {
					errMsg, _ = e.(string)
				}
			}

			reqSummary := ""
			respSummary := ""
			if event == "api_request" {
				b, _ := json.Marshal(data)
				reqSummary = string(b)
			}
			if event == "api_response" {
				b, _ := json.Marshal(data)
				respSummary = string(b)
			}

			apiLog := &domain.ScrapeAPILog{
				JobID:           jobID,
				CompanyName:     companyName,
				StepName:        stepName,
				RequestSummary:  reqSummary,
				ResponseSummary: respSummary,
				Status:          status,
				DurationMs:      durationMs,
				Error:           errMsg,
			}
			_ = s.logRepo.Create(apiLog)
		}
	}
}

// runPipeline ejecuta el pipeline completo (equivalente a runPipeline de pipeline.js).
func (s *leadScraperService) runPipeline(jobID, query, location string, maxResults int) {
	ctx := context.Background()

	emit := func(event string, data interface{}) {
		s.broadcast(jobID, event, data)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[LeadScraper] Pipeline panic para job %s: %v", jobID, r)
			_ = s.jobRepo.UpdateStatus(jobID, "error")
			emit("pipeline_error", map[string]interface{}{"jobId": jobID, "error": fmt.Sprintf("panic: %v", r)})
		}
	}()

	emit("pipeline_start", map[string]interface{}{
		"jobId": jobID, "query": query, "location": location, "maxResults": maxResults,
	})

	// Step 1: Buscar empresas en Apify
	companies, err := searchCompanies(ctx, query, location, maxResults, emit)
	if err != nil || len(companies) == 0 {
		errMsg := "No se encontraron empresas con esa búsqueda."
		if err != nil {
			errMsg = err.Error()
		}
		emit("pipeline_error", map[string]interface{}{"jobId": jobID, "error": errMsg})
		_ = s.jobRepo.UpdateStatus(jobID, "error")
		return
	}

	_ = s.jobRepo.UpdateProgress(jobID, len(companies), 0)

	// Emitir lista de empresas encontradas
	companyList := make([]map[string]string, len(companies))
	for i, c := range companies {
		companyList[i] = map[string]string{
			"name": c.Title, "category": c.CategoryName, "city": c.City,
		}
	}
	emit("companies_found", map[string]interface{}{
		"total": len(companies), "companies": companyList,
	})

	// Procesar cada empresa secuencialmente
	for i, rawCompany := range companies {
		companyName := rawCompany.Title
		if companyName == "" {
			companyName = fmt.Sprintf("Empresa %d", i+1)
		}

		emit("company_start", map[string]interface{}{
			"index": i, "total": len(companies), "name": companyName,
		})

		func() {
			defer func() {
				if r := recover(); r != nil {
					emit("company_error", map[string]interface{}{
						"index": i, "name": companyName, "error": fmt.Sprintf("panic: %v", r),
					})
					_ = s.jobRepo.IncrementProcessed(jobID)
				}
			}()

			// Step 2: Normalizar datos
			emp := normalizeCompany(rawCompany)

			webInfo := ""
			if emp.TieneWeb {
				webInfo = fmt.Sprintf("🌐 Web: %s", emp.Website)
			} else if emp.WebsiteIsSocial {
				webInfo = fmt.Sprintf("⚠️ \"Web\" es red social: %s", emp.WebsiteOriginal)
			} else {
				webInfo = "❌ Sin web"
			}

			normalizeDetails := fmt.Sprintf("%s — %s | %s | ⭐ %v (%d reseñas) | 📧 %d emails | 📱 RRSS: %v",
				emp.NombreEmpresa, emp.CategoriaGoogle, webInfo,
				emp.RatingGoogle, emp.NumReviews,
				len(emp.TodosLosEmails), emp.PresenciaRedesSociales)

			emit("step_done", map[string]interface{}{
				"step": "normalize", "label": "📋 Datos normalizados", "result": normalizeDetails,
			})

			// Step 2.5: Verificar duplicados
			existing, _ := s.leadRepo.FindExisting(emp.GooglePlaceID, emp.NombreEmpresa, emp.Ciudad)
			if existing != nil {
				emit("lead_duplicate", map[string]interface{}{
					"name": emp.NombreEmpresa, "score": existing.ScoreLead,
					"estado": existing.Estado, "razones_score": existing.RazonesScore,
				})
				_ = s.jobRepo.IncrementProcessed(jobID)
				emit("company_done", map[string]interface{}{
					"index": i, "total": len(companies), "name": companyName,
					"score": existing.ScoreLead, "estado": existing.Estado,
				})
				return
			}

			// Step 2.7: Pre-Calificación IA rápida para ahorrar créditos
			preScore := preQualifyWithGemini(ctx, emp, emit)
			if preScore < 25 {
				dbLead := &domain.ScrapedLead{
					JobID:           jobID,
					NombreEmpresa:   emp.NombreEmpresa,
					GooglePlaceID:   emp.GooglePlaceID,
					CategoriaGoogle: emp.CategoriaGoogle,
					Website:         emp.Website,
					Ciudad:          emp.Ciudad,
					Pais:            emp.Pais,
					TieneWeb:        emp.TieneWeb,
					ScoreLead:       preScore,
					RazonesScore:    "Descartado automáticamente por Filtro Básico IA (<25).",
					Estado:          "descartado",
				}
				_ = s.leadRepo.Create(dbLead)
				_ = s.jobRepo.IncrementProcessed(jobID)

				emit("lead_skipped", map[string]interface{}{
					"name": dbLead.NombreEmpresa, "score": dbLead.ScoreLead,
					"reason": "Descartado inicial (Score < 25)",
				})
				emit("company_done", map[string]interface{}{
					"index": i, "total": len(companies), "name": companyName,
					"score": dbLead.ScoreLead, "estado": dbLead.Estado,
				})
				return
			}

			// Step 3: LinkedIn
			linkedin := searchLinkedIn(ctx, emp.NombreEmpresa, emit)
			if linkedin.LinkedinURL != "" && emp.LinkedinURL == "" {
				emp.LinkedinURL = linkedin.LinkedinURL
			}

			// Step 4: Scrape website
			var webContent string
			origenPrompt := "sin_web"
			if emp.TieneWeb {
				webContent = scrapeWebsite(ctx, emp.Website, emit)
				origenPrompt = "con_web"
			} else if emp.WebsiteIsSocial {
				emit("step_done", map[string]interface{}{
					"step": "jina", "label": "🌐 Web Scrape",
					"result": fmt.Sprintf("⏭️ Omitido — \"%s\" es red social, no web real", emp.WebsiteOriginal),
				})
			} else {
				emit("step_done", map[string]interface{}{
					"step": "jina", "label": "🌐 Web Scrape", "result": "Sin website — omitido",
				})
			}

			// Step 5+6: Prompt + Gemini
			var prompt string
			if origenPrompt == "con_web" {
				prompt = buildPromptWithWeb(emp, webContent, linkedin.Snippet)
			} else {
				prompt = buildPromptWithoutWeb(emp, linkedin.Snippet)
			}

			geminiRaw, err := analyzeWithGemini(ctx, prompt, emit)
			if err != nil {
				emit("company_error", map[string]interface{}{
					"index": i, "name": companyName, "error": err.Error(),
				})
				_ = s.jobRepo.IncrementProcessed(jobID)
				return
			}

			// Step 7: Parse response
			lead := parseGeminiResponse(geminiRaw, emp, origenPrompt)

			// Convertir parsedLead a domain.ScrapedLead
			contactosJSON, _ := json.Marshal(lead.ContactosClasificados)
			emailsJSON, _ := json.Marshal(lead.EmailsAdicionales)

			dbLead := &domain.ScrapedLead{
				JobID:                      jobID,
				NombreEmpresa:              lead.NombreEmpresa,
				GooglePlaceID:              lead.GooglePlaceID,
				Email:                      lead.Email,
				Telefono:                   lead.Telefono,
				Website:                    lead.Website,
				Direccion:                  lead.Direccion,
				Ciudad:                     lead.Ciudad,
				Provincia:                  lead.Provincia,
				Pais:                       lead.Pais,
				CodigoPostal:               lead.CodigoPostal,
				Latitud:                    lead.Latitud,
				Longitud:                   lead.Longitud,
				CategoriaGoogle:            lead.CategoriaGoogle,
				RatingGoogle:               lead.RatingGoogle,
				NumReviews:                 lead.NumReviews,
				TieneWeb:                   lead.TieneWeb,
				PresenciaRedesSociales:     lead.PresenciaRedesSociales,
				LinkedinURL:                lead.LinkedinURL,
				InstagramURL:               lead.InstagramURL,
				FacebookURL:                lead.FacebookURL,
				AniosEnMercado:             lead.AniosEnMercado,
				DescripcionRaw:             lead.DescripcionRaw,
				NichoEspecifico:            lead.NichoEspecifico,
				DescripcionNegocio:         lead.DescripcionNegocio,
				MadurezDigital:             lead.MadurezDigital,
				NecesidadesRRHH:            lead.NecesidadesRRHH,
				NecesidadesVentas:          lead.NecesidadesVentas,
				AnguloTransformacion:       lead.AnguloTransformacion,
				GranPotencialEstrategico:   lead.GranPotencialEstrategico,
				ScoreLead:                  lead.ScoreLead,
				RazonesScore:               lead.RazonesScore,
				Icebreaker:                 lead.Icebreaker,
				ContactosClasificados:      contactosJSON,
				EmailsAdicionales:          emailsJSON,
				AnalisisCompleto:           lead.AnalisisCompleto,
				Fuente:                     lead.Fuente,
				Estado:                     lead.Estado,
			}

			// Score filter (como n8n: skip si ≤ 30)
			leadDetail := map[string]interface{}{
				"name":              lead.NombreEmpresa,
				"score":             lead.ScoreLead,
				"icebreaker":        lead.Icebreaker,
				"nicho":             lead.NichoEspecifico,
				"descripcion":       lead.DescripcionNegocio,
				"potencial":         lead.GranPotencialEstrategico,
				"madurez_digital":   lead.MadurezDigital,
				"razones_score":     lead.RazonesScore,
				"angulo":            lead.AnguloTransformacion,
				"necesidades_rrhh":  lead.NecesidadesRRHH,
				"necesidades_ventas": lead.NecesidadesVentas,
				"contactos":         lead.ContactosClasificados,
				"tiene_web":         lead.TieneWeb,
				"email":             lead.Email,
			}

			if lead.ScoreLead <= 30 {
				leadDetail["reason"] = "Score ≤ 30"
				emit("lead_skipped", leadDetail)
				dbLead.Estado = "descartado"
			} else {
				emit("lead_saved", leadDetail)
			}

			// Guardar en MySQL
			_ = s.leadRepo.Create(dbLead)
			_ = s.jobRepo.IncrementProcessed(jobID)

			emit("company_done", map[string]interface{}{
				"index": i, "total": len(companies), "name": companyName,
				"score": lead.ScoreLead, "estado": dbLead.Estado,
			})
		}()
	}

	now := time.Now()
	_ = s.jobRepo.UpdateStatus(jobID, "completed")
	// Update completedAt (handled via UpdateStatus)
	_ = now // referenced for clarity

	emit("pipeline_done", map[string]interface{}{
		"jobId": jobID, "totalProcessed": len(companies),
	})
}

// ═══════════════════════════════════════════════════════════
// CRUD — Jobs
// ═══════════════════════════════════════════════════════════

func (s *leadScraperService) GetJob(jobID string) (*domain.ScrapeJob, error) {
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return nil, ErrJobNotFound
	}
	return job, nil
}

func (s *leadScraperService) ListJobs(offset, limit int) ([]domain.ScrapeJob, int64, error) {
	return s.jobRepo.FindAll(offset, limit)
}

// ═══════════════════════════════════════════════════════════
// CRUD — Leads
// ═══════════════════════════════════════════════════════════

func (s *leadScraperService) GetLead(id uint) (*domain.ScrapedLead, error) {
	lead, err := s.leadRepo.FindByID(id)
	if err != nil {
		return nil, ErrLeadNotFound
	}
	return lead, nil
}

func (s *leadScraperService) GetLeadsByJob(jobID string) ([]domain.ScrapedLead, error) {
	return s.leadRepo.FindByJobID(jobID)
}

func (s *leadScraperService) ListLeads(offset, limit int) ([]domain.ScrapedLead, int64, error) {
	return s.leadRepo.FindAll(offset, limit)
}

func (s *leadScraperService) UpdateLeadEstado(id uint, estado string) error {
	_, err := s.leadRepo.FindByID(id)
	if err != nil {
		return ErrLeadNotFound
	}
	return s.leadRepo.UpdateEstado(id, estado)
}

func (s *leadScraperService) DeleteLead(id uint) error {
	_, err := s.leadRepo.FindByID(id)
	if err != nil {
		return ErrLeadNotFound
	}
	return s.leadRepo.Delete(id)
}

// ═══════════════════════════════════════════════════════════
// API Logs
// ═══════════════════════════════════════════════════════════

func (s *leadScraperService) GetAPILogs(jobID string) ([]domain.ScrapeAPILog, error) {
	return s.logRepo.FindByJobID(jobID)
}

// ═══════════════════════════════════════════════════════════
// Importar Lead al CRM (ScrapedLead → Company + Contact)
// ═══════════════════════════════════════════════════════════

func (s *leadScraperService) ImportToCRM(leadID, userID uint) (*domain.Company, *domain.Contact, error) {
	lead, err := s.leadRepo.FindByID(leadID)
	if err != nil {
		return nil, nil, ErrLeadNotFound
	}

	// Crear Company
	company := &domain.Company{
		UserID:  userID,
		Name:    lead.NombreEmpresa,
		Sector:  lead.CategoriaGoogle,
		Status:  "prospect",
		Website: lead.Website,
		Phone:   lead.Telefono,
		Address: lead.Direccion,
	}
	if err := s.companyRepo.Create(company); err != nil {
		return nil, nil, fmt.Errorf("error creando empresa: %w", err)
	}

	// Crear Contact (si hay email o teléfono)
	var contact *domain.Contact
	if lead.Email != "" || lead.Telefono != "" {
		contact = &domain.Contact{
			UserID:    userID,
			CompanyID: &company.ID,
			Name:      lead.NombreEmpresa,
			Email:     lead.Email,
			Phone:     lead.Telefono,
			Status:    "lead",
		}
		if err := s.contactRepo.Create(contact); err != nil {
			return company, nil, fmt.Errorf("empresa creada pero error creando contacto: %w", err)
		}
	}

	// Marcar lead como convertido
	_ = s.leadRepo.UpdateEstado(leadID, "convertido")

	return company, contact, nil
}
