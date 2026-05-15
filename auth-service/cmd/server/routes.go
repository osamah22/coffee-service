package main

import (
	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/auth-service/internal/handlers"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
)

func addRoutes(router *gin.Engine, middleware *sharedauth.Middleware, authHandler *handlers.AuthHandler) {
	authHandler.Register(router, middleware)
}
