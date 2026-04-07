package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

func (h *Handler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var clientData struct {
		Email   string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := utils.ReadJSON(w, r, &clientData)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

	// Check if this is the first user registering
	userCount, err := h.models.Users.CountUsers()
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	// Set role_id to 1 for the first user (admin), otherwise 3 (regular user)
	roleID := 3
	if userCount == 0 {
		roleID = 1
	}

	// copy the data from the clientData struct to a new db User struct
	user := &data.User{
		Email:    clientData.Email,
		Username: clientData.Username,
		Status: data.UserStatusPending, // default status for new users
		IsActivated: false, // new users are not activated by default
		RoleID: roleID, // set role based on user count
	}

	// hash the password and store it in the User struct
	err = user.Password.SetPassword(clientData.Password)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	// validate the user data
	v := validator.New()
	if data.ValidateUser(v, user); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	// insert the user into the database
	user, err = h.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "email address already in use")
			h.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrDuplicateUsername):
			v.AddError("username", "username already in use")
			h.failedValidationResponse(w, r, v.Errors)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	// Generate a new activation token which expires in 3 days
	token, err := h.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	response := utils.Envelope{"user": user}

	h.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
			"username":        user.Username,
		}

		err = h.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			h.Logger.Error(err.Error())
		}
	})

	err = utils.WriteJSON(w, http.StatusCreated, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}


func (h *Handler) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the body from the request and store in temporary struct
	var incomingData struct {
		TokenPlaintext string `json:"token"`
	}
	err := utils.ReadJSON(w, r, &incomingData)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

	// Validate the data
	v := validator.New()
	data.ValidateTokenPlaintext(v, incomingData.TokenPlaintext)
	if !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Let's check if the token provided belongs to the user
	user, err := h.models.Users.GetForToken(data.ScopeActivation,
		incomingData.TokenPlaintext)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			h.failedValidationResponse(w, r, v.Errors)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	// User provided the right token so activate them
	h.Logger.Info("Activating user", "user_id", user.ID, "username", user.Username, "email", user.Email)
	err = h.models.Users.UpdateActivation(user.ID, data.UserStatusActive, true)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "user not found")
			h.failedValidationResponse(w, r, v.Errors)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update the user object for the response
	user.Status = data.UserStatusActive
	user.IsActivated = true

	// User has been activated so delete the activation token to
	// prevent reuse.
	err = h.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	// Send a response
	data := utils.Envelope{
		"user": user,
	}

	err = utils.WriteJSON(w, http.StatusOK, data, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}


func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}

	// get the current user 
	currentUser := h.contextGetUser(r)

	// Check if the current user can access this user's data
	canAccess, err := h.models.Users.CanAccessUserData(currentUser, int(id))
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	if !canAccess {
		h.notPermittedResponse(w, r)
		return
	}
	
	// Try to get the user from the database
	user, err := h.models.Users.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	response := utils.Envelope{
		"user": user,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.Page = utils.GetSingleIntegerParameter(qs, "page", 1, v)
	input.Filters.PageSize = utils.GetSingleIntegerParameter(qs, "page_size", 20, v)
	input.Filters.Sort = utils.GetSingleQueryParameter(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "status", "is_activated", "-id", "-status", "-is_activated"}

	if data.ValidateFilters(v, input.Filters); !v.IsEmpty() {
				h.failedValidationResponse(w, r, v.Errors)
		return
	}

	users, metadata, err := h.models.Users.GetAll(input.Filters)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}
	
	response := utils.Envelope{
		"users": users,
		"metadata": metadata,
	}
	
	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// updates a user
func (h *Handler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}

	// get the current user 
	currentUser := h.contextGetUser(r)

	// Check if the current user can access this user's data
	canAccess, err := h.models.Users.CanAccessUserData(currentUser, int(id))
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	if !canAccess {
		h.notPermittedResponse(w, r)
		return
	}

	// Get the existing user from the database
	user, err := h.models.Users.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	// Parse the request body for updates
	var clientData struct {
		Password    *string `json:"password,omitempty"`
		Email       *string `json:"email"`
	}


	err = utils.ReadJSON(w, r, &clientData)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

		// Update only the fields that were provided
	if clientData.Email != nil {
		user.Email = *clientData.Email
	}

	// Update password if provided
	if clientData.Password != nil {
		err = user.Password.SetPassword(*clientData.Password)
		if err != nil {
			h.serverErrorResponse(w, r, err)
			return
		}
	}

	// Validate the updated user data
	v := validator.New()
	if data.ValidateUser(v, user); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update the user in the database
	user, err = h.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "email address already in use")
			h.failedValidationResponse(w, r, v.Errors)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	response := utils.Envelope{
		"user": user,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}


// deletes a user 
func (h *Handler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}

	// get the current user 
	currentUser := h.contextGetUser(r)

	// Check if the current user can access this user's data
	canAccess, err := h.models.Users.CanAccessUserData(currentUser, int(id))
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	if !canAccess {
		h.notPermittedResponse(w, r)
		return
	}

	// Try to delete the user from the database
	err = h.models.Users.Delete(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	response := utils.Envelope{
		"message": "user successfully deleted",
	}
	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
