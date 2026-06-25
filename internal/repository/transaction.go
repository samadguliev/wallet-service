package repository

import (
	"context"
	"errors"
	"wallet/internal/domain"
	"wallet/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

func (r *TransactionRepository) GetByUUIDForUpdate(
	ctx context.Context,
	tx *gorm.DB, transactionUUID uuid.UUID,
) (*models.Transaction, error) {
	var trx models.Transaction
	err := tx.WithContext(ctx).
		Model(&models.Transaction{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("transactions.uuid = ?", transactionUUID).
		First(&trx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &trx, nil
}

func (r *TransactionRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	t *models.Transaction,
) (bool, error) {
	res := tx.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "uuid"},
			},
			DoNothing: true,
		}).
		Create(t)

	if res.Error != nil {
		return false, res.Error
	}

	return res.RowsAffected == 1, nil
}

func (r *TransactionRepository) Cancel(ctx context.Context, tx *gorm.DB, transactionID uuid.UUID) error {
	result := tx.WithContext(ctx).
		Model(&models.Transaction{}).
		Where("uuid = ? AND NOT is_cancelled", transactionID).
		Update("is_cancelled", true)

	if result.Error != nil {
		return result.Error
	}
	return nil
}
