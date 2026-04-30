package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/osamah22/coffee-service/order-service/internal/handlers"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"github.com/osamah22/coffee-service/order-service/internal/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	port := envOrDefault("PORT", "8080")
	db := setupDatabase()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	addRoutes(
		router,
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

	return db
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
