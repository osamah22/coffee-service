package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Email        string          `gorm:"size:255;not null;uniqueIndex"`
	PasswordHash string          `gorm:"type:text;not null"`
	Role         sharedauth.Role `gorm:"size:32;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Role = sharedauth.ParseRole(string(u.Role))
	return nil
}

func (u *User) BeforeSave(_ *gorm.DB) error {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Role = sharedauth.ParseRole(string(u.Role))
	return nil
}
