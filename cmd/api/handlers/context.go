package handlers

import (
	"context"
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (h *Handler) contextSetUser(r *http.Request, user *data.User) *http.Request {
	// WithValue() expects the original context along with the new
	// key:value pair you want to update it with
    ctx := context.WithValue(r.Context(), userContextKey, user)
    return r.WithContext(ctx)
}

func (h *Handler) contextGetUser(r *http.Request) *data.User {
    user, ok := r.Context().Value(userContextKey).(*data.User)
    if !ok {
        panic("missing user value in request context")
    }

    return user
}
