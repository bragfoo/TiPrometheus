package simpleHTTP

import (
	"net/http"
)

const defaultHTTPEndpoint = ":12350"

// Server starts an HTTP listener on endpoint and waits for connections.
func Server(endpoint string) {
	// create a muxer and register the handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/write", RemoteWrtie)
	mux.HandleFunc("/read", RemoteRead)

	// configure our server
	if endpoint == "" {
		// fall back to the default endpoint if none given
		endpoint = defaultHTTPEndpoint
	}
	server := &http.Server{
		Addr:    endpoint,
		Handler: mux,
	}
		server.ListenAndServe()
}
