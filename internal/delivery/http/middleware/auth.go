package middleware

import (
	"fmt"
	"net/http"
	"wallet/internal/delivery/http/response"
)

func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				response.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
				return
			}

			if header != token {
				fmt.Println("header", header)
				fmt.Println("token", token)
				response.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
