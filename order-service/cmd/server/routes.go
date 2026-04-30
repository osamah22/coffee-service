package main

import (
	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/order-service/internal/authn"
	"github.com/osamah22/coffee-service/order-service/internal/handlers"
)

func addRoutes(
	router *gin.Engine,
	authMiddleware *authn.Middleware,
	authHandler *handlers.AuthHandler,
	productHandler *handlers.ProductHandler,
	orderHandler *handlers.OrderHandler,
) {
	authHandler.Register(router)

	protected := router.Group("/")
	protected.Use(authMiddleware.Authenticate(), authMiddleware.RateLimit())

	productHandler.Register(protected, authMiddleware)
	orderHandler.Register(protected, authMiddleware)
}
