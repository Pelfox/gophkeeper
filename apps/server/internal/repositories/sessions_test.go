package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Pelfox/gophkeeper/apps/server/internal/models"
	"github.com/google/uuid"
)

// TestSessionsRepositoryCreate verifies that Create stores a session and
// returns generated metadata.
func TestSessionsRepositoryCreate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewSessionsRepository(pool)

	userID := createTestUser(t, ctx, pool, "sessions@example.com")
	expiresAt := time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond)

	session, err := repo.Create(ctx, CreateSessionInput{
		UserID:      userID,
		AccessToken: "access-token",
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if session.ID == uuid.Nil ||
		session.UserID != userID ||
		session.AccessToken != "access-token" {
		t.Fatalf("unexpected session: %#v", session)
	}
}

// TestSessionsRepositoryCreateDuplicateToken verifies that Create preserves
// the unique access-token constraint.
func TestSessionsRepositoryCreateDuplicateToken(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewSessionsRepository(pool)
	userID := createTestUser(t, ctx, pool, "sessions@example.com")

	createTestSession(t, ctx, repo, userID, "access-token")
	_, err := repo.Create(ctx, CreateSessionInput{
		UserID:      userID,
		AccessToken: "access-token",
		ExpiresAt:   time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond),
	})
	assertPgErrorCode(t, err, "23505")
}

// TestSessionsRepositoryCreateUnknownUser verifies that Create preserves the
// foreign-key constraint on user_id.
func TestSessionsRepositoryCreateUnknownUser(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewSessionsRepository(pool)

	_, err := repo.Create(ctx, CreateSessionInput{
		UserID:      uuid.New(),
		AccessToken: "access-token",
		ExpiresAt:   time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond),
	})
	assertPgErrorCode(t, err, "23503")
}

// TestSessionsRepositoryGetForUser verifies that GetForUser returns sessions
// belonging to the supplied user.
func TestSessionsRepositoryGetForUser(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewSessionsRepository(pool)

	userID := createTestUser(t, ctx, pool, "sessions@example.com")
	session := createTestSession(t, ctx, repo, userID, "access-token")

	sessions, err := repo.GetForUser(ctx, userID)
	if err != nil {
		t.Fatalf("GetForUser returned error: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != session.ID {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}
}

// TestSessionsRepositoryGetForUserEmpty verifies that GetForUser returns an
// empty slice when the user has no sessions.
func TestSessionsRepositoryGetForUserEmpty(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewSessionsRepository(pool)
	userID := createTestUser(t, ctx, pool, "sessions@example.com")

	sessions, err := repo.GetForUser(ctx, userID)
	if err != nil {
		t.Fatalf("GetForUser returned error: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected no sessions, got %#v", sessions)
	}
}

// TestSessionsRepositoryFindByAccessToken verifies that FindByAccessToken
// returns the session tied to the supplied token.
func TestSessionsRepositoryFindByAccessToken(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewSessionsRepository(pool)

	userID := createTestUser(t, ctx, pool, "sessions@example.com")
	session := createTestSession(t, ctx, repo, userID, "access-token")

	found, err := repo.FindByAccessToken(ctx, "access-token")
	if err != nil {
		t.Fatalf("FindByAccessToken returned error: %v", err)
	}
	if found.ID != session.ID || found.UserID != userID {
		t.Fatalf("unexpected session: %#v", found)
	}
}

// TestSessionsRepositoryFindByAccessTokenNotFound verifies that
// FindByAccessToken returns ErrSessionNotFound for an unknown token.
func TestSessionsRepositoryFindByAccessTokenNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewSessionsRepository(pool)

	_, err := repo.FindByAccessToken(ctx, "missing-token")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrSessionNotFound)
	}
}

func createTestSession(
	t *testing.T,
	ctx context.Context,
	repo SessionsRepository,
	userID uuid.UUID,
	accessToken string,
) *models.Session {
	t.Helper()

	session, err := repo.Create(ctx, CreateSessionInput{
		UserID:      userID,
		AccessToken: accessToken,
		ExpiresAt:   time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond),
	})
	if err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}
	return session
}
