package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/osamah22/coffee-service/auth-service/internal/models"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
	"github.com/osamah22/coffee-service/shared/events"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoginAcceptsSeededHash(t *testing.T) {
	db := newTestDB(t)
	hash, err := HashPassword("user12345")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := models.User{
		Email:        "user@example.test",
		PasswordHash: hash,
		Role:         sharedauth.RoleUser,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	authService := NewAuthService(db)
	loggedIn, err := authService.Login(context.Background(), "user@example.test", "user12345")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loggedIn.Role != sharedauth.RoleUser {
		t.Fatalf("expected user role, got %q", loggedIn.Role)
	}
}

func TestRequestPasswordResetEnqueuesEvent(t *testing.T) {
	db := newTestDB(t)
	hash, err := HashPassword("barista123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := models.User{
		Email:        "barista@example.test",
		PasswordHash: hash,
		Role:         sharedauth.RoleBarista,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	authService := NewAuthService(db)
	if err := authService.RequestPasswordReset(context.Background(), user.Email); err != nil {
		t.Fatalf("request password reset: %v", err)
	}

	var outbox models.OutboxEvent
	if err := db.First(&outbox).Error; err != nil {
		t.Fatalf("expected outbox event, got %v", err)
	}
	if outbox.EventType != events.PasswordResetRequestedType {
		t.Fatalf("expected event type %q, got %q", events.PasswordResetRequestedType, outbox.EventType)
	}

	var payload events.PasswordResetRequested
	if err := json.Unmarshal([]byte(outbox.Payload), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Email != user.Email {
		t.Fatalf("expected email %q, got %q", user.Email, payload.Email)
	}
	if payload.Role != string(sharedauth.RoleBarista) {
		t.Fatalf("expected role %q, got %q", sharedauth.RoleBarista, payload.Role)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.OutboxEvent{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}
