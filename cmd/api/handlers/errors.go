package handlers

import (
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
)

// log an error message
func (h *Handler) logError(r *http.Request, err error) {

	method := r.Method
	uri := r.URL.RequestURI()
	h.Logger.Error(err.Error(), "method", method, "uri", uri)
}

// send an error response in JSON
func (h *Handler) errorResponseJSON(w http.ResponseWriter, r *http.Request, status int, message any) {
	errorData := utils.Envelope{"error": message}
	err := utils.WriteJSON(w, status, errorData, nil)
	if err != nil {
		h.logError(r, err)
		w.WriteHeader(500)
	}
}

// send an error response if our server messes up
func (h *Handler) serverErrorResponse(w http.ResponseWriter,
	r *http.Request,
	err error) {

	// first thing is to log error message
	h.logError(r, err)
	// prepare a response to send to the client
	message := "the server encountered a problem and could not process your request"
	h.errorResponseJSON(w, r, http.StatusInternalServerError, message)
}

// send an error response if our client messes up with a 404
func (h *Handler) notFoundResponse(w http.ResponseWriter,
	r *http.Request) {

	// we only log server errors, not client errors
	// prepare a response to send to the client
	message := "the requested resource could not be found"
	h.errorResponseJSON(w, r, http.StatusNotFound, message)
}

func (h *Handler) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	h.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
}

func (h *Handler) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	h.errorResponseJSON(w, r, http.StatusConflict, message)
}