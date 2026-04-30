package services

import (
	"context"
	"errors"
	"strings"

	"github.com/osamah22/coffee-service/order-service/internal/authn"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredential  = errors.New("invalid email or password")
	ErrTokenConfiguration = errors.New("token issuer is not configured")
)

type AuthService struct {
	DB     *gorm.DB
	issuer *authn.TokenIssuer
}

func NewAuthService(db *gorm.DB, issuer *authn.TokenIssuer) *AuthService {
	return &AuthService{DB: db, issuer: issuer}
}

func (svc *AuthService) Signup(ctx context.Context, name, email, password string) (*models.User, string, error) {
	email = normalizeEmail(email)
	var existing models.User
	result := svc.DB.WithContext(ctx).Where("email = ?", email).First(&existing)
	if result.Error == nil {
		return nil, "", ErrUserExists
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, "", result.Error
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{
		Name:         strings.TrimSpace(name),
		Email:        email,
		PasswordHash: hash,
		Role:         models.UserRoleUser,
	}
	if err := svc.DB.WithContext(ctx).Create(user).Error; err != nil {
		return nil, "", err
	}

	token, err := svc.issueToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (svc *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	var user models.User
	result := svc.DB.WithContext(ctx).Where("email = ?", normalizeEmail(email)).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, "", ErrInvalidCredential
	}
	if result.Error != nil {
		return nil, "", result.Error
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, "", ErrInvalidCredential
	}

	token, err := svc.issueToken(&user)
	if err != nil {
		return nil, "", err
	}
	return &user, token, nil
}

func (svc *AuthService) issueToken(user *models.User) (string, error) {
	if svc.issuer == nil {
		return "", ErrTokenConfiguration
	}
	return svc.issuer.Issue(user.ID.String(), user.Email, authn.Role(user.Role))
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
