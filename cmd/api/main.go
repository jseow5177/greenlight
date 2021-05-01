package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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
}

// Define an application struct to hold the dependencies for HTTP handlers, helpers,
// and middleware.
type application struct {
	config
	logger *log.Logger
}

func main() {
	// Declare an instance of the config struct
	var cfg config

	// Read the value of the port and env command-line flags into the config struct
	// The port number defaults to 4000
	// The environment defaults to "development"\
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

	// Declare an instance of the application struct, containing the config struct and the logger.
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Declare a custom ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// Declare a HTTP server with sensible timeout settings.
	// The server listens on the port provided in the config struct and uses the ServeMux
	// created above as the handler.
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.port),
		Handler: mux,
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
	}

	// Start the HTTP server
	logger.Printf("Starting %s server on port %d", cfg.env, cfg.port)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}