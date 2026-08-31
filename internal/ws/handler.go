package ws

import (
	"fready/internal/auth/session"
	"fready/internal/group"
	"fready/internal/response"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func HandleWS(hub *Hub, sessionSvc session.Service, groupService group.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := sessionSvc.CurrentUserID(r)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "Not authenticated")
			return
		}

		groups, err := groupService.ListMyGroups(r.Context(), userID)
		if err != nil {
			slog.Error("failed to list groups for websocket connection", "error", err, "user_id", userID)
			response.Error(w, http.StatusInternalServerError, "Failed to retrieve groups")
			return
		}

		groupIDs := make([]uuid.UUID, 0, len(groups))
		for i := range groups {
			groupIDs = append(groupIDs, groups[i].ID)
		}

		hub.ServeHTTP(w, r, userID, groupIDs)
	}
}