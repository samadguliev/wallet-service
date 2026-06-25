package router

import (
	"net/http"
	"time"
	"wallet/internal/config"
	"wallet/internal/delivery/http/handler"
	"wallet/internal/delivery/http/middleware"

	"github.com/go-chi/chi/v5"
)

func New(transactionHandler *handler.TransactionHandler, playerHandler *handler.PlayerHandler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.TimeoutMiddleware(10 * time.Second))
	r.Use(middleware.Auth(config.Config.AuthToken))
	r.Get("/players/{id}/balance", playerHandler.GetBalance)
	r.Post(
		"/transactions",
		transactionHandler.Apply,
	)
	r.Delete("/transactions/{transactionId}", transactionHandler.Cancel)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	return r

}
