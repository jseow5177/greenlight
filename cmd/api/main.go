package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/jseow5177/greenlight/internal/data"
	"github.com/jseow5177/greenlight/internal/jsonlog"
	"github.com/jseow5177/greenlight/internal/mailer"
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
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64 // Request per second limiter
		burst   int     // Burst value for limiter
		enabled bool    // Boolean value to enable or disable rate limitting
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// Define an application struct to hold the dependencies for HTTP handlers, helpers,
// and middleware.
type application struct {
	config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	psqlPass := os.Getenv("POSTGRESS_PASSWORD")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")

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

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port (25|465|587|2525)")
	flag.StringVar(&cfg.smtp.username, "smtp-username", smtpUser, "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", smtpPass, "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.net>", "SMTP sender")

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
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// Start the server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
