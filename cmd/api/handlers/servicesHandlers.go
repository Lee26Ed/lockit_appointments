package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

func (h *Handler) CreateServiceHandler(w http.ResponseWriter, r *http.Request) {

	// Get the current user from context
	currentUser := h.contextGetUser(r)

	// Check if the current user has a business
	business, err := h.models.Businesses.GetByOwnerID(currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	h.Logger.Info("creating service for business", "business", business)

	var input struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
		Description string `json:"description,omitempty"`
		DurationMinutes int `json:"duration,omitempty"`
		DownTimeMinutes int `json:"downtime,omitempty"`
	}

	err = utils.ReadJSON(w, r, &input)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	
	service := &data.Service{
		Name: input.Name,
		Price: float64(input.Price),
		Description: input.Description,
		Duration: input.DurationMinutes,
		DownTime: input.DownTimeMinutes,
		Active: true, // default to active when creating a new service
		BusinessID: business.ID,
	}

	v := validator.New()
	if data.ValidateService(v, service); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	service, err = h.models.Services.Insert(service)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/services/%d", service.ID))

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"service": service}, headers)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) GetAllServicesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.Page = utils.GetSingleIntegerParameter(qs, "page", 1, v)
	input.Filters.PageSize = utils.GetSingleIntegerParameter(qs, "page_size", 20, v)
	input.Filters.Sort = utils.GetSingleQueryParameter(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "price", "name", "-id", "-price", "-name"}

	if data.ValidateFilters(v, input.Filters); !v.IsEmpty() {
				h.failedValidationResponse(w, r, v.Errors)
		return
	}

	services, metadata, err := h.models.Services.GetAll(input.Filters)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK,utils.Envelope{"services": services, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) GetServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}
	
	// Try to get the service from the database
	service, err := h.models.Services.Get(int(id))
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
		"service": service,
	}

	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) UpdateServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}

	// get the current user 
	currentUser := h.contextGetUser(r)

	// try to get Service data
	service, err := h.models.Services.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	canAccess, err := h.models.Businesses.CanAccessBusinessData(currentUser, service.BusinessID)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	if !canAccess {
		h.notPermittedResponse(w, r)
		return
	}
	
	var input struct {
		Name  *string `json:"name,omitempty"`
		Price *int    `json:"price,omitempty"`
		Description *string `json:"description,omitempty"`
		DurationMinutes *int `json:"duration,omitempty"`
		DownTimeMinutes *int `json:"downtime,omitempty"`
	}

	err = utils.ReadJSON(w, r, &input)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		service.Name = *input.Name
	}
	if input.Price != nil {
		service.Price = float64(*input.Price)
	}
	if input.Description != nil {
		service.Description = *input.Description
	}
	if input.DurationMinutes != nil {
		service.Duration = *input.DurationMinutes
	}
	if input.DownTimeMinutes != nil {
		service.DownTime = *input.DownTimeMinutes
	}

	v := validator.New()
	if data.ValidateService(v, service); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = h.models.Services.Update(service)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	response := utils.Envelope{
		"service": service,
	}

	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

func (h *Handler) DeleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}

	currentUser := h.contextGetUser(r)

	service, err := h.models.Services.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	canAccess, err := h.models.Businesses.CanAccessBusinessData(currentUser, service.BusinessID)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	if !canAccess {
		h.notPermittedResponse(w, r)
		return
	}
	
	// Try to delete the service from the database
	err = h.models.Services.Delete(int(id))
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
		"message": "service successfully deleted",
	}

	err = utils.WriteJSON(w, http.StatusOK, response, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}