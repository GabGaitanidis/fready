package location

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

type Service interface {
	UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) (*LocationPing, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) (*LocationPing, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID is required")
	}
	if lat < -90 || lat > 90 {
		return nil, errors.New("invalid latitude")
	}
	if lng < -180 || lng > 180 {
		return nil, errors.New("invalid longitude")
	}

	ping, err := s.repo.Upsert(ctx, userID, lat, lng)
	if err != nil {
		slog.Error("failed to upsert location", "error", err, "user_id", userID)
		return nil, err
	}

	return ping, nil
}