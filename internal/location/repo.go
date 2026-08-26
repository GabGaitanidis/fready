package location

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
	Upsert(ctx context.Context, userID uuid.UUID, lat, lng float64) (*LocationPing, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Upsert(ctx context.Context, userID uuid.UUID, lat, lng float64) (*LocationPing, error) {
	var p LocationPing
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO location_pings (user_id, lat, lng, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id) DO UPDATE
		SET lat = $2, lng = $3, updated_at = now()
		RETURNING user_id, lat, lng, updated_at`,
		userID, lat, lng).
		Scan(&p.UserID, &p.Lat, &p.Lng, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}