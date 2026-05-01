package events

type OrderCreated struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	UserEmail string `json:"user_email"`
	CreatedAt string `json:"created_at"`
}
