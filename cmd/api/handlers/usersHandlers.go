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