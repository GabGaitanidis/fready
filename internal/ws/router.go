package ws

import (
	"fready/internal/auth/session"
	"fready/internal/group"
	"net/http"
)

type Router struct {}

func NewRouter() *Router {
    return &Router{}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux, hub *Hub,s session.Service, groupService group.Service) {
	mux.HandleFunc("GET /ws", HandleWS(hub, s, groupService))
}