package seed

import (
	"context"
	"os"
	"strings"

	"github.com/osamah22/coffee-service/auth-service/internal/models"
	"github.com/osamah22/coffee-service/auth-service/internal/services"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func DemoUsers(ctx context.Context, db *gorm.DB) error {
	for _, user := range usersFromEnv() {
		hash, err := services.HashPassword(user.Password)
		if err != nil {
			return err
		}
		row := models.User{
			Email:        user.Email,
			PasswordHash: hash,
			Role:         user.Role,
		}
		if err := db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "email"}},
			DoUpdates: clause.AssignmentColumns([]string{"password_hash", "role", "updated_at"}),
		}).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

type seededUser struct {
	Email    string
	Password string
	Role     sharedauth.Role
}

func usersFromEnv() []seededUser {
	raw := strings.TrimSpace(os.Getenv("AUTH_DEMO_USERS"))
	if raw == "" {
		return defaultUsers()
	}

	parts := strings.Split(raw, ",")
	users := make([]seededUser, 0, len(parts))
	for _, part := range parts {
		fields := strings.Split(strings.TrimSpace(part), ":")
		if len(fields) < 3 {
			continue
		}
		email := strings.TrimSpace(fields[0])
		password := strings.TrimSpace(fields[1])
		role := sharedauth.ParseRole(fields[2])
		if email == "" || password == "" {
			continue
		}
		users = append(users, seededUser{
			Email:    email,
			Password: password,
			Role:     role,
		})
	}
	if len(users) == 0 {
		return defaultUsers()
	}
	return users
}

func defaultUsers() []seededUser {
	return []seededUser{
		{Email: "customer@example.com", Password: "customer123", Role: sharedauth.RoleUser},
		{Email: "barista@coffee.local", Password: "barista123", Role: sharedauth.RoleBarista},
		{Email: "admin@coffee.local", Password: "admin123", Role: sharedauth.RoleAdmin},
	}
}
