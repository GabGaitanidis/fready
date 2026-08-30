package location

import (
	"context"
	"encoding/json"
	"errors"
	"fready/internal/ws"
	"log/slog"

	"github.com/google/uuid"
)

type Service interface {
	UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lon float64) (*LocationPing, error)
	StartConsumer(ctx context.Context, incoming <-chan ws.LocationEvent, hub *ws.Hub)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lon float64) (*LocationPing, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID is required")
	}
	if lat < -90 || lat > 90 {
		return nil, errors.New("invalid latitude")
	}
	if lon < -180 || lon > 180 {
		return nil, errors.New("invalid longitude")
	}

	ping, err := s.repo.Upsert(ctx, userID, lat, lon)
	if err != nil {
		slog.Error("failed to upsert location", "error", err, "user_id", userID)
		return nil, err
	}

	return ping, nil
}

func (s *service) StartConsumer(ctx context.Context, incoming <-chan ws.LocationEvent, hub *ws.Hub) {
    go func() {
        for event := range incoming {
            ping, err := s.UpdateLocation(ctx, event.UserID, event.Update.Lat, event.Update.Lon)
            if err != nil {
                slog.Error("consumer failed to update location", "error", err, "user_id", event.UserID)
                continue
            }
            
            payload, _ := json.Marshal(ping)

            for _, groupID := range event.GroupIDs {
                hub.Broadcast <- ws.GroupBroadcast{
                    GroupID:       groupID,
                    ExcludeUserID: event.UserID,
                    Payload:       payload,
                }
            }
        }
    }()
}