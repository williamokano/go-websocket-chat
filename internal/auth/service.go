package auth

import (
	"context"
	"fmt"

	"github.com/williamokano/example-websocket-chat/internal/user"
)

type Service struct {
	userRepo   user.Repo
	jwtService *JWTService
}

func NewService(userRepo user.Repo, jwtService *JWTService) *Service {
	return &Service{userRepo: userRepo, jwtService: jwtService}
}

type AuthResponse struct {
	Token string    `json:"token"`
	User  user.User `json:"user"`
}

func (s *Service) Register(ctx context.Context, username, password string) (*AuthResponse, error) {
	if len(username) < 3 || len(username) > 50 {
		return nil, fmt.Errorf("username must be between 3 and 50 characters")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	hashed, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	u, err := s.userRepo.Create(ctx, username, hashed)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	token, err := s.jwtService.GenerateToken(u.ID, u.Username)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	u.Password = ""
	return &AuthResponse{Token: token, User: *u}, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*AuthResponse, error) {
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !CheckPassword(password, u.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.jwtService.GenerateToken(u.ID, u.Username)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	u.Password = ""
	return &AuthResponse{Token: token, User: *u}, nil
}

func (s *Service) GetUser(ctx context.Context, userID string) (*user.User, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	u.Password = ""
	return u, nil
}
