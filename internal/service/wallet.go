package service

import (
	"context"
	"errors"
	"wallet/internal/domain"
	"wallet/internal/models"
	"wallet/internal/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletService struct {
	db              *gorm.DB
	playerRepo      repository.PlayerRepositoryI
	transactionRepo repository.TransactionRepositoryI
}

func NewWalletService(
	db *gorm.DB,
	playerRepo repository.PlayerRepositoryI,
	txRepo repository.TransactionRepositoryI,
) *WalletService {
	return &WalletService{
		db:              db,
		playerRepo:      playerRepo,
		transactionRepo: txRepo,
	}
}

func (s *WalletService) GetBalance(ctx context.Context, playerUUID uuid.UUID) (BalanceResult, error) {
	player, err := s.playerRepo.GetByUUID(ctx, playerUUID)
	if err != nil {
		return BalanceResult{}, err
	}
	return BalanceResult{
		Balance: player.Balance,
	}, nil
}

func (s *WalletService) Apply(ctx context.Context, req ApplyRequest) (ApplyResult, error) {
	var result ApplyResult

	player, err := s.playerRepo.GetByUUID(ctx, req.PlayerID)
	if err != nil {
		return result, err
	}

	// смена валюты не предусмотрена, безопасная проверка
	if player.Currency != req.Currency {
		return result, domain.ErrIncorrectCurrency
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		created, err := s.transactionRepo.Create(ctx, tx, &models.Transaction{
			UUID:     req.TransactionID,
			PlayerID: player.ID,
			Type:     models.GetType(req.Type),
			Amount:   req.Amount,
			Currency: req.Currency,
		})
		if err != nil {
			return err
		}

		player, err := s.playerRepo.GetByUUIDForUpdate(ctx, tx, req.PlayerID)
		if err != nil {
			return err
		}

		if !created {
			result.Balance = player.Balance
			return nil
		}

		var newBalance decimal.Decimal
		if req.Type == models.TransactionTypeWithdrawStr {
			if player.Balance.LessThan(req.Amount) {
				return domain.ErrInsufficientFunds
			}
			newBalance = player.Balance.Sub(req.Amount)
		} else {
			newBalance = player.Balance.Add(req.Amount)
		}

		if err := s.playerRepo.UpdateBalance(ctx, tx, player.ID, newBalance); err != nil {
			return err
		}

		result.Balance = newBalance
		return nil
	})

	if err != nil {
		return ApplyResult{}, err
	}

	return result, nil
}

func (s *WalletService) Cancel(ctx context.Context, transactionID uuid.UUID) (*CancelResult, error) {
	var result CancelResult

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing, err := s.transactionRepo.GetByUUIDForUpdate(ctx, tx, transactionID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		if existing.Type == models.TransactionTypeDeposit {
			return domain.ErrDepositCannotCancel
		}

		player, err := s.playerRepo.GetByIDForUpdate(ctx, tx, existing.PlayerID)
		if err != nil {
			return err
		}

		if existing.IsCancelled {
			result.Balance = player.Balance
			return nil
		}

		newBalance := player.Balance.Add(existing.Amount)

		if err := s.transactionRepo.Cancel(ctx, tx, transactionID); err != nil {
			return err
		}

		if err := s.playerRepo.UpdateBalance(ctx, tx, player.ID, newBalance); err != nil {
			return err
		}

		result.Balance = newBalance
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}
