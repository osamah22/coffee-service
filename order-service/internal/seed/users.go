package seed

import (
	"context"
	"errors"

	"github.com/osamah22/coffee-service/order-service/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	defaultAdminEmail    = "admin@example.com"
	defaultAdminPassword = "admin"
)

func AdminUser(ctx context.Context, db *gorm.DB) error {
	var existing models.User
	result := db.WithContext(ctx).Where("email = ?", defaultAdminEmail).First(&existing)
	if result.Error == nil {
		if existing.Role != models.UserRoleAdmin {
			existing.Role = models.UserRoleAdmin
			return db.WithContext(ctx).Save(&existing).Error
		}
		return nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Create(&models.User{
		Name:         "Order Service Admin",
		Email:        defaultAdminEmail,
		PasswordHash: string(hash),
		Role:         models.UserRoleAdmin,
	}).Error
}
