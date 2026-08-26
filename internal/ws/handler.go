package ws

import (
	"fready/internal/auth/session"
	"net/http"

	"github.com/google/uuid"
)

func HandleWS(hub *Hub, sessionSvc session.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := sessionSvc.CurrentUserID(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		groupIDs := []uuid.UUID{}

		hub.ServeHTTP(w, r, userID, groupIDs)
	}
}