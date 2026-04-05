package handlers

import (
	"errors"
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

func (h *Handler) CreateBusinessHandler(w http.ResponseWriter, r *http.Request) {
	var clientData struct {
		Name 	  string `json:"name"`
		Bio 	  string `json:"bio"`
		Email 	  string `json:"email"`
		Phone 	  string `json:"phone"`
		OwnerID   int    `json:"owner_id"`
	}

	err := utils.ReadJSON(w, r, &clientData)
	if err != nil {
		h.Logger.Error("failed to read JSON", "error", err)
		h.badRequestResponse(w, r, err)
		return
	}

	Business := &data.Business{
		Name: clientData.Name,
		Bio: clientData.Bio,
		Email: clientData.Email,
		Phone: clientData.Phone,
		OwnerID: clientData.OwnerID,
		Status: data.BusinessStatusActive, // default status for new businesses
		LogoURL: "https://via.placeholder.com/150", // set temp logo_url until we implement file uploads

	}

	// generate a unique slug for the Business based on its name
	slug, err := h.models.Businesses.GenerateUniqueSlug(Business.Name)
	if err != nil {
		h.Logger.Error("failed to generate unique slug", "error", err)
		h.serverErrorResponse(w, r, err)
		return
	}
	Business.Slug = slug

	// validate the Business data
	v := validator.New()
	if data.ValidateBusiness(v, Business); !v.IsEmpty() {
		h.Logger.Error("failed validation", "errors", v.Errors)
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	Business, err = h.models.Businesses.Insert(Business)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "email address already in use")
			h.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrDuplicateSlug):
			v.AddError("slug", "slug already in use")
			h.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrOwnerIDInvalid):
			v.AddError("owner_id", "owner_id does not reference a valid user")
			h.failedValidationResponse(w, r, v.Errors)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	response := utils.Envelope{"business": Business}
	err = utils.WriteJSON(w, http.StatusCreated, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}