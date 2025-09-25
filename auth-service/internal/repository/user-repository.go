package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leandrowiemesfilho/auth-service/internal/model"
	"github.com/leandrowiemesfilho/auth-service/internal/util"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrDuplicateEmail    = errors.New("email already registered")
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
        INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		util.Error("Failed to create user", map[string]interface{}{
			"error": err,
			"email": user.Email,
		})

		// Check for duplicate email error
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return ErrDuplicateEmail
		}

		return fmt.Errorf("failed to create user: %w", err)
	}

	util.Info("User created successfully", map[string]interface{}{
		"email":   user.Email,
		"user_id": user.ID,
	})
	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
        SELECT id, email, password_hash, name, created_at, updated_at
        FROM users 
        WHERE email = $1
    `

	var user model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		util.Error("Failed to get user by email", map[string]interface{}{
			"error": err,
			"email": email,
		})
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	query := `
        SELECT id, email, password_hash, name, created_at, updated_at
        FROM users 
        WHERE id = $1
    `

	var user model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		util.Error("Failed to get user by ID", map[string]interface{}{
			"error":   err,
			"user_id": id,
		})

		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}
