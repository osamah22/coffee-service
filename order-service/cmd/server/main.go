package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/order-service/internal/handlers"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"github.com/osamah22/coffee-service/order-service/internal/seed"
	"github.com/osamah22/coffee-service/order-service/internal/services"
	sharedauth "github.com/osamah22/coffee-service/shared/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	port := envOrDefault("PORT", "8080")
	db := setupDatabase()
	authConfig := sharedauth.ConfigFromEnv()
	if err := sharedauth.InitSuperTokens(authConfig); err != nil {
		log.Fatal("supertokens setup failed:", err)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(sharedauth.CORS(authConfig), sharedauth.SuperTokens())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	authMiddleware := sharedauth.NewMiddleware(authConfig)

	addRoutes(
		router,
		authMiddleware,
		sharedauth.NewHandlerSet(authConfig),
		handlers.NewProductHandler(services.NewProductService(db)),
		handlers.NewOrderHandler(services.NewOrderService(db), services.NewProductService(db)),
	)

	log.Printf("starting order-service on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}

func setupDatabase() *gorm.DB {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=coffee port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	if err := db.AutoMigrate(&models.Product{}, &models.Order{}, &models.LineItem{}); err != nil {
		log.Fatal("auto migration failed:", err)
	}

	if err := seed.CoffeeMenu(context.Background(), db); err != nil {
		log.Fatal("product seeding failed:", err)
	}
	return db
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
