package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const claimsContextKey = "auth_claims"

var (
	errMissingBearerToken = errors.New("missing bearer token")
	errInvalidBearerToken = errors.New("invalid bearer token")
)

type Role string

const (
	RoleGuest    Role = "guest"
	RoleCustomer Role = "customer"
	RoleStaff    Role = "staff"
	RoleAdmin    Role = "admin"
)

type Claims struct {
	Subject string `json:"subject"`
	Email   string `json:"email,omitempty"`
	Role    Role   `json:"role"`
}

type jwtClaims struct {
	Email string `json:"email,omitempty"`
	Role  Role   `json:"role"`
	jwt.RegisteredClaims
}

type Middleware struct {
	cfg Config
}

func NewMiddleware(cfg Config) *Middleware {
	return &Middleware{cfg: cfg}
}

func CORS(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (origin == cfg.WebsiteDomain || origin == cfg.APIDomain) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", strings.Join([]string{
			"Accept",
			"Authorization",
			"Content-Type",
		}, ","))
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (m *Middleware) AuthenticateOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.claimsFromRequest(c)
		if err != nil && !errors.Is(err, errMissingBearerToken) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		if err == nil {
			c.Set(claimsContextKey, claims)
		}
		c.Next()
	}
}

func (m *Middleware) AuthenticateRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.claimsFromRequest(c)
		if err != nil {
			status := http.StatusUnauthorized
			message := "authentication required"
			if !errors.Is(err, errMissingBearerToken) {
				message = "invalid_token"
			}
			c.AbortWithStatusJSON(status, gin.H{"error": message})
			return
		}
		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func (m *Middleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func (m *Middleware) RequireRole(roles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := CurrentClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		for _, role := range roles {
			if claims.Role == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
	}
}

func (m *Middleware) IssueToken(user User) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.cfg.JWTTTL)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Subject,
			Issuer:    m.cfg.JWTIssuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	signed, err := token.SignedString([]byte(m.cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (m *Middleware) claimsFromRequest(c *gin.Context) (Claims, error) {
	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if header == "" {
		return Claims{}, errMissingBearerToken
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return Claims{}, fmt.Errorf("%w: malformed authorization header", errInvalidBearerToken)
	}

	token, err := jwt.ParseWithClaims(parts[1], &jwtClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.cfg.JWTSecret), nil
	}, jwt.WithIssuer(m.cfg.JWTIssuer))
	if err != nil {
		return Claims{}, err
	}

	typedClaims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return Claims{}, errors.New("invalid token claims")
	}

	return Claims{
		Subject: typedClaims.Subject,
		Email:   typedClaims.Email,
		Role:    ParseRole(string(typedClaims.Role)),
	}, nil
}

func CurrentClaims(c *gin.Context) (Claims, bool) {
	claims, ok := c.Get(claimsContextKey)
	if !ok {
		return Claims{}, false
	}
	typed, ok := claims.(Claims)
	return typed, ok
}

func ParseRole(value string) Role {
	switch Role(strings.ToLower(strings.TrimSpace(value))) {
	case RoleStaff:
		return RoleStaff
	case RoleAdmin:
		return RoleAdmin
	case RoleGuest:
		return RoleGuest
	default:
		return RoleCustomer
	}
}
