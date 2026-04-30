package seed

import (
	"context"

	"github.com/osamah22/coffee-service/order-service/internal/models"
	"gorm.io/gorm"
)

var CoffeeProducts = []models.Product{
	{
		Name:         "Espresso",
		Category:     models.Hot,
		PriceInKurus: 4500,
		ImagePath:    "/images/products/espresso.jpg",
		Available:    true,
	},
	{
		Name:         "Caffe Americano",
		Category:     models.Hot,
		PriceInKurus: 5500,
		ImagePath:    "/images/products/caffe-americano.jpg",
		Available:    true,
	},
	{
		Name:         "Caffe Latte",
		Category:     models.Hot,
		PriceInKurus: 7200,
		ImagePath:    "/images/products/caffe-latte.jpg",
		Available:    true,
	},
	{
		Name:         "Cappuccino",
		Category:     models.Hot,
		PriceInKurus: 7200,
		ImagePath:    "/images/products/cappuccino.jpg",
		Available:    true,
	},
	{
		Name:         "Flat White",
		Category:     models.Hot,
		PriceInKurus: 7800,
		ImagePath:    "/images/products/flat-white.jpg",
		Available:    true,
	},
	{
		Name:         "Caramel Macchiato",
		Category:     models.Hot,
		PriceInKurus: 8800,
		ImagePath:    "/images/products/caramel-macchiato.jpg",
		Available:    true,
	},
	{
		Name:         "Caffe Mocha",
		Category:     models.Hot,
		PriceInKurus: 8600,
		ImagePath:    "/images/products/caffe-mocha.jpg",
		Available:    true,
	},
	{
		Name:         "Cold Brew",
		Category:     models.Cold,
		PriceInKurus: 7600,
		ImagePath:    "/images/products/cold-brew.jpg",
		Available:    true,
	},
	{
		Name:         "Iced Latte",
		Category:     models.Cold,
		PriceInKurus: 7600,
		ImagePath:    "/images/products/iced-latte.jpg",
		Available:    true,
	},
	{
		Name:         "Iced Caramel Macchiato",
		Category:     models.Cold,
		PriceInKurus: 9200,
		ImagePath:    "/images/products/iced-caramel-macchiato.jpg",
		Available:    true,
	},
}

func CoffeeMenu(ctx context.Context, db *gorm.DB) error {
	for _, product := range CoffeeProducts {
		var existing models.Product
		result := db.WithContext(ctx).
			Where(&models.Product{Name: product.Name}).
			Attrs(product).
			FirstOrCreate(&existing)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
