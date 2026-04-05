package main

import (
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/handlers"
	"github.com/julienschmidt/httprouter"
)

func (app *applicationDependencies) Routes() http.Handler{

	const apiv = "/api/v1"

	router := httprouter.New()
	h := handlers.NewHandler(app.config, app.logger, app.models)

	// * General routes
	router.HandlerFunc(http.MethodGet, apiv+"/healthcheck", h.HealthcheckHandler)
	router.HandlerFunc(http.MethodGet, apiv+"/metrics", h.MetricsHandler)

	
	// * User routes
	router.HandlerFunc(http.MethodPost, apiv+"/users", h.CreateUserHandler)
	router.HandlerFunc(http.MethodGet, apiv+"/users", h.GetAllUsersHandler)
	router.HandlerFunc(http.MethodGet, apiv+"/users/:id", h.GetUserHandler)
	router.HandlerFunc(http.MethodPut, apiv+"/users/:id", h.UpdateUserHandler)
	router.HandlerFunc(http.MethodDelete, apiv+"/users/:id", h.DeleteUserHandler)

	// * Businesses routes
	router.HandlerFunc(http.MethodPost, apiv+"/businesses", h.CreateBusinessHandler)
	
	// * Appointment routes
	router.HandlerFunc(http.MethodGet, "/appointments", h.GetAppointmentsHandler)
	router.HandlerFunc(http.MethodPost, "/appointments/create", h.CreateAppointmentHandler)

	// * Services routes
	router.HandlerFunc(http.MethodPost, "/services/", h.CreateServiceHandler)
	router.HandlerFunc(http.MethodGet, "/services", h.GetAllServicesHandler)

	// wrap router with middleware
	handler := h.Metrics(router)
	handler = h.EnableCORS(handler)
	handler = h.RateLimit(handler)
	handler = h.LoggingMiddleware(handler)
	handler = h.GzipMiddleware(handler)
	return handler

}