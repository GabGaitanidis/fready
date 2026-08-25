package auth

import (
	"fready/internal/auth/session"
	"fready/internal/user"
	"net/http"
)

type Router struct{}

func NewRouter() *Router {
    return &Router{}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux, userService user.Service, sessionService session.Service) {
    authHandler := NewHandler(userService, sessionService)
    mux.HandleFunc("POST /login", authHandler.LoginHandler)
	mux.HandleFunc("POST /register", authHandler.RegisterHandler)
	mux.HandleFunc("POST /logout", authHandler.LogoutHandler)
	mux.HandleFunc("GET /me", authHandler.GetMe)
}