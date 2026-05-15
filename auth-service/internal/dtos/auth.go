package dtos

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}
