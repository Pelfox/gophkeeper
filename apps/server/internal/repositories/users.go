package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/Pelfox/gophkeeper/apps/server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrDuplicateUser is returned when a user with the given email already
	// exists.
	ErrDuplicateUser = errors.New("user with this email already exists")
	// ErrUserNotFound is returned when no user matches the query.
	ErrUserNotFound = errors.New("user not found")
)

// CreateUserInput contains the data required to create a user.
type CreateUserInput struct {
	// Email is user's email.
	Email string
	// PasswordHash is the hash of the user's password.
	PasswordHash string
}

// UpdateUserInput describes parameters used to update user's password.
type UpdateUserInput struct {
	// Email is new user's email.
	Email *string
	// Password hash is new hasth of the user's password.
	PasswordHash *string
}

// UsersRepository stores and retrieves users.
type UsersRepository interface {
	// Create saves a new user.
	Create(ctx context.Context, input CreateUserInput) (*models.User, error)
	// FindByEmail finds user by the supplied email. It returns ErrUserNotFound
	// if no matching user exists.
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	// Update updates existing user with the new information.
	Update(
		ctx context.Context,
		id uuid.UUID,
		input UpdateUserInput,
	) (*models.User, error)
}

type usersRepository struct {
	pool *pgxpool.Pool
	sq   squirrel.StatementBuilderType
}

// NewUsersRepository creates a user repository backed by PostgreSQL.
func NewUsersRepository(pool *pgxpool.Pool) UsersRepository {
	return &usersRepository{
		pool: pool,
		sq:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *usersRepository) Create(
	ctx context.Context,
	input CreateUserInput,
) (*models.User, error) {
	query, args, err := r.sq.Insert("users").
		Columns("email", "password_hash").
		Values(input.Email, input.PasswordHash).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	user := models.User{
		Email:        input.Email,
		PasswordHash: input.PasswordHash,
	}

	err = r.pool.QueryRow(ctx, query, args...).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return nil, ErrDuplicateUser
			}
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &user, nil
}

func (r *usersRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	query, args, err := r.sq.
		Select("id", "password_hash", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	user := models.User{Email: email}
	err = r.pool.QueryRow(ctx, query, args...).
		Scan(&user.ID, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &user, nil
}

func (r *usersRepository) Update(
	ctx context.Context,
	id uuid.UUID,
	input UpdateUserInput,
) (*models.User, error) {
	builder := r.sq.Update("users")

	if input.Email != nil {
		builder = builder.Set("email", input.Email)
	}
	if input.PasswordHash != nil {
		builder = builder.Set("password_hash", input.PasswordHash)
	}

	query, args, err := builder.Set("updated_at", squirrel.Expr("NOW()")).
		Suffix("RETURNING email, password_hash, created_at, updated_at").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	user := models.User{ID: id}

	err = r.pool.QueryRow(ctx, query, args...).
		Scan(&user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &user, nil
}
