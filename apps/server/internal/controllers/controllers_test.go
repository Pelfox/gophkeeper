package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pelfox/gophkeeper/apps/server/internal"
	"github.com/Pelfox/gophkeeper/apps/server/internal/services"
	"github.com/Pelfox/gophkeeper/shared/protocol"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
}

func testSession(userID uuid.UUID) *services.SessionResult {
	return &services.SessionResult{
		User: services.User{
			ID:        userID,
			Email:     "user@example.com",
			CreatedAt: fixedTime(),
			UpdatedAt: fixedTime().Add(time.Minute),
		},
		AccessToken: "access-token",
		ExpiresAt:   fixedTime().Add(time.Hour),
	}
}

func performJSONRequest(
	router http.Handler,
	method string,
	path string,
	body any,
) *httptest.ResponseRecorder {
	var requestBody bytes.Buffer
	switch value := body.(type) {
	case nil:
	case string:
		requestBody.WriteString(value)
	default:
		if err := json.NewEncoder(&requestBody).Encode(value); err != nil {
			panic(err)
		}
	}

	request := httptest.NewRequest(method, path, &requestBody)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder
}

func decodeResponse[T any](
	t *testing.T,
	recorder *httptest.ResponseRecorder,
) T {
	t.Helper()

	var response T
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf(
			"failed to decode response: %v; body=%s",
			err,
			recorder.Body.String(),
		)
	}

	return response
}

func decodeProtocolError(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
) protocol.ProtocolError {
	t.Helper()

	return decodeResponse[protocol.ProtocolError](t, recorder)
}

func newAuthRouter(authService services.AuthService) *gin.Engine {
	router := gin.New()
	NewAuthController(authService).RegisterRoutes(router.Group("/auth"))
	return router
}

func newVaultsRouter(
	vaultsService services.VaultsService,
	vaultItemsService services.VaultItemsService,
	session *services.SessionResult,
) *gin.Engine {
	router := gin.New()

	group := router.Group("/vaults")
	group.Use(func(ctx *gin.Context) {
		request := ctx.Request.WithContext(
			internal.WithSession(ctx.Request.Context(), session),
		)
		ctx.Request = request
		ctx.Next()
	})

	NewVaultsController(vaultsService, vaultItemsService).RegisterRoutes(group)
	return router
}

func newVaultsRouterWithoutSession(
	vaultsService services.VaultsService,
	vaultItemsService services.VaultItemsService,
) *gin.Engine {
	router := gin.New()
	NewVaultsController(vaultsService, vaultItemsService).
		RegisterRoutes(router.Group("/vaults"))
	return router
}

type fakeAuthService struct {
	registerCalls  int
	registerInput  services.RegisterInput
	registerResult *services.SessionResult
	registerErr    error

	loginCalls  int
	loginInput  services.LoginInput
	loginResult *services.SessionResult
	loginErr    error
}

func (s *fakeAuthService) Register(
	_ context.Context,
	input services.RegisterInput,
) (*services.SessionResult, error) {
	s.registerCalls++
	s.registerInput = input
	return s.registerResult, s.registerErr
}

func (s *fakeAuthService) Login(
	_ context.Context,
	input services.LoginInput,
) (*services.SessionResult, error) {
	s.loginCalls++
	s.loginInput = input
	return s.loginResult, s.loginErr
}

func (s *fakeAuthService) VerifySession(
	context.Context,
	string,
) (*services.SessionResult, error) {
	panic("unexpected VerifySession call")
}

type fakeVaultsService struct {
	createCalls  int
	createInput  services.CreateVaultInput
	createResult *services.VaultResult
	createErr    error

	getForUserCalls int
	getForUserID    uuid.UUID
	getForUserItems []services.VaultResult
	getForUserErr   error

	getKeyringCalls int
	getKeyringUser  uuid.UUID
	getKeyringVault uuid.UUID
	getKeyringItem  *services.KeyringResult
	getKeyringErr   error

	updateCalls int
	updateVault uuid.UUID
	updateUser  uuid.UUID
	updateInput services.UpdateVaultInput
	updateItem  *services.VaultResult
	updateErr   error

	deleteCalls int
	deleteVault uuid.UUID
	deleteUser  uuid.UUID
	deleteErr   error
}

func (s *fakeVaultsService) Create(
	_ context.Context,
	input services.CreateVaultInput,
) (*services.VaultResult, error) {
	s.createCalls++
	s.createInput = input
	return s.createResult, s.createErr
}

func (s *fakeVaultsService) GetForUser(
	_ context.Context,
	userID uuid.UUID,
) ([]services.VaultResult, error) {
	s.getForUserCalls++
	s.getForUserID = userID
	return s.getForUserItems, s.getForUserErr
}

func (s *fakeVaultsService) GetKeyring(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
) (*services.KeyringResult, error) {
	s.getKeyringCalls++
	s.getKeyringUser = userID
	s.getKeyringVault = vaultID
	return s.getKeyringItem, s.getKeyringErr
}

func (s *fakeVaultsService) Update(
	_ context.Context,
	id uuid.UUID,
	userID uuid.UUID,
	input services.UpdateVaultInput,
) (*services.VaultResult, error) {
	s.updateCalls++
	s.updateVault = id
	s.updateUser = userID
	s.updateInput = input
	return s.updateItem, s.updateErr
}

func (s *fakeVaultsService) Delete(
	_ context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	s.deleteCalls++
	s.deleteVault = id
	s.deleteUser = userID
	return s.deleteErr
}

type fakeVaultItemsService struct {
	createCalls int
	createUser  uuid.UUID
	createVault uuid.UUID
	createInput services.VaultItemInput
	createItem  *services.VaultItemResult
	createErr   error

	updateCalls int
	updateUser  uuid.UUID
	updateVault uuid.UUID
	updateID    uuid.UUID
	updateInput services.VaultItemInput
	updateItem  *services.VaultItemResult
	updateErr   error

	getByIDCalls int
	getByIDUser  uuid.UUID
	getByIDVault uuid.UUID
	getByIDID    uuid.UUID
	getByIDItem  *services.VaultItemResult
	getByIDErr   error

	getForVaultCalls int
	getForVaultUser  uuid.UUID
	getForVaultID    uuid.UUID
	getForVaultItems []services.VaultItemResult
	getForVaultErr   error

	deleteCalls int
	deleteUser  uuid.UUID
	deleteVault uuid.UUID
	deleteID    uuid.UUID
	deleteErr   error
}

func (s *fakeVaultItemsService) Create(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	input services.VaultItemInput,
) (*services.VaultItemResult, error) {
	s.createCalls++
	s.createUser = userID
	s.createVault = vaultID
	s.createInput = input
	return s.createItem, s.createErr
}

func (s *fakeVaultItemsService) Update(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	itemID uuid.UUID,
	input services.VaultItemInput,
) (*services.VaultItemResult, error) {
	s.updateCalls++
	s.updateUser = userID
	s.updateVault = vaultID
	s.updateID = itemID
	s.updateInput = input
	return s.updateItem, s.updateErr
}

func (s *fakeVaultItemsService) GetByID(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	id uuid.UUID,
) (*services.VaultItemResult, error) {
	s.getByIDCalls++
	s.getByIDUser = userID
	s.getByIDVault = vaultID
	s.getByIDID = id
	return s.getByIDItem, s.getByIDErr
}

func (s *fakeVaultItemsService) GetForVault(
	_ context.Context,
	userID uuid.UUID,
	id uuid.UUID,
) ([]services.VaultItemResult, error) {
	s.getForVaultCalls++
	s.getForVaultUser = userID
	s.getForVaultID = id
	return s.getForVaultItems, s.getForVaultErr
}

func (s *fakeVaultItemsService) Delete(
	_ context.Context,
	userID uuid.UUID,
	vaultID uuid.UUID,
	id uuid.UUID,
) error {
	s.deleteCalls++
	s.deleteUser = userID
	s.deleteVault = vaultID
	s.deleteID = id
	return s.deleteErr
}
