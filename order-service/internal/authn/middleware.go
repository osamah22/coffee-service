package authn

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const claimsContextKey = "auth_claims"

type Middleware struct {
	cfg      Config
	verifier *Verifier
	limiter  *roleLimiter
}

func NewMiddleware(cfg Config) (*Middleware, error) {
	verifier, err := NewVerifier(cfg)
	if err != nil {
		return nil, err
	}
	return &Middleware{
		cfg:      cfg,
		verifier: verifier,
		limiter:  newRoleLimiter(cfg),
	}, nil
}

func (m *Middleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.cfg.Enabled {
			c.Set(claimsContextKey, Claims{Subject: "dev", Role: RoleAdmin, Roles: []string{string(RoleAdmin)}})
			c.Next()
			return
		}

		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.Set(claimsContextKey, guestClaims(c.ClientIP()))
			c.Next()
			return
		}

		claims, err := m.verifier.Verify(c.Request.Context(), strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			return
		}

		c.Set(claimsContextKey, claims)
		c.Next()
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
		Roles:   []string{string(RoleGuest)},
	}
}
