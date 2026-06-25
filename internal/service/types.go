package service

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ApplyRequest struct {
	TransactionID uuid.UUID
	PlayerID      uuid.UUID
	Type          string
	Amount        decimal.Decimal
	Currency      string
}

type ApplyResult struct {
	Balance decimal.Decimal `json:"balance"`
}

type BalanceResult struct {
	Balance decimal.Decimal `json:"balance"`
}

type CancelResult struct {
	Balance decimal.Decimal `json:"balance"`
}
