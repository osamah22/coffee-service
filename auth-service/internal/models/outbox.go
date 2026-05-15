package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxEvent struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	EventType     string     `gorm:"not null;index"`
	AggregateType string     `gorm:"not null"`
	AggregateID   string     `gorm:"not null;index"`
	RoutingKey    string     `gorm:"not null"`
	Payload       string     `gorm:"type:text;not null"`
	Attempts      int        `gorm:"not null;default:0"`
	LastError     string     `gorm:"type:text"`
	OccurredAt    time.Time  `gorm:"not null;index"`
	PublishedAt   *time.Time `gorm:"index"`
	CreatedAt     time.Time
}

func (e *OutboxEvent) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	return nil
}
