package rxtspot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// Config holds the configuration for the Rackspace Spot API client
type Config struct {
	BaseURL      string
	OAuthURL     string
	AccessToken  string
	RefreshToken string
	HTTPClient   *http.Client // Optional custom HTTP client
	
	// Retry configuration
	MaxRetries   int           // Maximum number of retries (default: 3)
	RetryWaitMin time.Duration // Minimum time between retries (default: 1s)
	RetryWaitMax time.Duration // Maximum time between retries (default: 30s)

	// HTTP client configuration
	RequestTimeout  time.Duration // Overall request timeout (default: 30s)
	IdleConnTimeout time.Duration // Maximum idle connection timeout (default: 90s)
	MaxIdleConns    int           // Maximum idle connections across all hosts (default: 100)
}

type RackspaceSpotClient struct {
	BaseURL      string
	OAuthURL     string
	HTTPClient   *http.Client
	Token        string
	RefreshToken string
	// New fields for retry and timeout
	MaxRetries      int           // Maximum number of retries (default: 3)
	RetryWaitMin    time.Duration // Minimum time to wait between retries (default: 1s)
	RetryWaitMax    time.Duration // Maximum time to wait between retries (default: 30s)
	RequestTimeout  time.Duration // Timeout for each HTTP request (default: 30s)
	IdleConnTimeout time.Duration // Maximum idle connection timeout (default: 90s)
}

type HTTPStatusError struct {
	StatusCode int
	Body       string
}

// NewSpotClient creates a new Rackspace Spot API client with secure defaults
func NewSpotClient(cfg *Config) (*RackspaceSpotClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	// Set default values if not provided
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryWaitMin == 0 {
		cfg.RetryWaitMin = time.Second
	}
	if cfg.RetryWaitMax == 0 {
		cfg.RetryWaitMax = 30 * time.Second
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 30 * time.Second
	}
	if cfg.IdleConnTimeout == 0 {
		cfg.IdleConnTimeout = 90 * time.Second
	}

	// Set default values if not provided
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 100
	}

	// Clone the default transport and override specific settings
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.IdleConnTimeout = cfg.IdleConnTimeout
	transport.MaxIdleConns = cfg.MaxIdleConns

	// Set dial timeout to match request timeout
	dialer := &net.Dialer{
		Timeout:   cfg.RequestTimeout,
		KeepAlive: 30 * time.Second,
	}
	transport.DialContext = dialer.DialContext

	// Create the HTTP client with timeouts
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Transport: transport,
			Timeout:   cfg.RequestTimeout,
		}
	}

	// Create HTTP client if not provided
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Transport: transport,
			Timeout:   cfg.RequestTimeout,
		}
	}

	return &RackspaceSpotClient{
		BaseURL:      cfg.BaseURL,
		OAuthURL:     cfg.OAuthURL,
		HTTPClient:   client,
		Token:        cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		MaxRetries:   cfg.MaxRetries,
		RetryWaitMin: cfg.RetryWaitMin,
		RetryWaitMax: cfg.RetryWaitMax,
	}, nil
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

// handleAPIError processes HTTP errors and returns user-friendly error messages
func (c *RackspaceSpotClient) handleAPIError(err error, resourceType, resourceName string, operation string) error {
	if err == nil {
		return nil
	}

	httpErr, ok := err.(*HTTPStatusError)
	if !ok {
		return fmt.Errorf("failed to %s %s: %w", operation, resourceType, err)
	}

	switch httpErr.StatusCode {
	case http.StatusForbidden, http.StatusUnauthorized:
		// Try to extract the detailed error message from the response body
		var apiErr struct {
			Message string `json:"message"`
		}
		if httpErr.Body != "" {
			if json.Unmarshal([]byte(httpErr.Body), &apiErr) == nil && apiErr.Message != "" {
				return fmt.Errorf("access denied: %s", apiErr.Message)
			}
		}
		return fmt.Errorf("access denied: you do not have permission to %s the %s '%s'", operation, resourceType, resourceName)

	case http.StatusNotFound:
		return fmt.Errorf("%s '%s' not found", resourceType, resourceName)

	case http.StatusBadRequest, http.StatusConflict:
		if httpErr.Body != "" {
			return fmt.Errorf("invalid request: %s", httpErr.Body)
		}
		return fmt.Errorf("invalid request: failed to %s %s", operation, resourceType)

	default:
		if httpErr.Body != "" {
			return fmt.Errorf("API error (HTTP %d): %s", httpErr.StatusCode, httpErr.Body)
		}
		return fmt.Errorf("API error (HTTP %d): failed to %s %s", httpErr.StatusCode, operation, resourceType)
	}
}

// validateURL performs basic validation of a URL
func validateURL(urlStr string) error {
	// Check for empty URL
	if urlStr == "" {
		return errors.New("URL cannot be empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme (http or https)
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s, must be http or https", scheme)
	}

	// Check hostname
	if parsedURL.Host == "" {
		return errors.New("URL must contain a host")
	}

	// Validate hostname
	hostname := parsedURL.Hostname()
	if !isValidHostname(hostname) {
		return fmt.Errorf("invalid hostname: %s", hostname)
	}

	// Check for invalid characters in path
	if strings.ContainsAny(parsedURL.Path, "\x00\r\n") {
		return errors.New("URL path contains invalid characters")
	}

	// Validate port if present
	if port := parsedURL.Port(); port != "" {
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			return fmt.Errorf("invalid port number")
		}
	}

	return nil
}

// calculateBackoff calculates the backoff duration using exponential backoff with jitter
func (c *RackspaceSpotClient) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff
	min := float64(c.RetryWaitMin)
	max := float64(c.RetryWaitMax)

	// Exponential backoff with jitter
	backoff := min * math.Pow(2, float64(attempt))
	if backoff > max {
		backoff = max
	}

	// Add jitter
	jitter := 0.2 * backoff * (rand.Float64()*2 - 1) // ±20% jitter
	backoff += jitter

	// Ensure we don't exceed max wait time
	if backoff > max {
		backoff = max
	}

	return time.Duration(backoff)
}

// doRequest performs an HTTP request with the given method, URL, body, and headers.
// It handles authentication, rate limiting, and error handling.
func (c *RackspaceSpotClient) doRequest(ctx context.Context, method, url string, body []byte, headers map[string]string, out interface{}) error {
	if err := validateURL(url); err != nil {
		return err
	}

	var lastErr error

	// create a new request for every new attempt
	for attempt := 1; attempt <= c.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			klog.Errorf("doRequest: failed to create request: %v", err)
			return err
		}

		// Add headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}
		if method == http.MethodPatch {
			req.Header.Set("Content-Type", "application/merge-patch+json")
		}

		klog.V(1).Infof("[%s] %s (attempt %d/%d)", method, url, attempt, c.MaxRetries)

		// Calculate backoff once per attempt
		backoff := c.calculateBackoff(attempt)

		// Perform HTTP request
		start := time.Now()
		resp, err := c.HTTPClient.Do(req)
		duration := time.Since(start)

		// Handle request errors
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed after %v: %w", duration, err)
			klog.Warningf("Request failed (attempt %d/%d): %v", attempt, c.MaxRetries, lastErr)

			if attempt < c.MaxRetries {
				klog.Warningf("Retrying in %v...", backoff)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					continue
				}
			}
			return lastErr
		}

			// Handle response
		if resp == nil {
			lastErr = fmt.Errorf("received nil response from server")
			if attempt < c.MaxRetries {
				klog.Warningf("Retrying in %v...", backoff)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					continue
				}
			}
			return lastErr
		}

		defer resp.Body.Close()

		// Log request completion
		klog.V(2).Infof("Request completed in %v with status: %d %s",
			duration, resp.StatusCode, http.StatusText(resp.StatusCode))

		// Handle successful response (2xx)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out != nil {
				if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}
				klog.V(4).Infof("Decoded object: %+v", out)
			}
			return nil
		}

		// Read response body for error cases
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response: %w", err)
		}

		// For 4xx errors, don't retry - return the error immediately
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return &HTTPStatusError{
				StatusCode: resp.StatusCode,
				Body:       string(body),
			}
		}

		// For 5xx errors, retry if we have attempts left
		lastErr = &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}

		if attempt < c.MaxRetries {
			klog.Warningf("Server error (attempt %d/%d), retrying in %v: %v",
				attempt, c.MaxRetries, backoff, lastErr)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}

		return fmt.Errorf("request failed after %d attempts: %w", c.MaxRetries, lastErr)
	}

	return lastErr
}

// isAlphanumeric returns true if the rune is an ASCII letter or digit
func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// isValidHostname checks if a string is a valid hostname (supports both domains and IPs)
func isValidHostname(hostname string) bool {
	// First try to parse as IP address
	if ip := net.ParseIP(hostname); ip != nil {
		return true
	}

	// Then check as domain name
	if len(hostname) > 253 || len(hostname) == 0 {
		return false
	}

	// Check each label in the hostname
	for _, label := range strings.Split(hostname, ".") {
		// Check label length
		if len(label) > 63 || len(label) == 0 {
			return false
		}

		// Check each character in the label
		for i, r := range label {
			// Allow alphanumeric and hyphens (but not as first or last character)
			if !isAlphanumeric(r) && r != '-' {
				return false
			}

			// First and last character must be alphanumeric
			if (i == 0 || i == len(label)-1) && !isAlphanumeric(r) {
				return false
			}
		}
	}

	return true
}

// APIError represents an error response from the API
type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
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
