package auth

import (
	"sync"
	"time"
)

type Session struct {
	UserId string
	CreatedAt time.Time
}

type SessionManager struct {
	mu sync.RWMutex
	sessions map[string]Session
}