package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/websocket"
)

var (
	WriteWait = 10 * time.Second
	PongWait = 60 * time.Second
	PingPeriod = (PongWait * 9) / 10
	MaxMessageSize int64 = 64 * 1024
)


type connection struct {
    ws     *websocket.Conn
    send   chan []byte
    userID uuid.UUID
    hub    *Hub
    once   sync.Once 
	groupIDs []uuid.UUID
}
type LocationEvent struct {
	UserID uuid.UUID
	GroupIDs []uuid.UUID
	Update LocationUpdate
}

type LocationUpdate struct {
    Lat float64 `json:"lat"`
    Lon float64 `json:"lon"` 
}

func (c *connection) close() {
    c.once.Do(func() {
        if err := c.ws.Close(); err != nil {
            slog.Warn("websocket close error", "error", err, "user_id", c.userID)
        }
        close(c.send)
    })
}
func (c *connection) listenRead() {
	defer func() {
		c.hub.unregister <- c
	}()

	c.ws.SetReadLimit(MaxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(PongWait))
	c.ws.SetPongHandler(func(string) error {
		return c.ws.SetReadDeadline(time.Now().Add(PongWait))
	})

	for {
		_, message, err := c.ws.ReadMessage()
		slog.Info("Message red")
		if err != nil {
			slog.Info("websocket read closed", "user_id", c.userID, "error", err)
			break
		}

		var update LocationUpdate
		if err := json.Unmarshal(message, &update); err != nil {
			slog.Warn("invalid location update payload", "user_id", c.userID)
			continue	
		}

		c.hub.IncomingLocation <- LocationEvent{UserID: c.userID, Update: update}
	}
}

func (c *connection) listenWrite() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.ws.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				slog.Warn("websocket write failed", "user_id", c.userID, "error", err)
				return
			}
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}