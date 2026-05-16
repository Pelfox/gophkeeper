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
