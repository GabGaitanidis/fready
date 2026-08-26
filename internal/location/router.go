package location

import (
	"database/sql"
	"net/http"
)

type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux, db *sql.DB, sessionService SessionService) Service {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service, sessionService)

	mux.HandleFunc("POST /location", handler.UpdateLocation)

	return service
}