package response

import (
	"errors"
	"net/http"
	"wallet/internal/domain"
)

func WriteFromError(w http.ResponseWriter, r *http.Request, err error) {
	code, errCode, message := mapError(err)
	WriteError(w, r, code, errCode, message)
}

func mapError(err error) (httpCode int, errCode string, message string) {
	switch {
	case errors.Is(err, domain.ErrInsufficientFunds):
		return http.StatusUnprocessableEntity, "INSUFFICIENT_FUNDS", "insufficient funds for withdrawal"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "resource not found"
	case errors.Is(err, domain.ErrDepositCannotCancel):
		return http.StatusUnprocessableEntity, "DEPOSIT_CANNOT_CANCEL", "deposit transactions cannot be cancelled"
	case errors.Is(err, domain.ErrTransactionCancelled):
		return http.StatusUnprocessableEntity, "TRANSACTION_CANCELLED", "transaction is already cancelled"
	case errors.Is(err, domain.ErrIncorrectCurrency):
		return http.StatusUnprocessableEntity, "INCORRECT_CURRENCY", "incorrect player currency"
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest, "VALIDATION_ERROR", err.Error()
	default:
		return http.StatusInternalServerError, "TRANSACTION_FAILED", "internal server error"
	}
}
