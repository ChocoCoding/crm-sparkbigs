package storage

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlUserRepository struct {
	db *gorm.DB
}

// NewMysqlUserRepository construye el repositorio de usuarios con MySQL/GORM.
func NewMysqlUserRepository(db *gorm.DB) ports.UserRepository {
	return &mysqlUserRepository{db: db}
}

func (r *mysqlUserRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *mysqlUserRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("License").First(&user, id).Error
	return &user, err
}

func (r *mysqlUserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("License").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *mysqlUserRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *mysqlUserRepository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

func (r *mysqlUserRepository) List(offset, limit int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	r.db.Model(&domain.User{}).Count(&total)
	err := r.db.Preload("License").Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}
