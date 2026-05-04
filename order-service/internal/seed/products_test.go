package seed

import (
	"context"
	"testing"

	"github.com/osamah22/coffee-service/order-service/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCoffeeMenuSeedsProducts(t *testing.T) {
	db := newTestDB(t)

	if err := CoffeeMenu(context.Background(), db); err != nil {
		t.Fatalf("expected seeding to succeed, got %v", err)
	}

	var products []models.Product
	if err := db.Order("name").Find(&products).Error; err != nil {
		t.Fatalf("expected products query to succeed, got %v", err)
	}

	if len(products) != len(CoffeeProducts) {
		t.Fatalf("expected %d products, got %d", len(CoffeeProducts), len(products))
	}

	for _, product := range products {
		if product.ImagePath == "" {
			t.Fatalf("expected product %q to have an image path", product.Name)
		}
		if product.PriceInKurus < 12000 {
			t.Fatalf("expected product %q to cost at least 120 TL, got %d kurus", product.Name, product.PriceInKurus)
		}
	}
}

func TestCoffeeMenuIsIdempotent(t *testing.T) {
	db := newTestDB(t)

	if err := CoffeeMenu(context.Background(), db); err != nil {
		t.Fatalf("expected first seed to succeed, got %v", err)
	}
	if err := CoffeeMenu(context.Background(), db); err != nil {
		t.Fatalf("expected second seed to succeed, got %v", err)
	}

	var count int64
	if err := db.Model(&models.Product{}).Count(&count).Error; err != nil {
		t.Fatalf("expected count query to succeed, got %v", err)
	}

	if count != int64(len(CoffeeProducts)) {
		t.Fatalf("expected %d products after reseed, got %d", len(CoffeeProducts), count)
	}
}

func TestCoffeeMenuUpdatesExistingProducts(t *testing.T) {
	db := newTestDB(t)

	if err := db.Create(&models.Product{
		Name:         "Espresso",
		Category:     models.Hot,
		PriceInKurus: 8000,
		ImagePath:    "/old.jpg",
		Available:    true,
	}).Error; err != nil {
		t.Fatalf("expected stale product insert to succeed, got %v", err)
	}

	if err := CoffeeMenu(context.Background(), db); err != nil {
		t.Fatalf("expected seed update to succeed, got %v", err)
	}

	var product models.Product
	if err := db.Where("name = ?", "Espresso").First(&product).Error; err != nil {
		t.Fatalf("expected espresso query to succeed, got %v", err)
	}

	if product.PriceInKurus != 12000 {
		t.Fatalf("expected espresso price to be updated to 12000 kurus, got %d", product.PriceInKurus)
	}
	if product.ImagePath != "/images/products/espresso.jpg" {
		t.Fatalf("expected espresso image path to be refreshed, got %q", product.ImagePath)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&models.Product{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}
