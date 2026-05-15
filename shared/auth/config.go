package auth

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	APIDomain      string
	WebsiteDomain  string
	JWTSecret      string
	JWTIssuer      string
	JWTTTL         time.Duration
	DefaultUsers   []User
	usersByEmail   map[string]User
	usersBySubject map[string]User
}

type User struct {
	Subject  string
	Email    string
	Password string
	Role     Role
}

func ConfigFromEnv() Config {
	users := usersFromEnv()
	cfg := Config{
		APIDomain:     strings.TrimRight(envOrDefault("API_DOMAIN", "http://localhost:8080"), "/"),
		WebsiteDomain: strings.TrimRight(envOrDefault("WEBSITE_DOMAIN", "http://localhost:5173"), "/"),
		JWTSecret:     envOrDefault("JWT_SECRET", "coffee-service-local-jwt-secret"),
		JWTIssuer:     envOrDefault("JWT_ISSUER", "coffee-service"),
		JWTTTL:        time.Duration(envIntOrDefault("JWT_TTL_MINUTES", 480)) * time.Minute,
		DefaultUsers:  users,
	}
	cfg.indexUsers()
	return cfg
}

func (c Config) Authenticate(email, password string) (User, bool) {
	user, ok := c.usersByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok || user.Password != password {
		return User{}, false
	}
	return user, true
}

func (c Config) UserBySubject(subject string) (User, bool) {
	user, ok := c.usersBySubject[strings.TrimSpace(subject)]
	return user, ok
}

func (c *Config) indexUsers() {
	c.usersByEmail = make(map[string]User, len(c.DefaultUsers))
	c.usersBySubject = make(map[string]User, len(c.DefaultUsers))
	for _, user := range c.DefaultUsers {
		user.Email = strings.TrimSpace(user.Email)
		user.Subject = strings.TrimSpace(user.Subject)
		if user.Subject == "" {
			user.Subject = "user:" + strings.ToLower(user.Email)
		}
		c.usersByEmail[strings.ToLower(user.Email)] = user
		c.usersBySubject[user.Subject] = user
	}
}

func usersFromEnv() []User {
	raw := strings.TrimSpace(os.Getenv("DEMO_USERS"))
	if raw == "" {
		return defaultUsers()
	}

	parts := strings.Split(raw, ",")
	users := make([]User, 0, len(parts))
	for _, part := range parts {
		fields := strings.Split(strings.TrimSpace(part), ":")
		if len(fields) < 3 {
			continue
		}
		user := User{
			Email:    strings.TrimSpace(fields[0]),
			Password: strings.TrimSpace(fields[1]),
			Role:     ParseRole(fields[2]),
		}
		if len(fields) > 3 {
			user.Subject = strings.TrimSpace(fields[3])
		}
		if user.Email == "" || user.Password == "" {
			continue
		}
		users = append(users, user)
	}
	if len(users) == 0 {
		return defaultUsers()
	}
	return users
}

func defaultUsers() []User {
	return []User{
		{Subject: "customer-1", Email: "customer@example.com", Password: "customer123", Role: RoleCustomer},
		{Subject: "staff-1", Email: "staff@coffee.local", Password: "staff123", Role: RoleStaff},
		{Subject: "admin-1", Email: "admin@coffee.local", Password: "admin123", Role: RoleAdmin},
	}
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
