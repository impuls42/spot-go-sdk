package rxtspot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rackerlabs/spot-sdk/rxtspot/httpclient" // Updated import path
)

// RackspaceSpotClient is the main client for interacting with the Rackspace Spot API.
type RackspaceSpotClient struct {
	BaseURL      string
	OAuthURL     string
	HTTPClient   *http.Client
	AccessToken  string
	RefreshToken string
	Timeout      time.Duration
}

// NewClient creates a new RackspaceSpotClient with the given refresh token.
func NewClient(refreshToken string) *RackspaceSpotClient {
	return &RackspaceSpotClient{
		BaseURL:      "https://spot.rackspace.com",
		OAuthURL:     "https://login.spot.rackspace.com",
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
		RefreshToken: refreshToken,
		Timeout:      30 * time.Second,
	}
}

// Authenticate exchanges the refresh token for an access token.
func (c *RackspaceSpotClient) Authenticate(ctx context.Context) error {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", "mwG3lUMV8KyeMqHe4fJ5Bb3nM1vBvRNa")
	form.Set("refresh_token", c.RefreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.OAuthURL+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("authentication failed: " + resp.Status)
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
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

// ListOrganizations retrieves all organizations accessible by the user.
func (c *RackspaceSpotClient) ListOrganizations(ctx context.Context) ([]Organization, error) {
	url := fmt.Sprintf("%s/apis/auth.ngpc.rxt.io/v1/organizations", c.BaseURL) // Corrected endpoint
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list organizations: %s", string(b))
	}

	// Adjust unmarshaling to handle the correct JSON structure
	var response struct {
		Organizations []Organization `json:"organizations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Organizations, nil
}

// ListCloudspaces retrieves all cloudspaces in a namespace.
func (c *RackspaceSpotClient) ListCloudspaces(ctx context.Context, namespace string) ([]Cloudspace, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/cloudspaces", c.BaseURL, namespace)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("access denied: you do not have permission to list cloudspaces in the organization '%s'", namespace)
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list cloudspaces: %s", string(b))
	}
	var cloudspaces []Cloudspace
	if err := json.NewDecoder(resp.Body).Decode(&cloudspaces); err != nil {
		return nil, err
	}
	return cloudspaces, nil
}

// CreateCloudspace creates a new cloudspace in the given namespace.
func (c *RackspaceSpotClient) CreateCloudspace(ctx context.Context, cs Cloudspace) (*Cloudspace, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/cloudspaces", c.BaseURL, cs.Namespace)
	body, err := json.Marshal(cs)
	if err != nil {
		return nil, err
	}
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodPost, url, body, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create cloudspace: %s", string(b))
	}
	var created Cloudspace
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, err
	}
	return &created, nil
}

// DeleteCloudspace deletes a cloudspace by name in the given namespace.
func (c *RackspaceSpotClient) DeleteCloudspace(ctx context.Context, namespace, name string) error {
	url := fmt.Sprintf("%s/api/namespaces/%s/cloudspaces/%s", c.BaseURL, namespace, name)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodDelete, url, nil, c.authHeader())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete cloudspace: %s", string(b))
	}
	return nil
}

// GetCloudspace retrieves a cloudspace by name in the given namespace.
func (c *RackspaceSpotClient) GetCloudspace(ctx context.Context, namespace, name string) (*Cloudspace, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/cloudspaces/%s", c.BaseURL, namespace, name)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get cloudspace: %s", string(b))
	}
	var cs Cloudspace
	if err := json.NewDecoder(resp.Body).Decode(&cs); err != nil {
		return nil, err
	}
	return &cs, nil
}

// ListSpotNodePools retrieves all spot node pools in a namespace.
func (c *RackspaceSpotClient) ListSpotNodePools(ctx context.Context, namespace string) ([]SpotNodePool, error) {
	// TODO: Implement API call to /api/namespaces/{namespace}/spotnodepools
	return nil, nil
}

// CreateSpotNodePool creates a new spot node pool in the given namespace.
func (c *RackspaceSpotClient) CreateSpotNodePool(ctx context.Context, pool SpotNodePool) (*SpotNodePool, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/spotnodepools", c.BaseURL, pool.Namespace)
	body, err := json.Marshal(pool)
	if err != nil {
		return nil, err
	}
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodPost, url, body, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create spot node pool: %s", string(b))
	}
	var created SpotNodePool
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, err
	}
	return &created, nil
}

// DeleteSpotNodePool deletes a spot node pool by name in the given namespace.
func (c *RackspaceSpotClient) DeleteSpotNodePool(ctx context.Context, namespace, name string) error {
	url := fmt.Sprintf("%s/api/namespaces/%s/spotnodepools/%s", c.BaseURL, namespace, name)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodDelete, url, nil, c.authHeader())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete spot node pool: %s", string(b))
	}
	return nil
}

// GetSpotNodePool retrieves a spot node pool by name in the given namespace.
func (c *RackspaceSpotClient) GetSpotNodePool(ctx context.Context, namespace, name string) (*SpotNodePool, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/spotnodepools/%s", c.BaseURL, namespace, name)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get spot node pool: %s", string(b))
	}
	var pool SpotNodePool
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, err
	}
	return &pool, nil
}

// ListOnDemandNodePools retrieves all on-demand node pools in a namespace.
func (c *RackspaceSpotClient) ListOnDemandNodePools(ctx context.Context, namespace string) ([]OnDemandNodePool, error) {
	// TODO: Implement API call to /api/namespaces/{namespace}/ondemandnodepools
	return nil, nil
}

// CreateOnDemandNodePool creates a new on-demand node pool in the given namespace.
func (c *RackspaceSpotClient) CreateOnDemandNodePool(ctx context.Context, pool OnDemandNodePool) (*OnDemandNodePool, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/ondemandnodepools", c.BaseURL, pool.Namespace)
	body, err := json.Marshal(pool)
	if err != nil {
		return nil, err
	}
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodPost, url, body, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create on-demand node pool: %s", string(b))
	}
	var created OnDemandNodePool
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, err
	}
	return &created, nil
}

// DeleteOnDemandNodePool deletes an on-demand node pool by name in the given namespace.
func (c *RackspaceSpotClient) DeleteOnDemandNodePool(ctx context.Context, namespace, name string) error {
	url := fmt.Sprintf("%s/api/namespaces/%s/ondemandnodepools/%s", c.BaseURL, namespace, name)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodDelete, url, nil, c.authHeader())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete on-demand node pool: %s", string(b))
	}
	return nil
}

// GetOnDemandNodePool retrieves an on-demand node pool by name in the given namespace.
func (c *RackspaceSpotClient) GetOnDemandNodePool(ctx context.Context, namespace, name string) (*OnDemandNodePool, error) {
	url := fmt.Sprintf("%s/api/namespaces/%s/ondemandnodepools/%s", c.BaseURL, namespace, name)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get on-demand node pool: %s", string(b))
	}
	var pool OnDemandNodePool
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, err
	}
	return &pool, nil
}

// ListRegions retrieves all available regions.
func (c *RackspaceSpotClient) ListRegions(ctx context.Context) ([]Region, error) {
	url := fmt.Sprintf("%s/api/regions", c.BaseURL)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list regions: %s", string(b))
	}
	var regions []Region
	if err := json.NewDecoder(resp.Body).Decode(&regions); err != nil {
		return nil, err
	}
	return regions, nil
}

// GetRegion retrieves a region by name.
func (c *RackspaceSpotClient) GetRegion(ctx context.Context, name string) (*Region, error) {
	url := fmt.Sprintf("%s/api/regions/%s", c.BaseURL, name)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get region: %s", string(b))
	}
	var region Region
	if err := json.NewDecoder(resp.Body).Decode(&region); err != nil {
		return nil, err
	}
	return &region, nil
}

// ListServerClasses retrieves all available server classes.
func (c *RackspaceSpotClient) ListServerClasses(ctx context.Context) ([]ServerClassInfo, error) {
	url := fmt.Sprintf("%s/api/serverclasses", c.BaseURL)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list server classes: %s", string(b))
	}
	var classes []ServerClassInfo
	if err := json.NewDecoder(resp.Body).Decode(&classes); err != nil {
		return nil, err
	}
	return classes, nil
}

// GetServerClass retrieves a server class by name.
func (c *RackspaceSpotClient) GetServerClass(ctx context.Context, name string) (*ServerClassInfo, error) {
	url := fmt.Sprintf("%s/api/serverclasses/%s", c.BaseURL, name)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get server class: %s", string(b))
	}
	var class ServerClassInfo
	if err := json.NewDecoder(resp.Body).Decode(&class); err != nil {
		return nil, err
	}
	return &class, nil
}

// GetPriceHistory retrieves the price history for a server class.
func (c *RackspaceSpotClient) GetPriceHistory(ctx context.Context, serverClass string) (*PriceHistory, error) {
	url := fmt.Sprintf("%s/api/serverclasses/%s/pricehistory", c.BaseURL, serverClass)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get price history: %s", string(b))
	}
	var history PriceHistory
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, err
	}
	return &history, nil
}
