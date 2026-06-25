package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	TransactionTypeDeposit  uint8 = 0
	TransactionTypeWithdraw uint8 = 1

	TransactionTypeDepositStr  = "deposit"
	TransactionTypeWithdrawStr = "withdraw"
)

type Transaction struct {
	gorm.Model
	UUID        uuid.UUID       `gorm:"type:uuid;uniqueIndex"`
	PlayerID    uint            `gorm:"index"`
	Player      *Player         `gorm:"->;foreignKey:PlayerID" json:"-"`
	Type        uint8           `gorm:"not null"`
	Amount      decimal.Decimal `gorm:"type:numeric(18,2);not null"`
	Currency    string          `gorm:"not null"`
	IsCancelled bool            `gorm:"default:false"`
	CancelledAt *time.Time
}

func GetType(strType string) uint8 {
	if strType == "deposit" {
		return TransactionTypeDeposit
	}
	return TransactionTypeWithdraw
}
