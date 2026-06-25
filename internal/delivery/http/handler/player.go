package handler

import (
	"net/http"
	"wallet/internal/delivery/http/response"
	"wallet/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PlayerHandler struct {
	service *service.WalletService
}

func NewPlayerHandler(s *service.WalletService) *PlayerHandler {
	return &PlayerHandler{service: s}
}

func (h *PlayerHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	playerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid player id")
		return
	}

	result, err := h.service.GetBalance(r.Context(), playerID)
	if err != nil {
		response.WriteFromError(w, r, err)
		return
	}

	response.WriteOK(w, r, result)
}
