package domain

import "errors"

var (
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrNotFound             = errors.New("not found")
	ErrDepositCannotCancel  = errors.New("deposit cannot be cancelled")
	ErrTransactionCancelled = errors.New("transaction already cancelled")
	ErrIncorrectCurrency    = errors.New("incorrect currency")
	ErrValidation           = errors.New("validation error")
)

// ValidationError позволяет прокинуть конкретное сообщение из валидации
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}
