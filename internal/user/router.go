package user

import (
	"database/sql"
	"net/http"
)

type Router struct{}

func NewRouter() *Router {
    return &Router{}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux, db *sql.DB) Service {
    userRepo := NewRepository(db)
    userService := NewService(userRepo)
    userHandler := NewHandler(userService)

    mux.HandleFunc("GET /users", userHandler.GetUsers)

    return userService
}