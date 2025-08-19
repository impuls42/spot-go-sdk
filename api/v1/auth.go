package rxtspot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Authenticate exchanges the refresh token for an access token.
func (c *RackspaceSpotClient) Authenticate(ctx context.Context) error {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", "mwG3lUMV8KyeMqHe4fJ5Bb3nM1vBvRNa")
	form.Set("refresh_token", c.RefreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.OAuthURL+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed: %s", resp.Status)
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if tokenResp.IDToken == "" {
		return errors.New("no id_token in authentication response")
	}
	c.AccessToken = tokenResp.IDToken
	return nil
}

// authHeader returns the Authorization header for the current access token.
func (c *RackspaceSpotClient) authHeader() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + c.AccessToken,
		"Content-Type":  "application/json",
	}
}
