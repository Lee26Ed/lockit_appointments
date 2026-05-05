package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

// createRoleHandler handles POST /v1/roles
func (h *Handler) CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RoleName string `json:"role_name"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

	role := &data.Role{
		RoleName: input.RoleName,
	}

	v := validator.New()
	if data.ValidateRole(v, role); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = h.models.Roles.Insert(role)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/roles/%d", role.ID))

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"role": role}, headers)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// getRoleHandler handles GET /v1/roles/:id
func (h *Handler) GetRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}

	role, err := h.models.Roles.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"role": role}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// getAllRolesHandler handles GET /v1/roles
func (h *Handler) GetAllRolesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.Page = utils.GetSingleIntegerParameter(qs, "page", 1, v)
	input.Filters.PageSize = utils.GetSingleIntegerParameter(qs, "page_size", 20, v)
	input.Filters.Sort = utils.GetSingleQueryParameter(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "role", "-id", "-role"}

	if data.ValidateFilters(v, input.Filters); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	roles, metadata, err := h.models.Roles.GetAll(input.Filters)
	if err != nil {
		h.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"roles": roles, "metadata": metadata}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// updateRoleHandler handles PATCH /v1/roles/:id
func (h *Handler) UpdateRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}

	role, err := h.models.Roles.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		RoleName *string `json:"role_name"`
	}

	err = utils.ReadJSON(w, r, &input)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

	if input.RoleName != nil {
		role.RoleName = *input.RoleName
	}

	v := validator.New()
	if data.ValidateRole(v, role); !v.IsEmpty() {
		h.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = h.models.Roles.Update(role)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"role": role}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}

// deleteRoleHandler handles DELETE /v1/roles/:id
func (h *Handler) DeleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := utils.ReadIDParam(r)
	if err != nil {
		h.notFoundResponse(w, r)
		return
	}

	err = h.models.Roles.Delete(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			h.notFoundResponse(w, r)
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "role successfully deleted"}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
