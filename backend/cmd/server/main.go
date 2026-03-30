package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/joho/godotenv"
	"github.com/sparkbigs/crm/internal/adapters/handler"
	"github.com/sparkbigs/crm/internal/adapters/middleware"
	"github.com/sparkbigs/crm/internal/adapters/storage"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/services"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// ── 1. Cargar variables de entorno ──────────────────────────
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: no se encontró .env, usando variables del sistema")
	}

	// ── 2. Conexión a la base de datos ───────────────────────────
	dsn := mustEnv("MYSQL_DSN")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Error conectando a MySQL: %v", err)
	}

	// ⚠️ AutoMigrate solo para desarrollo/MVP.
	// En producción usar scripts SQL en /migrations con golang-migrate.
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.License{},
		&domain.RefreshToken{},
		&domain.Company{},
		&domain.Contact{},
		&domain.Deal{},
		&domain.Meeting{},
		&domain.Subscription{},
		&domain.Setting{},
		&domain.APIKey{},
	); err != nil {
		log.Fatalf("Error en AutoMigrate: %v", err)
	}

	// ── 3. Repositorios (adaptadores de salida) ──────────────────
	userRepo         := storage.NewMysqlUserRepository(db)
	licenseRepo      := storage.NewMysqlLicenseRepository(db)
	refreshTokenRepo := storage.NewMysqlRefreshTokenRepository(db)
	companyRepo      := storage.NewMysqlCompanyRepository(db)
	contactRepo      := storage.NewMysqlContactRepository(db)
	dealRepo         := storage.NewMysqlDealRepository(db)
	meetingRepo      := storage.NewMysqlMeetingRepository(db)
	subscriptionRepo := storage.NewMysqlSubscriptionRepository(db)
	settingRepo      := storage.NewMysqlSettingRepository(db)
	dashboardRepo    := storage.NewMysqlDashboardRepository(db)
	apiKeyRepo       := storage.NewMysqlAPIKeyRepository(db)

	// ── 4. Servicios core (lógica de negocio) ────────────────────
	jwtSecret := mustEnv("JWT_SECRET")

	authSvc         := services.NewAuthService(userRepo, refreshTokenRepo, licenseRepo, jwtSecret)
	adminSvc        := services.NewAdminService(userRepo, licenseRepo)
	companySvc      := services.NewCompanyService(companyRepo)
	contactSvc      := services.NewContactService(contactRepo)
	dealSvc         := services.NewDealService(dealRepo)
	meetingSvc      := services.NewMeetingService(meetingRepo)
	subscriptionSvc := services.NewSubscriptionService(subscriptionRepo)
	settingSvc      := services.NewSettingService(settingRepo)
	dashboardSvc    := services.NewDashboardService(dashboardRepo)
	apiKeySvc       := services.NewAPIKeyService(apiKeyRepo)

	// ── Seed ─────────────────────────────────────────────────────
	seedDatabase(db, authSvc)

	// ── 5. Handlers (adaptadores de entrada) ─────────────────────
	authHandler         := handler.NewAuthHandler(authSvc)
	adminHandler        := handler.NewAdminHandler(adminSvc)
	companyHandler      := handler.NewCompanyHandler(companySvc)
	contactHandler      := handler.NewContactHandler(contactSvc)
	dealHandler         := handler.NewDealHandler(dealSvc)
	meetingHandler      := handler.NewMeetingHandler(meetingSvc)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionSvc)
	settingHandler      := handler.NewSettingHandler(settingSvc)
	dashboardHandler    := handler.NewDashboardHandler(dashboardSvc)
	apiKeyHandler       := handler.NewAPIKeyHandler(apiKeySvc)
	webhookHandler      := handler.NewWebhookHandler(companySvc, contactSvc, meetingSvc, subscriptionSvc)

	// ── 6. Fiber + Middleware global ─────────────────────────────
	app := fiber.New(fiber.Config{
		AppName: "CRM SparkBigs v1.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
			})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: mustEnv("CORS_ORIGINS"),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// JWT protege todas las rutas /api/v1/* (bypass automático de /webhooks/)
	app.Use(middleware.NewJWTMiddleware(authSvc))

	// ── 7. Rutas API (protegidas por JWT) ─────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"status": "ok"}})
	})

	authHandler.RegisterRoutes(app)
	adminHandler.RegisterRoutes(app)
	companyHandler.RegisterRoutes(app)
	contactHandler.RegisterRoutes(app)
	dealHandler.RegisterRoutes(app)
	meetingHandler.RegisterRoutes(app)
	subscriptionHandler.RegisterRoutes(app)
	settingHandler.RegisterRoutes(app)
	dashboardHandler.RegisterRoutes(app)
	apiKeyHandler.RegisterRoutes(app)

	// ── 8. Rutas Webhook (protegidas por API Key + rate limiter) ──
	// Rate limit: 120 peticiones/minuto por clave — protección contra abuso
	webhookRateLimiter := limiter.New(limiter.Config{
		Max:        120,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Limitar por API Key, no por IP (puede haber NAT compartido)
			key := c.Get("X-API-Key")
			if len(key) >= 12 {
				return key[:12] // usar solo el prefijo para el rate limit
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "RATE_LIMIT_EXCEEDED", "message": "Demasiadas peticiones. Límite: 120/minuto"},
			})
		},
	})

	apiKeyMw := middleware.NewAPIKeyMiddleware(apiKeySvc)
	webhookHandler.RegisterRoutes(app, fiber.Handler(func(c *fiber.Ctx) error {
		// Aplicar rate limiter primero, luego validar la API Key
		if err := webhookRateLimiter(c); err != nil {
			return err
		}
		return apiKeyMw(c)
	}))

	// ── 9. Arrancar servidor ─────────────────────────────────────
	port := getEnvOrDefault("PORT", "8080")
	log.Printf("Servidor escuchando en :%s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Variable de entorno requerida no definida: %s", key)
	}
	return val
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
