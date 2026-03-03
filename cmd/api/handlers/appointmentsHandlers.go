package handlers

import (
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/validator"
)

func (h *Handler) GetAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	appointments := utils.Envelope{"appointments": []utils.Envelope{{"id": 1, "title": "Doctor's Appointment", "date": "2024-07-01T10:00:00Z"},{"id": 2, "title": "Meeting with Bob", "date": "2024-07-02T14:00:00Z"}}}
	utils.WriteJSON(w, http.StatusOK, appointments, nil)
}

func (h *Handler) CreateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		Date  string `json:"date"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		errorData := utils.Envelope{"error": err.Error()}
		utils.WriteJSON(w, http.StatusBadRequest, errorData, nil)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"message": "Appointment created successfully"}, nil)
	if err != nil {
		h.Logger.Error("Failed to write JSON response", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create appointment"}, nil)
	}
}

func (h *Handler) GetAllAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.Page = utils.GetSingleIntegerParameter(qs, "page", 1, v)
	input.Filters.PageSize = utils.GetSingleIntegerParameter(qs, "page_size", 20, v)
	input.Filters.Sort = utils.GetSingleQueryParameter(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "date", "-id", "-title", "-date"}

	if data.ValidateFilters(v, input.Filters); !v.IsEmpty() {
				h.failedValidationResponse(w, r, v.Errors)
		return
	}
}