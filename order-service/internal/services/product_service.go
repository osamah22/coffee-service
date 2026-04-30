package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/osamah22/coffee-service/order-service/internal/models"
	"gorm.io/gorm"
)

const productNotFound = "product_not_found"

var ErrProductNotFound = errors.New(productNotFound)

type ProductService struct {
	DB *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{DB: db}
}

func (svc *ProductService) ListAll(
	ctx context.Context,
) ([]models.Product, error) {
	var products []models.Product

	if err := svc.DB.WithContext(ctx).
		Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (svc *ProductService) Find(
	ctx context.Context,
	id uuid.UUID,
) (*models.Product, error) {
	var product models.Product
	tx := svc.DB.WithContext(ctx).
		Where("id = ?", id).
		First(&product)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, ErrProductNotFound
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &product, nil
}

func (svc *ProductService) AddProduct(
	ctx context.Context,
	product *models.Product,
) (*models.Product, error) {
	if err := models.ValidateProduct(product); err != nil {
		return nil, err
	}

	tx := svc.DB.WithContext(ctx).Create(product)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return product, nil
}

func (svc *ProductService) UpdateProduct(
	ctx context.Context,
	product *models.Product,
) (*models.Product, error) {
	if err := models.ValidateProduct(product); err != nil {
		return nil, err
	}

	var existing models.Product
	err := svc.DB.WithContext(ctx).
		First(&existing, "id = ?", product.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	if err := svc.DB.WithContext(ctx).Save(product).Error; err != nil {
		return nil, err
	}

	return product, nil
}

func (svc *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	result := svc.DB.WithContext(ctx).
		Delete(&models.Product{}, "id = ?", id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}
