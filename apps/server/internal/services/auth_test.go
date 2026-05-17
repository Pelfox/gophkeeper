package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Pelfox/gophkeeper/apps/server/internal/crypto"
	"github.com/Pelfox/gophkeeper/apps/server/internal/models"
	"github.com/Pelfox/gophkeeper/apps/server/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TestAuthServiceCreateAccessToken verifies that access tokens are signed with
// expected claims and expiration.
func TestAuthServiceCreateAccessToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	service := NewAuthService(
		&fakeUsersRepository{},
		&fakeSessionsRepository{},
		secret,
		testLogger(),
	).(*authService)

	result, err := service.createAccessToken(userID)
	if err != nil {
		t.Fatalf("createAccessToken returned error: %v", err)
	}
	if result.signedToken == "" {
		t.Fatal("expected signed token to be populated")
	}

	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		result.signedToken,
		&claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				t.Fatalf("unexpected signing method: %v", token.Method)
			}
			return []byte(secret), nil
		},
	)
	if err != nil {
		t.Fatalf("failed to parse access token: %v", err)
	}
	if !token.Valid {
		t.Fatal("expected access token to be valid")
	}
	if claims.Subject != userID.String() {
		t.Fatalf(
			"unexpected token subject: got %q want %q",
			claims.Subject,
			userID,
		)
	}
	if result.expiresAt.Before(time.Now().Add(AccessTokenLifetime - time.Minute)) {
		t.Fatalf("token expiry is too early: %s", result.expiresAt)
	}
	if result.expiresAt.After(time.Now().Add(AccessTokenLifetime + time.Minute)) {
		t.Fatalf("token expiry is too late: %s", result.expiresAt)
	}
}

// TestAuthServiceRegisterSuccess verifies that registration hashes the
// password, creates a user and creates a session.
func TestAuthServiceRegisterSuccess(t *testing.T) {
	user := &models.User{
		ID:        uuid.New(),
		Email:     "user@example.com",
		CreatedAt: fixedTime(),
		UpdatedAt: fixedTime().Add(time.Minute),
	}
	users := &fakeUsersRepository{createUser: user}
	sessions := &fakeSessionsRepository{}
	service := NewAuthService(users, sessions, "secret", testLogger())

	result, err := service.Register(context.Background(), RegisterInput{
		Email:    user.Email,
		Password: "password",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if users.createCalls != 1 {
		t.Fatalf(
			"expected users.Create to be called once, got %d",
			users.createCalls,
		)
	}
	if users.createInput.Email != user.Email {
		t.Fatalf(
			"unexpected create email: got %q want %q",
			users.createInput.Email,
			user.Email,
		)
	}

	matches, err := crypto.VerifyPassword(
		"password",
		users.createInput.PasswordHash,
	)
	if err != nil {
		t.Fatalf("stored password hash cannot be verified: %v", err)
	}
	if !matches {
		t.Fatal("stored password hash does not match plaintext password")
	}
	if sessions.createCalls != 1 {
		t.Fatalf(
			"expected sessions.Create to be called once, got %d",
			sessions.createCalls,
		)
	}
	if sessions.createInput.UserID != user.ID {
		t.Fatalf(
			"unexpected session user: got %s want %s",
			sessions.createInput.UserID,
			user.ID,
		)
	}
	if sessions.createInput.AccessToken == "" {
		t.Fatal("expected session access token to be populated")
	}
	if result.User.ID != user.ID || result.User.Email != user.Email {
		t.Fatalf("unexpected result user: %#v", result.User)
	}
	if result.AccessToken != sessions.createInput.AccessToken {
		t.Fatalf(
			"unexpected access token: got %q want %q",
			result.AccessToken,
			sessions.createInput.AccessToken,
		)
	}
}

// TestAuthServiceRegisterRepositoryErrors verifies that repository failures are
// translated to service-level registration errors.
func TestAuthServiceRegisterRepositoryErrors(t *testing.T) {
	tests := []struct {
		name          string
		users         *fakeUsersRepository
		sessions      *fakeSessionsRepository
		expectedError error
	}{
		{
			name: "duplicate user",
			users: &fakeUsersRepository{
				createErr: repositories.ErrDuplicateUser,
			},
			sessions:      &fakeSessionsRepository{},
			expectedError: ErrDuplicateUser,
		},
		{
			name: "user creation failure",
			users: &fakeUsersRepository{
				createErr: errors.New("insert failed"),
			},
			sessions:      &fakeSessionsRepository{},
			expectedError: ErrUserCreationFailed,
		},
		{
			name: "session creation failure",
			users: &fakeUsersRepository{
				createUser: &models.User{
					ID:    uuid.New(),
					Email: "user@example.com",
				},
			},
			sessions: &fakeSessionsRepository{
				createErr: errors.New("insert session failed"),
			},
			expectedError: ErrSessionCreationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAuthService(
				tt.users,
				tt.sessions,
				"secret",
				testLogger(),
			)

			result, err := service.Register(context.Background(), RegisterInput{
				Email:    "user@example.com",
				Password: "password",
			})
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("unexpected error: got %v want %v", err, tt.expectedError)
			}
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
		})
	}
}

// TestAuthServiceLoginSuccess verifies that a valid email and password creates
// a new authenticated session.
func TestAuthServiceLoginSuccess(t *testing.T) {
	passwordHash, err := crypto.HashPassword("password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := &models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: passwordHash,
		CreatedAt:    fixedTime(),
		UpdatedAt:    fixedTime().Add(time.Minute),
	}
	users := &fakeUsersRepository{findByEmailUser: user}
	sessions := &fakeSessionsRepository{}
	service := NewAuthService(users, sessions, "secret", testLogger())

	result, err := service.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "password",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if users.findByEmailCalls != 1 || users.findByEmailEmail != user.Email {
		t.Fatalf(
			"unexpected FindByEmail call: calls=%d email=%q",
			users.findByEmailCalls,
			users.findByEmailEmail,
		)
	}
	if sessions.createCalls != 1 {
		t.Fatalf(
			"expected sessions.Create to be called once, got %d",
			sessions.createCalls,
		)
	}
	if result.User.ID != user.ID || result.User.Email != user.Email {
		t.Fatalf("unexpected result user: %#v", result.User)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token to be populated")
	}
}

// TestAuthServiceLoginErrors verifies that login failures are translated to the
// expected service-level errors.
func TestAuthServiceLoginErrors(t *testing.T) {
	passwordHash, err := crypto.HashPassword("password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	tests := []struct {
		name          string
		users         *fakeUsersRepository
		sessions      *fakeSessionsRepository
		password      string
		expectedError error
	}{
		{
			name: "user not found",
			users: &fakeUsersRepository{
				findByEmailErr: repositories.ErrUserNotFound,
			},
			sessions:      &fakeSessionsRepository{},
			password:      "password",
			expectedError: ErrLoginFailed,
		},
		{
			name: "user query failed",
			users: &fakeUsersRepository{
				findByEmailErr: errors.New("select failed"),
			},
			sessions:      &fakeSessionsRepository{},
			password:      "password",
			expectedError: ErrUserQueryFailed,
		},
		{
			name: "invalid password hash",
			users: &fakeUsersRepository{
				findByEmailUser: &models.User{
					ID:           uuid.New(),
					PasswordHash: "not an argon hash",
				},
			},
			sessions:      &fakeSessionsRepository{},
			password:      "password",
			expectedError: ErrPasswordVerificationFailed,
		},
		{
			name: "wrong password",
			users: &fakeUsersRepository{
				findByEmailUser: &models.User{
					ID:           uuid.New(),
					PasswordHash: passwordHash,
				},
			},
			sessions:      &fakeSessionsRepository{},
			password:      "wrong-password",
			expectedError: ErrLoginFailed,
		},
		{
			name: "session creation failed",
			users: &fakeUsersRepository{
				findByEmailUser: &models.User{
					ID:           uuid.New(),
					PasswordHash: passwordHash,
				},
			},
			sessions: &fakeSessionsRepository{
				createErr: errors.New("insert session failed"),
			},
			password:      "password",
			expectedError: ErrSessionCreationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAuthService(
				tt.users,
				tt.sessions,
				"secret",
				testLogger(),
			)

			result, err := service.Login(context.Background(), LoginInput{
				Email:    "user@example.com",
				Password: tt.password,
			})
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf(
					"unexpected error: got %v want %v",
					err,
					tt.expectedError,
				)
			}
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
		})
	}
}

// TestAuthServiceVerifySessionSuccess verifies that an access token resolves to
// its session and user.
func TestAuthServiceVerifySessionSuccess(t *testing.T) {
	user := &models.User{
		ID:        uuid.New(),
		Email:     "user@example.com",
		CreatedAt: fixedTime(),
		UpdatedAt: fixedTime().Add(time.Minute),
	}
	session := &models.Session{
		ID:          uuid.New(),
		UserID:      user.ID,
		AccessToken: "token",
		ExpiresAt:   fixedTime().Add(time.Hour),
		CreatedAt:   fixedTime(),
	}
	users := &fakeUsersRepository{getByIDUser: user}
	sessions := &fakeSessionsRepository{findByAccessTokenItem: session}
	service := NewAuthService(users, sessions, "secret", testLogger())

	result, err := service.VerifySession(context.Background(), "token")
	if err != nil {
		t.Fatalf("VerifySession returned error: %v", err)
	}
	if sessions.findByAccessTokenCalls != 1 ||
		sessions.findByAccessTokenValue != "token" {
		t.Fatalf(
			"unexpected session lookup: calls=%d token=%q",
			sessions.findByAccessTokenCalls,
			sessions.findByAccessTokenValue,
		)
	}
	if users.getByIDCalls != 1 || users.getByIDID != user.ID {
		t.Fatalf(
			"unexpected user lookup: calls=%d id=%s",
			users.getByIDCalls,
			users.getByIDID,
		)
	}
	if result.User.ID != user.ID ||
		result.AccessToken != session.AccessToken ||
		!result.ExpiresAt.Equal(session.ExpiresAt) {
		t.Fatalf("unexpected result: %#v", result)
	}
}

// TestAuthServiceVerifySessionErrors verifies session lookup and user lookup
// error handling.
func TestAuthServiceVerifySessionErrors(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		users         *fakeUsersRepository
		sessions      *fakeSessionsRepository
		expectedError error
	}{
		{
			name:  "session not found",
			users: &fakeUsersRepository{},
			sessions: &fakeSessionsRepository{
				findByAccessTokenErr: repositories.ErrSessionNotFound,
			},
			expectedError: ErrSessionNotFound,
		},
		{
			name:  "session query failed",
			users: &fakeUsersRepository{},
			sessions: &fakeSessionsRepository{
				findByAccessTokenErr: errors.New("select session failed"),
			},
			expectedError: ErrSessionQueryFailed,
		},
		{
			name: "session user no longer exists",
			users: &fakeUsersRepository{
				getByIDErr: repositories.ErrUserNotFound,
			},
			sessions: &fakeSessionsRepository{
				findByAccessTokenItem: &models.Session{
					ID:     uuid.New(),
					UserID: userID,
				},
			},
			expectedError: ErrSessionNotFound,
		},
		{
			name: "user query failed",
			users: &fakeUsersRepository{
				getByIDErr: errors.New("select user failed"),
			},
			sessions: &fakeSessionsRepository{
				findByAccessTokenItem: &models.Session{
					ID:     uuid.New(),
					UserID: userID,
				},
			},
			expectedError: ErrUserQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAuthService(
				tt.users,
				tt.sessions,
				"secret",
				testLogger(),
			)

			result, err := service.VerifySession(context.Background(), "token")
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf(
					"unexpected error: got %v want %v",
					err,
					tt.expectedError,
				)
			}
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
		})
	}
}
