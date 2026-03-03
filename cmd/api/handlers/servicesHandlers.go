package handlers

import (
	"fmt"
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

func (h *Handler) CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
		Description string `json:"description,omitempty"`
		DurationMinutes int `json:"duration,omitempty"`
		Active bool `json:"active,omitempty"`
		BusinessID int `json:"business_id,omitempty"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	
	service := &data.Service{
		Name: input.Name,
		Price: float64(input.Price),
		Description: input.Description,
		Duration: input.DurationMinutes,
		Active: input.Active,
		BusinessID: input.BusinessID,
	}

	v := validator.New()
	if data.ValidateService(v, service); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = h.models.Services.Insert(service)
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
