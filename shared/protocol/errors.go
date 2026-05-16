package protocol

// ProtocolErrorCode is a type that's used for errors' codes.
type ProtocolErrorCode string

const (
	// ProtocolErrorUserNotFound is an error code that indicates that the
	// requested user was not found.
	ProtocolErrorUserNotFound ProtocolErrorCode = "USER_NOT_FOUND"
	// ProtocolErrorInvalidPassword is an error code that indicates that the
	// provided password is invalid.
	ProtocolErrorInvalidPassword ProtocolErrorCode = "INVALID_PASSWORD"
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
