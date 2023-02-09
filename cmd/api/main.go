package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/narinderv/blockbuster/internal/data"
)

// Database constants
const MAX_OPEN_CONNS = 25
const MAX_IDLE_CONNS = 25
const MAX_IDLE_TIME = "15m"

// Application version
const version = "1.0.0"

/// Configuration structure
type configuration struct {
	port      int
	env       string
	dbDetails struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Common information for all handlers
type application struct {
	config configuration
	logger log.Logger
	models data.Models
}

func main() {

	var config configuration

	// Read the port and environment
	flag.IntVar(&config.port, "port", 8080, "Application server port")
	flag.StringVar(&config.env, "env", "development", "Current API environment (development|staging|production)")
	flag.StringVar(&config.dbDetails.dsn, "dsn", "postgres://bbuster:bbuster@localhost/blockbuster?sslmode=disable", "Postgres connection details (postgres://<user>:<pass>@<host/databaseName>)?sslmode=disable")

	flag.Parse()

	config.dbDetails.maxIdleConns = MAX_IDLE_CONNS
	config.dbDetails.maxOpenConns = MAX_OPEN_CONNS
	config.dbDetails.maxIdleTime = MAX_IDLE_TIME

	// Create a logger
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)

	// Create a database connection
	dbConn, err := connectToDatabase(config)
	if err != nil {
		logger.Fatal(err)
	}

	defer dbConn.Close()

	logger.Println("Connected to the database")

	// Create and fill an application structure instance
	app := &application{
		config: config,
		logger: *logger,
		models: data.NewModel(dbConn),
	}

	// Create a new HTTP Server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("Starting the %s server on port %d", config.env, config.port)

	err = httpServer.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}

func connectToDatabase(conf configuration) (*sql.DB, error) {
	db, err := sql.Open("postgres", conf.dbDetails.dsn)
	if err != nil {
		return nil, err
	}

	// Set the connection timeout values
	db.SetMaxOpenConns(conf.dbDetails.maxOpenConns)
	db.SetMaxIdleConns(conf.dbDetails.maxIdleConns)

	duration, err := time.ParseDuration(conf.dbDetails.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	// Create a context with 5 second timeout
	ctxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to the database using the context
	if err = db.PingContext(ctxt); err != nil {
		return nil, err
	}

	return db, nil
}
