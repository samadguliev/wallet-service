package service

import (
	"context"
	"testing"
	"wallet/internal/domain"
	"wallet/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ─── Моки ────────────────────────────────────────────────────────────────────

type mockPlayerRepo struct{ mock.Mock }

func (m *mockPlayerRepo) GetByUUID(ctx context.Context, playerUUID uuid.UUID) (*models.Player, error) {
	args := m.Called(ctx, playerUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Player), args.Error(1)
}
func (m *mockPlayerRepo) GetByUUIDForUpdate(ctx context.Context, tx *gorm.DB, playerUUID uuid.UUID) (*models.Player, error) {
	args := m.Called(ctx, tx, playerUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Player), args.Error(1)
}
func (m *mockPlayerRepo) GetByID(ctx context.Context, id uint) (*models.Player, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Player), args.Error(1)
}
func (m *mockPlayerRepo) GetByIDForUpdate(ctx context.Context, tx *gorm.DB, id uint) (*models.Player, error) {
	args := m.Called(ctx, tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Player), args.Error(1)
}
func (m *mockPlayerRepo) UpdateBalance(ctx context.Context, tx *gorm.DB, id uint, balance decimal.Decimal) error {
	return m.Called(ctx, tx, id, balance).Error(0)
}

type mockTransactionRepo struct{ mock.Mock }

func (m *mockTransactionRepo) Create(ctx context.Context, tx *gorm.DB, t *models.Transaction) (bool, error) {
	args := m.Called(ctx, tx, t)
	return args.Bool(0), args.Error(1)
}
func (m *mockTransactionRepo) GetByUUIDForUpdate(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*models.Transaction, error) {
	args := m.Called(ctx, tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}
func (m *mockTransactionRepo) Cancel(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	return m.Called(ctx, tx, id).Error(0)
}

// ─── Хелперы ─────────────────────────────────────────────────────────────────

func newService(t *testing.T) (*WalletService, *mockPlayerRepo, *mockTransactionRepo, *gorm.DB) {
	// Используем sqlmock для gorm.DB — нужен только чтобы пробросить Transaction()
	sqlDB, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	assert.NoError(t, err)

	playerRepo := &mockPlayerRepo{}
	txRepo := &mockTransactionRepo{}
	svc := NewWalletService(db, playerRepo, txRepo)

	return svc, playerRepo, txRepo, db
}

func newPlayer(balance float64) *models.Player {
	return &models.Player{
		UUID:     uuid.New(),
		Balance:  decimal.NewFromFloat(balance),
		Currency: "eur",
	}
}

func newApplyReq(playerID uuid.UUID, txType string, amount float64) ApplyRequest {
	return ApplyRequest{
		TransactionID: uuid.New(),
		PlayerID:      playerID,
		Type:          txType,
		Amount:        decimal.NewFromFloat(amount),
		Currency:      "eur",
	}
}

func matchDecimal(expected decimal.Decimal) interface{} {
	return mock.MatchedBy(func(actual decimal.Decimal) bool {
		return actual.Equal(expected)
	})
}

// ─── Apply: withdraw ──────────────────────────────────────────────────────────

func TestApply_Withdraw_Success(t *testing.T) {
	svc, playerRepo, txRepo, _ := newService(t)
	ctx := context.Background()

	player := newPlayer(100)
	req := newApplyReq(player.UUID, models.TransactionTypeWithdrawStr, 30)

	playerRepo.On("GetByUUID", ctx, player.UUID).Return(player, nil)
	playerRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, player.UUID).Return(player, nil)
	playerRepo.On("UpdateBalance", ctx, mock.Anything, player.ID, decimal.NewFromFloat(70)).Return(nil)
	txRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(true, nil)

	result, err := svc.Apply(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, decimal.NewFromFloat(70), result.Balance)
	playerRepo.AssertExpectations(t)
	txRepo.AssertExpectations(t)
}

func TestApply_Deposit_Success(t *testing.T) {
	svc, playerRepo, txRepo, _ := newService(t)
	ctx := context.Background()

	player := newPlayer(100)
	req := newApplyReq(player.UUID, models.TransactionTypeDepositStr, 50)

	playerRepo.On("GetByUUID", ctx, player.UUID).Return(player, nil)
	playerRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, player.UUID).Return(player, nil)
	playerRepo.On("UpdateBalance", ctx, mock.Anything, player.ID, decimal.NewFromFloat(150)).Return(nil)
	txRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(true, nil)

	result, err := svc.Apply(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, decimal.NewFromFloat(150), result.Balance)
}

// ─── Apply: insufficient funds ────────────────────────────────────────────────

func TestApply_Withdraw_InsufficientFunds(t *testing.T) {
	svc, playerRepo, txRepo, _ := newService(t)
	ctx := context.Background()

	player := newPlayer(20)
	req := newApplyReq(player.UUID, models.TransactionTypeWithdrawStr, 50)

	playerRepo.On("GetByUUID", ctx, player.UUID).Return(player, nil)
	playerRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, player.UUID).Return(player, nil)
	txRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(true, nil)

	_, err := svc.Apply(ctx, req) // результат игнорируем

	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
	playerRepo.AssertNotCalled(t, "UpdateBalance") // баланс не трогали
}

func TestApply_Withdraw_ZeroBalance(t *testing.T) {
	svc, playerRepo, txRepo, _ := newService(t)
	ctx := context.Background()

	player := newPlayer(0)
	req := newApplyReq(player.UUID, models.TransactionTypeWithdrawStr, 0.01)

	playerRepo.On("GetByUUID", ctx, player.UUID).Return(player, nil)
	playerRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, player.UUID).Return(player, nil)
	txRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(true, nil)

	_, err := svc.Apply(ctx, req)

	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestApply_Withdraw_ZeroAmount_Success(t *testing.T) {
	svc, playerRepo, txRepo, _ := newService(t)
	ctx := context.Background()

	player := newPlayer(100)
	req := newApplyReq(player.UUID, models.TransactionTypeWithdrawStr, 0)

	playerRepo.On("GetByUUID", ctx, player.UUID).Return(player, nil)
	playerRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, player.UUID).Return(player, nil)
	txRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(true, nil)
	playerRepo.On("UpdateBalance", ctx, mock.Anything, player.ID, matchDecimal(decimal.NewFromFloat(100))).Return(nil)

	result, err := svc.Apply(ctx, req)

	assert.NoError(t, err)
	assert.True(t, result.Balance.Equal(decimal.NewFromFloat(100))) // тоже через .Equal()
}

// ─── Apply: идемпотентность ───────────────────────────────────────────────────

func TestApply_Idempotency_SameTransactionID(t *testing.T) {
	svc, playerRepo, txRepo, _ := newService(t)
	ctx := context.Background()

	player := newPlayer(100)
	req := newApplyReq(player.UUID, models.TransactionTypeWithdrawStr, 30)

	playerRepo.On("GetByUUID", ctx, player.UUID).Return(player, nil)
	// Create возвращает false — транзакция уже существует
	playerRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, player.UUID).Return(player, nil)
	txRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(false, nil)

	result, err := svc.Apply(ctx, req)

	assert.NoError(t, err)
	// Баланс не изменился — вернули текущий
	assert.Equal(t, decimal.NewFromFloat(100), result.Balance)
	// UpdateBalance не вызывался
	playerRepo.AssertNotCalled(t, "UpdateBalance")
}

// ─── Apply: валюта ────────────────────────────────────────────────────────────

func TestApply_CurrencyMismatch(t *testing.T) {
	svc, playerRepo, _, _ := newService(t)
	ctx := context.Background()

	player := newPlayer(100) // currency = eur
	req := ApplyRequest{
		TransactionID: uuid.New(),
		PlayerID:      player.UUID,
		Type:          models.TransactionTypeWithdrawStr,
		Amount:        decimal.NewFromFloat(10),
		Currency:      "usd", // не совпадает
	}

	playerRepo.On("GetByUUID", ctx, player.UUID).Return(player, nil)

	_, err := svc.Apply(ctx, req)

	assert.ErrorIs(t, err, domain.ErrIncorrectCurrency)
}

// ─── Cancel ───────────────────────────────────────────────────────────────────

func TestCancel_Withdraw_Success(t *testing.T) {
	svc, playerRepo, txRepo, _ := newService(t)
	ctx := context.Background()

	player := newPlayer(70)
	tx := &models.Transaction{
		UUID:        uuid.New(),
		PlayerID:    player.ID,
		Type:        models.TransactionTypeWithdraw,
		Amount:      decimal.NewFromFloat(30),
		IsCancelled: false,
	}

	txRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, tx.UUID).Return(tx, nil)
	playerRepo.On("GetByIDForUpdate", ctx, mock.Anything, player.ID).Return(player, nil)
	txRepo.On("Cancel", ctx, mock.Anything, tx.UUID).Return(nil)
	playerRepo.On("UpdateBalance", ctx, mock.Anything, player.ID, matchDecimal(decimal.NewFromFloat(100))).Return(nil)

	result, err := svc.Cancel(ctx, tx.UUID)

	assert.NoError(t, err)
	assert.True(t, result.Balance.Equal(decimal.NewFromFloat(100)))
	txRepo.AssertExpectations(t)
	playerRepo.AssertExpectations(t)
}

func TestCancel_Deposit_NotAllowed(t *testing.T) {
	svc, _, txRepo, _ := newService(t)
	ctx := context.Background()

	tx := &models.Transaction{
		UUID:     uuid.New(),
		PlayerID: 1,
		Type:     models.TransactionTypeDeposit,
		Amount:   decimal.NewFromFloat(50),
	}

	txRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, tx.UUID).Return(tx, nil)

	_, err := svc.Cancel(ctx, tx.UUID)

	assert.ErrorIs(t, err, domain.ErrDepositCannotCancel)
}

func TestCancel_Idempotency_AlreadyCancelled(t *testing.T) {
	svc, playerRepo, txRepo, _ := newService(t)
	ctx := context.Background()

	player := newPlayer(100)
	tx := &models.Transaction{
		UUID:        uuid.New(),
		PlayerID:    player.ID,
		Type:        models.TransactionTypeWithdraw,
		Amount:      decimal.NewFromFloat(30),
		IsCancelled: true, // уже отменена
	}

	txRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, tx.UUID).Return(tx, nil)
	playerRepo.On("GetByIDForUpdate", ctx, mock.Anything, player.ID).Return(player, nil)

	result, err := svc.Cancel(ctx, tx.UUID)

	assert.NoError(t, err)
	// Баланс не изменился
	assert.Equal(t, decimal.NewFromFloat(100), result.Balance)
	// Cancel и UpdateBalance не вызывались
	txRepo.AssertNotCalled(t, "Cancel")
	playerRepo.AssertNotCalled(t, "UpdateBalance")
}

func TestCancel_NotFound(t *testing.T) {
	svc, _, txRepo, _ := newService(t)
	ctx := context.Background()

	txRepo.On("GetByUUIDForUpdate", ctx, mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)

	_, err := svc.Cancel(ctx, uuid.New())

	assert.ErrorIs(t, err, domain.ErrNotFound)
}
