package handlers

import (
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/helpers"
)

func (h *Handler) GetAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	appointments := helpers.Envelope{"appointments": []helpers.Envelope{{"id": 1, "title": "Doctor's Appointment", "date": "2024-07-01T10:00:00Z"},{"id": 2, "title": "Meeting with Bob", "date": "2024-07-02T14:00:00Z"}}}
	helpers.WriteJSON(w, http.StatusOK, appointments, nil)
}

func (h *Handler) CreateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		Date  string `json:"date"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errorData := helpers.Envelope{"error": err.Error()}
		helpers.WriteJSON(w, http.StatusBadRequest, errorData, nil)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"message": "Appointment created successfully"}, nil)
	if err != nil {
		h.Logger.Error("Failed to write JSON response", "error", err)
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.Envelope{"error": "Failed to create appointment"}, nil)
	}
}