package internal

import (
	"context"

	"github.com/Pelfox/gophkeeper/apps/server/internal/services"
)

type sessionContextKey struct{}

// WithSession is a small wrapper that populates context.Context with the given
// value and session.
func WithSession(
	ctx context.Context,
	session *services.SessionResult,
) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, session)
}

// SessionFromContext retrieves session from the given context.Context.
func SessionFromContext(ctx context.Context) (*services.SessionResult, bool) {
	session, ok := ctx.Value(sessionContextKey{}).(*services.SessionResult)
	return session, ok
}
