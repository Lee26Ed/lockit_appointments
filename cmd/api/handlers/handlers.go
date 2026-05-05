package handlers

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/types"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/mailer"
)

// Handler struct holds the configuration and logger for the API handlers.
type Handler struct {
	Config types.ServerConfig
	Logger *slog.Logger
	models *data.Models
	mailer *mailer.Mailer
	wg     *sync.WaitGroup
}

// NewHandler function creates a new Handler instance with the provided configuration and logger.
func NewHandler(cfg types.ServerConfig, logger *slog.Logger, models *data.Models, wg *sync.WaitGroup, mailer *mailer.Mailer) *Handler {
	return &Handler{Config: cfg, Logger: logger, models: models, mailer: mailer, wg: wg}
}


// Accept a function and run it in the background also recover from any panic
func (h *Handler) background(fn func()) {
    h.wg.Add(1) // Use a wait group to ensure all goroutines finish before we exit
    go func() {
        defer h.wg.Done()     // signal goroutine is done
        defer func() {
           err := recover() 
           if err != nil {
                h.Logger.Error(fmt.Sprintf("%v", err))
            }
        }()
       fn()     // Run the actual function
   }()
}