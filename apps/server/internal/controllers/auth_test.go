package controllers

import (
	"net/http"
	"testing"

	"github.com/Pelfox/gophkeeper/apps/server/internal/services"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/google/uuid"
)

// TestAuthControllerRegister verifies that register maps a valid protocol
// request to the auth service and returns a session response.
func TestAuthControllerRegister(t *testing.T) {
	userID := uuid.New()
	authService := &fakeAuthService{registerResult: testSession(userID)}
	router := newAuthRouter(authService)

	recorder := performJSONRequest(
		router,
		http.MethodPost,
		"/auth/register",
		protocol.RegisterRequest{
			Email:    "user@example.com",
			Password: "password",
		},
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusCreated,
		)
	}
	if authService.registerCalls != 1 ||
		authService.registerInput.Email != "user@example.com" ||
		authService.registerInput.Password != "password" {
		t.Fatalf("unexpected register input: %#v", authService.registerInput)
	}

	response := decodeResponse[protocol.RegisterResponse](t, recorder)
	if response.User.ID != userID || response.AccessToken != "access-token" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

// TestAuthControllerRegisterDuplicateUser verifies that duplicate-user service
// errors are mapped to conflict responses.
func TestAuthControllerRegisterDuplicateUser(t *testing.T) {
	authService := &fakeAuthService{registerErr: services.ErrDuplicateUser}
	router := newAuthRouter(authService)

	recorder := performJSONRequest(
		router,
		http.MethodPost,
		"/auth/register",
		protocol.RegisterRequest{
			Email:    "user@example.com",
			Password: "password",
		},
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusConflict,
		)
	}

	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorDuplicateUser {
		t.Fatalf(
			"unexpected error code: got %s want %s",
			response.Code,
			protocol.ProtocolErrorDuplicateUser,
		)
	}
}

// TestAuthControllerLogin verifies that login maps a valid protocol request to
// the auth service and returns a session response.
func TestAuthControllerLogin(t *testing.T) {
	userID := uuid.New()
	authService := &fakeAuthService{loginResult: testSession(userID)}
	router := newAuthRouter(authService)

	recorder := performJSONRequest(
		router,
		http.MethodPost,
		"/auth/login",
		protocol.LoginRequest{
			Email:    "user@example.com",
			Password: "password",
		},
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusOK,
		)
	}
	if authService.loginCalls != 1 ||
		authService.loginInput.Email != "user@example.com" ||
		authService.loginInput.Password != "password" {
		t.Fatalf("unexpected login input: %#v", authService.loginInput)
	}

	response := decodeResponse[protocol.LoginResponse](t, recorder)
	if response.User.ID != userID || response.AccessToken != "access-token" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

// TestAuthControllerLoginUnsuccessful verifies that login failures from the
// service are mapped to bad request responses.
func TestAuthControllerLoginUnsuccessful(t *testing.T) {
	authService := &fakeAuthService{loginErr: services.ErrLoginFailed}
	router := newAuthRouter(authService)

	recorder := performJSONRequest(
		router,
		http.MethodPost,
		"/auth/login",
		protocol.LoginRequest{
			Email:    "user@example.com",
			Password: "password",
		},
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"unexpected status: got %d want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}

	response := decodeProtocolError(t, recorder)
	if response.Code != protocol.ProtocolErrorLoginUnsuccessful {
		t.Fatalf(
			"unexpected error code: got %s want %s",
			response.Code,
			protocol.ProtocolErrorLoginUnsuccessful,
		)
	}
}
