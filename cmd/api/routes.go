package main

import "net/http"

func (app *applicationDependencies) Routes() *http.ServeMux{
	mux := http.NewServeMux()

	mux.HandleFunc("/healthcheck", app.HealthcheckHandler)
	mux.HandleFunc("/appointments", app.GetAppointmentsHandler)
	mux.HandleFunc("/appointments/create", app.CreateAppointmentHandler)
	return mux

}