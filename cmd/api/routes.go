package main

import (
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/handlers"
	"github.com/julienschmidt/httprouter"
)

func (app *applicationDependencies) Routes() http.Handler{
	router := httprouter.New()

	h := handlers.NewHandler(app.config, app.logger, app.models)

	router.HandlerFunc(http.MethodGet, "/healthcheck", h.HealthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/appointments", h.GetAppointmentsHandler)
	router.HandlerFunc(http.MethodPost, "/appointments/create", h.CreateAppointmentHandler)

	// * User routes
	router.HandlerFunc(http.MethodPost, "/users", h.CreateUserHandler)
	router.HandlerFunc(http.MethodGet, "/users/:id", h.GetUserHandler)
	router.HandlerFunc(http.MethodGet, "/users", h.GetAllUsersHandler)
	router.HandlerFunc(http.MethodPut, "/users/:id", h.UpdateUserHandler)
	router.HandlerFunc(http.MethodDelete, "/users/:id", h.DeleteUserHandler)
	return router

}