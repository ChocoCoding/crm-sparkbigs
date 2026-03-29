package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
		&domain.Contact{},
		&domain.Deal{},
	); err != nil {
		log.Fatalf("Error en AutoMigrate: %v", err)
	}

	// ── 3. Repositorios (adaptadores de salida) ──────────────────
	userRepo         := storage.NewMysqlUserRepository(db)
	licenseRepo      := storage.NewMysqlLicenseRepository(db)
	refreshTokenRepo := storage.NewMysqlRefreshTokenRepository(db)
	contactRepo      := storage.NewMysqlContactRepository(db)
	dealRepo         := storage.NewMysqlDealRepository(db)

	// ── 4. Servicios core (lógica de negocio) ────────────────────
	jwtSecret := mustEnv("JWT_SECRET")

	authSvc    := services.NewAuthService(userRepo, refreshTokenRepo, licenseRepo, jwtSecret)
	adminSvc   := services.NewAdminService(userRepo, licenseRepo)
	contactSvc := services.NewContactService(contactRepo)
	dealSvc    := services.NewDealService(dealRepo)

	// ── 5. Handlers (adaptadores de entrada) ─────────────────────
	authHandler    := handler.NewAuthHandler(authSvc)
	adminHandler   := handler.NewAdminHandler(adminSvc)
	contactHandler := handler.NewContactHandler(contactSvc)
	dealHandler    := handler.NewDealHandler(dealSvc)

	// ── 6. Fiber + Middleware global ─────────────────────────────
	app := fiber.New(fiber.Config{
		AppName: "CRM SparkBigs v1.0",
		// Ocultar stack traces en producción
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": err.Error()},
			})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: mustEnv("CORS_ORIGINS"), // NUNCA usar "*" en producción
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Use(middleware.NewJWTMiddleware(authSvc))

	// ── 7. Rutas ─────────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"status": "ok"}})
	})

	authHandler.RegisterRoutes(app)
	adminHandler.RegisterRoutes(app)
	contactHandler.RegisterRoutes(app)
	dealHandler.RegisterRoutes(app)

	// ── 8. Arrancar servidor ─────────────────────────────────────
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
