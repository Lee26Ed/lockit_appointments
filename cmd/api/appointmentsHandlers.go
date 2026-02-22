package main

import "net/http"

func (app *applicationDependencies) GetAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	appointments := envelope{"appointments": []envelope{{"id": 1, "title": "Doctor's Appointment", "date": "2024-07-01T10:00:00Z"},{"id": 2, "title": "Meeting with Bob", "date": "2024-07-02T14:00:00Z"}}}
	app.writeJSON(w, http.StatusOK, appointments, nil)
}

func (app *applicationDependencies) CreateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		Date  string `json:"date"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		errorData := envelope{"error": err.Error()}
		app.writeJSON(w, http.StatusBadRequest, errorData, nil)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "Appointment created successfully"}, nil)
	if err != nil {
		app.logger.Error("Failed to write JSON response", "error", err)
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": "Failed to create appointment"}, nil)
	}
}