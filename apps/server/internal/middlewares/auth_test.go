package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	serverinternal "github.com/Pelfox/gophkeeper/apps/server/internal"
	"github.com/Pelfox/gophkeeper/apps/server/internal/services"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeAuthService struct {
	verifyCalls int
	verifyToken string
	session     *services.SessionResult
	err         error
}

func (s *fakeAuthService) Register(
	context.Context,
	services.RegisterInput,
) (*services.SessionResult, error) {
	panic("unexpected Register call")
}

func (s *fakeAuthService) Login(
	context.Context,
	services.LoginInput,
) (*services.SessionResult, error) {
	panic("unexpected Login call")
}

func (s *fakeAuthService) VerifySession(
	_ context.Context,
	accessToken string,
) (*services.SessionResult, error) {
	s.verifyCalls++
	s.verifyToken = accessToken
	return s.session, s.err
}

func performAuthMiddlewareRequest(
	authService services.AuthService,
	authorizationHeader string,
) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuthMiddleware(authService))
	router.GET("/protected", func(ctx *gin.Context) {
		session, ok := serverinternal.SessionFromContext(ctx.Request.Context())
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "session not found"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"user_id": session.User.ID.String()})
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authorizationHeader != "" {
		request.Header.Set("Authorization", authorizationHeader)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeProtocolError(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
) protocol.ProtocolError {
	t.Helper()

	var response protocol.ProtocolError
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode protocol error: %v", err)
	}
	return response
}

// TestAuthMiddlewareRejectsMissingAuthorizationHeader verifies that requests
// without authorization are rejected before session validation.
func TestAuthMiddlewareRejectsMissingAuthorizationHeader(t *testing.T) {
	authService := &fakeAuthService{}
	recorder := performAuthMiddlewareRequest(authService, "")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusUnauthorized,
		)
	}
	if authService.verifyCalls != 0 {
		t.Fatalf(
			"expected VerifySession not to be called, got %d calls",
			authService.verifyCalls,
		)
	}
	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorUnauthorized {
		t.Fatalf(
			"unexpected error code: got %q want %q",
			response.Code,
			protocol.ProtocolErrorUnauthorized,
		)
	}
}

// TestAuthMiddlewareRejectsUnknownSession verifies that missing sessions are
// returned as unauthorized responses.
func TestAuthMiddlewareRejectsUnknownSession(t *testing.T) {
	authService := &fakeAuthService{err: services.ErrSessionNotFound}
	recorder := performAuthMiddlewareRequest(authService, "Bearer token")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusUnauthorized,
		)
	}
	if authService.verifyCalls != 1 || authService.verifyToken != "token" {
		t.Fatalf(
			"unexpected VerifySession call: calls=%d token=%q",
			authService.verifyCalls,
			authService.verifyToken,
		)
	}
	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorUnauthorized {
		t.Fatalf(
			"unexpected error code: got %q want %q",
			response.Code,
			protocol.ProtocolErrorUnauthorized,
		)
	}
}

// TestAuthMiddlewareHandlesSessionValidationFailure verifies that unexpected
// session validation failures are returned as internal server errors.
func TestAuthMiddlewareHandlesSessionValidationFailure(t *testing.T) {
	authService := &fakeAuthService{err: errors.New("session store failed")}
	recorder := performAuthMiddlewareRequest(authService, "Bearer token")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusInternalServerError,
		)
	}
	if authService.verifyCalls != 1 || authService.verifyToken != "token" {
		t.Fatalf(
			"unexpected VerifySession call: calls=%d token=%q",
			authService.verifyCalls,
			authService.verifyToken,
		)
	}
	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorSessionValidationFailed {
		t.Fatalf(
			"unexpected error code: got %q want %q",
			response.Code,
			protocol.ProtocolErrorSessionValidationFailed,
		)
	}
}

// TestAuthMiddlewareRejectsExpiredSession verifies that expired sessions are
// rejected after successful session validation.
func TestAuthMiddlewareRejectsExpiredSession(t *testing.T) {
	authService := &fakeAuthService{
		session: &services.SessionResult{
			User: services.User{
				ID:    uuid.New(),
				Email: "user@example.com",
			},
			AccessToken: "token",
			ExpiresAt:   time.Now().Add(-time.Minute),
		},
	}
	recorder := performAuthMiddlewareRequest(authService, "Bearer token")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusUnauthorized,
		)
	}
	if authService.verifyCalls != 1 || authService.verifyToken != "token" {
		t.Fatalf(
			"unexpected VerifySession call: calls=%d token=%q",
			authService.verifyCalls,
			authService.verifyToken,
		)
	}
	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorUnauthorized {
		t.Fatalf(
			"unexpected error code: got %q want %q",
			response.Code,
			protocol.ProtocolErrorUnauthorized,
		)
	}
}

// TestAuthMiddlewareAllowsValidSession verifies that a valid session is stored
// in request context and the protected handler is executed.
func TestAuthMiddlewareAllowsValidSession(t *testing.T) {
	userID := uuid.New()
	authService := &fakeAuthService{
		session: &services.SessionResult{
			User: services.User{
				ID:    userID,
				Email: "user@example.com",
			},
			AccessToken: "token",
			ExpiresAt:   time.Now().Add(time.Hour),
		},
	}
	recorder := performAuthMiddlewareRequest(authService, "Bearer token")

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d want %d; body=%s",
			recorder.Code,
			http.StatusOK,
			recorder.Body.String(),
		)
	}
	if authService.verifyCalls != 1 || authService.verifyToken != "token" {
		t.Fatalf(
			"unexpected VerifySession call: calls=%d token=%q",
			authService.verifyCalls,
			authService.verifyToken,
		)
	}

	var response struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.UserID != userID.String() {
		t.Fatalf("unexpected user ID: got %q want %q", response.UserID, userID)
	}
}
