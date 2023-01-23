package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Application version
const version = "1.0.0"

/// Configuration structure
type configuration struct {
	port int
	env  string
}

// Common information for all handlers
type application struct {
	config configuration
	logger log.Logger
}

func main() {

	var config configuration

	// Read the port and environment
	flag.IntVar(&config.port, "port", 8080, "Application server port")
	flag.StringVar(&config.env, "env", "development", "Current API environment (development|staging|production)")
	flag.Parse()

	// Create a logger
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)

	// Create and fill an application structure instance
	app := &application{
		config: config,
		logger: *logger,
	}

	// Create a new HTTP Server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("Starting the %s server on port %d", config.env, config.port)

	err := httpServer.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}
