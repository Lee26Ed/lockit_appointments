package handlers

import (
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
)

func (h *Handler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var userData struct {
		Email   string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := utils.ReadJSON(w, r, &userData)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}

	
}