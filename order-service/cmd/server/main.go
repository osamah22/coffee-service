package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(sharedauth.CORS(authConfig))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	authMiddleware := sharedauth.NewMiddleware(authConfig)
	authHandlers := sharedauth.NewHandlerSet(authConfig)

	addRoutes(
		router,
		authMiddleware,
		authHandlers,
		handlers.NewProductHandler(services.NewProductService(db)),
		handlers.NewOrderHandler(services.NewOrderService(db), services.NewProductService(db)),
	)

	startOutboxDispatcher(db)

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

	if err := db.AutoMigrate(&models.Product{}, &models.Order{}, &models.LineItem{}, &models.OutboxEvent{}); err != nil {
		log.Fatal("auto migration failed:", err)
	}

	if err := seed.CoffeeMenu(context.Background(), db); err != nil {
		log.Fatal("product seeding failed:", err)
	}
	return db
}

func startOutboxDispatcher(db *gorm.DB) {
	rabbitURL := envOrDefault("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/")
	dispatcher := services.NewOutboxDispatcher(db, rabbitURL, 2*time.Second, 10)
	go dispatcher.Run(context.Background())
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
