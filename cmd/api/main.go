package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/narinderv/blockbuster/internal/data"
	"github.com/narinderv/blockbuster/internal/jsonlog"
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
	// Structure for rate limiting
	rateLimiter struct {
		tps        float64
		burstLimit int
		enabled    bool
	}
}

// Common information for all handlers
type application struct {
	config configuration
	logger *jsonlog.Logger
	models data.Models
}

func main() {

	var config configuration

	// Read the port and environment
	flag.IntVar(&config.port, "port", 8080, "Application server port")
	flag.StringVar(&config.env, "env", "development", "Current API environment (development|staging|production)")

	// Database details
	flag.StringVar(&config.dbDetails.dsn, "dsn", "postgres://bbuster:bbuster@localhost/blockbuster?sslmode=disable", "Postgres connection details (postgres://<user>:<pass>@<host/databaseName>)?sslmode=disable")

	// Rate Limiting
	flag.Float64Var(&config.rateLimiter.tps, "tps", 2, "Transaction per second")
	flag.IntVar(&config.rateLimiter.burstLimit, "burst", 4, "Max. spike limit")
	flag.BoolVar(&config.rateLimiter.enabled, "enable-rate-limit", true, "Enable rate limiting")

	flag.Parse()

	config.dbDetails.maxIdleConns = MAX_IDLE_CONNS
	config.dbDetails.maxOpenConns = MAX_OPEN_CONNS
	config.dbDetails.maxIdleTime = MAX_IDLE_TIME

	// Create a logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Create a database connection
	dbConn, err := connectToDatabase(config)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer dbConn.Close()

	logger.PrintInfo("Connected to the database", nil)

	// Create and fill an application structure instance
	app := &application{
		config: config,
		logger: logger,
		models: data.NewModel(dbConn),
	}

	err = app.startServer()
	if err != nil {
		logger.PrintFatal(err, nil)
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
