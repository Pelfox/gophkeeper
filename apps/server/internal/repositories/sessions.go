package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/Pelfox/gophkeeper/apps/server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateSessionInput describes the parameters used for session creation.
type CreateSessionInput struct {
	// UserID is a unique identifier of the owner of the session.
	UserID uuid.UUID
	// AccessToken is session's access token.
	AccessToken string
	// ExpiresAt is a timestamp when this session expires.
	ExpiresAt time.Time
}

// SessionsRepository describes sessions-related operations.
type SessionsRepository interface {
	// Create creates a new session.
	Create(
		ctx context.Context,
		input CreateSessionInput,
	) (*models.Session, error)
	// GetForUser returns all session associated with the given user.
	GetForUser(ctx context.Context, userID uuid.UUID) ([]models.Session, error)
}

type sessionsRepository struct {
	pool *pgxpool.Pool
	sq   squirrel.StatementBuilderType
}

// NewSessionsRepository creates a new sessions repository, backed by Postgres.
func NewSessionsRepository(pool *pgxpool.Pool) SessionsRepository {
	return &sessionsRepository{
		pool: pool,
		sq:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *sessionsRepository) Create(
	ctx context.Context,
	input CreateSessionInput,
) (*models.Session, error) {
	query, args, err := r.sq.Insert("sessions").
		Columns("user_id", "access_token", "expires_at").
		Values(input.UserID, input.AccessToken, input.ExpiresAt).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	session := models.Session{
		UserID:      input.UserID,
		AccessToken: input.AccessToken,
		ExpiresAt:   input.ExpiresAt,
	}

	err = r.pool.QueryRow(ctx, query, args...).
		Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &session, nil
}

func (r *sessionsRepository) GetForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Session, error) {
	query, args, err := r.sq.
		Select("id", "access_token", "expires_at", "created_at").
		From("sessions").
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	sessions := make([]models.Session, 0)
	for rows.Next() {
		session := models.Session{UserID: userID}
		err = rows.Scan(
			&session.ID,
			&session.AccessToken,
			&session.ExpiresAt,
			&session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan rows: %w", err)
	}

	return sessions, nil
}
