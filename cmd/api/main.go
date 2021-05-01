package main

import "log"

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
	
}