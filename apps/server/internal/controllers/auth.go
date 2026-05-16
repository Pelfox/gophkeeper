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

// register godoc
// @Summary Registers a new user.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body protocol.RegisterRequest true "Registration payload"
// @Success 201 {object} protocol.RegisterResponse
// @Failure 400 {object} protocol.ProtocolError
// @Failure 409 {object} protocol.ProtocolError
// @Failure 422 {object} protocol.ProtocolError
// @Failure 500 {object} protocol.ProtocolError
// @Router /auth/register [post]
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
			ctx.AbortWithStatusJSON(http.StatusConflict, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorDuplicateUser,
				Message: "User with the same email address already exist.",
			})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
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

// login godoc
// @Summary Logins user into their account.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body protocol.LoginRequest true "Login payload"
// @Success 200 {object} protocol.LoginResponse
// @Failure 400 {object} protocol.ProtocolError
// @Failure 422 {object} protocol.ProtocolError
// @Failure 500 {object} protocol.ProtocolError
// @Router /auth/login [post]
func (c *AuthController) login(ctx *gin.Context) {
	var request protocol.LoginRequest
	if !bindAndValidate(ctx, &request) {
		return
	}

	session, err := c.authService.Login(ctx.Request.Context(), services.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, services.ErrLoginFailed) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorLoginUnsuccessful,
				Message: "User not found or password is invalid.",
			})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
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
