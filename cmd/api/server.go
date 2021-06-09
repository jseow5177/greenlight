package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Graceful shutdown shuts down the server w/o interrupting any active connections.
// It works by first closing all open listeners, then closing all idle connections, and then wait
// indefinitely for connections to return to idle and then shutdown.

// Basically, it instructs our server to stop receiving new HTTP requests. Our server also gives any in-flight requests
// a 5-second period to complete before the application is terminated.

func (app *application) serve() error {
	// Declare a HTTP server
	// The server listens on the port provided in the config struct and uses the ServeMux
	// created above as the handler.
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
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
		// Create a new Go log.logger instance with the log.New() function, passing our own logger as the underlying
		// io.Writer. The "" and 0 indicate that the log.logger instance should not use any prefix or flags.
		ErrorLog: log.New(app.logger, "", 0),
	}

	// Create a shutdownError channel. This is used to receive any errors returned by the graceful Shutdown() method.
	shutdownError := make(chan error)

	go func() {
		// Create a quit channel which carries os.Signal.
		// Use a buffered channel with size 1.
		// A buffered channel is required because signal.Notify() does not wait for a receiver
		// to be available when sending a signal to the quit channel. If an unbuffered channel is used,
		// the signal could be missed if the quit channel is not ready to receive at the exact moment the signal is sent.
		// See the empty default case: https://github.com/golang/go/blob/bc7e4d9257693413d57ad467814ab71f1585a155/src/os/signal/signal.go#L243
		quit := make(chan os.Signal, 1)

		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and relay them to the
		// quit channel. Any other signals will not be caught by signal.Notify() and will retain their default
		// behavior.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will block until a signal is received.
		s := <-quit

		// Log a message to say that the signal has been caught.
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		// Create a 5-second timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Call Shutdown() on our server, passing in the 5-second timeout context.
		// Shutdown() returns nil if the graceful shutdown was successful.
		// An error may happen if there is a problem of closing the listeners, or because
		// the shutdown didn't complete before the 5-second context deadline is hit.
		// The error is relayed to the shutdownError channel.
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// Log a message to say that we're waiting for any background routines to
		// complete their tasks.
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})

		// Call Wait() to block until the WaitGroup counter is zero -- essentially blocking until
		// the background routines have finished. Then we return nil on the shutdownError channel, to indicate that
		// the shutdown completed without any issues.
		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// Calling Shutdown() causes ListenAndServe to immediately return a http.ErrServerClosed error.
	// This error indicates that the graceful shutdown has started (a good thing).
	// We return any error that is NOT ErrServerClosed.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise, wait to receive the return value from Shutdown() on the shutdownError channel. If there is an error,
	// we know there was a problem with the graceful shutdown.
	err = <-shutdownError
	if err != nil {
		return err
	}

	// At this point, we know the graceful shutdown is successful.
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
