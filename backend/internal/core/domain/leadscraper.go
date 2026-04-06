package domain

import (
	"time"

	"gorm.io/datatypes"
)

// ═══════════════════════════════════════════════════════════
// Lead Scraper — Modelos de dominio
// Pipeline autónomo de generación de leads (Sincron-IA)
// ═══════════════════════════════════════════════════════════

// ScrapeJob representa una ejecución del pipeline de scraping.
type ScrapeJob struct {
	ID                  string     `gorm:"primarykey;size:36" json:"id"` // UUID
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	SearchQuery         string     `gorm:"size:500;not null" json:"search_query"`
	Location            string     `gorm:"size:255;not null" json:"location"`
	MaxResults          int        `gorm:"default:5" json:"max_results"`
	Status              string     `gorm:"size:50;default:pending" json:"status"` // "pending" | "running" | "completed" | "error"
	TotalCompanies      int        `gorm:"default:0" json:"total_companies"`
	ProcessedCompanies  int        `gorm:"default:0" json:"processed_companies"`
	CompletedAt         *time.Time `json:"completed_at"`

	// Relaciones
	Leads []ScrapedLead  `gorm:"foreignKey:JobID" json:"leads,omitempty"`
	Logs  []ScrapeAPILog `gorm:"foreignKey:JobID" json:"logs,omitempty"`
}

// ScrapedLead es un lead generado por el pipeline de scraping.
// Contiene toda la información recopilada de Google Maps + LinkedIn + Web + IA.
type ScrapedLead struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	JobID string `gorm:"size:36;index;not null" json:"job_id"`

	// ─── Datos de empresa (Google Maps) ──────────────────────────
	NombreEmpresa  string  `gorm:"size:500" json:"nombre_empresa"`
	GooglePlaceID  string  `gorm:"size:255;index" json:"google_place_id"`
	Email          string  `gorm:"size:255" json:"email"`
	Telefono       string  `gorm:"size:100" json:"telefono"`
	Website        string  `gorm:"size:500" json:"website"`
	Direccion      string  `gorm:"size:500" json:"direccion"`
	Ciudad         string  `gorm:"size:255" json:"ciudad"`
	Provincia      string  `gorm:"size:255" json:"provincia"`
	Pais           string  `gorm:"size:10;default:ES" json:"pais"`
	CodigoPostal   string  `gorm:"size:20" json:"codigo_postal"`
	Latitud        *float64 `json:"latitud"`
	Longitud       *float64 `json:"longitud"`
	CategoriaGoogle string `gorm:"size:255" json:"categoria_google"`
	RatingGoogle   *float64 `json:"rating_google"`
	NumReviews     int     `gorm:"default:0" json:"num_reviews"`
	TieneWeb       bool    `gorm:"default:false" json:"tiene_web"`

	// ─── Redes sociales ──────────────────────────────────────────
	PresenciaRedesSociales bool   `gorm:"default:false" json:"presencia_redes_sociales"`
	LinkedinURL            string `gorm:"size:500" json:"linkedin_url"`
	InstagramURL           string `gorm:"size:500" json:"instagram_url"`
	FacebookURL            string `gorm:"size:500" json:"facebook_url"`

	// ─── Métricas estimadas ──────────────────────────────────────
	AniosEnMercado int `gorm:"default:1" json:"anios_en_mercado"`

	// ─── Texto raw ───────────────────────────────────────────────
	DescripcionRaw string `gorm:"type:text" json:"descripcion_raw"`

	// ─── Análisis IA (Gemini) ────────────────────────────────────
	NichoEspecifico            string `gorm:"type:text" json:"nicho_especifico"`
	DescripcionNegocio         string `gorm:"type:text" json:"descripcion_negocio"`
	MadurezDigital             string `gorm:"size:50" json:"madurez_digital"`
	NecesidadesRRHH            string `gorm:"type:text" json:"necesidades_rrhh"`
	NecesidadesVentas          string `gorm:"type:text" json:"necesidades_ventas"`
	AnguloTransformacion       string `gorm:"type:text" json:"angulo_transformacion"`
	GranPotencialEstrategico   bool   `gorm:"default:false" json:"gran_potencial_estrategico"`
	ScoreLead                  int    `gorm:"default:0" json:"score_lead"`
	RazonesScore               string `gorm:"type:text" json:"razones_score"`
	Icebreaker                 string `gorm:"type:text" json:"icebreaker"`
	AnalisisCompleto           string `gorm:"type:text" json:"analisis_completo"`

	// ─── Contactos clasificados (JSON array) ─────────────────────
	ContactosClasificados datatypes.JSON `json:"contactos_clasificados"`
	EmailsAdicionales     datatypes.JSON `json:"emails_adicionales"`

	// ─── Meta ────────────────────────────────────────────────────
	Fuente string `gorm:"size:100" json:"fuente"` // "google_maps_web_deep_scrape" | "google_maps_only"
	Estado string `gorm:"size:50;default:nuevo" json:"estado"` // "nuevo" | "contactado" | "descartado" | "convertido"
}

// ScrapeAPILog registra cada llamada a API externa durante el pipeline.
type ScrapeAPILog struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	JobID           string    `gorm:"size:36;index;not null" json:"job_id"`
	CompanyName     string    `gorm:"size:500" json:"company_name"`
	StepName        string    `gorm:"size:100;not null" json:"step_name"`
	StepIndex       int       `gorm:"default:0" json:"step_index"`
	RequestSummary  string    `gorm:"type:text" json:"request_summary"`
	ResponseSummary string    `gorm:"type:text" json:"response_summary"`
	Status          string    `gorm:"size:50;default:pending" json:"status"` // "success" | "error" | "pending"
	DurationMs      int       `gorm:"default:0" json:"duration_ms"`
	Error           string    `gorm:"type:text" json:"error"`
}

// LeadScraperEvent es un evento emitido durante el pipeline para SSE.
type LeadScraperEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}
