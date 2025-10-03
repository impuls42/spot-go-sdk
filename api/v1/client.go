package rxtspot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/kelseyhightower/envconfig"
)

type RetryConfig struct {
	MaxRetries   int           `envconfig:"RXTSPOT_MAX_RETRIES" default:"3"`
	RetryWaitMax time.Duration `envconfig:"RXTSPOT_RETRY_WAIT_MAX" default:"30s"`
	RetryWaitMin time.Duration `envconfig:"RXTSPOT_RETRY_WAIT_MIN" default:"1s"`
}

// Config holds the configuration for the Rackspace Spot API client
// Config holds the configuration for the Rackspace Spot API client
type Config struct {
	BaseURL      string       `envconfig:"RXTSPOT_BASE_URL" default:"https://spot.rackspace.com"`
	OAuthURL     string       `envconfig:"RXTSPOT_OAUTH_URL" default:"https://login.spot.rackspace.com"`
	AccessToken  string       `envconfig:"RXTSPOT_ACCESS_TOKEN"`
	RefreshToken string       `envconfig:"RXTSPOT_REFRESH_TOKEN"`
	HTTPClient   *http.Client `ignored:"true"` // Custom HTTP client (not configurable via env)

	// Retry configuration
	RetryConfig RetryConfig `envconfig:"RXTSPOT_RETRY_CONFIG"`

	// HTTP client configuration
	RequestTimeout  time.Duration `envconfig:"RXTSPOT_REQUEST_TIMEOUT" default:"30s"`
	IdleConnTimeout time.Duration `envconfig:"RXTSPOT_IDLE_CONN_TIMEOUT" default:"90s"`
	MaxIdleConns    int           `envconfig:"RXTSPOT_MAX_IDLE_CONNS" default:"100"`
}

type RackspaceSpotClient struct {
	BaseURL      string
	OAuthURL     string
	HTTPClient   *http.Client
	Token        string
	RefreshToken string
	// Retry configuration
	RetryConfig RetryConfig

	RequestTimeout  time.Duration // Timeout for each HTTP request (default: 30s)
	IdleConnTimeout time.Duration // Maximum idle connection timeout (default: 90s)
}

type HTTPStatusError struct {
	StatusCode int
	Body       string
}

// NewSpotClient creates a new Rackspace Spot API client with secure defaults.
// Configuration is loaded in the following order:
// 1. Values provided in the config parameter
// 2. Environment variables (with RXTSPOT_ prefix)
// 3. Default values (for fields with default tags)
func NewSpotClient(cfg *Config) (*RackspaceSpotClient, error) {
	// Process environment variables if config is nil
	if cfg == nil {
		cfg = &Config{}
	}

	// Load configuration from environment variables
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment config: %w", err)
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required (set RXTSPOT_BASE_URL)")
	}

	// Set up HTTP transport
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.IdleConnTimeout = cfg.IdleConnTimeout
	transport.MaxIdleConns = cfg.MaxIdleConns

	// Configure dialer with timeout
	dialer := &net.Dialer{
		Timeout:   cfg.RequestTimeout,
		KeepAlive: 30 * time.Second,
	}
	transport.DialContext = dialer.DialContext

	// Create or configure HTTP client
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Transport: transport,
			Timeout:   cfg.RequestTimeout,
		}
	}

	return &RackspaceSpotClient{
		BaseURL:         cfg.BaseURL,
		OAuthURL:        cfg.OAuthURL,
		HTTPClient:      client,
		Token:           cfg.AccessToken,
		RefreshToken:    cfg.RefreshToken,
		RetryConfig:     cfg.RetryConfig,
		RequestTimeout:  cfg.RequestTimeout,
		IdleConnTimeout: cfg.IdleConnTimeout,
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

// doRequest performs an HTTP request with the given method, URL, body, and headers.
// It handles authentication, rate limiting, and error handling.
func (c *RackspaceSpotClient) doRequest(ctx context.Context, method, url string, body []byte, headers map[string]string, out interface{}) error {
	if err := validateURL(url); err != nil {
		return err
	}

	operation := func() (string, error) {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return "", backoff.Permanent(fmt.Errorf("failed to create request: %w", err))
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}
		if method == http.MethodPatch {
			req.Header.Set("Content-Type", "application/merge-patch+json")
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			// retryable error
			return "", fmt.Errorf("http request failed: %w", err)
		}
		defer resp.Body.Close()

		// Treat 4xx as permanent errors (no retry)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return "", backoff.Permanent(&HTTPStatusError{
				StatusCode: resp.StatusCode,
				Body:       string(bodyBytes),
			})
		}

		// Treat 5xx as retryable errors
		if resp.StatusCode >= 500 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("server error: %d %s", resp.StatusCode, string(bodyBytes))
		}

		// Success case
		if out != nil {
			if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
				return "", backoff.Permanent(fmt.Errorf("failed to decode response: %w", err))
			}
		}
		return "", nil
	}

	_, err := backoff.Retry(ctx, operation, backoff.WithMaxTries(uint(c.RetryConfig.MaxRetries)), backoff.WithMaxElapsedTime(time.Duration(c.RetryConfig.RetryWaitMax)))
	if err != nil {
		return err
	}

	return nil
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
