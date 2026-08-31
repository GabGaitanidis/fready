package session

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, sessionID string, userID uuid.UUID, expiresAt time.Time) error
	GetSessionByID(ctx context.Context, sessionID string) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteExpiredSessions(ctx context.Context) error
}

type postgresSessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &postgresSessionRepository{db: db}
}

func (r *postgresSessionRepository) CreateSession(ctx context.Context, sessionID string, userID uuid.UUID, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)", sessionID, userID, expiresAt)
	return err
}

func (r *postgresSessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*Session, error) {	
	var s Session
	err := r.db.QueryRowContext(ctx, "SELECT id, user_id, created_at, expires_at FROM sessions WHERE id = $1 AND expires_at > NOW()", sessionID).
		Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *postgresSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = $1", sessionID)
	return err
}

func (r *postgresSessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= $1`, time.Now())
	return err
}