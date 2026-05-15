package config

import (
	"os"
	"strconv"
)

type Config struct {
	RabbitMQURL   string
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	SMTPFrom      string
	FallbackEmail string
}

func FromEnv() Config {
	return Config{
		RabbitMQURL:   envOrDefault("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      envIntOrDefault("SMTP_PORT", 1025),
		SMTPUsername:  os.Getenv("SMTP_USERNAME"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:      envOrDefault("SMTP_FROM", "Coffee Service <orders@coffee.local>"),
		FallbackEmail: envOrDefault("NOTIFICATION_FALLBACK_EMAIL", "dev@coffee.local"),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
