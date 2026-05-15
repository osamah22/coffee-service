package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/auth-service/internal/dtos"
	"github.com/osamah22/coffee-service/auth-service/internal/services"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
)

type AuthHandler struct {
	authService *services.AuthService
	middleware  *sharedauth.Middleware
}

func NewAuthHandler(authService *services.AuthService, middleware *sharedauth.Middleware) *AuthHandler {
	return &AuthHandler{authService: authService, middleware: middleware}
}

func (h *AuthHandler) Register(router gin.IRouter, middleware *sharedauth.Middleware) {
	auth := router.Group("/auth")
	auth.POST("/login", h.login)
	auth.GET("/me", middleware.AuthenticateRequired(), h.me)
	auth.POST("/password-reset-requests", h.passwordResetRequest)
}

func (h *AuthHandler) login(c *gin.Context) {
	var req dtos.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_login_request"})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if !errors.Is(err, services.ErrInvalidCredentials) {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	token, expiresAt, err := h.middleware.IssueToken(sharedauth.Claims{
		Subject: user.ID.String(),
		Email:   user.Email,
		Role:    user.Role,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_issue_failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_at":   expiresAt.Format(time.RFC3339),
		"user": gin.H{
			"id":    user.ID.String(),
			"email": user.Email,
			"role":  string(user.Role),
		},
	})
}

func (h *AuthHandler) me(c *gin.Context) {
	claims, ok := sharedauth.CurrentClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	user, err := h.authService.FindByID(c.Request.Context(), claims.Subject)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, services.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID.String(),
		"email": user.Email,
		"role":  string(user.Role),
	})
}

func (h *AuthHandler) passwordResetRequest(c *gin.Context) {
	var req dtos.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_password_reset_request"})
		return
	}

	if err := h.authService.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password_reset_request_failed"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}
