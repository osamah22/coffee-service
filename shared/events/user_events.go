package events

import "time"

const (
	UserExchange               = "coffee.auth"
	PasswordResetRequestedType = "password_reset.requested"
)

type PasswordResetRequested struct {
	EventID     string    `json:"event_id"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	RequestedAt time.Time `json:"requested_at"`
}
