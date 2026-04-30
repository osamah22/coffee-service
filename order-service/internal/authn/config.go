package authn

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Enabled             bool
	Issuer              string
	Audience            string
	JWKSURL             string
	RoleClaim           string
	DefaultRole         Role
	AdminRoles          map[string]struct{}
	UserRoles           map[string]struct{}
	UserLimitPerSecond  int
	AdminLimitPerSecond int
}

func ConfigFromEnv() Config {
	return Config{
		Enabled:             envBool("AUTH_ENABLED", true),
		Issuer:              strings.TrimRight(os.Getenv("AUTH_ISSUER"), "/"),
		Audience:            os.Getenv("AUTH_AUDIENCE"),
		JWKSURL:             os.Getenv("AUTH_JWKS_URL"),
		RoleClaim:           envOrDefault("AUTH_ROLE_CLAIM", "groups"),
		DefaultRole:         RoleUser,
		AdminRoles:          roleSet(envOrDefault("AUTH_ADMIN_ROLES", "admin,admins,order-service-admin")),
		UserRoles:           roleSet(envOrDefault("AUTH_USER_ROLES", "user,users,order-service-user")),
		UserLimitPerSecond:  envInt("AUTH_USER_RPS", 10),
		AdminLimitPerSecond: envInt("AUTH_ADMIN_RPS", 1000),
	}
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Issuer == "" {
		return ErrMissingIssuer
	}
	if c.Audience == "" {
		return ErrMissingAudience
	}
	if c.JWKSURL == "" {
		return ErrMissingJWKSURL
	}
	if c.UserLimitPerSecond < 1 {
		c.UserLimitPerSecond = 1
	}
	if c.AdminLimitPerSecond < 1 {
		c.AdminLimitPerSecond = 1
	}
	return nil
}

func roleSet(csv string) map[string]struct{} {
	values := strings.Split(csv, ",")
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			set[value] = struct{}{}
		}
	}
	return set
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func now() time.Time {
	return time.Now()
}
