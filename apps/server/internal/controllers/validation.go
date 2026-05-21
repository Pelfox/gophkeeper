package controllers

import (
	"errors"
	"net/http"

	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func bindAndValidate[T any](ctx *gin.Context, dst *T) bool {
	if err := ctx.ShouldBindJSON(dst); err != nil {
		if validationErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
			// TODO: Improve response with actual validator error in details.
			ctx.JSON(http.StatusBadRequest, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorValidationFailed,
				Message: validationErrs.Error(),
				Details: map[string]any{},
			})
			return false
		}

		ctx.JSON(http.StatusUnprocessableEntity, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidRequest,
			Message: "Received request is invalid or malformed.",
		})
		return false
	}

	return true
}
