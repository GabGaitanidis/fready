package location

import (
	"encoding/json"
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
	Lng float64 `json:"lon"`
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionService.CurrentUserID(r)
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var req updateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	ping, err := h.service.UpdateLocation(r.Context(), userID, req.Lat, req.Lng)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ping)
}