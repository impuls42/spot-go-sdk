package rxtspot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"k8s.io/klog/v2"
)

// Configuration struct for RackspaceSpotClient
type Config struct {
	BaseURL      string
	OAuthURL     string
	HTTPClient   *http.Client
	RefreshToken string
	Timeout      time.Duration
	LogLevel     int
}

// RackspaceSpotClient is the main client for interacting with the Rackspace Spot API.
type RackspaceSpotClient struct {
	BaseURL      string
	OAuthURL     string
	HTTPClient   *http.Client
	AccessToken  string
	RefreshToken string
	Timeout      time.Duration
}

type HTTPStatusError struct {
	StatusCode int
	Body       string
}

// NewClient creates a new RackspaceSpotClient with the given configuration.
func NewClient(cfg Config) *RackspaceSpotClient {
	return &RackspaceSpotClient{
		BaseURL:      cfg.BaseURL,
		OAuthURL:     cfg.OAuthURL,
		HTTPClient:   cfg.HTTPClient,
		RefreshToken: cfg.RefreshToken,
		Timeout:      cfg.Timeout,
	}
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

func (c *RackspaceSpotClient) doRequest(ctx context.Context, method, url string, body []byte, headers map[string]string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		klog.Errorf("doRequest: failed to create request: %v", err)
		return err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// ----- Request logging -----
	klog.V(1).Infof("[%s] %s", method, url)
	if len(headers) > 0 {
		klog.V(2).Infof("Request headers: %+v", headers)
	}
	if len(body) > 0 {
		klog.V(3).Infof("Request body: %s", string(body))
	}

	// ----- Perform HTTP request -----
	start := time.Now()

	//fmt.Printf("url - %v\n", url)
	//fmt.Printf("request body - %v\n", string(body))
	resp, err := c.HTTPClient.Do(req)
	duration := time.Since(start)
	if err != nil {
		klog.Errorf("HTTP request failed after %v: %v", duration, err)
		return err
	}
	defer resp.Body.Close()

	klog.V(2).Infof("Response status: %d %s (duration: %v)", resp.StatusCode, http.StatusText(resp.StatusCode), duration)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return &HTTPStatusError{StatusCode: resp.StatusCode, Body: string(b)}
	}

	// ----- Response logging -----
	if klog.V(3).Enabled() {
		klog.V(3).Infof("Response headers: %+v", resp.Header)
	}
	if klog.V(4).Enabled() {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		klog.V(4).Infof("Response body: %s", string(respBody))
	}

	// ----- Decode if out != nil -----
	if out != nil {
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(out); err != nil && err != io.EOF {
			klog.Errorf("Failed to decode JSON: %v", err)
			return fmt.Errorf("decode json: %w", err)
		}
		klog.V(4).Infof("Decoded object ****: %+v", out)
	}

	return nil
}

// Helper matchers for consumers (like your CLI)
func IsNotFound(err error) bool {
	var e *HTTPStatusError
	if errors.As(err, &e) {
		return e.StatusCode == http.StatusNotFound
	}
	return false
}

func IsForbidden(err error) bool {
	var e *HTTPStatusError
	if errors.As(err, &e) {
		return e.StatusCode == http.StatusForbidden
	}
	return false
}

func IsConflict(err error) bool {
	var e *HTTPStatusError
	if errors.As(err, &e) {
		return e.StatusCode == http.StatusConflict
	}
	return false
}
