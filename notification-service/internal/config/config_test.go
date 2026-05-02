package config

import "testing"

func TestFromEnvUsesDefaults(t *testing.T) {
	t.Setenv("RABBITMQ_URL", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM", "")
	t.Setenv("NOTIFICATION_FALLBACK_EMAIL", "")

	cfg := FromEnv()

	if cfg.RabbitMQURL != "amqp://guest:guest@localhost:5672/" {
		t.Fatalf("unexpected rabbit url: %q", cfg.RabbitMQURL)
	}
	if cfg.SMTPHost != "" {
		t.Fatalf("expected empty SMTP host, got %q", cfg.SMTPHost)
	}
	if cfg.SMTPPort != 1025 {
		t.Fatalf("expected default SMTP port 1025, got %d", cfg.SMTPPort)
	}
	if cfg.SMTPFrom != "Coffee Service <orders@coffee.local>" {
		t.Fatalf("unexpected SMTP from: %q", cfg.SMTPFrom)
	}
	if cfg.FallbackEmail != "dev@coffee.local" {
		t.Fatalf("unexpected fallback email: %q", cfg.FallbackEmail)
	}
}

func TestFromEnvReadsOverrides(t *testing.T) {
	t.Setenv("RABBITMQ_URL", "amqp://example/")
	t.Setenv("SMTP_HOST", "smtp.example.test")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_USERNAME", "user")
	t.Setenv("SMTP_PASSWORD", "secret")
	t.Setenv("SMTP_FROM", "Orders <orders@example.test>")
	t.Setenv("NOTIFICATION_FALLBACK_EMAIL", "fallback@example.test")

	cfg := FromEnv()

	if cfg.RabbitMQURL != "amqp://example/" {
		t.Fatalf("unexpected rabbit url: %q", cfg.RabbitMQURL)
	}
	if cfg.SMTPHost != "smtp.example.test" {
		t.Fatalf("unexpected SMTP host: %q", cfg.SMTPHost)
	}
	if cfg.SMTPPort != 2525 {
		t.Fatalf("expected SMTP port 2525, got %d", cfg.SMTPPort)
	}
	if cfg.SMTPUsername != "user" || cfg.SMTPPassword != "secret" {
		t.Fatalf("unexpected SMTP credentials")
	}
	if cfg.SMTPFrom != "Orders <orders@example.test>" {
		t.Fatalf("unexpected SMTP from: %q", cfg.SMTPFrom)
	}
	if cfg.FallbackEmail != "fallback@example.test" {
		t.Fatalf("unexpected fallback email: %q", cfg.FallbackEmail)
	}
}

func TestFromEnvFallsBackOnInvalidSMTPPort(t *testing.T) {
	t.Setenv("SMTP_PORT", "not-a-number")

	cfg := FromEnv()

	if cfg.SMTPPort != 1025 {
		t.Fatalf("expected default SMTP port 1025, got %d", cfg.SMTPPort)
	}
}
