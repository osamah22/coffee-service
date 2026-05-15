package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/osamah22/coffee-service/shared/events"
)

type OrderNotificationService struct {
	sender        EmailSender
	fallbackEmail string
}

func NewOrderNotificationService(sender EmailSender, fallbackEmail string) *OrderNotificationService {
	return &OrderNotificationService{
		sender:        sender,
		fallbackEmail: fallbackEmail,
	}
}

func (svc *OrderNotificationService) SendOrderCreated(ctx context.Context, event events.OrderCreated) error {
	recipient := svc.recipient(event.CustomerEmail)
	return svc.sender.Send(ctx, Email{
		To:      recipient,
		Subject: fmt.Sprintf("Coffee order %s received", shortID(event.OrderID)),
		Text:    createdBody(event),
	})
}

func (svc *OrderNotificationService) SendOrderStatusUpdated(ctx context.Context, event events.OrderStatusUpdated) error {
	recipient := svc.recipient(event.CustomerEmail)
	return svc.sender.Send(ctx, Email{
		To:      recipient,
		Subject: fmt.Sprintf("Coffee order %s is %s", shortID(event.OrderID), event.Status),
		Text:    statusBody(event),
	})
}

func (svc *OrderNotificationService) SendPasswordResetRequested(ctx context.Context, event events.PasswordResetRequested) error {
	recipient := svc.recipient(event.Email)
	return svc.sender.Send(ctx, Email{
		To:      recipient,
		Subject: "Coffee Control password reset request received",
		Text:    passwordResetBody(event),
	})
}

func (svc *OrderNotificationService) recipient(email string) string {
	email = strings.TrimSpace(email)
	if email != "" {
		return email
	}
	return svc.fallbackEmail
}

func createdBody(event events.OrderCreated) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Your coffee order is in.\n\n")
	fmt.Fprintf(&b, "Order: %s\n", event.OrderID)
	fmt.Fprintf(&b, "Status: %s\n", event.Status)
	fmt.Fprintf(&b, "Total: %s\n\n", formatPrice(event.Total))
	fmt.Fprintf(&b, "Items:\n")

	for _, item := range event.Items {
		name := item.ProductName
		if name == "" {
			name = item.ProductID
		}
		lineTotal := int64(item.Quantity) * item.PriceInKurus
		fmt.Fprintf(&b, "- %dx %s (%s)\n", item.Quantity, name, formatPrice(lineTotal))
	}

	fmt.Fprintf(&b, "\nThanks for ordering from Coffee Control.\n")
	return b.String()
}

func statusBody(event events.OrderStatusUpdated) string {
	return fmt.Sprintf(
		"Your coffee order status changed.\n\nOrder: %s\nPrevious status: %s\nCurrent status: %s\n",
		event.OrderID,
		event.PreviousStatus,
		event.Status,
	)
}

func passwordResetBody(event events.PasswordResetRequested) string {
	return fmt.Sprintf(
		"We received a password reset request.\n\nUser: %s\nRole: %s\nRequested at: %s\n",
		event.Email,
		event.Role,
		event.RequestedAt.Format(time.RFC3339),
	)
}

func formatPrice(kurus int64) string {
	return fmt.Sprintf("%.2f TRY", float64(kurus)/100)
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return strings.ToUpper(id[:8])
}
