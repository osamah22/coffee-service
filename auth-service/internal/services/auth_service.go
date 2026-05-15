package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/osamah22/coffee-service/auth-service/internal/models"
	"github.com/osamah22/coffee-service/shared/events"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrUserNotFound       = errors.New("user_not_found")
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.findByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *AuthService) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.findByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	eventID := uuid.New()
	occurredAt := time.Now().UTC()
	payload, err := json.Marshal(events.PasswordResetRequested{
		EventID:     eventID.String(),
		UserID:      user.ID.String(),
		Email:       user.Email,
		Role:        string(user.Role),
		RequestedAt: occurredAt,
	})
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(&models.OutboxEvent{
			ID:            eventID,
			EventType:     events.PasswordResetRequestedType,
			AggregateType: "user",
			AggregateID:   user.ID.String(),
			RoutingKey:    events.PasswordResetRequestedType,
			Payload:       string(payload),
			OccurredAt:    occurredAt,
		}).Error
	})
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *AuthService) findByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email)).
		First(&user).
		Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
