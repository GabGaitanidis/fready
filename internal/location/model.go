package location

import (
	"time"

	"github.com/google/uuid"
)

type LocationPing struct {
	UserID    uuid.UUID `json:"user_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	UpdatedAt time.Time `json:"updated_at"`
}