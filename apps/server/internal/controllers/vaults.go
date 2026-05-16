package controllers

import (
	"errors"
	"net/http"

	"github.com/Pelfox/gophkeeper/apps/server/internal"
	"github.com/Pelfox/gophkeeper/apps/server/internal/services"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VaultsController handles all vaults-related requests.
type VaultsController struct {
	vaultsService services.VaultsService
}

// NewVaultsController creates a new vaults controller, backed by a vaults
// service.
func NewVaultsController(vaultsService services.VaultsService) *VaultsController {
	return &VaultsController{
		vaultsService: vaultsService,
	}
}

// RegisterRoutes registers all vaults routes.
func (c *VaultsController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/", c.create)
	group.GET("/", c.list)
	group.DELETE("/:id", c.delete)
}

// create godoc
// @Summary Creates a new vault.
// @Tags vaults
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body protocol.CreateVaultRequest true "Vault creation payload"
// @Success 201 {object} protocol.CreateVaultResponse
// @Failure 400 {object} protocol.ProtocolError
// @Failure 401 {object} protocol.ProtocolError
// @Failure 422 {object} protocol.ProtocolError
// @Failure 500 {object} protocol.ProtocolError
// @Router /vaults/ [post]
func (c *VaultsController) create(ctx *gin.Context) {
	var request protocol.CreateVaultRequest
	if !bindAndValidate(ctx, &request) {
		return
	}

	session, ok := internal.SessionFromContext(ctx.Request.Context())
	if !ok {
		return
	}

	vault, err := c.vaultsService.Create(ctx.Request.Context(), services.CreateVaultInput{
		OwnerID:               session.User.ID,
		Name:                  request.Name,
		EncryptionSalt:        request.EncryptionSalt,
		EncryptionNonce:       request.EncryptionNonce,
		EncryptedMasterKey:    request.EncryptedMasterKey,
		EncryptionTimeCost:    request.EncryptionTimeCost,
		EncryptionMemoryCost:  request.EncryptionMemoryCost,
		EncryptionParallelism: request.EncryptionParallelism,
		EncryptionKeySize:     request.EncryptionKeySize,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorVaultCreationFailed,
			Message: "Something went wrong.",
		})
		return
	}

	ctx.JSON(http.StatusCreated, protocol.CreateVaultResponse{
		ProtocolVault: protocol.ProtocolVault{
			ID:        vault.ID,
			Name:      vault.Name,
			CreatedAt: vault.CreatedAt,
			UpdatedAt: vault.UpdatedAt,
		},
	})
}

// list godoc
// @Summary List user's vaults.
// @Tags vaults
// @Produce json
// @Security BearerAuth
// @Success 200 {array} protocol.ProtocolVault
// @Failure 401 {object} protocol.ProtocolError
// @Failure 500 {object} protocol.ProtocolError
// @Router /vaults/ [get]
func (c *VaultsController) list(ctx *gin.Context) {
	session, ok := internal.SessionFromContext(ctx.Request.Context())
	if !ok {
		return
	}

	vaults, err := c.vaultsService.GetForUser(
		ctx.Request.Context(),
		session.User.ID,
	)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorVaultsRetrievalFailed,
			Message: "Something went wrong.",
		})
		return
	}

	userVaults := make([]protocol.ProtocolVault, len(vaults))
	for i, vault := range vaults {
		userVaults[i] = protocol.ProtocolVault{
			ID:        vault.ID,
			Name:      vault.Name,
			CreatedAt: vault.CreatedAt,
			UpdatedAt: vault.UpdatedAt,
		}
	}

	ctx.JSON(http.StatusOK, userVaults)
}

// delete godoc
// @Summary Deletes a specific vault.
// @Tags vaults
// @Produce json
// @Security BearerAuth
// @Param id path string true "Vault ID"
// @Success 204
// @Failure 400 {object} protocol.ProtocolError
// @Failure 401 {object} protocol.ProtocolError
// @Failure 404 {object} protocol.ProtocolError
// @Router /vaults/{id} [delete]
func (c *VaultsController) delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidID,
			Message: "Provided vault ID is invalid.",
		})
		return
	}

	session, ok := internal.SessionFromContext(ctx.Request.Context())
	if !ok {
		return
	}

	err = c.vaultsService.Delete(ctx.Request.Context(), id, session.User.ID)
	if err != nil {
		if errors.Is(err, services.ErrVaultNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorVaultNotFound,
				Message: "Vault with the given ID was not found.",
			})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorVaultDeletionFailed,
			Message: "Something went wrong.",
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}
