package repository

import (
	"context"
	"wallet/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PlayerRepositoryI interface {
	GetByUUID(
		ctx context.Context,
		userID uuid.UUID,
	) (*models.Player, error)
	GetByID(
		ctx context.Context,
		userID uint,
	) (*models.Player, error)
	GetByUUIDForUpdate(
		ctx context.Context,
		tx *gorm.DB,
		userID uuid.UUID,
	) (*models.Player, error)
	GetByIDForUpdate(
		ctx context.Context,
		tx *gorm.DB,
		userID uint,
	) (*models.Player, error)
	UpdateBalance(ctx context.Context, tx *gorm.DB, playerID uint, newBalance decimal.Decimal) error
}

type TransactionRepositoryI interface {
	GetByUUIDForUpdate(ctx context.Context, tx *gorm.DB, transactionID uuid.UUID) (*models.Transaction, error)
	Create(ctx context.Context, tx *gorm.DB, transaction *models.Transaction) (bool, error)
	Cancel(ctx context.Context, tx *gorm.DB, transactionID uuid.UUID) error
}
