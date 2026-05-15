package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestConfigAuthenticateMatchesDefaultUser(t *testing.T) {
	cfg := ConfigFromEnv()

	user, ok := cfg.Authenticate("customer@example.com", "customer123")

	if !ok {
		t.Fatal("expected default customer credentials to authenticate")
	}
	if user.Role != RoleCustomer {
		t.Fatalf("expected customer role, got %q", user.Role)
	}
}

func TestLoginExchangesBasicAuthForBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := ConfigFromEnv()
	middleware := NewMiddleware(cfg)

	router := gin.New()
	NewHandlerSet(cfg).Register(router, middleware)

	request := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("staff@coffee.local:staff123")))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	var body struct {
		AccessToken string `json:"access_token"`
		User        struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if body.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if body.User.Role != string(RoleStaff) {
		t.Fatalf("expected staff role, got %q", body.User.Role)
	}
}

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
	token, _, err := middleware.IssueToken(User{
		Subject: "admin-1",
		Email:   "admin@coffee.local",
		Role:    RoleAdmin,
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	router := gin.New()
	router.Use(middleware.AuthenticateOptional())
	router.GET("/staff", middleware.RequireRole(RoleStaff, RoleAdmin), func(c *gin.Context) {
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
		JWTSecret:      "test-secret",
		JWTIssuer:      "coffee-test",
		JWTTTL:         time.Minute,
		DefaultUsers:   defaultUsers(),
		usersByEmail:   map[string]User{},
		usersBySubject: map[string]User{},
	}
	cfg.indexUsers()

	_, expiresAt, err := NewMiddleware(cfg).IssueToken(cfg.DefaultUsers[0])
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if time.Until(expiresAt) <= 0 {
		t.Fatal("expected future expiration time")
	}
}
