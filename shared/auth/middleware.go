package auth

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const claimsContextKey = "auth_claims"

type Role string

const (
	RoleGuest Role = "guest"
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Claims struct {
	Subject string `json:"subject"`
	Email   string `json:"email,omitempty"`
	Role    Role   `json:"role"`
}

type Middleware struct {
	cfg     Config
	limiter *roleLimiter
}

func NewMiddleware(cfg Config) *Middleware {
	return &Middleware{
		cfg:     cfg,
		limiter: newRoleLimiter(cfg),
	}
}

func SuperTokens() gin.HandlerFunc {
	return func(c *gin.Context) {
		handled := false
		Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			c.Next()
			handled = true
		})).ServeHTTP(c.Writer, c.Request)

		if !handled || c.Writer.Written() {
			c.Abort()
		}
	}
}

func CORS(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (origin == cfg.WebsiteDomain || origin == cfg.APIDomain) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", strings.Join(append([]string{"Content-Type"}, CORSHeaders()...), ","))
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (m *Middleware) AuthenticateOptional() gin.HandlerFunc {
	return m.authenticate(false)
}

func (m *Middleware) AuthenticateRequired() gin.HandlerFunc {
	return m.authenticate(true)
}

func (m *Middleware) authenticate(sessionRequired bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		VerifySession(sessionRequired, func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			session := SessionFromRequest(r)
			if session == nil {
				c.Set(claimsContextKey, guestClaims(c.ClientIP()))
				c.Next()
				return
			}

			payload := session.GetAccessTokenPayload()
			role := RoleUser
			if value, ok := payload["role"].(string); ok && value == string(RoleAdmin) {
				role = RoleAdmin
			}

			email, _ := payload["email"].(string)
			c.Set(claimsContextKey, Claims{
				Subject: session.GetUserID(),
				Email:   email,
				Role:    role,
			})
			c.Next()
		}).ServeHTTP(c.Writer, c.Request)

		if c.Writer.Written() {
			c.Abort()
		}
	}
}

func (m *Middleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := CurrentClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		if !m.limiter.Allow(claims) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

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

func CurrentClaims(c *gin.Context) (Claims, bool) {
	claims, ok := c.Get(claimsContextKey)
	if !ok {
		return Claims{}, false
	}
	typed, ok := claims.(Claims)
	return typed, ok
}

type roleLimiter struct {
	guestLimit int
	userLimit  int
	adminLimit int
	mu         sync.Mutex
	buckets    map[string]*bucket
}

type bucket struct {
	second int64
	count  int
}

func newRoleLimiter(cfg Config) *roleLimiter {
	return &roleLimiter{
		guestLimit: max(1, cfg.GuestLimitPerSecond),
		userLimit:  max(1, cfg.UserLimitPerSecond),
		adminLimit: max(1, cfg.AdminLimitPerSecond),
		buckets:    map[string]*bucket{},
	}
}

func (l *roleLimiter) Allow(claims Claims) bool {
	limit := l.userLimit
	switch claims.Role {
	case RoleGuest:
		limit = l.guestLimit
	case RoleAdmin:
		limit = l.adminLimit
	}

	key := claims.Subject + ":" + string(claims.Role)
	currentSecond := time.Now().Unix()

	l.mu.Lock()
	defer l.mu.Unlock()

	b := l.buckets[key]
	if b == nil || b.second != currentSecond {
		l.buckets[key] = &bucket{second: currentSecond, count: 1}
		return true
	}

	if b.count >= limit {
		return false
	}
	b.count++
	return true
}

func guestClaims(ip string) Claims {
	if ip == "" {
		ip = "unknown"
	}
	return Claims{
		Subject: "guest:" + ip,
		Role:    RoleGuest,
	}
}
