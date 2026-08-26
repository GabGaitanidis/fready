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
	broadcast chan groupBroadcast
	register   chan registration
	unregister chan *connection
	incomingLocation chan locationEvent
}

type groupBroadcast struct {
	groupID       uuid.UUID
	excludeUserID uuid.UUID
	payload       []byte
}

func New() *Hub {
	return &Hub{
		connections: make(map[*connection]uuid.UUID),
		groups:      make(map[uuid.UUID]map[*connection]bool),
		register:    make(chan registration),
		unregister:  make(chan *connection),
		broadcast:   make(chan groupBroadcast),
		incomingLocation: make(chan locationEvent),
		wsConnFactory: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, userID uuid.UUID, groupIDs []uuid.UUID) {
	ws, err := h.wsConnFactory.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	c := &connection{
		send:   make(chan []byte, 256),
		ws:     ws,
		hub:    h,
		userID: userID,
	}

	h.register <- registration{conn: c, groupIDs: groupIDs}

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

	for groupID, conns := range h.groups {
		if ok := conns[c]; ok {
			delete(conns, c)
			if len(conns) == 0 {
				delete(h.groups, groupID)
			}
		}
	}


}

func (h *Hub) doBroadcast(b groupBroadcast) {
	h.Lock()
	defer h.Unlock()

	conns, ok := h.groups[b.groupID]
	if !ok {
		return
	}

	for c := range conns {
		if c.userID == b.excludeUserID {
			continue
		}
		select {
		case c.send <- b.payload:
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
		case b := <-h.broadcast:
			h.doBroadcast(b)
		}
	}
}