package auth

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type HandlerSet struct {
	cfg Config
}

func NewHandlerSet(cfg Config) *HandlerSet {
	return &HandlerSet{cfg: cfg}
}

func (h *HandlerSet) Register(router gin.IRouter, middleware *Middleware) {
	auth := router.Group("/auth")
	auth.POST("/login", h.login(middleware))
	auth.GET("/me", middleware.AuthenticateRequired(), h.me)
}

func (h *HandlerSet) login(middleware *Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, password, ok := decodeBasicAuth(c.GetHeader("Authorization"))
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="coffee-service"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "basic_auth_required"})
			return
		}

		user, valid := h.cfg.Authenticate(email, password)
		if !valid {
			c.Header("WWW-Authenticate", `Basic realm="coffee-service"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
			return
		}

		token, expiresAt, err := middleware.IssueToken(user)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "token_issue_failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_at":   expiresAt.Format(time.RFC3339),
			"user": gin.H{
				"id":    user.Subject,
				"email": user.Email,
				"role":  string(user.Role),
			},
		})
	}
}

func (h *HandlerSet) me(c *gin.Context) {
	claims, ok := CurrentClaims(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	user, found := h.cfg.UserBySubject(claims.Subject)
	if found {
		claims.Email = user.Email
		claims.Role = user.Role
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    claims.Subject,
		"email": claims.Email,
		"role":  string(claims.Role),
	})
}

func decodeBasicAuth(header string) (string, string, bool) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Basic") {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}
	username, password, ok := strings.Cut(string(decoded), ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(username), password, true
}
