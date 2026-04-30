package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/order-service/internal/dtos"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"github.com/osamah22/coffee-service/order-service/internal/services"
)

type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(router gin.IRouter) {
	auth := router.Group("/auth")
	auth.POST("/signup", h.signup)
	auth.POST("/login", h.login)
}

func (h *AuthHandler) signup(c *gin.Context) {
	var req dtos.SignupRequest
	if !bind(c, &req) {
		return
	}

	user, token, err := h.svc.Signup(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, services.ErrUserExists) {
			status = http.StatusConflict
		}
		respondError(c, err, status)
		return
	}

	c.JSON(http.StatusCreated, authResponse(user, token))
}

func (h *AuthHandler) login(c *gin.Context) {
	var req dtos.LoginRequest
	if !bind(c, &req) {
		return
	}

	user, token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, services.ErrInvalidCredential) {
			status = http.StatusUnauthorized
		}
		respondError(c, err, status)
		return
	}

	c.JSON(http.StatusOK, authResponse(user, token))
}

func authResponse(user *models.User, token string) dtos.AuthResponse {
	return dtos.AuthResponse{
		Token: token,
		User: dtos.UserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
			Role:  string(user.Role),
		},
	}
}
