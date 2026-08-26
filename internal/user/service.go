package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	GetUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email, password string) (*User, error)
	RegisterUser(ctx context.Context, name, email, password string) (*User, error)
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

func (s *service) GetUser(ctx context.Context, userId uuid.UUID) (*User, error) {
    if userId == uuid.Nil {
        return nil, errors.New("user ID is required")
    } 
	return s.repo.GetByID(ctx, userId)
}

func (s *service) GetUserByEmail(ctx context.Context, email, password string) (*User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return u, nil
}

func (s *service) RegisterUser(ctx context.Context, name, email, password string) (*User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("name, email, and password are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        slog.Error("bcrypt hash generation failed", "error", err)
        return nil, err
    }

	u := &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, u); err != nil {
        slog.Error("failed to create user", "error", err, "email", email)
        return nil, err
    }
	
	return u, nil
}