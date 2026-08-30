package location

import (
	"encoding/json"
	"fready/internal/response"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	service        Service
	sessionService SessionService
}

type SessionService interface {
	CurrentUserID(r *http.Request) (uuid.UUID, error)
}

func NewHandler(s Service, ss SessionService) *Handler {
	return &Handler{service: s, sessionService: ss}
}

type updateLocationRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionService.CurrentUserID(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req updateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid body")
		return
	}

	ping, err := h.service.UpdateLocation(r.Context(), userID, req.Lat, req.Lng)
	if err != nil {
		slog.Warn("location update rejected", "error", err, "user_id", userID)
		response.Error(w, http.StatusBadRequest, "Invalid location update")
		return
	}

	response.Success(w, http.StatusOK, ping)
}