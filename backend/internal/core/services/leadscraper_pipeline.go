package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════════════
// Lead Scraper Pipeline — Traducción directa de pipeline.js
// ═══════════════════════════════════════════════════════════

// pipelineEmitter es una función callback para emitir eventos SSE.
type pipelineEmitter func(event string, data interface{})

// timedFetch hace un HTTP request con medición de duración.
func timedFetch(ctx context.Context, method, urlStr string, body io.Reader, headers map[string]string) ([]byte, int, int, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, 0, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	duration := int(time.Since(start).Milliseconds())
	return data, resp.StatusCode, duration, err
}

// ─── Step 1: Search companies via Apify ───────────────────────────

type apifyRequest struct {
	SearchStringsArray           []string `json:"searchStringsArray"`
	LocationQuery                string   `json:"locationQuery"`
	MaxCrawledPlacesPerSearch    int      `json:"maxCrawledPlacesPerSearch"`
	Language                     string   `json:"language"`
	SearchMatching               string   `json:"searchMatching"`
	Website                      string   `json:"website"`
	SkipClosedPlaces             bool     `json:"skipClosedPlaces"`
	ScrapePlaceDetailPage        bool     `json:"scrapePlaceDetailPage"`
	ScrapeContacts               bool     `json:"scrapeContacts"`
	MaxReviews                   int      `json:"maxReviews"`
	MaxImages                    int      `json:"maxImages"`
	IncludeWebResults            bool     `json:"includeWebResults"`
}

type apifyResult struct {
	Title        string   `json:"title"`
	PlaceID      string   `json:"placeId"`
	CategoryName string   `json:"categoryName"`
	Address      string   `json:"address"`
	City         string   `json:"city"`
	CountryCode  string   `json:"countryCode"`
	PostalCode   string   `json:"postalCode"`
	Lat          *float64 `json:"lat"`
	Lon          *float64 `json:"lon"`
	Phone        string   `json:"phone"`
	Website      string   `json:"website"`
	TotalScore   *float64 `json:"totalScore"`
	ReviewsCount int      `json:"reviewsCount"`
	Description  string   `json:"description"`
	Emails       []string `json:"emails"`
	LinkedIns    []string `json:"linkedIns"`
	Instagrams   []string `json:"instagrams"`
	Facebooks    []string `json:"facebooks"`
}

func searchCompanies(ctx context.Context, query, location string, maxResults int, emit pipelineEmitter) ([]apifyResult, error) {
	emit("step_start", map[string]interface{}{
		"step": "apify", "label": "📍 Buscando empresas en Google Maps",
		"detail": fmt.Sprintf("\"%s\" en %s (máx. %d)", query, location, maxResults),
	})

	apifyToken := os.Getenv("APIFY_TOKEN")
	apiURL := fmt.Sprintf("https://api.apify.com/v2/acts/compass~crawler-google-places/run-sync-get-dataset-items?token=%s&format=json", apifyToken)

	reqBody := apifyRequest{
		SearchStringsArray:        []string{query},
		LocationQuery:             location,
		MaxCrawledPlacesPerSearch: maxResults,
		Language:                  "es",
		SearchMatching:            "all",
		Website:                   "allPlaces",
		SkipClosedPlaces:          false,
		ScrapePlaceDetailPage:     false,
		ScrapeContacts:            true,
		MaxReviews:                0,
		MaxImages:                 0,
		IncludeWebResults:         false,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	emit("api_request", map[string]interface{}{
		"step": "apify", "method": "POST",
		"url": strings.Replace(apiURL, apifyToken, "***", 1),
	})

	data, status, duration, err := timedFetch(ctx, "POST", apiURL, strings.NewReader(string(bodyBytes)),
		map[string]string{"Content-Type": "application/json"})
	if err != nil {
		emit("step_error", map[string]interface{}{"step": "apify", "error": err.Error()})
		return nil, err
	}

	var results []apifyResult
	if err := json.Unmarshal(data, &results); err != nil {
		// Apify puede devolver un objeto en vez de array en caso de error
		emit("step_error", map[string]interface{}{"step": "apify", "error": "respuesta inesperada de Apify"})
		return nil, fmt.Errorf("error parsing Apify response: %w", err)
	}

	emit("api_response", map[string]interface{}{
		"step": "apify", "status": status, "duration": duration,
		"summary": fmt.Sprintf("%d empresas encontradas", len(results)),
	})
	emit("step_done", map[string]interface{}{
		"step": "apify", "label": "📍 Google Maps",
		"result": fmt.Sprintf("%d empresas", len(results)),
	})

	return results, nil
}

// ─── Social media detection ──────────────────────────────────────

var socialMediaDomains = []string{
	"instagram.com", "facebook.com", "fb.com", "twitter.com", "x.com",
	"tiktok.com", "youtube.com", "youtu.be", "linkedin.com",
	"pinterest.com", "threads.net", "wa.me", "whatsapp.com",
}

func isSocialMediaURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	hostname := strings.TrimPrefix(parsed.Hostname(), "www.")
	for _, domain := range socialMediaDomains {
		if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
			return true
		}
	}
	return false
}

// ─── Step 2: Normalize company data ──────────────────────────────

type normalizedCompany struct {
	NombreEmpresa          string   `json:"nombre_empresa"`
	GooglePlaceID          string   `json:"google_place_id"`
	CategoriaGoogle        string   `json:"categoria_google"`
	Direccion              string   `json:"direccion"`
	Ciudad                 string   `json:"ciudad"`
	Provincia              string   `json:"provincia"`
	Pais                   string   `json:"pais"`
	CodigoPostal           string   `json:"codigo_postal"`
	Latitud                *float64 `json:"latitud"`
	Longitud               *float64 `json:"longitud"`
	Telefono               string   `json:"telefono"`
	Website                string   `json:"website"`
	WebsiteOriginal        string   `json:"website_original"`
	WebsiteIsSocial        bool     `json:"_website_is_social"`
	Email                  string   `json:"email"`
	EmailsAdicionales      []string `json:"emails_adicionales"`
	TodosLosEmails         []string `json:"todos_los_emails"`
	LinkedinURL            string   `json:"linkedin_url"`
	InstagramURL           string   `json:"instagram_url"`
	FacebookURL            string   `json:"facebook_url"`
	PresenciaRedesSociales bool     `json:"presencia_redes_sociales"`
	TieneWeb               bool     `json:"tiene_web"`
	RatingGoogle           *float64 `json:"rating_google"`
	NumReviews             int      `json:"num_reviews"`
	AniosEnMercado         int      `json:"anios_en_mercado"`
	DescripcionRaw         string   `json:"descripcion_raw"`
}

func normalizeCompany(item apifyResult) normalizedCompany {
	// Filtrar emails válidos
	emails := make([]string, 0)
	for _, e := range item.Emails {
		if e != "" && strings.Contains(e, "@") {
			emails = append(emails, e)
		}
	}

	// Redes sociales
	linkedin := ""
	if len(item.LinkedIns) > 0 {
		linkedin = item.LinkedIns[0]
	}
	instagram := ""
	if len(item.Instagrams) > 0 {
		instagram = item.Instagrams[0]
	}
	facebook := ""
	if len(item.Facebooks) > 0 {
		facebook = item.Facebooks[0]
	}

	// Estimar años en mercado por reviews
	reviews := item.ReviewsCount
	anios := 1
	switch {
	case reviews >= 500:
		anios = 10
	case reviews >= 200:
		anios = 7
	case reviews >= 50:
		anios = 4
	case reviews >= 10:
		anios = 2
	}

	// Filtrar si la "web" es red social
	rawWebsite := item.Website
	isSocial := isSocialMediaURL(rawWebsite)
	realWebsite := rawWebsite
	if isSocial {
		realWebsite = ""
	}

	email := ""
	emailsAdicionales := make([]string, 0)
	if len(emails) > 0 {
		email = emails[0]
		emailsAdicionales = emails[1:]
	}

	// Instagram/Facebook fallback si la web es social
	if isSocial && strings.Contains(rawWebsite, "instagram") && instagram == "" {
		instagram = rawWebsite
	}
	if isSocial && strings.Contains(rawWebsite, "facebook") && facebook == "" {
		facebook = rawWebsite
	}

	pais := item.CountryCode
	if pais == "" {
		pais = "ES"
	}

	return normalizedCompany{
		NombreEmpresa:          item.Title,
		GooglePlaceID:          item.PlaceID,
		CategoriaGoogle:        item.CategoryName,
		Direccion:              item.Address,
		Ciudad:                 item.City,
		Provincia:              item.City,
		Pais:                   pais,
		CodigoPostal:           item.PostalCode,
		Latitud:                item.Lat,
		Longitud:               item.Lon,
		Telefono:               item.Phone,
		Website:                realWebsite,
		WebsiteOriginal:        rawWebsite,
		WebsiteIsSocial:        isSocial,
		Email:                  email,
		EmailsAdicionales:      emailsAdicionales,
		TodosLosEmails:         emails,
		LinkedinURL:            linkedin,
		InstagramURL:           instagram,
		FacebookURL:            facebook,
		PresenciaRedesSociales: linkedin != "" || instagram != "" || facebook != "" || isSocial,
		TieneWeb:               realWebsite != "",
		RatingGoogle:           item.TotalScore,
		NumReviews:             reviews,
		AniosEnMercado:         anios,
		DescripcionRaw:         item.Description,
	}
}

// ─── Step 2.5: Pre-calificación Gemini (Ahorro de créditos) ──────────────

type preQualifyResponse struct {
	ScoreLead int `json:"score_lead"`
}

func preQualifyWithGemini(ctx context.Context, emp normalizedCompany, emit pipelineEmitter) int {
	emit("step_start", map[string]interface{}{
		"step": "prequalify", "label": "🤖 Filtro Básico IA (Ahorro Créditos)",
		"detail": emp.NombreEmpresa,
	})

	geminiKey := os.Getenv("GEMINI_API_KEY")
	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", geminiKey)

	rating := 0.0
	if emp.RatingGoogle != nil {
		rating = *emp.RatingGoogle
	}

	prompt := fmt.Sprintf(`Eres un analista B2B experto. Evalúa rápidamente del 0 al 100 si esta empresa es un negocio legítimo y estructurado para ofrecerle consultoría de Negocios, Recursos Humanos y Marketing, o si es un micro-negocio/quiosco sin valor B2B.
- Empresa: %s
- Categoría: %s
- Reseñas: %d (%.1f estrellas)
- Tiene Web: %v

Devuelve EXCLUSIVAMENTE un JSON válido con la forma: {"score_lead": 80}`, emp.NombreEmpresa, emp.CategoriaGoogle, emp.NumReviews, rating, emp.TieneWeb)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: geminiGenConfig{
			Temperature:     0.1,
			MaxOutputTokens: 60,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	emit("api_request", map[string]interface{}{
		"step": "prequalify", "method": "POST", "url": "***",
	})

	geminiCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	data, status, duration, err := timedFetch(geminiCtx, "POST", apiURL, strings.NewReader(string(bodyBytes)),
		map[string]string{"Content-Type": "application/json"})
	if err != nil {
		emit("step_error", map[string]interface{}{"step": "prequalify", "error": err.Error()})
		return 50 // Por defecto pasa si falla Google
	}

	var resp geminiResponse
	json.Unmarshal(data, &resp)

	rawText := "{}"
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		rawText = strings.ReplaceAll(resp.Candidates[0].Content.Parts[0].Text, "```json", "")
		rawText = strings.ReplaceAll(rawText, "```", "")
	}

	var pq preQualifyResponse
	json.Unmarshal([]byte(rawText), &pq)

	score := pq.ScoreLead
	if score == 0 {
		score = 50 // fallback a tolerante si falla parse
	}

	emit("api_response", map[string]interface{}{
		"step": "prequalify", "status": status, "duration": duration,
		"summary": fmt.Sprintf("Pre-Score: %d", score),
	})
	emit("step_done", map[string]interface{}{
		"step": "prequalify", "label": "🤖 Filtro IA", "result": fmt.Sprintf("Score pre-calculado: %d", score),
	})

	return score
}

// ─── Step 3: LinkedIn search via Serper ──────────────────────────

type serperResponse struct {
	Organic []struct {
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic"`
}

type linkedinResult struct {
	Snippet     string
	LinkedinURL string
}

func searchLinkedIn(ctx context.Context, companyName string, emit pipelineEmitter) linkedinResult {
	emit("step_start", map[string]interface{}{
		"step": "serper", "label": "🔎 Buscando LinkedIn", "detail": companyName,
	})

	serperKey := os.Getenv("SERPER_API_KEY")
	apiURL := "https://google.serper.dev/search"

	body := map[string]string{
		"q":  fmt.Sprintf(`site:linkedin.com/company/ "%s"`, companyName),
		"gl": "es",
		"hl": "es",
	}
	bodyBytes, _ := json.Marshal(body)

	emit("api_request", map[string]interface{}{
		"step": "serper", "method": "POST", "url": apiURL,
	})

	data, status, duration, err := timedFetch(ctx, "POST", apiURL, strings.NewReader(string(bodyBytes)),
		map[string]string{
			"X-API-KEY":    serperKey,
			"Content-Type": "application/json",
		})
	if err != nil {
		emit("step_error", map[string]interface{}{"step": "serper", "error": err.Error()})
		return linkedinResult{Snippet: "Error al buscar LinkedIn"}
	}

	var resp serperResponse
	json.Unmarshal(data, &resp)

	result := linkedinResult{Snippet: "No se encontró información en LinkedIn."}
	if len(resp.Organic) > 0 {
		result.Snippet = resp.Organic[0].Snippet
		result.LinkedinURL = resp.Organic[0].Link
	}

	summarySnippet := result.Snippet
	if len(summarySnippet) > 120 {
		summarySnippet = summarySnippet[:120]
	}
	emit("api_response", map[string]interface{}{
		"step": "serper", "status": status, "duration": duration, "summary": summarySnippet,
	})

	label := "No encontrado"
	if result.LinkedinURL != "" {
		label = "Perfil encontrado"
	}
	emit("step_done", map[string]interface{}{
		"step": "serper", "label": "🔎 LinkedIn", "result": label,
	})

	return result
}

// ─── Step 4: Scrape website via Jina ─────────────────────────────

func scrapeWebsite(ctx context.Context, websiteURL string, emit pipelineEmitter) string {
	emit("step_start", map[string]interface{}{
		"step": "jina", "label": "🌐 Escaneando website", "detail": websiteURL,
	})

	jinaURL := fmt.Sprintf("https://r.jina.ai/%s", websiteURL)

	emit("api_request", map[string]interface{}{
		"step": "jina", "method": "GET", "url": jinaURL,
	})

	jinaCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	data, status, duration, err := timedFetch(jinaCtx, "GET", jinaURL, nil,
		map[string]string{"Accept": "text/plain"})
	if err != nil {
		emit("step_error", map[string]interface{}{"step": "jina", "error": err.Error()})
		return "No se pudo extraer contenido de la web o el scrape falló."
	}

	content := string(data)
	if len(content) > 15000 {
		content = content[:15000]
	}

	emit("api_response", map[string]interface{}{
		"step": "jina", "status": status, "duration": duration,
		"summary": fmt.Sprintf("%d caracteres extraídos", len(content)),
	})
	emit("step_done", map[string]interface{}{
		"step": "jina", "label": "🌐 Web Scrape",
		"result": fmt.Sprintf("%d chars", len(content)),
	})

	return content
}

// ─── Step 5: Build Gemini prompts (idénticos al n8n original) ────

func buildPromptWithWeb(emp normalizedCompany, webContent, linkedinSnippet string) string {
	allEmails := "Ninguno"
	if len(emp.TodosLosEmails) > 0 {
		allEmails = strings.Join(emp.TodosLosEmails, ", ")
	}
	ratingStr := "N/A"
	if emp.RatingGoogle != nil {
		ratingStr = fmt.Sprintf("%.1f", *emp.RatingGoogle)
	}
	descRaw := emp.DescripcionRaw
	if descRaw == "" {
		descRaw = "N/A"
	}

	return fmt.Sprintf(`Actúa como un Consultor Estratégico Senior y analista de M&A para Sincron-IA. Somos una firma B2B especializada en acelerar el crecimiento empresarial mediante:
1. Optimización y reestructuración de Recursos Humanos (HR).
2. Estrategias avanzadas de Marketing B2B/B2C y posicionamiento.
3. Transformación organizacional profunda.
4. Escalamiento de Operaciones de Ventas (Sales).

EVALUACIÓN DE CANDIDATO ICP (Ideal Customer Profile) CON ANÁLISIS WEB PROFUNDO
-----------------------------------------------------------------------------
Datos de Empresa (Google Maps):
- Empresa: %s
- Sector/Categoría: %s
- Ciudad: %s, %s
- Descripción Maps: %s
- Reputación: %s estrellas (%d reseñas)
- Correos encontrados previamente: %s

-----------------------------------------------------------------------------
INFORMACIÓN DE SU PERFIL DE LINKEDIN (Snippet Google):
- %s

-----------------------------------------------------------------------------
CONTENIDO EXTRAÍDO DE SU PÁGINA WEB (%s):
%s

INSTRUCCIONES DE ANÁLISIS:
Examina minuciosamente el contenido de esta web en conjunto con sus datos de Google Maps y su información de LinkedIn.
Lee "entre líneas" sobre cómo se comunican, la complejidad de sus servicios, si mencionan equipo o tecnología, y su propuesta de valor. Identifica dónde fallan (¿mala comunicación? ¿web anticuada? ¿falta de foco B2B? ¿dificultades de talento evidentes?) y cómo Sincron-IA puede transformar su negocio.

Formato de respuesta: DEVUELVE EXCLUSIVAMENTE UN JSON VÁLIDO. Sin bloques de markdown.
{
  "nicho_especifico": "Descripción extremadamente detallada del nicho basada en su propia copia web",
  "descripcion_negocio": "Qué hacen exactamente (modelado desde su web) y su principal promesa de valor",
  "madurez_digital": "Evalúa su madurez (Baja/Media/Alta) basado en cómo está estructurada y redactada su web",
  "necesidades_rrhh": "¿Hay signos de que necesiten optimizar HR?",
  "necesidades_ventas": "¿Dejan claro cómo comprar o hay fricción comercial?",
  "angulo_transformacion": "Qué servicio de Sincron-IA sería más crítico venderles HOY",
  "gran_potencial_estrategico": true o false,
  "score_lead": 0 a 100 (100 = lead perfecto ICP),
  "razones_score": "2 oraciones justificando estratégicamente el score usando evidencia explícita de la web",
  "dificultad": "Alta, Media o Baja. ¿Cuán difícil sería venderles nuestros servicios? (Ej: Si su marca y web son de nivel de clase mundial, la dificultad es Alta. Si su marca es mala o no existe, es Baja)",
  "viabilidad_razon": "Justificación de 1 o 2 líneas explicando por qué crees que son viables o difíciles para nuestros servicios.",
  "icebreaker_hiper_personalizado": "Línea de apertura hiper-personalizada para cold email.",
  "email_detectado": "Email de contacto principal si aparece en el sitio web (o null)",
  "contactos_clasificados": [
    { "email": "ejemplo@empresa.com", "departamento": "Recursos Humanos / Ventas / General / Dirección / Otro" }
  ]
}`,
		emp.NombreEmpresa, emp.CategoriaGoogle, emp.Ciudad, emp.Pais,
		descRaw, ratingStr, emp.NumReviews, allEmails,
		linkedinSnippet,
		emp.Website, webContent)
}

func buildPromptWithoutWeb(emp normalizedCompany, linkedinSnippet string) string {
	allEmails := "Ninguno"
	if len(emp.TodosLosEmails) > 0 {
		allEmails = strings.Join(emp.TodosLosEmails, ", ")
	}
	ratingStr := "N/A"
	if emp.RatingGoogle != nil {
		ratingStr = fmt.Sprintf("%.1f", *emp.RatingGoogle)
	}
	descRaw := emp.DescripcionRaw
	if descRaw == "" {
		descRaw = "N/A"
	}
	rrss := "No"
	if emp.PresenciaRedesSociales {
		rrss = "Sí"
	}
	if linkedinSnippet == "" {
		linkedinSnippet = "No disponible"
	}

	return fmt.Sprintf(`Actúa como un Consultor Estratégico Senior y analista de M&A para Sincron-IA. Somos una firma B2B especializada en acelerar el crecimiento empresarial mediante:
1. Optimización y reestructuración de Recursos Humanos (HR).
2. Estrategias avanzadas de Marketing B2B/B2C y posicionamiento.
3. Transformación organizacional profunda.
4. Escalamiento de Operaciones de Ventas (Sales).

EVALUACIÓN DE CANDIDATO ICP (Ideal Customer Profile)
--------------------------------------------------
Datos disponibles (Google Maps):
- Empresa: %s
- Sector/Categoría: %s
- Ubicación: %s, %s
- Descripción: %s
- Reputación: %s estrellas (basado en %d reseñas)
- Redes sociales detectadas: %s
- Trayectoria estimada: ~%d años
- Correos encontrados: %s
INFORMACIÓN DE LINKEDIN (Snippet de Google):
- %s

Nota: Esta empresa NO tiene página web o no fue detectada.

INSTRUCCIONES DE ANÁLISIS:
Examina esta empresa para determinar su valor estratégico como cliente para Sincron-IA. Considera lo que implica no tener página web reportada en 2026. Buscamos pymes consolidadas, empresas de servicios, B2B, comercializadoras de nicho, o empresas con capacidad de contratación, NO micro-negocios.

Formato de respuesta: DEVUELVE EXCLUSIVAMENTE UN JSON VÁLIDO. Sin bloques de markdown.
{
  "nicho_especifico": "Descripción muy específica del nicho y modelo de negocio deducido",
  "descripcion_negocio": "Qué hacen y a quién venden",
  "madurez_digital": "Baja, Media o Alta",
  "necesidades_rrhh": "Hipótesis sobre retos de contratación, retención o cultura",
  "necesidades_ventas": "Hipótesis sobre sus cuellos de botella comerciales",
  "angulo_transformacion": "Qué servicio de Sincron-IA sería más crítico venderles",
  "gran_potencial_estrategico": true o false,
  "score_lead": 0 a 100,
  "razones_score": "2 oraciones justificando estratégicamente el score",
  "dificultad": "Alta, Media o Baja. Cuán difícil sería mejorarles comercialmente y venderles.",
  "viabilidad_razon": "Explicación breve de su viabilidad comercial actual.",
  "icebreaker_hiper_personalizado": "Línea de apertura para email B2B.",
  "contactos_clasificados": [
    { "email": "ejemplo@empresa.com", "departamento": "Recursos Humanos / Ventas / General / Dirección / Otro" }
  ]
}`,
		emp.NombreEmpresa, emp.CategoriaGoogle, emp.Ciudad, emp.Pais,
		descRaw, ratingStr, emp.NumReviews, rrss,
		emp.AniosEnMercado, allEmails, linkedinSnippet)
}

// ─── Step 6: Analyze with Gemini ─────────────────────────────────

type geminiRequest struct {
	Contents         []geminiContent  `json:"contents"`
	GenerationConfig geminiGenConfig  `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	Temperature    float64 `json:"temperature"`
	MaxOutputTokens int    `json:"maxOutputTokens"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func analyzeWithGemini(ctx context.Context, prompt string, emit pipelineEmitter) (string, error) {
	emit("step_start", map[string]interface{}{
		"step": "gemini", "label": "🤖 Analizando con Gemini AI",
		"detail": "Evaluación ICP estratégica",
	})

	geminiKey := os.Getenv("GEMINI_API_KEY")
	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", geminiKey)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: geminiGenConfig{
			Temperature:     0.2,
			MaxOutputTokens: 4096,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	emit("api_request", map[string]interface{}{
		"step": "gemini", "method": "POST",
		"url": strings.Replace(apiURL, geminiKey, "***", 1),
	})

	geminiCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	data, status, duration, err := timedFetch(geminiCtx, "POST", apiURL, strings.NewReader(string(bodyBytes)),
		map[string]string{"Content-Type": "application/json"})
	if err != nil {
		emit("step_error", map[string]interface{}{"step": "gemini", "error": err.Error()})
		return "", err
	}

	var resp geminiResponse
	json.Unmarshal(data, &resp)

	rawText := "{}"
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		rawText = resp.Candidates[0].Content.Parts[0].Text
	}

	emit("api_response", map[string]interface{}{
		"step": "gemini", "status": status, "duration": duration,
		"summary": fmt.Sprintf("Respuesta IA recibida (%d chars)", len(rawText)),
	})
	emit("step_done", map[string]interface{}{
		"step": "gemini", "label": "🤖 Gemini AI", "result": "Análisis completado",
	})

	return rawText, nil
}

// ─── Step 7: Parse Gemini response ───────────────────────────────

type geminiAnalysis struct {
	NichoEspecifico            string                   `json:"nicho_especifico"`
	Nicho                      string                   `json:"nicho"`
	DescripcionNegocio         string                   `json:"descripcion_negocio"`
	MadurezDigital             string                   `json:"madurez_digital"`
	NecesidadesRRHH            string                   `json:"necesidades_rrhh"`
	NecesidadesVentas          string                   `json:"necesidades_ventas"`
	AnguloTransformacion       string                   `json:"angulo_transformacion"`
	GranPotencialEstrategico   bool                     `json:"gran_potencial_estrategico"`
	ScoreLead                  json.Number              `json:"score_lead"`
	RazonesScore               string                   `json:"razones_score"`
	RazonScore                 string                   `json:"razon_score"`
	Dificultad                 string                   `json:"dificultad"`
	ViabilidadRazon            string                   `json:"viabilidad_razon"`
	IcebreakerPersonalizado    string                   `json:"icebreaker_hiper_personalizado"`
	Icebreaker                 string                   `json:"icebreaker"`
	EmailDetectado             *string                  `json:"email_detectado"`
	ContactosClasificados      []map[string]string      `json:"contactos_clasificados"`
}

type parsedLead struct {
	// Datos empresa
	NombreEmpresa          string
	GooglePlaceID          string
	Email                  string
	EmailsAdicionales      []string
	Telefono               string
	Website                string
	Direccion              string
	Ciudad                 string
	Provincia              string
	Pais                   string
	CodigoPostal           string
	Latitud                *float64
	Longitud               *float64
	CategoriaGoogle        string
	RatingGoogle           *float64
	NumReviews             int
	TieneWeb               bool
	PresenciaRedesSociales bool
	LinkedinURL            string
	InstagramURL           string
	FacebookURL            string
	AniosEnMercado         int
	DescripcionRaw         string
	// Análisis IA
	NichoEspecifico          string
	DescripcionNegocio       string
	MadurezDigital           string
	NecesidadesRRHH          string
	NecesidadesVentas        string
	AnguloTransformacion     string
	GranPotencialEstrategico bool
	ScoreLead                int
	RazonesScore             string
	Dificultad               string
	ViabilidadRazon          string
	Icebreaker               string
	AnalisisCompleto         string
	ContactosClasificados    []map[string]string
	Fuente                   string
	Estado                   string
}

func parseGeminiResponse(rawText string, emp normalizedCompany, origenPrompt string) parsedLead {
	// Limpiar bloques markdown
	cleaned := strings.ReplaceAll(rawText, "```json", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")
	cleaned = strings.TrimSpace(cleaned)

	var ai geminiAnalysis
	if err := json.Unmarshal([]byte(cleaned), &ai); err != nil {
		ai = geminiAnalysis{
			NichoEspecifico:          emp.CategoriaGoogle,
			DescripcionNegocio:       emp.DescripcionRaw,
			MadurezDigital:           "Baja",
			GranPotencialEstrategico: false,
			ScoreLead:                "30",
			RazonesScore:             "Error de IA al parsear JSON",
			IcebreakerPersonalizado:  "Hola, me encantaría conocer más sobre su empresa.",
		}
	}

	// Resolver campos con alias
	nicho := ai.NichoEspecifico
	if nicho == "" {
		nicho = ai.Nicho
	}
	razonesScore := ai.RazonesScore
	if razonesScore == "" {
		razonesScore = ai.RazonScore
	}
	icebreaker := ai.IcebreakerPersonalizado
	if icebreaker == "" {
		icebreaker = ai.Icebreaker
	}

	// Score clamped 0-100
	scoreInt, _ := ai.ScoreLead.Int64()
	score := int(math.Min(100, math.Max(0, float64(scoreInt))))
	if score == 0 {
		score = 30
	}

	contactos := ai.ContactosClasificados
	if contactos == nil {
		contactos = []map[string]string{}
	}

	// Email: priorizar el de empresa, luego el detectado por IA
	email := emp.Email
	if email == "" && ai.EmailDetectado != nil {
		email = *ai.EmailDetectado
	}

	// Construir análisis completo
	contactoStr := ""
	if len(contactos) > 0 {
		lines := make([]string, len(contactos))
		for i, c := range contactos {
			lines[i] = fmt.Sprintf("- %s (%s)", c["email"], c["departamento"])
		}
		contactoStr = "\n\n📧 CONTACTOS CLASIFICADOS:\n" + strings.Join(lines, "\n")
	}

	analisis := fmt.Sprintf("🎯 Nicho: %s | 💻 Madurez Digital: %s\n💡 Ángulo de Venta: %s\n🤝 RRHH: %s\n📈 Ventas: %s\n📝 Razones Score: %s%s",
		nicho, ai.MadurezDigital,
		ai.AnguloTransformacion,
		ai.NecesidadesRRHH,
		ai.NecesidadesVentas,
		razonesScore, contactoStr)

	fuente := "google_maps_only"
	if origenPrompt == "con_web" {
		fuente = "google_maps_web_deep_scrape"
	}

	return parsedLead{
		NombreEmpresa:            emp.NombreEmpresa,
		GooglePlaceID:            emp.GooglePlaceID,
		Email:                    email,
		EmailsAdicionales:        emp.EmailsAdicionales,
		Telefono:                 emp.Telefono,
		Website:                  emp.Website,
		Direccion:                emp.Direccion,
		Ciudad:                   emp.Ciudad,
		Provincia:                emp.Provincia,
		Pais:                     emp.Pais,
		CodigoPostal:             emp.CodigoPostal,
		Latitud:                  emp.Latitud,
		Longitud:                 emp.Longitud,
		CategoriaGoogle:          emp.CategoriaGoogle,
		RatingGoogle:             emp.RatingGoogle,
		NumReviews:               emp.NumReviews,
		TieneWeb:                 emp.TieneWeb,
		PresenciaRedesSociales:   emp.PresenciaRedesSociales,
		LinkedinURL:              emp.LinkedinURL,
		InstagramURL:             emp.InstagramURL,
		FacebookURL:              emp.FacebookURL,
		AniosEnMercado:           emp.AniosEnMercado,
		DescripcionRaw:           emp.DescripcionRaw,
		NichoEspecifico:          nicho,
		DescripcionNegocio:       ai.DescripcionNegocio,
		MadurezDigital:           ai.MadurezDigital,
		NecesidadesRRHH:          ai.NecesidadesRRHH,
		NecesidadesVentas:        ai.NecesidadesVentas,
		AnguloTransformacion:     ai.AnguloTransformacion,
		GranPotencialEstrategico: ai.GranPotencialEstrategico,
		ScoreLead:                score,
		RazonesScore:             razonesScore,
		Dificultad:               ai.Dificultad,
		ViabilidadRazon:          ai.ViabilidadRazon,
		Icebreaker:               icebreaker,
		ContactosClasificados:    contactos,
		AnalisisCompleto:         analisis,
		Fuente:                   fuente,
		Estado:                   "nuevo",
	}
}
