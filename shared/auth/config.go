package auth

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	APIDomain     string
	AuthAPIDomain string
	WebsiteDomain string
	JWTSecret     string
	JWTIssuer     string
	JWTTTL        time.Duration
}

func ConfigFromEnv() Config {
	return Config{
		APIDomain:     strings.TrimRight(envOrDefault("API_DOMAIN", "http://localhost:8080"), "/"),
		AuthAPIDomain: strings.TrimRight(envOrDefault("AUTH_API_DOMAIN", "http://localhost:8081"), "/"),
		WebsiteDomain: strings.TrimRight(envOrDefault("WEBSITE_DOMAIN", "http://localhost:5173"), "/"),
		JWTSecret:     envOrDefault("JWT_SECRET", "coffee-service-local-jwt-secret"),
		JWTIssuer:     envOrDefault("JWT_ISSUER", "coffee-service"),
		JWTTTL:        time.Duration(envIntOrDefault("JWT_TTL_MINUTES", 480)) * time.Minute,
	}
}

func (c Config) AllowedOrigins() []string {
	seen := map[string]struct{}{}
	origins := make([]string, 0, 3)
	for _, origin := range []string{c.WebsiteDomain, c.APIDomain, c.AuthAPIDomain} {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	return origins
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envIntOrDefault(key string, fallback int) int {
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
