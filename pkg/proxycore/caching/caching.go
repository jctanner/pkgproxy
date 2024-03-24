package caching

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

    "github.com/jctanner/pkgproxy/pkg/proxycore/hashing"
)

const (
	cacheDir   = "/src/packages"
	listenAddr = ":3128"
)

func URLToContentType(fullUrl string) string {
	if strings.Contains(fullUrl, "pypi.org/simple") {
		return "text/html"
	}
	if strings.Contains(fullUrl, ".whl.metadata") {
		return "text/plain; charset=UTF-8"
	}
	if strings.Contains(fullUrl, ".whl") {
		return "application/octet-stream"
	}
	return "application/octet-stream"
}

func GetFullURL(req *http.Request) string {
	// Default to http unless determined otherwise
	proto := "http"

	// Check if the request is over TLS
	if req.TLS != nil || strings.Contains(req.Host, "pypi.org") || strings.Contains(req.Host, "files.pythonhosted.org") {
		proto = "https"
	}

	// Alternatively, or in addition, check the X-Forwarded-Proto header,
	// which can be set by proxies to indicate the original protocol
	if forwardedProto := req.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		proto = forwardedProto
	}

	// Construct the full URL
	host := req.Host
	path := req.URL.Path
	query := req.URL.RawQuery

	fullURL := fmt.Sprintf("%s://%s%s", proto, host, path)
	if query != "" {
		fullURL += "?" + query
	}

	return fullURL
}

func GetCachedUrl(req *http.Request) (string, *http.Response, error) {

	//log.Printf("get cache file for host: %s, path: %s", req.Host, req.URL.Path)
	//log.Printf("get cache file for %s", req.URL.String())
	fullUrl := GetFullURL(req)
	log.Printf("get cache file for %s", fullUrl)

	var cacheFilePath string

	// Construct the local cache file path based on whether the URL is an RPM package
	if strings.HasSuffix(fullUrl, ".rpm") {
		cacheFilePath = filepath.Join(cacheDir, filepath.Base(fullUrl))
	} else if strings.Contains(fullUrl, ".whl.metadata") {
		cacheFilePath = filepath.Join(cacheDir, filepath.Base(fullUrl))
	} else if strings.Contains(fullUrl, ".whl") {
		cacheFilePath = filepath.Join(cacheDir, filepath.Base(fullUrl))
	} else {
		urlHash := hashing.Sha1Hash(fullUrl)
		cacheFileName := fmt.Sprintf("%s_%s", urlHash, filepath.Base(fullUrl))
		cacheFilePath = filepath.Join(cacheDir, cacheFileName)
	}

	// Check for a lock file and wait if it exists
	lockFilePath := cacheFilePath + ".lock"
	for {
		if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Check if the cache file already exists and return it if so
	if _, err := os.Stat(cacheFilePath); err == nil {
		/*
			_, err := os.Open(cacheFilePath)
			if err != nil {
				return cacheFilePath, nil, err
			}
		*/
		log.Printf("found %s and returning it", cacheFilePath)
		return cacheFilePath, nil, nil // Assuming no need to return the HTTP response if served from cache
	}

	// Create a lock file to signal that the cache file is being written
	log.Printf("creating lockfile %s", lockFilePath)
	lockFile, err := os.Create(lockFilePath)
	if err != nil {
		return cacheFilePath, nil, err
	}
	lockFile.Close()
	//defer os.Remove(lockFilePath)

	// Fetch the content from the URL
	log.Printf("get %s", fullUrl)
	resp, err := http.Get(fullUrl)
	if err != nil {
		return cacheFilePath, nil, err
	}
	//defer resp.Body.Close()

	// Create the cache file
	log.Printf("create %s", cacheFilePath)
	cacheFile, err := os.Create(cacheFilePath)
	if err != nil {
		return cacheFilePath, nil, err
	}
	//defer cacheFile.Close()

	// Copy the response body to the cache file
	log.Printf("copy response to %s", cacheFilePath)
	if _, err := io.Copy(cacheFile, resp.Body); err != nil {
		log.Printf("Failed to write data to cache file: %v", err)
		return cacheFilePath, nil, err
	}

	resp.Body.Close()
	cacheFile.Close()
	os.Remove(lockFilePath)

	// Return the cache file and the response
	// The caller needs to handle re-opening the cache file if necessary
	log.Printf("return %s handle", cacheFilePath)
	return cacheFilePath, resp, nil
}
