package auth

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ConnectionURI       string
	AppName             string
	APIDomain           string
	WebsiteDomain       string
	APIBasePath         string
	WebsiteBasePath     string
	AdminEmails         map[string]struct{}
	GuestLimitPerSecond int
	UserLimitPerSecond  int
	AdminLimitPerSecond int
}

func ConfigFromEnv() Config {
	return Config{
		ConnectionURI:       envOrDefault("SUPERTOKENS_CONNECTION_URI", "http://localhost:3567"),
		AppName:             envOrDefault("SUPERTOKENS_APP_NAME", "Coffee Service"),
		APIDomain:           strings.TrimRight(envOrDefault("SUPERTOKENS_API_DOMAIN", "http://localhost:8080"), "/"),
		WebsiteDomain:       strings.TrimRight(envOrDefault("SUPERTOKENS_WEBSITE_DOMAIN", "http://localhost:5173"), "/"),
		APIBasePath:         envOrDefault("SUPERTOKENS_API_BASE_PATH", "/auth"),
		WebsiteBasePath:     envOrDefault("SUPERTOKENS_WEBSITE_BASE_PATH", "/auth"),
		AdminEmails:         emailSet(envOrDefault("SUPERTOKENS_ADMIN_EMAILS", "admin@example.com")),
		GuestLimitPerSecond: envInt("AUTH_GUEST_RPS", 10),
		UserLimitPerSecond:  envInt("AUTH_USER_RPS", 10),
		AdminLimitPerSecond: envInt("AUTH_ADMIN_RPS", 1000),
	}
}

func (c Config) RoleForEmail(email string) Role {
	if _, ok := c.AdminEmails[strings.ToLower(strings.TrimSpace(email))]; ok {
		return RoleAdmin
	}
	return RoleUser
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
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

func emailSet(csv string) map[string]struct{} {
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
