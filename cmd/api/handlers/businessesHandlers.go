package handlers

import (
	"errors"
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

func (h *Handler) CreateBusinessHandler(w http.ResponseWriter, r *http.Request) {
	// Get the current user from context and create a business with the current user as the owner
	currentUser := h.contextGetUser(r)

	var clientData struct {
		Name 	  string `json:"name"`
		Bio 	  string `json:"bio"`
		Email 	  string `json:"email"`
		Phone 	  string `json:"phone"`
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
		OwnerID: currentUser.ID,
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

	// update the user's role to business owner if they aren't already
	if currentUser.RoleID == 3 {
		currentUser.RoleID = 2 // set role to business owner
		_, err = h.models.Users.Update(currentUser)
		if err != nil {
			h.Logger.Error("failed to update user role", "error", err)
			h.serverErrorResponse(w, r, err)
			return
		}
	}

	response := utils.Envelope{"business": Business}
	err = utils.WriteJSON(w, http.StatusCreated, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) GetAllBusinessesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.Page = utils.GetSingleIntegerParameter(qs, "page", 1, v)
	input.Filters.PageSize = utils.GetSingleIntegerParameter(qs, "page_size", 20, v)
	input.Filters.Sort = utils.GetSingleQueryParameter(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "-id", "-name"}

	if data.ValidateFilters(v, input.Filters); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	businesses, metadata, err := h.models.Businesses.GetAll(input.Filters)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"businesses": businesses, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) GetBusinessHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	
	// Try to get the business from the database
	business, err := h.models.Businesses.Get(int(id))
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
		"business": business,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) UpdateBusinessHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	
	// Try to get the business from the database
	business, err := h.models.Businesses.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	// parse the updated business data from the request body
	var clientData struct {
		Name 	  *string `json:"name"`
		Bio 	  *string `json:"bio"`
		Email 	  *string `json:"email"`
		Phone 	  *string `json:"phone"`
	}

	err = utils.ReadJSON(w, r, &clientData)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	
	// update the business fields if they were provided in the request body
	if clientData.Name != nil {
		business.Name = *clientData.Name
	}
	if clientData.Bio != nil {
		business.Bio = *clientData.Bio
	}
	if clientData.Email != nil {
		business.Email = *clientData.Email
	}
	if clientData.Phone != nil {
		business.Phone = *clientData.Phone
	}

	// validate the updated business data
	v := validator.New()
	if data.ValidateBusiness(v, business); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	// update the business in the database
	err = h.models.Businesses.Update(business)
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
		"business": business,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) DeleteBusinessHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	
	// Try to delete the business from the database
	err = h.models.Businesses.Delete(int(id))
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
		"message": "business successfully deleted",
	}
	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}