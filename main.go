package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux() // Create a new mux (multiplexer)

	// Define the healthcheck route
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		// Send response to the client
		fmt.Fprintf(w, "hello, world!")
	})

	// Create the HTTP server with timeouts
	srv := &http.Server{
		Addr:                         ":8080", // The address to listen on
		Handler:                      mux,     // Attach the mux handler
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

	fmt.Println("Server has been closed") // Will execute after the server is closed
}
