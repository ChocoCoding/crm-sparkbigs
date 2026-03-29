package services

import (
	"errors"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

var ErrSettingForbidden = errors.New("no tienes permiso sobre este ajuste")

// defaultSettings define los ajustes iniciales que se crean para cada usuario nuevo.
var defaultSettings = []domain.Setting{
	{Category: "general", Key: "default_currency", Value: "EUR", Label: "Moneda por defecto", InputType: "select"},
	{Category: "general", Key: "company_name", Value: "", Label: "Nombre de la empresa", InputType: "text"},
	{Category: "general", Key: "timezone", Value: "Europe/Madrid", Label: "Zona horaria", InputType: "select"},
	{Category: "notifications", Key: "renewal_alert_days", Value: "30", Label: "Días de alerta antes de renovación", InputType: "number"},
	{Category: "notifications", Key: "email_notifications", Value: "true", Label: "Notificaciones por email", InputType: "toggle"},
	{Category: "integrations", Key: "webhook_url", Value: "", Label: "URL de Webhook", InputType: "text"},
}

type settingService struct {
	repo ports.SettingRepository
}

func NewSettingService(repo ports.SettingRepository) ports.SettingService {
	return &settingService{repo: repo}
}

func (s *settingService) GetSettings(userID uint) ([]domain.Setting, error) {
	return s.repo.FindByUserID(userID)
}

func (s *settingService) GetCategory(userID uint, category string) ([]domain.Setting, error) {
	return s.repo.FindByCategory(userID, category)
}

func (s *settingService) Upsert(setting *domain.Setting) error {
	return s.repo.Upsert(setting)
}

func (s *settingService) SeedDefaults(userID uint) error {
	existing, err := s.repo.FindByUserID(userID)
	if err != nil {
		return err
	}
	existingKeys := make(map[string]bool, len(existing))
	for _, st := range existing {
		existingKeys[st.Key] = true
	}
	for _, def := range defaultSettings {
		if existingKeys[def.Key] {
			continue
		}
		entry := def
		entry.UserID = userID
		if err := s.repo.Upsert(&entry); err != nil {
			return err
		}
	}
	return nil
}

func (s *settingService) DeleteSetting(id, userID uint) error {
	settings, err := s.repo.FindByUserID(userID)
	if err != nil {
		return err
	}
	for _, st := range settings {
		if st.ID == id {
			return s.repo.Delete(id)
		}
	}
	return ErrSettingForbidden
}
