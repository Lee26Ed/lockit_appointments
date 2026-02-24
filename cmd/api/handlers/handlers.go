package handlers

import (
	"log/slog"

	"github.com/Lee26Ed/lockit_appointments/cmd/api/types"
	"github.com/Lee26Ed/lockit_appointments/cmd/internal/data"
)

// Handler struct holds the configuration and logger for the API handlers.
type Handler struct {
	Config types.ServerConfig
	Logger *slog.Logger
	models *data.Models
}

// NewHandler function creates a new Handler instance with the provided configuration and logger.
func NewHandler(cfg types.ServerConfig, logger *slog.Logger, models *data.Models) *Handler {
	return &Handler{Config: cfg, Logger: logger, models: models}
}