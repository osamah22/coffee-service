package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/osamah22/coffee-service/notification-service/internal/config"
	"github.com/osamah22/coffee-service/notification-service/internal/messaging"
	"github.com/osamah22/coffee-service/notification-service/internal/services"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.FromEnv()
	sender := emailSender(cfg)
	notifications := services.NewOrderNotificationService(sender, cfg.FallbackEmail)
	consumer := messaging.NewOrderConsumer(cfg.RabbitMQURL, notifications)

	consumer.Run(ctx)
}

func emailSender(cfg config.Config) services.EmailSender {
	if cfg.SMTPHost == "" {
		log.Println("SMTP_HOST is not set; notification emails will be logged")
		return services.NewLogEmailSender()
	}

	log.Printf("sending notification emails through %s:%d", cfg.SMTPHost, cfg.SMTPPort)
	return services.NewSMTPEmailSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)
}
