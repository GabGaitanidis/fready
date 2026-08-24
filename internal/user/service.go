package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service interface {
    GetUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
    RegisterUser(ctx context.Context, name, email string) (*User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetUsers(ctx context.Context) ([]User, error) {
    return s.repo.GetUsers(ctx)
}

func (s *service) GetUser(ctx context.Context,id uuid.UUID) (*User,error) {
	return  s.repo.GetByID(ctx, id)
}
func (s *service) RegisterUser(ctx context.Context, name, email string) (*User, error) {
    if name == "" || email == "" {
        return nil, errors.New("name and email are required")
    }
    u := &User{ID: uuid.New(),Name: name, Email: email}
    if err := s.repo.Create(ctx, u); err != nil {
        return nil, err
    }
    return u, nil
}