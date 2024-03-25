package handlers

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

    "github.com/jctanner/pkgproxy/pkg/proxycore/caching"
    "github.com/jctanner/pkgproxy/pkg/proxycore/generator"
)

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Log every request
	log.Printf("Request: %s %s", r.Method, r.URL.String())

	if r.Method == "CONNECT" {
		// Assuming HandleConnectRequest is implemented elsewhere
		HandleConnectRequest(w, r)
		return
	}

	// Attempt to retrieve the cached content or fetch it if not cached
	cacheFileName, resp, err := caching.GetCachedUrl(r)
	if err != nil {
		// Corrected the Printf function call and added a missing parenthesis
		log.Printf("Unable to get the cached file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cacheFile, _ := os.Open(cacheFileName)
	defer cacheFile.Close()

	// If the content was just fetched and cached, write the headers and status code to the response
	if resp != nil {
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
	}

	// Seek to the beginning of the cache file before reading
	if _, err := cacheFile.Seek(0, io.SeekStart); err != nil {
		log.Printf("Failed to seek cache file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy the content of the cache file to the HTTP response writer
	if _, err := io.Copy(w, cacheFile); err != nil {
		log.Printf("Failed to write cached data to response: %v", err)
		// Note: At this point, some data might have already been written to w,
		// so it might not be correct to send an HTTP error.
	}

	cacheFile.Close()
}

func HandleConnectRequest(w http.ResponseWriter, req *http.Request) {
	destConn, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		http.Error(w, "Failed to connect to destination", http.StatusInternalServerError)
		return
	}
	defer destConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "HTTP Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection Established to the client
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		log.Printf("Error sending connection established response: %v", err)
		return
	}

	// Extract the hostname from the request URL
	host := strings.Split(req.URL.Host, ":")[0]

	// Generate a TLS certificate for the requested host
	cert, err := generator.GenerateTLSCertificateForHost(host)
	if err != nil {
		log.Printf("Failed to generate TLS certificate: %v", err)
		return
	}

	// Create a TLS config using the generated certificate
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Establish a TLS connection with the client using the generated certificate
	tlsConn := tls.Server(clientConn, tlsConfig)
	err = tlsConn.Handshake()
	if err != nil {
		log.Printf("TLS handshake error: %v", err)
		return
	}

	// Now, you have a TLS connection with the client (tlsConn)
	// and a TCP connection with the destination server (destConn).
	// You can start forwarding data between client and server.
	// This involves reading from one connection and writing to the other, and vice versa.
	HandleAndForwardHTTPRequest(tlsConn, destConn)
}

func HandleAndForwardHTTPRequest(clientConn *tls.Conn, destConn net.Conn) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		// Wrap the client connection in a bufio.Reader to use with http.ReadRequest
		clientReader := bufio.NewReader(clientConn)

		// Parse the HTTP request from the client
		clientReq, err := http.ReadRequest(clientReader)
		if err != nil {
			log.Printf("Failed to read request from client: %v", err)
			return
		}
		defer clientReq.Body.Close()

		fullUrl := caching.GetFullURL(clientReq)

		// You now have the client's request, including the method, host, and path
		log.Printf("Client requested %s", fullUrl)

		// Use GetCachedUrl to either fetch the content or retrieve it from cache
		cacheFileName, _, err := caching.GetCachedUrl(clientReq)
		if err != nil {
			log.Printf("Error using GetCachedUrl: %v", err)
			return
		}
		//if cacheFileName != nil {
		//	defer cacheFile.Close()
		//}

		contentType := caching.URLToContentType(fullUrl)
		log.Printf("CONTENT-TYPE %s", contentType)

		// Write the status line and headers back to the client
		_, err = clientConn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			log.Printf("Error writing response status to client: %v", err)
			return
		}

		_, err = clientConn.Write([]byte("Content-Type: " + contentType + "\r\n\r\n")) // End of headers
		if err != nil {
			log.Printf("Error writing headers to client: %v", err)
			return
		}

		cacheFile, _ := os.Open(cacheFileName)

		// Stream the cached content back to the client
		if _, err = io.Copy(clientConn, cacheFile); err != nil {
			log.Printf("Error streaming cached content to client: %v", err)
		}

		cacheFile.Close()

	}()

	wg.Wait()
	clientConn.Close()
}
