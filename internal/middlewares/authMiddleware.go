package middlewares

import (
	"context"
	"fready/internal/auth/session"
	"net/http"
)

func AuthMiddleware(next http.Handler, s session.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login", "/register", "/health":
			next.ServeHTTP(w, r)
			return
		}

		userID, err := s.CurrentUserID(r)
		if err != nil {
			_ = s.Logout(w, r)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}