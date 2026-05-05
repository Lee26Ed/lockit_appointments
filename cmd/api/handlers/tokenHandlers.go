// Filename: cmd/api/tokenHandlers.go
package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

// createAuthTokenHandler handles POST /v1/tokens/authentication
func (h *Handler) CreateAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username    string        `json:"username"`
		Password string        `json:"password"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.Username != "", "username", "must be provided")
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Default TTL to 24 hours if not provided
	ttl := 24 * time.Hour

	// Is there an associated user for the provided username?
    user, err := h.models.Users.GetByUsername(input.Username)

    if err != nil {
        switch {
            case errors.Is(err, data.ErrRecordNotFound):
                h.invalidCredentialsResponse(w, r)
            default:
                h.serverErrorResponse(w, r, err)
        }
        return
    }

	// The user is found. Does their password match?
	match, err := user.Password.Matches(input.Password)
    if err != nil {
        h.serverErrorResponse(w, r, err)
        return
    }

	// Wrong password
	if !match {
		h.invalidCredentialsResponse(w, r)
		return
	}

	// Is the user activated?
	if !user.IsActivated {
		h.inactiveAccountResponse(w, r)
		return
	}

	// Create the token
	token, err := h.models.Tokens.New(user.ID, ttl, data.ScopeAuthentication)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	// Return both token and user data for the client
	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{
		"token": token,
		"user":  user,
	}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// createActivationTokenHandler handles POST /v1/tokens/activation
func (h *Handler) CreateActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int `json:"user_id"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.UserID > 0, "user_id", "must be provided")

	if !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Activation tokens typically have shorter TTL (e.g., 3 days)
	token, err := h.models.Tokens.New(input.UserID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"token": token}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// deleteAllTokensForUserHandler handles DELETE /v1/tokens/user/:user_id
func (h *Handler) DeleteAllTokensForUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ReadIDParam(r)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

	// Get scope from query parameter (default to authentication)
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = data.ScopeAuthentication
	}

	v := validator.New()
	v.Check(scope == data.ScopeActivation || scope == data.ScopeAuthentication, "scope", "must be 'activation' or 'authentication'")

	if !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = h.models.Tokens.DeleteAllForUser(scope, int(userID))
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "tokens successfully deleted"}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
