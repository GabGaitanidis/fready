package auth

import (
	"encoding/json"
	"fready/internal/auth/session"
	"fready/internal/response"
	"fready/internal/user"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	userService    user.Service
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
	if _, err := h.sessionService.CurrentUserID(r); err == nil {
		response.Error(w, http.StatusBadRequest, "Already authenticated")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid body")
		return
	}

	u, err := h.userService.GetUserByEmail(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	sess, err := h.sessionService.CreateSession(r.Context(), u.ID)
	if err != nil {
		slog.Error("failed to create session", "error", err, "user_id", u.ID)
		response.Error(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sess.ID,
		Expires:  sess.ExpiresAt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	response.Success(w, http.StatusOK, LoginResponse{
		User:      u,
		SessionID: sess.ID,
	})
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid body")
		return
	}

	u, err := h.userService.RegisterUser(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		slog.Error("failed to register user", "error", err, "email", req.Email)
		response.Error(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	sess, err := h.sessionService.CreateSession(r.Context(), u.ID)
	if err != nil {
		slog.Error("failed to create session", "error", err, "user_id", u.ID)
		response.Error(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sess.ID,
		Expires:  sess.ExpiresAt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	response.Success(w, http.StatusCreated, LoginResponse{
		User:      u,
		SessionID: sess.ID,
	})
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.sessionService.RevokeSession(r.Context(), cookie.Value); err != nil {
		slog.Error("failed to revoke session", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	sess, err := h.sessionService.GetSession(r.Context(), cookie.Value)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid or expired session")
		return
	}

	u, err := h.userService.GetUser(r.Context(), sess.UserID)
	if err != nil {
		slog.Error("failed to fetch user for session", "error", err, "user_id", sess.UserID)
		response.Error(w, http.StatusInternalServerError, "User not found")
		return
	}

	response.Success(w, http.StatusOK, u)
}