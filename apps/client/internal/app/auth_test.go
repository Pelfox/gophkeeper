package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Pelfox/gophkeeper/apps/client/internal/api"
	"github.com/Pelfox/gophkeeper/shared/protocol"
)

type authClientMock struct {
	unsupportedClient

	login func(
		context.Context,
		protocol.LoginRequest,
	) (*api.PostAuthLoginResponse, error)
	register func(
		context.Context,
		protocol.RegisterRequest,
	) (*api.PostAuthRegisterResponse, error)
}

func (m *authClientMock) PostAuthLoginWithResponse(
	ctx context.Context,
	body api.PostAuthLoginJSONRequestBody,
	_ ...api.RequestEditorFn,
) (*api.PostAuthLoginResponse, error) {
	if m.login == nil {
		return nil, m.unsupportedMethod()
	}

	return m.login(ctx, body)
}

func (m *authClientMock) PostAuthRegisterWithResponse(
	ctx context.Context,
	body api.PostAuthRegisterJSONRequestBody,
	_ ...api.RequestEditorFn,
) (*api.PostAuthRegisterResponse, error) {
	if m.register == nil {
		return nil, m.unsupportedMethod()
	}

	return m.register(ctx, body)
}

// TestLoginStoresSession verifies that successful login persists the session
// details returned by the server.
func TestLoginStoresSession(t *testing.T) {
	expiresAt := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	var loginRequest protocol.LoginRequest
	client := &authClientMock{
		login: func(
			_ context.Context,
			request protocol.LoginRequest,
		) (*api.PostAuthLoginResponse, error) {
			loginRequest = request

			return &api.PostAuthLoginResponse{
				HTTPResponse: testHTTPResponse(200),
				JSON200: &protocol.LoginResponse{
					SessionCreatedResponse: protocol.SessionCreatedResponse{
						AccessToken: "access-token",
						ExpiresAt:   expiresAt,
					},
				},
			}, nil
		},
	}
	app := newTestApp(t, client)

	resp, err := app.Login(context.Background(), "user@example.com", "password")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if resp.AccessToken != "access-token" {
		t.Fatalf("unexpected access token: got %q", resp.AccessToken)
	}
	if loginRequest.Email != "user@example.com" {
		t.Fatalf("unexpected login email: got %q", loginRequest.Email)
	}
	if app.Config.AccessToken == nil || *app.Config.AccessToken != "access-token" {
		t.Fatalf("expected access token to be stored in config")
	}
	if app.Config.SessionExpiresAt == nil ||
		!app.Config.SessionExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected session expiration to be stored in config")
	}
}

// TestLoginFailureDoesNotMutateConfig verifies that rejected login keeps the
// existing session details untouched.
func TestLoginFailureDoesNotMutateConfig(t *testing.T) {
	existingToken := "existing-token"
	client := &authClientMock{
		login: func(
			context.Context,
			protocol.LoginRequest,
		) (*api.PostAuthLoginResponse, error) {
			return &api.PostAuthLoginResponse{
				HTTPResponse: testHTTPResponse(400),
				JSON400: &protocol.ProtocolError{
					Message: "bad credentials",
				},
			}, nil
		},
	}
	app := newTestApp(t, client)
	app.Config.AccessToken = &existingToken

	_, err := app.Login(context.Background(), "user@example.com", "bad")
	if err == nil {
		t.Fatal("expected Login to fail")
	}
	if app.Config.AccessToken == nil || *app.Config.AccessToken != existingToken {
		t.Fatalf("expected existing access token to be preserved")
	}
}

// TestLoginRequestError verifies that transport failures are returned with
// useful context.
func TestLoginRequestError(t *testing.T) {
	client := &authClientMock{
		login: func(
			context.Context,
			protocol.LoginRequest,
		) (*api.PostAuthLoginResponse, error) {
			return nil, errors.New("network failed")
		},
	}
	app := newTestApp(t, client)

	_, err := app.Login(context.Background(), "user@example.com", "password")
	if err == nil {
		t.Fatal("expected Login to fail")
	}
	if !strings.Contains(err.Error(), "failed to perform login request") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRegisterStoresSession verifies that successful registration persists the
// session details returned by the server.
func TestRegisterStoresSession(t *testing.T) {
	expiresAt := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	var registerRequest protocol.RegisterRequest
	client := &authClientMock{
		register: func(
			_ context.Context,
			request protocol.RegisterRequest,
		) (*api.PostAuthRegisterResponse, error) {
			registerRequest = request

			return &api.PostAuthRegisterResponse{
				HTTPResponse: testHTTPResponse(201),
				JSON201: &protocol.RegisterResponse{
					SessionCreatedResponse: protocol.SessionCreatedResponse{
						AccessToken: "access-token",
						ExpiresAt:   expiresAt,
					},
				},
			}, nil
		},
	}
	app := newTestApp(t, client)

	resp, err := app.Register(
		context.Background(),
		"user@example.com",
		"password",
	)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if resp.AccessToken != "access-token" {
		t.Fatalf("unexpected access token: got %q", resp.AccessToken)
	}
	if registerRequest.Email != "user@example.com" {
		t.Fatalf("unexpected register email: got %q", registerRequest.Email)
	}
	if app.Config.AccessToken == nil || *app.Config.AccessToken != "access-token" {
		t.Fatalf("expected access token to be stored in config")
	}
	if app.Config.SessionExpiresAt == nil ||
		!app.Config.SessionExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected session expiration to be stored in config")
	}
}

// TestRegisterFailureDoesNotMutateConfig verifies that failed registration
// keeps existing session details untouched.
func TestRegisterFailureDoesNotMutateConfig(t *testing.T) {
	existingToken := "existing-token"
	client := &authClientMock{
		register: func(
			context.Context,
			protocol.RegisterRequest,
		) (*api.PostAuthRegisterResponse, error) {
			return &api.PostAuthRegisterResponse{
				HTTPResponse: testHTTPResponse(409),
				JSON409: &protocol.ProtocolError{
					Message: "already exists",
				},
			}, nil
		},
	}
	app := newTestApp(t, client)
	app.Config.AccessToken = &existingToken

	_, err := app.Register(context.Background(), "user@example.com", "password")
	if err == nil {
		t.Fatal("expected Register to fail")
	}
	if app.Config.AccessToken == nil || *app.Config.AccessToken != existingToken {
		t.Fatalf("expected existing access token to be preserved")
	}
}
