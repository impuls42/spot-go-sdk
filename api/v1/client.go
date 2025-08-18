package rxtspot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
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

// ListOrganizations retrieves all organizations accessible by the user.
func (c *RackspaceSpotClient) ListOrganizations(ctx context.Context) ([]Organization, error) {
	url := fmt.Sprintf("%s/apis/auth.ngpc.rxt.io/v1/organizations", c.BaseURL) // Correct URL

	// Structure for decoding
	var response struct {
		Organizations []Organization `json:"organizations"`
	}

	// Pass &response to doRequest so it decodes automatically
	err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &response)
	if err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("access denied: you do not have permission to list organizations")
		}
		return nil, err
	}
	return response.Organizations, nil
}

func (c *RackspaceSpotClient) getOrgIDIfExists(ctx context.Context, orgName string) (bool, string, error) {
	url := fmt.Sprintf("%s/apis/auth.ngpc.rxt.io/v1/organizations", c.BaseURL) // Correct URL

	// Structure for decoding
	var response struct {
		Organizations []Organization `json:"organizations"`
	}

	// Pass &response to doRequest so it decodes automatically
	err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &response)
	if err != nil {
		// Specific 403 handling if doRequest returns your custom HTTPStatusError
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return false, "", fmt.Errorf("access denied: you do not have permission to list organizations")
		}
		return false, "", err
	}

	for _, org := range response.Organizations {
		if org.Name == orgName {
			org.ID = strings.ReplaceAll(org.ID, "_", "-")
			org.ID = strings.ToLower(org.ID)
			return true, org.ID, nil
		}
	}
	return false, "", nil
}

// ListCloudspaces retrieves all cloudspaces in a namespace.
func (c *RackspaceSpotClient) ListCloudspaces(ctx context.Context, org string) (*CloudSpaceList, error) {
	exists, orgID, err := c.getOrgIDIfExists(ctx, org)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", org)
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces", c.BaseURL, orgID)

	// Pass &interm to be populated by doRequest JSON decoding
	var interm cloudSpaceListResponse
	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("access denied: you do not have permission to list cloudspaces in the organization '%s'", org)
		}
		return nil, err
	}
	var finalList CloudSpaceList
	for _, cs := range interm.Items {
		// Spot node pools
		spotNodePools, err := c.ListSpotNodePools(ctx, org, cs.Metadata.Name)
		if err != nil {
			return nil, err
		}

		// OnDemand node pools
		onDemandNodePools, err := c.ListOnDemandNodePools(ctx, org, cs.Metadata.Name)
		if err != nil {
			return nil, err
		}
		finalList.Items = append(finalList.Items, CloudSpace{
			Name:                 cs.Metadata.Name,
			Org:                  org,
			CreationTimestamp:    cs.Metadata.CreationTimestamp,
			Cni:                  cs.Spec.Cni,
			DeploymentType:       cs.Spec.DeploymentType,
			GpuEnabled:           cs.Spec.GpuEnabled,
			KubernetesVersion:    cs.Spec.KubernetesVersion,
			Region:               cs.Spec.Region,
			PreemptionWebhookURL: cs.Spec.Webhook,
			APIServerEndpoint:    cs.Status.APIServerEndpoint,
			AssignedServers:      cs.Status.AssignedServers,
			SpotNodepools:        spotNodePools,
			OnDemandNodePools:    onDemandNodePools,
			Health:               cs.Status.Health,
		})
	}
	return &finalList, nil
}

// CreateCloudspace creates a new cloudspace in the given namespace.
func (c *RackspaceSpotClient) CreateCloudspace(ctx context.Context, cs CloudSpace) error {
	exists, orgID, err := c.getOrgIDIfExists(ctx, cs.Org)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", cs.Org)
	}

	cloudspaceCreateRequestBody := CloudSpaceCreateRequestBody{
		APIVersion: "ngpc.rxt.io/v1",
		Kind:       "CloudSpace",
		Metadata: struct {
			Name        string `json:"name"`
			Namespace   string `json:"namespace"`
			Annotations struct {
			} `json:"annotations"`
		}{
			Name:      cs.Name,
			Namespace: orgID,
			Annotations: struct {
			}{},
		},
		Spec: struct {
			DeploymentType    string `json:"deploymentType"`
			Cloud             string `json:"cloud"`
			Region            string `json:"region"`
			Webhook           string `json:"webhook"`
			Cni               string `json:"cni"`
			KubernetesVersion string `json:"kubernetesVersion"`
			HAControlPlane    bool   `json:"HAControlPlane"`
			GpuEnabled        bool   `json:"gpuEnabled"`
		}{
			DeploymentType:    "gen2",
			Cloud:             "default",
			Region:            cs.Region,
			Webhook:           cs.PreemptionWebhookURL,
			Cni:               cs.Cni,
			KubernetesVersion: cs.KubernetesVersion,
			HAControlPlane:    false,
			GpuEnabled:        cs.GpuEnabled,
		},
	}

	body, err := json.Marshal(cloudspaceCreateRequestBody)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces", c.BaseURL, orgID)

	if err := c.doRequest(ctx, http.MethodPost, url, body, c.authHeader(), nil); err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("access denied: you do not have permission to create the cloudspace '%s'", cs.Name)
		}
		return err
	}
	return nil
}

// DeleteCloudspace deletes a cloudspace by name in the given namespace.
func (c *RackspaceSpotClient) DeleteCloudspace(ctx context.Context, org, name string) error {
	exists, orgID, err := c.getOrgIDIfExists(ctx, org)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces/%s", c.BaseURL, orgID, name)
	err = c.doRequest(ctx, http.MethodDelete, url, nil, c.authHeader(), nil)
	if err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("access denied: you do not have permission to delete the cloudspace '%s'", name)
		} else {
			if httpErr.StatusCode == http.StatusNotFound {
				return fmt.Errorf("cloudspace '%s' not found", name)
			}
			return err
		}
	}
	return nil
}

// GetCloudspace retrieves a cloudspace by name in the given namespace.
func (c *RackspaceSpotClient) GetCloudspace(ctx context.Context, org, name string) (*CloudSpace, error) {
	exists, orgID, err := c.getOrgIDIfExists(ctx, org)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", org)
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces/%s", c.BaseURL, orgID, name)
	fmt.Printf("url -----: %s\n", url)
	var interm cloudSpaceGetResponse

	// Pass &interm so doRequest will JSON unmarshal into it
	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("access denied: you do not have permission to get the cloudspace '%s'", name)
		}
		return nil, err
	}

	// Spot node pools
	spotNodePools, err := c.ListSpotNodePools(ctx, org, interm.Metadata.Name)
	if err != nil {
		return nil, err
	}

	// OnDemand node pools
	onDemandNodePools, err := c.ListOnDemandNodePools(ctx, org, interm.Metadata.Name)
	if err != nil {
		return nil, err
	}

	finalList := CloudSpace{
		Name:                 interm.Metadata.Name,
		Org:                  org,
		CreationTimestamp:    interm.Metadata.CreationTimestamp,
		Cni:                  interm.Spec.Cni,
		DeploymentType:       interm.Spec.DeploymentType,
		GpuEnabled:           interm.Spec.GpuEnabled,
		KubernetesVersion:    interm.Spec.KubernetesVersion,
		Region:               interm.Spec.Region,
		PreemptionWebhookURL: interm.Spec.Webhook,
		APIServerEndpoint:    interm.Status.APIServerEndpoint,
		AssignedServers:      interm.Status.AssignedServers,
		SpotNodepools:        spotNodePools,
		OnDemandNodePools:    onDemandNodePools,
		Health:               interm.Status.Health,
	}

	return &finalList, nil
}

// ListSpotNodePools retrieves all spot node pools in a namespace.
func (c *RackspaceSpotClient) ListSpotNodePools(ctx context.Context, org, cloudspaceName string) ([]*SpotNodePool, error) {

	exists, orgID, err := c.getOrgIDIfExists(ctx, org)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", org)
	}

	labelKey := "ngpc.rxt.io/cloudspace"
	labelSelector := fmt.Sprintf("%s=%s", labelKey, cloudspaceName)
	encodedSelector := url.QueryEscape(labelSelector)

	url := fmt.Sprintf(
		"%s/apis/ngpc.rxt.io/v1/namespaces/%s/spotnodepools?labelSelector=%s",
		c.BaseURL, orgID, encodedSelector,
	)

	var pool SpotNodePoolListResponse
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &pool); err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("access denied: you do not have permission to list spot node pools in the namespace '%s'", org)
		}
		return nil, err
	}

	var finalList []*SpotNodePool
	for _, item := range pool.Items {
		finalList = append(finalList, &SpotNodePool{
			Name:        item.Metadata.Name,
			Org:         org,
			Cloudspace:  item.Spec.CloudSpace,
			ServerClass: item.Spec.ServerClass,
			Desired:     item.Spec.Desired,
			BidPrice:    item.Spec.BidPrice + "$",
		})
	}
	return finalList, nil
}

// CreateSpotNodePool creates a new spot node pool in the given namespace.
func (c *RackspaceSpotClient) CreateSpotNodePool(ctx context.Context, org string, pool SpotNodePool) error {
	exists, orgID, err := c.getOrgIDIfExists(ctx, org)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/spotnodepools", c.BaseURL, orgID)

	spotNodePoolCreateRequestBody := SpotNodePoolCreateRequestBody{
		APIVersion: "ngpc.rxt.io/v1",
		Kind:       "SpotNodePool",
		Metadata: struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Labels    struct {
				NgpcRxtIoCloudspace string `json:"ngpc.rxt.io/cloudspace"`
			} `json:"labels"`
		}{
			Name:      uuid.New().String(),
			Namespace: orgID,
			Labels: struct {
				NgpcRxtIoCloudspace string `json:"ngpc.rxt.io/cloudspace"`
			}{
				NgpcRxtIoCloudspace: pool.Cloudspace,
			},
		},
		Spec: struct {
			ServerClass string `json:"serverClass"`
			Desired     int    `json:"desired"`
			BidPrice    string `json:"bidPrice"`
			CloudSpace  string `json:"cloudSpace"`
			Autoscaling struct {
				Enabled  bool  `json:"enabled"`
				MinNodes int64 `json:"minNodes"`
				MaxNodes int64 `json:"maxNodes"`
			} `json:"autoscaling"`
		}{
			ServerClass: pool.ServerClass,
			Desired:     pool.Desired,
			BidPrice:    pool.BidPrice,
			CloudSpace:  pool.Cloudspace,
			Autoscaling: struct {
				Enabled  bool  `json:"enabled"`
				MinNodes int64 `json:"minNodes"`
				MaxNodes int64 `json:"maxNodes"`
			}{
				Enabled:  pool.Autoscaling.Enabled,
				MinNodes: pool.Autoscaling.MinNodes,
				MaxNodes: pool.Autoscaling.MaxNodes,
			},
		},
	}

	body, err := json.Marshal(spotNodePoolCreateRequestBody)
	if err != nil {
		return err
	}

	err = c.doRequest(ctx, http.MethodPost, url, body, c.authHeader(), nil)
	if err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("access denied: you do not have permission to create the spot node pool '%s'", pool.Name)
		}
		return err
	}
	return nil
}

// DeleteSpotNodePool deletes a spot node pool by name in the given namespace.
func (c *RackspaceSpotClient) DeleteSpotNodePool(ctx context.Context, org, name string) error {
	// 	url := fmt.Sprintf("%s/api/namespaces/%s/spotnodepools/%s", c.BaseURL, namespace, name)
	// 	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodDelete, url, nil, c.authHeader())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer resp.Body.Close()
	// 	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
	// 		b, _ := ioutil.ReadAll(resp.Body)
	// 		return fmt.Errorf("failed to delete spot node pool: %s", string(b))
	// 	}
	return nil
}

// GetSpotNodePool retrieves a spot node pool by name in the given namespace.
func (c *RackspaceSpotClient) GetSpotNodePool(ctx context.Context, namespace, cloudspaceName string) (*SpotNodePool, error) {
	// labelKey := "ngpc.rxt.io/cloudspace"
	// labelSelector := fmt.Sprintf("%s=%s", labelKey, cloudspaceName)
	// encodedSelector := url.QueryEscape(labelSelector)

	// url := fmt.Sprintf("%s/namespaces/%s/spotnodepools?labelSelector=%s", c.BaseURL, namespace, encodedSelector)
	// fmt.Println(url)

	// resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	// if err != nil {
	// 	return nil, err
	// }
	// defer resp.Body.Close()
	// if resp.StatusCode != http.StatusOK {
	// 	b, _ := ioutil.ReadAll(resp.Body)
	// 	return nil, fmt.Errorf("failed to get spot node pool: %s", string(b))
	// }
	// var pool SpotNodePool
	// if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
	// 	return nil, err
	// }
	return nil, nil
}

// ListOnDemandNodePools retrieves all on-demand node pools in a namespace.
func (c *RackspaceSpotClient) ListOnDemandNodePools(ctx context.Context, org, cloudspaceName string) ([]*OnDemandNodePool, error) {

	exists, orgID, err := c.getOrgIDIfExists(ctx, org)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", org)
	}

	labelKey := "ngpc.rxt.io/cloudspace"
	labelSelector := fmt.Sprintf("%s=%s", labelKey, cloudspaceName)
	encodedSelector := url.QueryEscape(labelSelector)

	url := fmt.Sprintf(
		"%s/apis/ngpc.rxt.io/v1/namespaces/%s/ondemandnodepools?labelSelector=%s",
		c.BaseURL, orgID, encodedSelector,
	)

	var pool SpotNodePoolListResponse
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &pool); err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("access denied: you do not have permission to list on-demand node pools in the namespace '%s'", org)
		}
		return nil, err
	}

	var finalList []*OnDemandNodePool
	for _, item := range pool.Items {
		finalList = append(finalList, &OnDemandNodePool{
			Name:        item.Metadata.Name,
			Org:         org,
			Cloudspace:  item.Spec.CloudSpace,
			ServerClass: item.Spec.ServerClass,
			Desired:     item.Spec.Desired,
		})
	}
	return finalList, nil
}

// CreateOnDemandNodePool creates a new on-demand node pool in the given namespace.
func (c *RackspaceSpotClient) CreateOnDemandNodePool(ctx context.Context, org string, pool OnDemandNodePool) error {

	exists, orgID, err := c.getOrgIDIfExists(ctx, org)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/ondemandnodepools", c.BaseURL, orgID)

	ondemandNodePoolCreateRequestBody := OnDemandNodePoolCreateRequestBody{
		APIVersion: "ngpc.rxt.io/v1",
		Kind:       "OnDemandNodePool",
		Metadata: struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Labels    struct {
				NgpcRxtIoCloudspace string `json:"ngpc.rxt.io/cloudspace"`
			} `json:"labels"`
		}{
			Name:      uuid.New().String(),
			Namespace: orgID,
			Labels: struct {
				NgpcRxtIoCloudspace string `json:"ngpc.rxt.io/cloudspace"`
			}{
				NgpcRxtIoCloudspace: pool.Cloudspace,
			},
		},
		Spec: struct {
			ServerClass string `json:"serverClass"`
			Desired     int    `json:"desired"`
			CloudSpace  string `json:"cloudSpace"`
			Autoscaling struct {
				Enabled  bool `json:"enabled"`
				MinNodes any  `json:"minNodes"`
				MaxNodes any  `json:"maxNodes"`
			} `json:"autoscaling"`
		}{
			ServerClass: pool.ServerClass,
			Desired:     pool.Desired,
			CloudSpace:  pool.Cloudspace,
			Autoscaling: struct {
				Enabled  bool `json:"enabled"`
				MinNodes any  `json:"minNodes"`
				MaxNodes any  `json:"maxNodes"`
			}{
				Enabled:  pool.Autoscaling.Enabled,
				MinNodes: pool.Autoscaling.MinNodes,
				MaxNodes: pool.Autoscaling.MaxNodes,
			},
		},
	}

	body, err := json.Marshal(ondemandNodePoolCreateRequestBody)
	if err != nil {
		return err
	}

	err = c.doRequest(ctx, http.MethodPost, url, body, c.authHeader(), nil)
	if err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("access denied: you do not have permission to create the ondemand node pool '%s'", pool.Name)
		}
		return err
	}
	return nil
}

// DeleteOnDemandNodePool deletes an on-demand node pool by name in the given namespace.
func (c *RackspaceSpotClient) DeleteOnDemandNodePool(ctx context.Context, namespace, name string) error {
	// 	url := fmt.Sprintf("%s/api/namespaces/%s/ondemandnodepools/%s", c.BaseURL, namespace, name)
	// 	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodDelete, url, nil, c.authHeader())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer resp.Body.Close()
	// 	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
	// 		b, _ := ioutil.ReadAll(resp.Body)
	// 		return fmt.Errorf("failed to delete on-demand node pool: %s", string(b))
	// 	}
	return nil
}

// GetOnDemandNodePool retrieves an on-demand node pool by name in the given namespace.
func (c *RackspaceSpotClient) GetOnDemandNodePool(ctx context.Context, namespace, name string) (*OnDemandNodePool, error) {
	// 	url := fmt.Sprintf("%s/api/namespaces/%s/ondemandnodepools/%s", c.BaseURL, namespace, name)
	// 	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	defer resp.Body.Close()
	// 	if resp.StatusCode != http.StatusOK {
	// 		b, _ := ioutil.ReadAll(resp.Body)
	// 		return nil, fmt.Errorf("failed to get on-demand node pool: %s", string(b))
	// 	}
	// 	var pool OnDemandNodePool
	// 	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
	// 		return nil, err
	// 	}
	return nil, nil
}

func (c *RackspaceSpotClient) GetCloudspaceConfig(ctx context.Context, namespace, name string) (string, error) {
	url := fmt.Sprintf("%s/apis/auth.ngpc.rxt.io/v1/generate-kubeconfig", c.BaseURL)
	reqBody := struct {
		OrganizationName string `json:"organization_name"`
		CloudspaceName   string `json:"cloudspace_name"`
		RefreshToken     string `json:"refresh_token"`
	}{
		OrganizationName: namespace,
		CloudspaceName:   name,
		RefreshToken:     c.RefreshToken, // actual token
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	var kubeConfigResponse KubeConfigResponse
	if err := c.doRequest(ctx, http.MethodPost, url, jsonBody, c.authHeader(), &kubeConfigResponse); err != nil {
		if httpErr, ok := err.(*HTTPStatusError); ok && httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return "", fmt.Errorf("access denied: you do not have permission to get the kubeconfig for - '%s'", name)
		}
		return "", err
	}
	return kubeConfigResponse.Data.Kubeconfig, nil
}

// ListRegions retrieves all available regions.
func (c *RackspaceSpotClient) ListRegions(ctx context.Context) ([]Region, error) {
	url := fmt.Sprintf("%s/api/regions", c.BaseURL)
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), nil); err != nil {
		return nil, err
	}
	var regions []Region
	// if err := json.NewDecoder(resp.Body).Decode(&regions); err != nil {
	// 	return nil, err
	// }
	return regions, nil
}

// GetRegion retrieves a region by name.
func (c *RackspaceSpotClient) GetRegion(ctx context.Context, name string) (*Region, error) {
	url := fmt.Sprintf("%s/api/regions/%s", c.BaseURL, name)
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), nil); err != nil {
		return nil, err
	}
	// defer resp.Body.Close()
	// if resp.StatusCode != http.StatusOK {
	// 	b, _ := ioutil.ReadAll(resp.Body)
	// 	return nil, fmt.Errorf("failed to get region: %s", string(b))
	// }
	var region Region
	// if err := json.NewDecoder(resp.Body).Decode(&region); err != nil {
	// 	return nil, err
	// }
	return &region, nil
}

// ListServerClasses retrieves all available server classes.
func (c *RackspaceSpotClient) ListServerClasses(ctx context.Context) (*ServerClassList, error) {
	url := fmt.Sprintf("%s/api/serverclasses", c.BaseURL)
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), nil); err != nil {
		return nil, err
	}
	// defer resp.Body.Close()
	// if resp.StatusCode != http.StatusOK {
	// 	b, _ := ioutil.ReadAll(resp.Body)
	// 	return nil, fmt.Errorf("failed to list server classes: %s", string(b))
	// }
	var classes []ServerClass
	// if err := json.NewDecoder(resp.Body).Decode(&classes); err != nil {
	// 	return nil, err
	return &ServerClassList{Items: classes}, nil
}

// GetServerClass retrieves a server class by name.
func (c *RackspaceSpotClient) GetServerClass(ctx context.Context, name string) (*ServerClass, error) {
	url := fmt.Sprintf("%s/api/serverclasses/%s", c.BaseURL, name)
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), nil); err != nil {
		return nil, err
	}
	// defer resp.Body.Close()
	// if resp.StatusCode != http.StatusOK {
	// 	b, _ := ioutil.ReadAll(resp.Body)
	// 	return nil, fmt.Errorf("failed to get server class: %s", string(b))
	// }
	var class ServerClass
	// if err := json.NewDecoder(resp.Body).Decode(&class); err != nil {
	// 	return nil, err
	//}
	return &class, nil
}

// GetPriceHistory retrieves the price history for a server class.
func (c *RackspaceSpotClient) GetPriceHistory(ctx context.Context, serverClass string) (*PriceHistory, error) {
	url := fmt.Sprintf("%s/api/serverclasses/%s/pricehistory", c.BaseURL, serverClass)
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), nil); err != nil {
		return nil, err
	}
	// defer resp.Body.Close()
	// if resp.StatusCode != http.StatusOK {
	// 	b, _ := ioutil.ReadAll(resp.Body)
	// 	return nil, fmt.Errorf("failed to get price history: %s", string(b))
	// }
	var history PriceHistory
	// if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
	// 	return nil, err
	//}
	return &history, nil
}

type HTTPStatusError struct {
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
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

	fmt.Printf("url - %v\n", url)
	fmt.Printf("request body - %v\n", string(body))
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
