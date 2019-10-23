package simpleHTTP

import (
	"net/http"
	"log"
	"io/ioutil"
	"crypto/x509"
	"crypto/tls"
)

const defaultHTTPEndpoint = ":12350"

// Server starts an HTTP listener on endpoint and waits for connections.
// This function simply calls ServerTLS without any certificates.
func Server(endpoint string) {
	ServerTLS(endpoint, "", "", "")
}

// ServerTLS starts an HTTP or HTTPS listener on endpoint and waits for connections.
//
// If endpoint is unspecified, ":12350" is used.
//
// If a server certificate is specified, HTTPS is enabled.
// Simple HTTP is used otherwise.
//
// If a CA certificate is specified, mTLS (mutual TLS) is enabled.
// Connecting clients must pass a valid client certificate signed by the CA.
func ServerTLS(endpoint string, caCertFile string, certFile string, keyFile string) {
	// create a muxer and register the handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/write", RemoteWrite)
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

	// check if mTLS is requested
	if certFile != "" && keyFile != "" && caCertFile != "" {
		caCert, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig := &tls.Config{
			ClientCAs: caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
		tlsConfig.BuildNameToCertificate()
		server := &http.Server{
			Addr:      endpoint,
			TLSConfig: tlsConfig,
		}
		server.TLSConfig = tlsConfig
	}

	// check if HTTPS is requested
	if certFile == "" || keyFile == "" {
		// simple HTTP
		server.ListenAndServe()
	} else {
		// HTTPS
		server.ListenAndServeTLS(certFile, keyFile)
	}
}
