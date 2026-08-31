package group

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, g *Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*Group, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]Group, error)

	AddMember(ctx context.Context, groupID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]MemberInfo, error)
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
	IsOwner(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, g *Group) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO groups (id, name, owner_id) VALUES ($1, $2, $3)",
		g.ID, g.Name, g.OwnerID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO group_members (group_id, user_id, is_owner) VALUES ($1, $2, $3)",
		g.ID, g.OwnerID, true)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Group, error) {
	var g Group
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, owner_id, created_at FROM groups WHERE id = $1", id).
		Scan(&g.ID, &g.Name, &g.OwnerID, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *postgresRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Group, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, g.owner_id, g.created_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []Group{}
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.OwnerID, &g.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *postgresRepository) AddMember(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)",
		groupID, userID)
	return err
}

func (r *postgresRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM group_members WHERE group_id = $1 AND user_id = $2",
		groupID, userID)
	return err
}

func (r *postgresRepository) ListMembers(ctx context.Context, groupID uuid.UUID) ([]MemberInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.name, u.email
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = $1`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []MemberInfo{}
	for rows.Next() {
		var m MemberInfo
		if err := rows.Scan(&m.UserID, &m.Name, &m.Email); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *postgresRepository) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
		groupID, userID).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) IsOwner(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var isOwner bool
	err := r.db.QueryRowContext(ctx,
		"SELECT is_owner FROM group_members WHERE group_id = $1 AND user_id = $2",
		groupID, userID).Scan(&isOwner)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return isOwner, err
}