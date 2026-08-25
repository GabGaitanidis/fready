package session

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateSession(ctx context.Context, userID uuid.UUID) (*Session, error)
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

type service struct {
	repo SessionRepository
}

func NewService(repo SessionRepository) Service {
	return &service{repo: repo}
}

func (s *service) CreateSession(ctx context.Context, userID uuid.UUID) (*Session, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID is required")
	}

	sessionID, err := GenerateSessionId()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	err = s.repo.CreateSession(ctx, sessionID, userID, expiresAt)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *service) GetSession(ctx context.Context, sessionID string) (*Session, error) {
    if sessionID == "" {
        return nil, errors.New("session ID is required")
    }

    sess, err := s.repo.GetSessionByID(ctx, sessionID)
    if err != nil {
        return nil, errors.New("invalid or expired session")
    }

    if CheckSessionLife(sess) {
        return nil, errors.New("session expired")
    }

    return sess, nil
}

func (s *service) RevokeSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	return s.repo.DeleteSession(ctx, sessionID)
}

func CheckSessionLife(s *Session) bool {
	return time.Now().After(s.ExpiresAt)
}