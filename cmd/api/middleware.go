package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) recoverFromPanic(nxtHandler http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		nxtHandler.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(nxtHandler http.Handler) http.Handler {

	// Structure to hold the rate limiters created for different clients
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Mutex and rate limiter map for holding rate limiting for each client
	var (
		mutx    sync.Mutex
		clients = make(map[string]*client, 0)
	)

	// A go routine for cleaning up old clients which have not sent request in last three seconds
	go func() {
		for {
			// Run once every minute
			time.Sleep(time.Minute)

			// Lock the mutex before accessing the clients map
			mutx.Lock()

			for ip, client := range clients {
				// Check if there has been no request since last three minutes
				// If no, remove the entry from the map
				if time.Since(client.lastSeen) > time.Minute*3 {
					delete(clients, ip)
				}
			}

			mutx.Unlock()
		}

	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get the ip of the client and check if this client is in the map or not.
		// If no, add it to the map and check if this request is within the define rate limit
		// Check if rate limiting is enabled
		if app.config.rateLimiter.enabled {

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverError(w, r, err)
				return
			}

			// Lock the mutext before accessing the clients map
			mutx.Lock()

			if _, found := clients[ip]; !found {
				// No earlier request, add to the map
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.rateLimiter.tps), app.config.rateLimiter.burstLimit),
				}

				clients[ip].lastSeen = time.Now()
			}

			// Check if request is within the configured rate
			if !clients[ip].limiter.Allow() {
				mutx.Unlock()
				app.tpsExceedResponse(w, r)
				return
			}

			mutx.Unlock()
		}

		nxtHandler.ServeHTTP(w, r)
	})
}
