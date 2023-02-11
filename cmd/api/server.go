package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) startServer() error {

	// Create a new HTTP Server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Create a channel to receive the response of Shutdown
	shutdownChannel := make(chan error)

	// Go routine for catching interrupt signals
	go func() {
		// Create a new channel for getting the signals
		sigChannel := make(chan os.Signal, 1)

		// Register for SIGTERM and SIGINT and relay these to the channel
		signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

		// Read signal from the channel. Block until signal is received
		sig := <-sigChannel

		// Log the details of the signal received and start the shutdown
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": sig.String(),
		})

		// Create a timeout context
		ctxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Use this contect to initiate shutdown of the server
		// and pass the return value to the shutdown channel
		shutdownChannel <- httpServer.Shutdown(ctxt)
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"addr":    httpServer.Addr,
		"env":     app.config.env,
		"tps":     fmt.Sprintf("%f", app.config.rateLimiter.tps),
		"burst":   fmt.Sprintf("%d", app.config.rateLimiter.burstLimit),
		"enabled": fmt.Sprintf("%v", app.config.rateLimiter.enabled),
	})

	// Start the server. Capture the return value which may be a normal error
	// or an error returned in case the server was shutdown. We pass back the error only
	// inc case it was not shut down by us calling the Shutdown function
	err := httpServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Wait to receive the response status of the Shutdown call
	err = <-shutdownChannel
	if err != nil {
		return err
	}

	// We reach here only on server shutdown. Log the status
	app.logger.PrintInfo("server shut down", map[string]string{
		"addr": httpServer.Addr,
	})

	return nil
}
