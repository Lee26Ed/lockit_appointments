package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *applicationDependencies) Serve() error {
	apiServer := &http.Server{
        Addr: fmt.Sprintf(":%d", app.config.Port),
        Handler: app.Routes(),
        IdleTimeout: time.Minute,
        ReadTimeout: 5 * time.Second,
        WriteTimeout: 10 * time.Second,
        ErrorLog: slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
    }

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		
		app.logger.Info("Shutting down server", "signal", s.String())
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		// we will only write to the error channel if there is an error
		err := apiServer.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}
		// Wait for background tasks to complete
		app.logger.Info("completing background tasks", "address", apiServer.Addr)
		app.wg.Wait()
		shutdownError <- nil           // successful shutdown
		}()

	app.logger.Info("Starting server", "address", apiServer.Addr, "env", app.config.Environment)

	err := apiServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}
	app.logger.Info("Server stopped", "addr", apiServer.Addr)
	return nil
}