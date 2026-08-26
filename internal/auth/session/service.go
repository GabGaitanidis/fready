package session

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateSession(ctx context.Context, userID uuid.UUID) (*Session, error)
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	CurrentUserID(r *http.Request) (uuid.UUID, error)
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

	if err := s.repo.CreateSession(ctx, sessionID, userID, expiresAt); err != nil {
        slog.Error("failed to persist session", "error", err, "user_id", userID)
        return nil, err
    }

    slog.Info("session created", "user_id", userID)
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

func (s *service) CurrentUserID(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return uuid.Nil, errors.New("not authenticated")
	}

	sess, err := s.GetSession(r.Context(), cookie.Value)
	if err != nil {
		return uuid.Nil, err
	}

	return sess.UserID, nil
}