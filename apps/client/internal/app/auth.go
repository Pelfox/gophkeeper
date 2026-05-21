package app

import (
	"context"
	"fmt"

	"github.com/Pelfox/gophkeeper/apps/client/internal/config"
	"github.com/Pelfox/gophkeeper/shared/protocol"
)

// Login authenticates the user and persists the received session details.
func (a *App) Login(
	ctx context.Context,
	email string,
	password string,
) (*protocol.LoginResponse, error) {
	resp, err := a.Client.PostAuthLoginWithResponse(ctx, protocol.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to perform login request: %w", err)
	}

	if resp.JSON200 != nil {
		a.Config.AccessToken = &resp.JSON200.AccessToken
		a.Config.SessionExpiresAt = &resp.JSON200.ExpiresAt

		if err := config.WriteConfig(a.Config); err != nil {
			return nil, fmt.Errorf("failed to write updated config: %w", err)
		}

		return resp.JSON200, nil
	}

	if resp.JSON400 != nil {
		return nil, fmt.Errorf("login failed: %s", resp.JSON400.Message)
	}

	if resp.JSON422 != nil {
		return nil, fmt.Errorf("invalid request: %s", resp.JSON422.Message)
	}

	if resp.JSON500 != nil {
		return nil, fmt.Errorf("server error: %s", resp.JSON500.Message)
	}

	return nil, fmt.Errorf("unexpected response from server: %s", resp.Status())
}

// Register creates a new account and persists the received session details.
func (a *App) Register(
	ctx context.Context,
	email string,
	password string,
) (*protocol.RegisterResponse, error) {
	resp, err := a.Client.PostAuthRegisterWithResponse(ctx, protocol.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to perform register request: %w", err)
	}

	if resp.JSON201 != nil {
		a.Config.AccessToken = &resp.JSON201.AccessToken
		a.Config.SessionExpiresAt = &resp.JSON201.ExpiresAt

		if err := config.WriteConfig(a.Config); err != nil {
			return nil, fmt.Errorf("failed to write updated config: %w", err)
		}

		return resp.JSON201, nil
	}

	if resp.JSON400 != nil {
		return nil, fmt.Errorf("registration failed: %s", resp.JSON400.Message)
	}

	if resp.JSON409 != nil {
		return nil, fmt.Errorf("registration failed: %s", resp.JSON409.Message)
	}

	if resp.JSON422 != nil {
		return nil, fmt.Errorf("invalid request: %s", resp.JSON422.Message)
	}

	if resp.JSON500 != nil {
		return nil, fmt.Errorf("server error: %s", resp.JSON500.Message)
	}

	return nil, fmt.Errorf("unexpected response from server: %s", resp.Status())
}
