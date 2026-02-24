package handlers

import (
	"net/http"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/utils"
)


func (h *Handler) HealthcheckHandler(w http.ResponseWriter,
                                               r *http.Request) {

	data := utils.Envelope{
                     "status": "available",
                     "system_info": utils.Envelope{
                             "environment": h.Config.Environment,
                             "version": h.Config.AppVersion,
                    },
   }

	err := utils.WriteJSON(w, http.StatusOK, data, nil)
	if err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
		return
	}

}
