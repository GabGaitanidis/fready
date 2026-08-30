package user

import (
	"fready/internal/response"
	"log/slog"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetUsers(r.Context())
	if err != nil {
		slog.Error("failed to list users", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	response.Success(w, http.StatusOK, users)
}