package services

import (
	"context"
	"errors"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"

	"github.com/google/uuid"
)

// ctxt is a dummy type for storing context values.
//
// For more information see the documentation of `context.WithValue`.
type ctxt string

const ctxvalue ctxt = "stageai.tech/sunshine/sunshine"

// ContextValue holds everything this project could store in a context.
type ContextValue struct {
	// Token ID
	ID   uuid.UUID
	User *models.User
}

// Authorized reports whether there's a non-nil UserID inside.
func (c ContextValue) Authorized() bool {
	return c.User != nil
}

// FromContext extracts ContextValue from given context.
func FromContext(ctx context.Context) ContextValue {
	v, _ := ctx.Value(ctxvalue).(ContextValue)
	return v
}

// WithContext returns a copy of parent with inserted ContextValue.
func WithContext(parent context.Context, token *models.Token) context.Context {
	return context.WithValue(parent, ctxvalue,
		ContextValue{
			ID:   token.ID,
			User: &token.User,
		})
}

// NewTestContext creates context with token for testing purposes.
func NewTestContext(t *testing.T, e *Env, user *models.Document) context.Context {
	if !e.Debug {
		sentry.Report(errors.New("NewTestContext called NOT in debug mode"))
		panic("not in debug mode")
	}

	ctx := context.Background()
	token, err := e.TokenStore.Create(ctx, models.SessionToken, user.ID)
	if err != nil {
		t.Fatalf("create test token: %v", err)
	}

	// re-fetch in order to take into account roles this user after.
	user, err = e.UserStore.Get(ctx, user.ID)
	if err != nil {
		t.Fatalf("re-fetch user: %v", err)
	}

	token.User = *user.Data.(*models.User)
	return WithContext(ctx, token)

}
