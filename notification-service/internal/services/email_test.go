package services

import (
	"context"
	"strings"
	"testing"
)

func TestSMTPEmailSenderRejectsMissingRecipient(t *testing.T) {
	sender := NewSMTPEmailSender("localhost", 1025, "", "", "Coffee Service <orders@coffee.local>")

	err := sender.Send(context.Background(), Email{
		To:      " ",
		Subject: "Subject",
		Text:    "Body",
	})

	if err == nil {
		t.Fatal("expected missing recipient error")
	}
	if !strings.Contains(err.Error(), "recipient") {
		t.Fatalf("expected recipient error, got %v", err)
	}
}

func TestSMTPEmailSenderHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sender := NewSMTPEmailSender("localhost", 1025, "", "", "Coffee Service <orders@coffee.local>")

	if err := sender.Send(ctx, Email{To: "customer@example.test"}); err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestBuildMessageIncludesHeadersAndBody(t *testing.T) {
	message := string(buildMessage(
		"Coffee Service <orders@coffee.local>",
		"customer@example.test",
		"Order ready",
		"Your drink is ready.",
	))

	expectedParts := []string{
		"From: Coffee Service <orders@coffee.local>\r\n",
		"To: customer@example.test\r\n",
		"Subject: Order ready\r\n",
		"Content-Type: text/plain; charset=\"utf-8\"\r\n",
		"\r\nYour drink is ready.\r\n",
	}

	for _, part := range expectedParts {
		if !strings.Contains(message, part) {
			t.Fatalf("expected message to contain %q, got:\n%s", part, message)
		}
	}
}

func TestEnvelopeAddressParsesDisplayName(t *testing.T) {
	address, err := envelopeAddress("Coffee Service <orders@coffee.local>")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if address != "orders@coffee.local" {
		t.Fatalf("expected parsed address, got %q", address)
	}
}

func TestEnvelopeAddressFallsBackForPlainInvalidValue(t *testing.T) {
	address, err := envelopeAddress("orders")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if address != "orders" {
		t.Fatalf("expected trimmed fallback address, got %q", address)
	}
}
