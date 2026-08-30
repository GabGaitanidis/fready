package group

import (
	"database/sql"
	"fready/internal/auth/session"
	"net/http"
)

type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux, db *sql.DB, sessionService session.Service) Service {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service, sessionService)

	mux.HandleFunc("POST /groups", handler.CreateGroup)
	mux.HandleFunc("GET /groups", handler.ListMyGroups)
	mux.HandleFunc("GET /groups/{id}", handler.GetGroup)
	mux.HandleFunc("GET /groups/{id}/members", handler.ListMembers)
	mux.HandleFunc("POST /groups/{id}/members", handler.AddMember)
	mux.HandleFunc("DELETE /groups/{id}/members/{userId}", handler.RemoveMember)

	return service
}