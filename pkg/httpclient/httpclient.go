package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

var verbose bool // Package-level variable to control verbosity

// SetVerbose sets the verbosity level for HTTP logging.
func SetVerbose(v bool) {
	verbose = v
}

// DoRequest performs an HTTP request and logs details if verbose is enabled.
func DoRequest(ctx context.Context, client *http.Client, method, url string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Add headers to the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	// Log the request if verbose is enabled
	if verbose {
		fmt.Printf("HTTP Request:\nMethod: %s\nURL: %s\nHeaders: %v\nBody: %s\n", method, url, headers, string(body))
	}

	// Perform the request
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	// Log the response if verbose is enabled
	if verbose {
		if err != nil {
			fmt.Printf("HTTP Response:\nError: %v\nDuration: %v\n", err, duration)
		} else {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewReader(respBody)) // Reassign body for further use
			fmt.Printf("HTTP Response:\nStatus: %s\nHeaders: %v\nBody: %s\nDuration: %v\n", resp.Status, resp.Header, string(respBody), duration)
		}
	}

	return resp, err
}
