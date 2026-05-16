package services

import (
	"context"
	"errors"
	"time"

	"github.com/Pelfox/gophkeeper/apps/server/internal/crypto"
	"github.com/Pelfox/gophkeeper/apps/server/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	// AccessTokenLifetime describes how long access token should live.
	AccessTokenLifetime = 30 * 24 * time.Hour // 1 month
)

var (
	// ErrPasswordHashingFailed is returned when service was unable to hash
	// user's password.
	ErrPasswordHashingFailed = errors.New("password hashing failed")
	// ErrDuplicateUser is returned when email that user has submitted already
	// taken by another user.
	ErrDuplicateUser = errors.New("user with the same email already exists")
	// ErrUserCreationFailed is returned when service was unable to create a
	// new user.
	ErrUserCreationFailed = errors.New("failed to create a new user")
	// ErrAccessTokenCreationFailed is returned when service was unable to
	// create a new access token.
	ErrAccessTokenCreationFailed = errors.New("failed to create a new access token")
	// ErrSessionCreationFailed is returned when service was unable to create a
	// new session.
	ErrSessionCreationFailed = errors.New("failed to create a new session")
	// ErrLoginFailed is returned when either user wasn't found, or their
	// password is invalid.
	ErrLoginFailed = errors.New("failed to login")
	// ErrUserQueryFailed is returned when service was unable to query a user.
	ErrUserQueryFailed = errors.New("failed to query the user")
	// ErrPasswordVerificationFailed is returned whtn service was unable to
	// verify user's password.
	ErrPasswordVerificationFailed = errors.New("failed to verify the password")
)

// User describes protocol-compatible user structure.
type User struct {
	// ID is user's identifier.
	ID uuid.UUID
	// Email is user's email.
	Email string
	// CreatedAt is a timestamp when this user was created.
	CreatedAt time.Time
	// UpdatedAt is a timestamp when this user was last updated.
	UpdatedAt time.Time
}

// RegisterInput describes input parameters for the register function.
type RegisterInput struct {
	// Email is user's email.
	Email string
	// Password is user's plaintext password.
	Password string
}

// RegisterResult describes the result of the registration.
type RegisterResult struct {
	// User is a registered user.
	User User
	// AccessToken is created session's access token.
	AccessToken string
	// ExpiresAt is a timestamp when this session expires.
	ExpiresAt time.Time
}

// LoginInput describes input parameters for logging user in.
type LoginInput struct {
	// Email is user's email.
	Email string
	// Password is user's plaintext password.
	Password string
}

// LoginResult describes the result of the login.
type LoginResult struct {
	// User is a logged in user.
	User User
	// AccessToken is created session's access token.
	AccessToken string
	// ExpiresAt is a timestamp when this session expires.
	ExpiresAt time.Time
}

// AuthService describes all auth-related operations.
type AuthService interface {
	// Register creates a new user as well as a new session and returns them.
	Register(ctx context.Context, request RegisterInput) (*RegisterResult, error)
	// Login logs user in, creating a new session.
	Login(ctx context.Context, request LoginInput) (*LoginResult, error)
}

type authService struct {
	usersRepository    repositories.UsersRepository
	sessionsRepository repositories.SessionsRepository
	jwtSecret          []byte
	logger             zerolog.Logger
}

// NewAuthService creates a new auth service.
func NewAuthService(
	usersRepository repositories.UsersRepository,
	sessionsRepository repositories.SessionsRepository,
	jwtSecret string,
	logger zerolog.Logger,
) AuthService {
	return &authService{
		usersRepository:    usersRepository,
		sessionsRepository: sessionsRepository,
		jwtSecret:          []byte(jwtSecret),
		logger:             logger.With().Str("service", "auth").Logger(),
	}
}

type accessTokenResult struct {
	signedToken string
	expiresAt   time.Time
}

func (s *authService) createAccessToken(
	userID uuid.UUID,
) (*accessTokenResult, error) {
	currentTime := time.Now()
	expiresAt := currentTime.Add(AccessTokenLifetime)

	claims := &jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		NotBefore: jwt.NewNumericDate(currentTime),
		IssuedAt:  jwt.NewNumericDate(currentTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	result := accessTokenResult{
		signedToken: signedToken,
		expiresAt:   expiresAt,
	}
	return &result, nil
}

func (s *authService) Register(
	ctx context.Context,
	request RegisterInput,
) (*RegisterResult, error) {
	passwordHash, err := crypto.HashPassword(request.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash the password")
		return nil, ErrPasswordHashingFailed
	}

	user, err := s.usersRepository.Create(ctx, repositories.CreateUserInput{
		Email:        request.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicateUser) {
			return nil, ErrDuplicateUser
		}
		s.logger.Error().Err(err).Msg("failed to create a new user")
		return nil, ErrUserCreationFailed
	}

	accessTokenResult, err := s.createAccessToken(user.ID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create a new access token")
		return nil, ErrAccessTokenCreationFailed
	}

	session, err := s.sessionsRepository.Create(ctx, repositories.CreateSessionInput{
		UserID:      user.ID,
		AccessToken: accessTokenResult.signedToken,
		ExpiresAt:   accessTokenResult.expiresAt,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create a new session")
		return nil, ErrSessionCreationFailed
	}

	result := RegisterResult{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		AccessToken: session.AccessToken,
		ExpiresAt:   session.ExpiresAt,
	}
	return &result, nil
}

func (s *authService) Login(
	ctx context.Context,
	request LoginInput,
) (*LoginResult, error) {
	user, err := s.usersRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrLoginFailed
		}
		s.logger.Error().Err(err).Msg("failed to query user")
		return nil, ErrUserQueryFailed
	}

	match, err := crypto.VerifyPassword(request.Password, user.PasswordHash)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to verify user's password")
		return nil, ErrPasswordVerificationFailed
	}

	if !match {
		return nil, ErrLoginFailed
	}

	accessTokenResult, err := s.createAccessToken(user.ID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create a new access token")
		return nil, ErrAccessTokenCreationFailed
	}

	session, err := s.sessionsRepository.Create(ctx, repositories.CreateSessionInput{
		UserID:      user.ID,
		AccessToken: accessTokenResult.signedToken,
		ExpiresAt:   accessTokenResult.expiresAt,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create a new session")
		return nil, ErrSessionCreationFailed
	}

	result := LoginResult{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		AccessToken: session.AccessToken,
		ExpiresAt:   session.ExpiresAt,
	}
	return &result, nil
}
