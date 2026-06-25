package repository

import (
	"context"
	"errors"
	"wallet/internal/domain"
	"wallet/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlayerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) *PlayerRepository {
	return &PlayerRepository{
		db: db,
	}
}

func (r *PlayerRepository) GetByUUID(ctx context.Context, playerUUID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.WithContext(ctx).Where("uuid = ?", playerUUID).First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepository) GetByID(ctx context.Context, playerID uint) (*models.Player, error) {
	var player models.Player
	err := r.db.WithContext(ctx).Where("id = ?", playerID).First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepository) GetByUUIDForUpdate(ctx context.Context, tx *gorm.DB, playerUUID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("uuid = ?", playerUUID).
		First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepository) GetByIDForUpdate(ctx context.Context, tx *gorm.DB, playerID uint) (*models.Player, error) {
	var player models.Player
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", playerID).
		First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepository) UpdateBalance(ctx context.Context, tx *gorm.DB, playerID uint, newBalance decimal.Decimal) error {
	result := tx.WithContext(ctx).
		Model(&models.Player{}).
		Where("id = ?", playerID).
		Update("balance", newBalance)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
