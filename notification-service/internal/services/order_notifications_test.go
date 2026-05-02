package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/osamah22/coffee-service/shared/events"
)

type recordingSender struct {
	emails []Email
	err    error
}

func (s *recordingSender) Send(_ context.Context, email Email) error {
	if s.err != nil {
		return s.err
	}
	s.emails = append(s.emails, email)
	return nil
}

func TestSendOrderCreatedBuildsCustomerEmail(t *testing.T) {
	sender := &recordingSender{}
	svc := NewOrderNotificationService(sender, "fallback@example.test")

	err := svc.SendOrderCreated(context.Background(), events.OrderCreated{
		OrderID:       "12345678-1234-1234-1234-123456789abc",
		CustomerEmail: " customer@example.test ",
		Status:        "preparing",
		Total:         12500,
		Items: []events.OrderItem{
			{
				ProductID:    "product-1",
				ProductName:  "Latte",
				Quantity:     2,
				PriceInKurus: 6250,
			},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(sender.emails) != 1 {
		t.Fatalf("expected one email, got %d", len(sender.emails))
	}
	email := sender.emails[0]
	if email.To != "customer@example.test" {
		t.Fatalf("expected trimmed customer email, got %q", email.To)
	}
	if email.Subject != "Coffee order 12345678 received" {
		t.Fatalf("unexpected subject: %q", email.Subject)
	}
	for _, part := range []string{"Status: preparing", "Total: 125.00 TRY", "- 2x Latte (125.00 TRY)"} {
		if !strings.Contains(email.Text, part) {
			t.Fatalf("expected body to contain %q, got:\n%s", part, email.Text)
		}
	}
}

func TestSendOrderCreatedUsesFallbackEmail(t *testing.T) {
	sender := &recordingSender{}
	svc := NewOrderNotificationService(sender, "fallback@example.test")

	err := svc.SendOrderCreated(context.Background(), events.OrderCreated{
		OrderID: "order-1",
		Status:  "preparing",
		Items: []events.OrderItem{
			{ProductID: "product-1", Quantity: 1, PriceInKurus: 5000},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sender.emails[0].To != "fallback@example.test" {
		t.Fatalf("expected fallback recipient, got %q", sender.emails[0].To)
	}
	if !strings.Contains(sender.emails[0].Text, "- 1x product-1 (50.00 TRY)") {
		t.Fatalf("expected product id fallback in body, got:\n%s", sender.emails[0].Text)
	}
}

func TestSendOrderStatusUpdatedBuildsEmail(t *testing.T) {
	sender := &recordingSender{}
	svc := NewOrderNotificationService(sender, "fallback@example.test")

	err := svc.SendOrderStatusUpdated(context.Background(), events.OrderStatusUpdated{
		OrderID:        "abcdef12-1234-1234-1234-123456789abc",
		CustomerEmail:  "customer@example.test",
		PreviousStatus: "preparing",
		Status:         "ready",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	email := sender.emails[0]
	if email.Subject != "Coffee order ABCDEF12 is ready" {
		t.Fatalf("unexpected subject: %q", email.Subject)
	}
	for _, part := range []string{"Previous status: preparing", "Current status: ready"} {
		if !strings.Contains(email.Text, part) {
			t.Fatalf("expected body to contain %q, got:\n%s", part, email.Text)
		}
	}
}

func TestOrderNotificationServiceReturnsSenderError(t *testing.T) {
	expected := errors.New("send failed")
	svc := NewOrderNotificationService(&recordingSender{err: expected}, "fallback@example.test")

	err := svc.SendOrderStatusUpdated(context.Background(), events.OrderStatusUpdated{})

	if !errors.Is(err, expected) {
		t.Fatalf("expected sender error, got %v", err)
	}
}
