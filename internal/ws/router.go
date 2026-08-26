package ws

import (
	"fready/internal/auth/session"
	"net/http"
)

type Router struct {}

func NewRouter() *Router {
    return &Router{}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux, hub *Hub,s session.Service) {
	http.HandleFunc("/ws", HandleWS(hub, s))
}