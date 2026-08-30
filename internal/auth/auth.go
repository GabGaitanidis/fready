package auth

import (
	"encoding/json"
	"fready/internal/auth/session"
	"fready/internal/user"
	"net/http"
	"time"
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
    _, err := h.sessionService.CurrentUserID(r)
    if err == nil {
        http.Error(w, "Already authenticated", http.StatusBadRequest)
        return
    }
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
		Secure:   false,
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


type RegisterRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid body", http.StatusBadRequest)
        return
    }

    u, err := h.userService.RegisterUser(r.Context(), req.Name, req.Email, req.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
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
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        Path:     "/",
    })


    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(LoginResponse{
        User:      u,
        SessionID: session.ID,
    })
}


func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_id")
    if err != nil {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    if err := h.sessionService.RevokeSession(r.Context(), cookie.Value); err != nil {
        http.Error(w, "Failed to log out", http.StatusInternalServerError)
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
        http.Error(w, "Not authenticated", http.StatusUnauthorized)
        return
    }

    sess, err := h.sessionService.GetSession(r.Context(), cookie.Value)
    if err != nil {
        http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
        return
    }

    u, err := h.userService.GetUser(r.Context(), sess.UserID)
    if err != nil {
        http.Error(w, "User not found", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(u)
}