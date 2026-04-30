package main

import (
	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/order-service/internal/handlers"
)

func addRoutes(router *gin.Engine, productHandler *handlers.ProductHandler, orderHandler *handlers.OrderHandler) {
	productHandler.Register(router)
	orderHandler.Register(router)
}
