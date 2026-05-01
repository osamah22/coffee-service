package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
)

type HandlerSet struct {
	cfg Config
}

func NewHandlerSet(cfg Config) *HandlerSet {
	return &HandlerSet{cfg: cfg}
}

func (h *HandlerSet) Register(router gin.IRouter, middleware *Middleware) {
	router.GET("/auth/me", middleware.AuthenticateRequired(), h.me)
}

func (h *HandlerSet) me(c *gin.Context) {
	claims, ok := CurrentClaims(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	user, err := emailpassword.GetUserByID(claims.Subject)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user lookup failed"})
		return
	}
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  string(h.cfg.RoleForEmail(user.Email)),
	})
}
