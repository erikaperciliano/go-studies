package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Higher-Order Function (HOF) | Log middleware
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		next.ServeHTTP(w, r)
		fmt.Println(r.URL.String(), r.Method, time.Since(begin))
	})
}

func main() {
	mux := http.NewServeMux() // Create a new mux (multiplexer)

	// Define the healthcheck route
	mux.HandleFunc("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Println(id)
		// Send response to the client
		fmt.Fprintf(w, "hello, world!")
	})

	// Create the HTTP server with timeouts
	srv := &http.Server{
		Addr:                         ":8080",  // The address to listen on
		Handler:                      Log(mux), // Attach the mux handler
		DisableGeneralOptionsHandler: false,
		ReadTimeout:                  10 * time.Second,
		WriteTimeout:                 10 * time.Second,
		IdleTimeout:                  1 * time.Second,
	}

	// Start the server and listen for requests
	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err) // Panic if server fails to start
		}
	}
}
