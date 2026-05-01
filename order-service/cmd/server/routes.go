package main

import (
	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/order-service/internal/handlers"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
)

func addRoutes(
	router *gin.Engine,
	authMiddleware *sharedauth.Middleware,
	authHandler *sharedauth.HandlerSet,
	productHandler *handlers.ProductHandler,
	orderHandler *handlers.OrderHandler,
) {
	authHandler.Register(router, authMiddleware)

	protected := router.Group("/")
	protected.Use(authMiddleware.AuthenticateOptional(), authMiddleware.RateLimit())

	productHandler.Register(protected, authMiddleware)
	orderHandler.Register(protected, authMiddleware)
}
