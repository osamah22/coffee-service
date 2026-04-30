package models

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category string

const (
	Cold Category = "cold"
	Hot  Category = "hot"
)

type Product struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Category     Category  `gorm:"type:varchar(10);not null" json:"category"`
	PriceInKurus int64     `gorm:"not null" json:"price_in_kurus"` // 2550 = TL 25.50
	Available    bool      `gorm:"default:true" json:"available"`
}

// BeforeCreate  automatically runs when inserting new record to the database
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	// create uuid if not initialized
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ValidateProduct ensure that product is in valid state
func ValidateProduct(p *Product) error {
	if p.Name == "" {
		return errors.New("name is required")
	}

	if p.PriceInKurus < 0 {
		return errors.New("price cannot be negative")
	}

	if p.Category != Cold && p.Category != Hot {
		return errors.New("invalid category")
	}

	return nil
}
