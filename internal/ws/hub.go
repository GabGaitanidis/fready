package ws

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	ReadBufferSize  int = 1024
	WriteBufferSize int = 1024
)

type registration struct {
	conn     *connection
	groupIDs []uuid.UUID
}

type Hub struct {
	sync.Mutex

	connections map[*connection]uuid.UUID          
	groups      map[uuid.UUID]map[*connection]bool

	wsConnFactory websocket.Upgrader
	Broadcast chan GroupBroadcast
	register   chan registration
	unregister chan *connection
	IncomingLocation chan LocationEvent
}

type GroupBroadcast struct {
	GroupID       uuid.UUID
	ExcludeUserID uuid.UUID
	Payload       []byte
}

func New(allowedOrigins []string) *Hub {
	return &Hub{
		connections: make(map[*connection]uuid.UUID),
		groups:      make(map[uuid.UUID]map[*connection]bool),
		register:    make(chan registration),
		unregister:  make(chan *connection),
		Broadcast:   make(chan GroupBroadcast),
		IncomingLocation: make(chan LocationEvent),
		wsConnFactory: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				allowedOrigins := []string{
					"http://localhost:8443", 
				}
				for _, allowed := range allowedOrigins {
					if allowed == origin {
						return true
					}
				}
				return false
			},
		},
	}
}
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, userID uuid.UUID, groupIDs []uuid.UUID) {
	ws, err := h.wsConnFactory.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	} else {
		slog.Info("Upgraded to websockets!!!")
	}

	c := &connection{
			send:     make(chan []byte, 256),
			ws:       ws,
			hub:      h,
			userID:   userID,
			groupIDs: groupIDs,
		}

	h.register <- registration{
		conn:     c,
		groupIDs: groupIDs,
	}
	go c.listenWrite()
	c.listenRead()
}
func (h *Hub) doRegister(c *connection, groupIDs []uuid.UUID) {
	h.Lock()
	defer h.Unlock()

	h.connections[c] = c.userID

	for _, groupID := range groupIDs {
		if h.groups[groupID] == nil {
			h.groups[groupID] = make(map[*connection]bool)
		} 
		h.groups[groupID][c] = true
	}
}

func (h *Hub) doUnregister(c *connection) {
	h.Lock()
	defer h.Unlock()

	delete(h.connections, c)
	for _, groupID := range c.groupIDs {
        if conns, ok := h.groups[groupID]; ok {
            delete(conns, c)
            if len(conns) == 0 {
                delete(h.groups, groupID)
            }
        }
    }
	c.close()

}

func (h *Hub) DoBroadcast(b GroupBroadcast) {
	h.Lock()
	defer h.Unlock()	
	conns, ok := h.groups[b.GroupID]
	if !ok {
		return
	}

	for c := range conns {
		if c.userID == b.ExcludeUserID {
			continue
		}
		select {
		case c.send <- b.Payload:
		default:
			slog.Warn("dropping message, send buffer full", "user_id", c.userID)
		}
	}
}
func (h *Hub) Run() {
	for {
		select {
		case reg := <-h.register:
			h.doRegister(reg.conn, reg.groupIDs)
		case c := <-h.unregister:
			h.doUnregister(c)
		case b := <-h.Broadcast:
			h.DoBroadcast(b)
		}
	}
}