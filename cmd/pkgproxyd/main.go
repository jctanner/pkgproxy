package main

import (
	"crypto/tls"
	"log"
	"os"

	"net/http"

    "github.com/jctanner/pkgproxy/pkg/proxycore/generator"
    "github.com/jctanner/pkgproxy/pkg/proxycore/handlers"
)

const (
	cacheDir   = "/src/packages"
	listenAddr = ":3128"
)

func main() {
	// Ensure the cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("Failed to create cache directory: %v", err)
	}

	// Prepare the HTTPS server with TLS configuration
	tlsConfig := &tls.Config{
		GetCertificate: generator.GetCertificateFunc(), // Assuming generator is defined elsewhere
	}

	httpsServer := &http.Server{
		Addr:      ":443", // Default port for HTTPS
		Handler:   http.HandlerFunc(handlers.ProxyHandler),
		TLSConfig: tlsConfig,
	}

	// Start the HTTPS server in a new goroutine
	go func() {
		log.Println("Starting HTTPS server on port 443")
		err := httpsServer.ListenAndServeTLS("", "") // Cert and key are provided by GetCertificate
		if err != nil {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	}()

	// Start the HTTP server on the main goroutine
	// No need for TLS configuration here
	httpServer := &http.Server{
		Addr:    ":80", // Default port for HTTP
		Handler: http.HandlerFunc(handlers.ProxyHandler),
	}

	log.Println("Starting HTTP server on port 80")
	err := httpServer.ListenAndServe() // This will block
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
