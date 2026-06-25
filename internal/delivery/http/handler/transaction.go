package handler

import (
	"encoding/json"
	"net/http"
	"wallet/internal/delivery/http/response"
	"wallet/internal/domain"
	"wallet/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionHandler struct {
	service *service.WalletService
}

func NewTransactionHandler(s *service.WalletService) *TransactionHandler {
	return &TransactionHandler{service: s}
}

type applyRequest struct {
	TransactionID uuid.UUID       `json:"transactionId"`
	PlayerID      uuid.UUID       `json:"playerId"`
	Type          string          `json:"type"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
}

func (h *TransactionHandler) Apply(w http.ResponseWriter, r *http.Request) {
	var req applyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	if err := validateApplyRequest(req); err != nil {
		response.WriteFromError(w, r, err)
		return
	}

	result, err := h.service.Apply(r.Context(), service.ApplyRequest{
		TransactionID: req.TransactionID,
		PlayerID:      req.PlayerID,
		Type:          req.Type,
		Amount:        req.Amount,
		Currency:      req.Currency,
	})
	if err != nil {
		response.WriteFromError(w, r, err)
		return
	}

	response.WriteOK(w, r, result)
}

func (h *TransactionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	transactionID, err := uuid.Parse(chi.URLParam(r, "transactionId"))
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid transactionId")
		return
	}

	result, err := h.service.Cancel(r.Context(), transactionID)
	if err != nil {
		response.WriteFromError(w, r, err)
		return
	}

	response.WriteOK(w, r, result)
}

func validateApplyRequest(req applyRequest) error {
	if req.TransactionID == uuid.Nil {
		return &domain.ValidationError{Field: "transactionId", Message: "required"}
	}
	if req.PlayerID == uuid.Nil {
		return &domain.ValidationError{Field: "playerId", Message: "required"}
	}
	if req.Type != "withdraw" && req.Type != "deposit" {
		return &domain.ValidationError{Field: "type", Message: "must be withdraw or deposit"}
	}
	if req.Amount.IsNegative() {
		return &domain.ValidationError{Field: "amount", Message: "must be non-negative"}
	}
	if req.Currency == "" {
		return &domain.ValidationError{Field: "currency", Message: "required"}
	}
	return nil
}
