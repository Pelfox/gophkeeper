package protocol

// ProtocolErrorCode is a type that's used for errors' codes.
type ProtocolErrorCode string

const (
	// ProtocolErrorValidationFailed is a generic error code that is returned
	// when validation of the request has failed.
	ProtocolErrorValidationFailed ProtocolErrorCode = "VALIDATION_ERROR"
	// ProtocolErrorInvalidRequest is a generic error code that is returned
	// when request is invalid (failed to parse, etc.)
	ProtocolErrorInvalidRequest ProtocolErrorCode = "INVALID_REQUEST"
	// ProtocolErrorInvalidID is returned when given ID for the resource is
	// invalid or unparseable.
	ProtocolErrorInvalidID ProtocolErrorCode = "INVALID_ID"

	// ProtocolErrorDuplicateUser is returned when user with the same email
	// already registered.
	ProtocolErrorDuplicateUser ProtocolErrorCode = "DUPLICATE_USER"
	// ProtocolErrorRegistrationFailed is returned when service was unable to
	// register user or create a new session for them.
	ProtocolErrorRegistrationFailed ProtocolErrorCode = "REGISTRATION_FAILED"
	// ProtocolErrorLoginUnsuccessful is returned when login request was
	// unsuccessful (either user not found or password is invalid).
	ProtocolErrorLoginUnsuccessful ProtocolErrorCode = "LOGIN_UNSUCCESSFUL"
	// ProtocolErrorLoginFailed is returned when service was unable to process
	// login request from the user.
	ProtocolErrorLoginFailed ProtocolErrorCode = "LOGIN_FAILED"
	// ProtocolErrorUnauthorized is returned when user was not authorized to
	// perform this request. This is a generic error code.
	ProtocolErrorUnauthorized ProtocolErrorCode = "UNAUTHORIZED"
	// ProtocolErrorSessionValidationFailed is returned when service was unable
	// to validate the session.
	ProtocolErrorSessionValidationFailed ProtocolErrorCode = "SESSION_VALIDATION_FAILED"

	// ProtocolErrorVaultCreationFailed is returned when service was not able
	// to create a new vault for the user.
	ProtocolErrorVaultCreationFailed ProtocolErrorCode = "VAULT_CREATION_FAILED"
	// ProtocolErrorVaultsRetrievalFailed is returned when service was not able
	// to retrieve user's vaults.
	ProtocolErrorVaultsRetrievalFailed ProtocolErrorCode = "VAULTS_RETRIEVAL_FAILED"
	// ProtocolErrorVaultNotFound is returned when no vault is found for the
	// given ID.
	ProtocolErrorVaultNotFound ProtocolErrorCode = "VAULT_NOT_FOUND"
	// ProtocolErrorVaultsDeletionFailed is returned when service was not able
	// to delete the vault.
	ProtocolErrorVaultDeletionFailed ProtocolErrorCode = "VAULT_DELETION_FAILED"
	// ProtocolErrorVaultUpdateFailed is returned when service was not able to
	// update the vault.
	ProtocolErrorVaultUpdateFailed ProtocolErrorCode = "VAULT_UPDATE_FAILED"
	// ProtocolErrorVaultKeyringRetrievalFailed is returned when service was
	// not able to retrieve vault's keyring.
	ProtocolErrorVaultKeyringRetrievalFailed ProtocolErrorCode = "VAULT_KEYRING_RETRIEVAL_FAILED"

	// ProtocolErrorVaultItemCreationFailed is returned when service was not
	// able to create a vault item.
	ProtocolErrorVaultItemCreationFailed ProtocolErrorCode = "VAULT_ITEM_CREATION_FAILED"
	// ProtocolErrorVaultItemUpdateFailed is returned when service was not able
	// to update a vault item.
	ProtocolErrorVaultItemUpdateFailed ProtocolErrorCode = "VAULT_ITEM_UPDATE_FAILED"
	// ProtocolErrorVaultItemNotFound is returned when no vault item is found
	// for the given ID.
	ProtocolErrorVaultItemNotFound ProtocolErrorCode = "VAULT_ITEM_NOT_FOUND"
	// ProtocolErrorVaultItemsRetrievalFailed is returned when service was not
	// able to retrieve vault items.
	ProtocolErrorVaultItemsRetrievalFailed ProtocolErrorCode = "VAULT_ITEMS_RETRIEVAL_FAILED"
	// ProtocolErrorVaultItemDeletionFailed is returned when service was not
	// able to delete a vault item.
	ProtocolErrorVaultItemDeletionFailed ProtocolErrorCode = "VAULT_ITEM_DELETION_FAILED"
)

// ProtocolError describes the payload of the error response.
type ProtocolError struct {
	// Code is a constant code for the error.
	Code ProtocolErrorCode `json:"code"`
	// Message is human-readable error message.
	Message string `json:"message"`
	// Details is an optional map with error details.
	Details map[string]any `json:"details"`
}
