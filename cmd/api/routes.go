package main

import (
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/handlers"
	"github.com/julienschmidt/httprouter"
)

func (app *applicationDependencies) Routes() http.Handler{

	const apiv = "/api/v1"

	router := httprouter.New()
	h := handlers.NewHandler(app.config, app.logger, app.models, &app.wg, &app.mailer)

	//* ----------------- General routes (public) ----------------- *//
	router.HandlerFunc(http.MethodGet, apiv+"/healthcheck", h.HealthcheckHandler)
	router.HandlerFunc(http.MethodGet, apiv+"/metrics", h.MetricsHandler)

	//* ----------------- User routes ----------------- *//
	router.HandlerFunc(http.MethodPost, apiv+"/users", h.CreateUserHandler)
	router.HandlerFunc(http.MethodPut, apiv+"/activate-user", h.ActivateUserHandler)
	router.HandlerFunc(http.MethodGet, apiv+"/users", 
		h.RequireRole("admin", h.GetAllUsersHandler))

	router.HandlerFunc(http.MethodGet, apiv+"/users/:id", 
		h.RequireActivatedUser(h.GetUserHandler))
	router.HandlerFunc(http.MethodPut, apiv+"/users/:id", 
		h.RequireActivatedUser(h.UpdateUserHandler))
	router.HandlerFunc(http.MethodDelete, apiv+"/users/:id", 
		h.RequireActivatedUser(h.DeleteUserHandler))

	//* ----------------- Role routes ----------------- *//
	router.HandlerFunc(http.MethodPost, apiv+"/roles", h.RequireRole("admin", h.CreateRoleHandler))
	router.HandlerFunc(http.MethodGet, apiv+"/roles", h.RequireRole("admin", h.GetAllRolesHandler))

	router.HandlerFunc(http.MethodGet, apiv+"/roles/:id", h.RequireRole("admin", h.GetRoleHandler))
	router.HandlerFunc(http.MethodPut, apiv+"/roles/:id", h.RequireRole("admin", h.UpdateRoleHandler))
	router.HandlerFunc(http.MethodDelete, apiv+"/roles/:id", h.RequireRole("admin", h.DeleteRoleHandler))

	//* ----------------- Businesses routes ----------------- *//
	router.HandlerFunc(http.MethodPost, apiv+"/businesses", 
		h.RequireActivatedUser(h.CreateBusinessHandler))
	router.HandlerFunc(http.MethodGet, apiv+"/businesses", h.GetAllBusinessesHandler) // public
	
	router.HandlerFunc(http.MethodGet, apiv+"/businesses/:id", h.GetBusinessHandler) // public
	router.HandlerFunc(http.MethodPut, apiv+"/businesses/:id", 
		h.RequireActivatedUser(h.UpdateBusinessHandler))
	router.HandlerFunc(http.MethodDelete, apiv+"/businesses/:id", 
		h.RequireActivatedUser(h.DeleteBusinessHandler))
		
	//* ----------------- Services routes ----------------- *//
	router.HandlerFunc(http.MethodPost, apiv+"/services/", 
		h.RequireRole("business", h.CreateServiceHandler))
	router.HandlerFunc(http.MethodGet, apiv+"/services", h.GetAllServicesHandler) // public
	
	router.HandlerFunc(http.MethodGet, apiv+"/services/:id", h.GetServiceHandler) // public
	router.HandlerFunc(http.MethodPut, apiv+"/services/:id", 
	h.RequireActivatedUser(h.UpdateServiceHandler))
	router.HandlerFunc(http.MethodDelete, apiv+"/services/:id", 
	h.RequireActivatedUser(h.DeleteServiceHandler))
		
	//* ----------------- Appointment routes ----------------- *//
	router.HandlerFunc(http.MethodGet, apiv+"/appointments", h.GetAppointmentsHandler)
	router.HandlerFunc(http.MethodPost, apiv+"/appointments/create", h.CreateAppointmentHandler)

	//* ----------------- Token routes ----------------- *//
	router.HandlerFunc(http.MethodPost, apiv+"/tokens/authenticate", h.CreateAuthTokenHandler)
	router.HandlerFunc(http.MethodPost, apiv+"/tokens/activate", h.CreateActivationTokenHandler)
	router.HandlerFunc(http.MethodDelete, apiv+"/tokens/user/:user_id", h.DeleteAllTokensForUserHandler)


	// wrap router with middleware
	handler := h.Metrics(router)
	handler = h.RecoverPanic(handler)
	handler = h.EnableCORS(handler)
	handler = h.RateLimit(handler)
	handler = h.LoggingMiddleware(handler)
	handler = h.Authenticate(handler)
	handler = h.GzipMiddleware(handler)
	return handler

}