package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leandrowiemesfilho/auth-service/internal/model"
	"github.com/leandrowiemesfilho/auth-service/internal/repository"
	"github.com/leandrowiemesfilho/auth-service/internal/util"
)

type AuthService interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	ValidateToken(ctx context.Context, token string) (*model.User, error)
}

type authService struct {
	userRepo     repository.UserRepository
	jwtUtil      util.JWTUtil
	passwordUtil util.PasswordUtil
	config       *JWTConfig
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtUtil util.JWTUtil,
	passwordUtil util.PasswordUtil,
	config *JWTConfig,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		jwtUtil:      jwtUtil,
		passwordUtil: passwordUtil,
		config:       config,
	}
}

func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already registered: %w", repository.ErrDuplicateEmail)
	}
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password
	hashedPassword, err := s.passwordUtil.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwtUtil.GenerateToken(user.ID.String(), user.Email, s.config.ExpirationHours)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Clear password hash for response
	user.PasswordHash = ""

	return &model.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if !s.passwordUtil.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.jwtUtil.GenerateToken(user.ID.String(), user.Email, s.config.ExpirationHours)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Clear password hash for response
	user.PasswordHash = ""

	return &model.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	claims, err := s.jwtUtil.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
