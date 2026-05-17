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
	vaultsService     services.VaultsService
	vaultItemsService services.VaultItemsService
}

// NewVaultsController creates a new vaults controller, backed by a vaults
// service.
func NewVaultsController(
	vaultsService services.VaultsService,
	vaultItemsService services.VaultItemsService,
) *VaultsController {
	return &VaultsController{
		vaultsService:     vaultsService,
		vaultItemsService: vaultItemsService,
	}
}

// RegisterRoutes registers all vaults routes.
func (c *VaultsController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/", c.create)
	group.GET("/", c.list)
	group.PATCH("/:id", c.update)
	group.DELETE("/:id", c.delete)
	group.POST("/:id/items", c.createItem)
	group.GET("/:id/items", c.listItems)
	group.GET("/:id/items/:item_id", c.getItem)
	group.PATCH("/:id/items/:item_id", c.updateItem)
	group.DELETE("/:id/items/:item_id", c.deleteItem)
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

// update godoc
// @Summary Updates a specific vault.
// @Tags vaults
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Vault ID"
// @Param request body protocol.UpdateVaultRequest true "Vault update payload"
// @Success 200 {object} protocol.UpdateVaultResponse
// @Failure 400 {object} protocol.ProtocolError
// @Failure 401 {object} protocol.ProtocolError
// @Failure 404 {object} protocol.ProtocolError
// @Failure 422 {object} protocol.ProtocolError
// @Failure 500 {object} protocol.ProtocolError
// @Router /vaults/{id} [patch]
func (c *VaultsController) update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidID,
			Message: "Provided vault ID is invalid.",
		})
		return
	}

	var request protocol.UpdateVaultRequest
	if !bindAndValidate(ctx, &request) {
		return
	}

	session, ok := internal.SessionFromContext(ctx.Request.Context())
	if !ok {
		return
	}

	vault, err := c.vaultsService.Update(ctx.Request.Context(), id, session.User.ID, services.UpdateVaultInput{
		Name: request.Name,
	})
	if err != nil {
		if errors.Is(err, services.ErrVaultNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorVaultNotFound,
				Message: "Vault with the given ID was not found.",
			})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorVaultUpdateFailed,
			Message: "Something went wrong.",
		})
		return
	}

	ctx.JSON(http.StatusOK, protocol.UpdateVaultResponse{
		ProtocolVault: protocol.ProtocolVault{
			ID:        vault.ID,
			Name:      vault.Name,
			CreatedAt: vault.CreatedAt,
			UpdatedAt: vault.UpdatedAt,
		},
	})
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

// createItem godoc
// @Summary Creates a new vault item.
// @Tags vault_items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Vault ID"
// @Param request body protocol.CreateVaultItemRequest true "Vault item creation payload"
// @Success 201 {object} protocol.CreateVaultItemResponse
// @Failure 400 {object} protocol.ProtocolError
// @Failure 401 {object} protocol.ProtocolError
// @Failure 404 {object} protocol.ProtocolError
// @Failure 422 {object} protocol.ProtocolError
// @Failure 500 {object} protocol.ProtocolError
// @Router /vaults/{id}/items [post]
func (c *VaultsController) createItem(ctx *gin.Context) {
	vaultID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidID,
			Message: "Provided vault ID is invalid.",
		})
		return
	}

	var request protocol.CreateVaultItemRequest
	if !bindAndValidate(ctx, &request) {
		return
	}

	session, ok := internal.SessionFromContext(ctx.Request.Context())
	if !ok {
		return
	}

	vaultItem, err := c.vaultItemsService.Create(ctx.Request.Context(), session.User.ID, vaultID, services.VaultItemInput{
		KeySalt:    request.KeySalt,
		ItemNonce:  request.ItemNonce,
		Ciphertext: request.Ciphertext,
	})
	if err != nil {
		if errors.Is(err, services.ErrVaultNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorVaultNotFound,
				Message: "Vault with the given ID was not found.",
			})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorVaultItemCreationFailed,
			Message: "Something went wrong.",
		})
		return
	}

	ctx.JSON(http.StatusCreated, protocol.CreateVaultItemResponse{
		ProtocolVaultItem: protocol.ProtocolVaultItem{
			ID:         vaultItem.ID,
			VaultID:    vaultItem.VaultID,
			KeySalt:    vaultItem.KeySalt,
			ItemNonce:  vaultItem.ItemNonce,
			Ciphertext: vaultItem.Ciphertext,
			CreatedAt:  vaultItem.CreatedAt,
			UpdatedAt:  vaultItem.UpdatedAt,
		},
	})
}

// listItems godoc
// @Summary Lists vault items.
// @Tags vault_items
// @Produce json
// @Security BearerAuth
// @Param id path string true "Vault ID"
// @Success 200 {array} protocol.ProtocolVaultItem
// @Failure 400 {object} protocol.ProtocolError
// @Failure 401 {object} protocol.ProtocolError
// @Failure 500 {object} protocol.ProtocolError
// @Router /vaults/{id}/items [get]
func (c *VaultsController) listItems(ctx *gin.Context) {
	vaultID, err := uuid.Parse(ctx.Param("id"))
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

	vaultItems, err := c.vaultItemsService.GetForVault(ctx.Request.Context(), session.User.ID, vaultID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorVaultItemsRetrievalFailed,
			Message: "Something went wrong.",
		})
		return
	}

	userVaultItems := make([]protocol.ProtocolVaultItem, len(vaultItems))
	for i, vaultItem := range vaultItems {
		userVaultItems[i] = protocol.ProtocolVaultItem{
			ID:         vaultItem.ID,
			VaultID:    vaultItem.VaultID,
			KeySalt:    vaultItem.KeySalt,
			ItemNonce:  vaultItem.ItemNonce,
			Ciphertext: vaultItem.Ciphertext,
			CreatedAt:  vaultItem.CreatedAt,
			UpdatedAt:  vaultItem.UpdatedAt,
		}
	}

	ctx.JSON(http.StatusOK, userVaultItems)
}

// getItem godoc
// @Summary Retrieves a vault item.
// @Tags vault_items
// @Produce json
// @Security BearerAuth
// @Param id path string true "Vault ID"
// @Param item_id path string true "Vault item ID"
// @Success 200 {object} protocol.GetVaultItemResponse
// @Failure 400 {object} protocol.ProtocolError
// @Failure 401 {object} protocol.ProtocolError
// @Failure 404 {object} protocol.ProtocolError
// @Failure 500 {object} protocol.ProtocolError
// @Router /vaults/{id}/items/{item_id} [get]
func (c *VaultsController) getItem(ctx *gin.Context) {
	vaultID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidID,
			Message: "Provided vault ID is invalid.",
		})
		return
	}

	itemID, err := uuid.Parse(ctx.Param("item_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidID,
			Message: "Provided vault item ID is invalid.",
		})
		return
	}

	session, ok := internal.SessionFromContext(ctx.Request.Context())
	if !ok {
		return
	}

	vaultItem, err := c.vaultItemsService.GetByID(ctx.Request.Context(), session.User.ID, vaultID, itemID)
	if err != nil {
		if errors.Is(err, services.ErrVaultItemNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorVaultItemNotFound,
				Message: "Vault item with the given ID was not found.",
			})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorVaultItemsRetrievalFailed,
			Message: "Something went wrong.",
		})
		return
	}

	ctx.JSON(http.StatusOK, protocol.GetVaultItemResponse{
		ProtocolVaultItem: protocol.ProtocolVaultItem{
			ID:         vaultItem.ID,
			VaultID:    vaultItem.VaultID,
			KeySalt:    vaultItem.KeySalt,
			ItemNonce:  vaultItem.ItemNonce,
			Ciphertext: vaultItem.Ciphertext,
			CreatedAt:  vaultItem.CreatedAt,
			UpdatedAt:  vaultItem.UpdatedAt,
		},
	})
}

// updateItem godoc
// @Summary Updates a vault item.
// @Tags vault_items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Vault ID"
// @Param item_id path string true "Vault item ID"
// @Param request body protocol.UpdateVaultItemRequest true "Vault item update payload"
// @Success 200 {object} protocol.UpdateVaultItemResponse
// @Failure 400 {object} protocol.ProtocolError
// @Failure 401 {object} protocol.ProtocolError
// @Failure 404 {object} protocol.ProtocolError
// @Failure 422 {object} protocol.ProtocolError
// @Failure 500 {object} protocol.ProtocolError
// @Router /vaults/{id}/items/{item_id} [patch]
func (c *VaultsController) updateItem(ctx *gin.Context) {
	vaultID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidID,
			Message: "Provided vault ID is invalid.",
		})
		return
	}

	itemID, err := uuid.Parse(ctx.Param("item_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidID,
			Message: "Provided vault item ID is invalid.",
		})
		return
	}

	var request protocol.UpdateVaultItemRequest
	if !bindAndValidate(ctx, &request) {
		return
	}

	session, ok := internal.SessionFromContext(ctx.Request.Context())
	if !ok {
		return
	}

	vaultItem, err := c.vaultItemsService.Update(ctx.Request.Context(), session.User.ID, vaultID, itemID, services.VaultItemInput{
		KeySalt:    request.KeySalt,
		ItemNonce:  request.ItemNonce,
		Ciphertext: request.Ciphertext,
	})
	if err != nil {
		if errors.Is(err, services.ErrVaultItemNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorVaultItemNotFound,
				Message: "Vault item with the given ID was not found.",
			})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorVaultItemUpdateFailed,
			Message: "Something went wrong.",
		})
		return
	}

	ctx.JSON(http.StatusOK, protocol.UpdateVaultItemResponse{
		ProtocolVaultItem: protocol.ProtocolVaultItem{
			ID:         vaultItem.ID,
			VaultID:    vaultItem.VaultID,
			KeySalt:    vaultItem.KeySalt,
			ItemNonce:  vaultItem.ItemNonce,
			Ciphertext: vaultItem.Ciphertext,
			CreatedAt:  vaultItem.CreatedAt,
			UpdatedAt:  vaultItem.UpdatedAt,
		},
	})
}

// deleteItem godoc
// @Summary Deletes a vault item.
// @Tags vault_items
// @Produce json
// @Security BearerAuth
// @Param id path string true "Vault ID"
// @Param item_id path string true "Vault item ID"
// @Success 204
// @Failure 400 {object} protocol.ProtocolError
// @Failure 401 {object} protocol.ProtocolError
// @Failure 404 {object} protocol.ProtocolError
// @Router /vaults/{id}/items/{item_id} [delete]
func (c *VaultsController) deleteItem(ctx *gin.Context) {
	vaultID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidID,
			Message: "Provided vault ID is invalid.",
		})
		return
	}

	itemID, err := uuid.Parse(ctx.Param("item_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorInvalidID,
			Message: "Provided vault item ID is invalid.",
		})
		return
	}

	session, ok := internal.SessionFromContext(ctx.Request.Context())
	if !ok {
		return
	}

	err = c.vaultItemsService.Delete(ctx.Request.Context(), session.User.ID, vaultID, itemID)
	if err != nil {
		if errors.Is(err, services.ErrVaultItemNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, protocol.ProtocolError{
				Code:    protocol.ProtocolErrorVaultItemNotFound,
				Message: "Vault item with the given ID was not found.",
			})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ProtocolError{
			Code:    protocol.ProtocolErrorVaultItemDeletionFailed,
			Message: "Something went wrong.",
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}
