package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAuthenticateOptionalRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := ConfigFromEnv()
	router := gin.New()
	router.Use(NewMiddleware(cfg).AuthenticateOptional())
	router.GET("/products", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	request.Header.Set("Authorization", "Bearer broken")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
}

func TestRequireRoleAcceptsIssuedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := ConfigFromEnv()
	middleware := NewMiddleware(cfg)
	token, _, err := middleware.IssueToken(Claims{
		Subject: "admin-1",
		Email:   "admin@coffee.local",
		Role:    RoleAdmin,
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	router := gin.New()
	router.Use(middleware.AuthenticateOptional())
	router.GET("/staff", middleware.RequireRole(RoleBarista, RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/staff", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.Code)
	}
}

func TestIssuedTokenExpires(t *testing.T) {
	cfg := Config{
		JWTSecret: "test-secret",
		JWTIssuer: "coffee-test",
		JWTTTL:    time.Minute,
	}

	_, expiresAt, err := NewMiddleware(cfg).IssueToken(Claims{
		Subject: "user-1",
		Email:   "user@example.test",
		Role:    RoleUser,
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if time.Until(expiresAt) <= 0 {
		t.Fatal("expected future expiration time")
	}
}

func TestParseRoleMapsLegacyAliases(t *testing.T) {
	if got := ParseRole("customer"); got != RoleUser {
		t.Fatalf("expected customer alias to map to user, got %q", got)
	}
	if got := ParseRole("staff"); got != RoleBarista {
		t.Fatalf("expected staff alias to map to barista, got %q", got)
	}
}
