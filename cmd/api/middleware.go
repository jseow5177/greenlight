package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rateLimit() middleware creates a new rate limiter and uses it for every request that it
// subsequently handles.
func (app *application) rateLimit(next http.Handler) http.Handler {

	// Define a client struct to hold the rate limiter and last seen time of each client
	type client struct {
		limiter *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine that removes old entries from the clients map once
	// every minute. This is to prevent the clients map from growing indefinitely.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place
			mu.Lock()

			// Loop through all the clients. If they haven't been seen in the last 3 minutes,
			// delete the corresponding entry from the map.
			for ip, client := range(clients) {
				if time.Since(client.lastSeen) > 3 * time.Minute {
					delete(clients, ip)
				}
			}

			// Unlock the mutex when the cleanup is complete
			mu.Unlock()
		}
	}()

	// The function returned closes over the initialized limiter.
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// Only carry out the check if rate limiting is enabled
		if app.config.limiter.enabled {
			// Extract the client's IP address from the request
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// Lock the mutex to prevent the following code from being executed concurrently
			mu.Lock()

			// Check if IP address already exists in the map.
			// If it doesn't, then initialize a new rate limiter and add the limiter to the map 
			// with the IP address as the key.
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					// Use limiter rps and burst from app config
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			// Update last seen
			clients[ip].lastSeen = time.Now()

			// limiter.Allow() checks if one event (request) can happen now.
			// It consumes one token.
			// If no token is available, it returns false.
			// Note that Allow() is protected by a mutex and safe for concurrent use.
			if !clients[ip].limiter.Allow() {
				mu.Unlock() // Unlock the mutex
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Unlock the mutex before calling the next handler in the chain.
			// Importantly, DO NOT defer the unlock of mutex.
			// Else, the mutex will not be unlocked until all the downstream handlers of this middleware have returned.
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

// recoverPanic() middleware recovers a panic in a go routine to
// return a 500 Internal Server Error response to the client
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (will be run as Go unwinds the stack in the go routine)
		// The function then checks if a panic has occured using the recover() function
		defer func() {
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				// This acts as a trigger to make Go's HTTP server automatically close the current connection
				// after a response has been sent.
				w.Header().Set("Connection", "close")
				// Use fmt.Errorf() to normalize err into an error and call the serverErrorResponse() helper.
				// This will log the error at the ERROR level and send the client a 500 Internal Server Error.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}