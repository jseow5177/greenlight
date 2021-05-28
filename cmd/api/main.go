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

	"github.com/joho/godotenv"
	"github.com/jseow5177/greenlight/internal/data"
	"github.com/jseow5177/greenlight/internal/jsonlog"
	_ "github.com/lib/pq"
)

// Declare a string containing the application version number.
// This will be generated automatically at build time later.
const version = "1.0.0"

// Define a config struct to hold all the configuration settings for our application.
// The configuration settings will be read from command-line flags when application starts.
// They will have sensible default values if not provided in command-line.
type config struct {
	port int
	env string
	db struct {
		dsn string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime string
	}
	limiter struct {
		rps float64 // Request per second limiter
		burst int // Burst value for limiter
		enabled bool // Boolean value to enable or disable rate limitting
	}
}

// Define an application struct to hold the dependencies for HTTP handlers, helpers,
// and middleware.
type application struct {
	config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	psqlPass := os.Getenv("POSTGRESS_PASSWORD")

	// Declare an instance of the config struct
	var cfg config

	// Read application configuration settings from command-line flags into the config struct
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", fmt.Sprintf("postgres://greenlight:%s@localhost/greenlight?sslmode=disable", psqlPass), "Postgres DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgresSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Postgres SQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgresSQL max connection time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	// Initialize a new jsonlog.Logger which writes any messages *at or above* the INFO
	// severity level to the standard output stream
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Call openDB() to create the connection pool, passing in the config struct.
	// If it returns an error, we log it and exit immediately.
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer a call to db.Close() so that the connection pool is closed before
	// the main() function exits.
	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	// Declare an instance of the application struct, containing the config struct and the logger.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db), // Add database models as application dependency
	}

	// Declare a custom ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// Declare a HTTP server with sensible timeout settings.
	// The server listens on the port provided in the config struct and uses the ServeMux
	// created above as the handler.
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
		// Go enables persistent HTTP connections by default to reduce latency.
		// By default, Go closes persistent connections after 3 minutes.
		// We can reduce this default with the IdleTimeout setting.
		IdleTimeout: time.Minute,
		// ReadTimeout covers the time from when request is accepted to when the request body is fully read
		// (If no body, until the end of headers)
		ReadTimeout: 10 * time.Second,
		// WriteTimeout covers the time from the end of the request header read to the end of the
		// response write (for HTTP).
		// For HTTPS, it covers the time from when request is accepted to the end of response write.
		WriteTimeout: 30 * time.Second,
		// Create a new Go log.logger instance with the log.New() function.
		// We pass in our custom json logger as the first parameter.
		// The "" and 0 indicate that the log.Logger instance should not use a prefix or any flags.
		ErrorLog: log.New(logger, "", 0),
	}

	// Start the HTTP server
	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env": cfg.env,
	})

	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}

// openDB() returns a sql.DB connection pool.
func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool.
	// Passing a value less than or equal to 0 means no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connections in the pool.
	// Passing a value less than or equal to 0 means no limit.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Use time.ParseDuration() to convert the idle timeout duration string
	// to a time.Duration
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing in the context
	// we created above as a parameter.
	// If the connection couldn't be established successfully within a 5 seconds deadline,
	// this will return an error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool
	return db, nil
}