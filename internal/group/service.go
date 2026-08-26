package group

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

type Service interface {
	CreateGroup(ctx context.Context, name string, ownerID uuid.UUID) (*Group, error)
	GetGroup(ctx context.Context, id uuid.UUID) (*Group, error)
	ListMyGroups(ctx context.Context, userID uuid.UUID) ([]Group, error)

	AddMember(ctx context.Context, groupID, requesterID, newMemberID uuid.UUID) error
	RemoveMember(ctx context.Context, groupID, requesterID, memberID uuid.UUID) error
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateGroup(ctx context.Context, name string, ownerID uuid.UUID) (*Group, error) {
	if name == "" {
		return nil, errors.New("group name is required")
	}
	if ownerID == uuid.Nil {
		return nil, errors.New("owner ID is required")
	}

	g := &Group{
		ID:      uuid.New(),
		Name:    name,
		OwnerID: ownerID,
	}

	if err := s.repo.Create(ctx, g); err != nil {
        slog.Error("failed to create group", "error", err, "owner_id", ownerID)
        return nil, err
    }

    slog.Info("group created", "group_id", g.ID, "owner_id", ownerID)
    return g, nil
}

func (s *service) GetGroup(ctx context.Context, id uuid.UUID) (*Group, error) {
	if id == uuid.Nil {
		return nil, errors.New("group ID is required")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListMyGroups(ctx context.Context, userID uuid.UUID) ([]Group, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID is required")
	}
	return s.repo.ListForUser(ctx, userID)
}

func (s *service) AddMember(ctx context.Context, groupID, requesterID, newMemberID uuid.UUID) error {
	isOwner, err := s.repo.IsOwner(ctx, groupID, requesterID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("only the group owner can add members")
	}

	alreadyMember, err := s.repo.IsMember(ctx, groupID, newMemberID)
	if err != nil {
		return err
	}
	if alreadyMember {
		return errors.New("user is already a member of this group")
	}

	if err := s.repo.AddMember(ctx, groupID, newMemberID); err != nil {
        slog.Error("failed to add group member", "error", err, "group_id", groupID)
        return err
    }

    slog.Info("member added to group", "group_id", groupID, "user_id", newMemberID, "added_by", requesterID)
    return nil
}

func (s *service) RemoveMember(ctx context.Context, groupID, requesterID, memberID uuid.UUID) error {
	isOwner, err := s.repo.IsOwner(ctx, groupID, requesterID)
	if err != nil {
		return err
	}

	if !isOwner && requesterID != memberID {
		return errors.New("not authorized to remove this member")
	}

	if memberID == requesterID && isOwner {
		return errors.New("group owner cannot leave — transfer ownership or delete the group instead")
	}

	return s.repo.RemoveMember(ctx, groupID, memberID)
}

func (s *service) ListMembers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	if groupID == uuid.Nil {
		return nil, errors.New("group ID is required")
	}
	return s.repo.ListMembers(ctx, groupID)
}