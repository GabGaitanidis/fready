package auth

import (
	"encoding/json"
	"fready/internal/auth/session"
	"fready/internal/user"
	"net/http"
)


type Handler struct {
	userService user.Service
	sessionService session.Service
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User      *user.User `json:"user"`
	SessionID string     `json:"session_id"`
}

func NewHandler(su user.Service, ss session.Service) *Handler {
	return &Handler{userService: su, sessionService: ss}
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	u, err := h.userService.GetUserByEmail(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized) 
		return 
	}

	session, err := h.sessionService.CreateSession(r.Context(), u.ID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		User:      u,
		SessionID: session.ID,
	})
}