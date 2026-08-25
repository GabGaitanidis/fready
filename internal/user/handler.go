package user

import (
	"encoding/json"
	"fready/internal/auth/session"
	"net/http"
)

type Handler struct {
	session session.Service
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	u, err := h.service.GetUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_id")
    if err != nil {
        http.Error(w, "Not authenticated", http.StatusUnauthorized)
        return
    }

    sess, err := h.session.GetSession(r.Context(), cookie.Value)
    if err != nil {
        http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
        return
    }

    u, err := h.service.GetUser(r.Context(), sess.UserID)
    if err != nil {
        http.Error(w, "User not found", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(u)
}


type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	u, err := h.service.RegisterUser(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}
