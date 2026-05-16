package controllers

import (
	"errors"
	"net/http"

	"github.com/Pelfox/gophkeeper/apps/server/internal/services"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/gin-gonic/gin"
)

// AuthController is a controller for all auth-related requests.
type AuthController struct {
	authService services.AuthService
}

// NewAuthController creates a new auth controller, backed by the given auth
// service.
func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// RegisterRoutes registers all routes for this controller using the provided
// group.
func (c *AuthController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/register", c.register)
	group.POST("/login", c.login)
}

func (c *AuthController) register(ctx *gin.Context) {
	var request protocol.RegisterRequest
	if !bindAndValidate(ctx, &request) {
		return
	}

	session, err := c.authService.Register(ctx.Request.Context(), services.RegisterInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, services.ErrDuplicateUser) {
			ctx.JSON(http.StatusConflict, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorDuplicateUser,
				Message: "User with the same email address already exist.",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorRegistrationFailed,
			Message: "Something went wrong.",
		})
		return
	}

	ctx.JSON(http.StatusCreated, protocol.RegisterResponse{
		SessionCreatedResponse: protocol.SessionCreatedResponse{
			User: protocol.ProtocolUser{
				ID:        session.User.ID,
				Email:     session.User.Email,
				CreatedAt: session.User.CreatedAt,
				UpdatedAt: session.User.UpdatedAt,
			},
			AccessToken: session.AccessToken,
			ExpiresAt:   session.ExpiresAt,
		},
	})
}

func (c *AuthController) login(ctx *gin.Context) {
	var request protocol.RegisterRequest
	if !bindAndValidate(ctx, &request) {
		return
	}

	session, err := c.authService.Login(ctx.Request.Context(), services.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, services.ErrLoginFailed) {
			ctx.JSON(http.StatusBadRequest, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorLoginUnsuccessful,
				Message: "User not found or password is invalid.",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorLoginFailed,
			Message: "Something went wrong.",
		})
		return
	}

	ctx.JSON(http.StatusOK, protocol.LoginResponse{
		SessionCreatedResponse: protocol.SessionCreatedResponse{
			User: protocol.ProtocolUser{
				ID:        session.User.ID,
				Email:     session.User.Email,
				CreatedAt: session.User.CreatedAt,
				UpdatedAt: session.User.UpdatedAt,
			},
			AccessToken: session.AccessToken,
			ExpiresAt:   session.ExpiresAt,
		},
	})
}
