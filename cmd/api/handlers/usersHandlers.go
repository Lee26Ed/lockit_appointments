package handlers

import (
	"errors"
	"net/http"

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

	// copy the data from the clientData struct to a new User struct
	user := &data.User{
		Email:    clientData.Email,
		Username: clientData.Username,
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
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	response := utils.Envelope{"user": user}
	err = utils.WriteJSON(w, http.StatusCreated, response, nil)
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
	users, err := h.models.Users.GetAll()
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}
	
	response := utils.Envelope{
		"users": users,
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
	var input struct {
		Username    *string `json:"username"`
		Password    *string `json:"password"`
		Email       *string `json:"email"`
		RoleID      *int    `json:"role_id"`
	}

	err = utils.ReadJSON(w, r, &input)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

		// Update only the fields that were provided
	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.RoleID != nil {
		user.RoleID = *input.RoleID
	}

	// Update password if provided
	if input.Password != nil {
		err = user.Password.SetPassword(*input.Password)
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
		h.serverErrorResponse(w, r, err)
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
