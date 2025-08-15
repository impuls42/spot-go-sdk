package httpclient

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"time"

// 	"k8s.io/klog"
// )

// var verbose bool // Package-level variable to control verbosity

// // SetVerbose sets the verbosity level for HTTP logging.
// func SetVerbose(v bool) {
// 	verbose = v
// }

// DoRequest performs an HTTP request and logs details if verbose is enabled.
// func DoRequest(ctx context.Context, client *http.Client, method, url string, body []byte, headers map[string]string) (*http.Response, error) {
// 	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
// 	if err != nil {
// 		return nil, err
// 	}
// 	klog.V(3).Infof("[%s] %s", method, url)

// 	// Add headers to the request
// 	for key, value := range headers {
// 		req.Header.Set(key, value)
// 	}
// 	// Log the request if verbose is enabled
// 	if verbose {
// 		fmt.Printf("HTTP Request:\nMethod: %s\nURL: %s\nHeaders: %v\nBody: %s\n", method, url, headers, string(body))
// 	}

// 	// Perform the request
// 	start := time.Now()
// 	resp, err := client.Do(req)
// 	duration := time.Since(start)

// 	// Log the response if verbose is enabled
// 	if verbose {
// 		if err != nil {
// 			fmt.Printf("HTTP Response:\nError: %v\nDuration: %v\n", err, duration)
// 		} else {
// 			respBody, _ := io.ReadAll(resp.Body)
// 			resp.Body = io.NopCloser(bytes.NewReader(respBody)) // Reassign body for further use
// 			fmt.Printf("HTTP Response:\nStatus: %s\nHeaders: %v\nBody: %s\nDuration: %v\n", resp.Status, resp.Header, string(respBody), duration)
// 		}
// 	}

// 	return resp, err
// }

// package httpclient

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"time"

// 	"k8s.io/klog/v2"
// )

// func (c *v1.RackspaceSpotClient) doRequest(ctx context.Context, method, url string, body []byte, headers map[string]string, out interface{}) error {
// 	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
// 	if err != nil {
// 		klog.Errorf("doRequest: failed to create request: %v", err)
// 		return err
// 	}

// 	// Add headers
// 	for key, value := range headers {
// 		req.Header.Set(key, value)
// 	}

// 	// ----- Request logging -----
// 	klog.V(1).Infof("[%s] %s", method, url)
// 	if len(headers) > 0 {
// 		klog.V(2).Infof("Request headers: %+v", headers)
// 	}
// 	if len(body) > 0 {
// 		klog.V(3).Infof("Request body: %s", string(body))
// 	}

// 	// ----- Perform HTTP request -----
// 	start := time.Now()
// 	resp, err := c.HTTPClient.Do(req)
// 	duration := time.Since(start)
// 	if err != nil {
// 		klog.Errorf("HTTP request failed after %v: %v", duration, err)
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	klog.V(2).Infof("Response status: %d %s (duration: %v)", resp.StatusCode, http.StatusText(resp.StatusCode), duration)

// 	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
// 		b, _ := io.ReadAll(resp.Body)
// 		klog.Warningf("Non-OK response (%d): %s", resp.StatusCode, string(b))
// 		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
// 	}

// 	// ----- Response logging -----
// 	if klog.V(3).Enabled() {
// 		klog.V(3).Infof("Response headers: %+v", resp.Header)
// 	}
// 	if klog.V(4).Enabled() {
// 		respBody, _ := io.ReadAll(resp.Body)
// 		resp.Body = io.NopCloser(bytes.NewReader(respBody))
// 		klog.V(4).Infof("Response body: %s", string(respBody))
// 	}

// 	// ----- Decode if out != nil -----
// 	if out != nil {
// 		dec := json.NewDecoder(resp.Body)
// 		if err := dec.Decode(out); err != nil && err != io.EOF {
// 			klog.Errorf("Failed to decode JSON: %v", err)
// 			return fmt.Errorf("decode json: %w", err)
// 		}
// 		klog.V(4).Infof("Decoded object: %+v", out)
// 	}

// 	return nil
// }
