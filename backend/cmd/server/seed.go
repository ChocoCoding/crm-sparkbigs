package main

import (
	"log"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

// seedDatabase crea el usuario administrador inicial si la base de datos está vacía.
// Solo se ejecuta en el arranque del servidor.
func seedDatabase(db *gorm.DB, authService ports.AuthService) {
	var count int64
	db.Model(&domain.User{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("Base de datos vacía — creando datos iniciales...")

	// ── Admin principal ──────────────────────────────────────────
	adminHash, err := authService.HashPassword("Admin1234!")
	if err != nil {
		log.Printf("Error generando hash de contraseña: %v", err)
		return
	}

	admin := &domain.User{
		Email:          "admin@sparkbigs.com",
		PasswordHash:   adminHash,
		Name:           "Admin SparkBIGS",
		Role:           "admin",
		IsActive:       true,
		MustChangePass: false,
	}

	if err := db.Create(admin).Error; err != nil {
		log.Printf("Error creando admin: %v", err)
		return
	}

	// Licencia enterprise para el admin
	db.Create(&domain.License{
		UserID:   admin.ID,
		Plan:     "enterprise",
		IsActive: true,
	})

	// ── Usuario de demo ──────────────────────────────────────────
	userHash, _ := authService.HashPassword("User1234!")
	demoUser := &domain.User{
		Email:          "demo@sparkbigs.com",
		PasswordHash:   userHash,
		Name:           "Usuario Demo",
		Role:           "user",
		IsActive:       true,
		MustChangePass: false,
	}

	if err := db.Create(demoUser).Error; err != nil {
		log.Printf("Error creando usuario demo: %v", err)
		return
	}

	db.Create(&domain.License{
		UserID:   demoUser.ID,
		Plan:     "pro",
		IsActive: true,
	})

	log.Println("✓ Seed completado:")
	log.Println("  → admin@sparkbigs.com  / Admin1234!  (role: admin)")
	log.Println("  → demo@sparkbigs.com   / User1234!   (role: user)")
}
