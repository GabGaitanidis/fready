package user

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
    GetUsers(ctx context.Context) ([]User, error)
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    Create(ctx context.Context, u *User) error
}

type postgresRepository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
    return &postgresRepository{db: db}
}


func (r *postgresRepository) GetUsers(ctx context.Context) ([]User, error) {
    rows, err := r.db.QueryContext(ctx, "SELECT id, name, email FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
            return nil, err
        }
        users = append(users, u)

    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil

}


func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    var u User
    err := r.db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = $1", id).
        Scan(&u.ID, &u.Name, &u.Email)
    if err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *postgresRepository) Create(ctx context.Context, u *User) error {
    _, err := r.db.ExecContext(ctx, "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)", u.ID, u.Name, u.Email)
    return err
}