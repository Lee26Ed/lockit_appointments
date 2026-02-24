package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (app *applicationDependencies) Serve() error {
	apiServer := &http.Server {
        Addr: fmt.Sprintf(":%d", app.config.Port),
        Handler: app.Routes(),
        IdleTimeout: time.Minute,
        ReadTimeout: 5 * time.Second,
        WriteTimeout: 10 * time.Second,
        ErrorLog: slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
    }

	app.logger.Info("Starting server", "address", apiServer.Addr, "env", app.config.Environment)

	err := apiServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}