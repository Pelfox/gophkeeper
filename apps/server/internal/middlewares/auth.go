package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Pelfox/gophkeeper/apps/server/internal"
	"github.com/Pelfox/gophkeeper/apps/server/internal/services"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware implements a basic middleware that verifies authorization of
// the request. It checks request's `Authorization` header and gets session for
// the given access token, if possible. Then it populates request's context
// with the received session.
func AuthMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := strings.TrimPrefix(ctx.GetHeader("authorization"), "Bearer ")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorUnauthorized,
				Message: "Authorization header wasn't provided.",
			})
			return
		}

		session, err := authService.VerifySession(ctx.Request.Context(), token)
		if err != nil {
			if errors.Is(err, services.ErrSessionNotFound) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, protocol.ProtocolError{
					Code:    protocol.ProtocolErrorUnauthorized,
					Message: "Unauthorized.",
				})
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorSessionValidationFailed,
				Message: "Something went wrong.",
			})
			return
		}

		// Verifying that session hasn't expired yet.
		if time.Now().After(session.ExpiresAt) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorUnauthorized,
				Message: "Session has expired.",
			})
			return
		}

		requestCtx := internal.WithSession(ctx.Request.Context(), session)
		ctx.Request = ctx.Request.WithContext(requestCtx)

		ctx.Next()
	}
}
