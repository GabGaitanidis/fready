package ws

import (
	"fready/internal/auth/session"
	"fready/internal/group"
	"net/http"

	"github.com/google/uuid"
)

func HandleWS(hub *Hub, sessionSvc session.Service, groupService group.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := sessionSvc.CurrentUserID(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		groups, err := groupService.ListMyGroups(r.Context(), userID)
		if err != nil {
			http.Error(w, "Failed to retrieve groups", http.StatusInternalServerError)
			return
		}

		groupIDs := make([]uuid.UUID, 0, len(groups))
		for i := range groups {
			groupIDs = append(groupIDs, groups[i].ID)
		}

		hub.ServeHTTP(w, r, userID, groupIDs)
	}
}