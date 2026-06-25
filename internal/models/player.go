package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	UUID     uuid.UUID       `gorm:"type:uuid;uniqueIndex;not null;default:gen_random_uuid()"`
	Currency string          `gorm:"not null"`
	Balance  decimal.Decimal `gorm:"type:numeric(18,2);not null"`
}
