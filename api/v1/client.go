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

	"github.com/google/uuid"
	"github.com/rackspace-spot/spot-go-sdk/pkg/httpclient"
)

// Configuration struct for RackspaceSpotClient
type Config struct {
	BaseURL      string
	OAuthURL     string
	HTTPClient   *http.Client
	RefreshToken string
	Timeout      time.Duration
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
func (c *RackspaceSpotClient) ListCloudspaces(ctx context.Context, namespace string) (*CloudSpaceList, error) {
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces", c.BaseURL, url.PathEscape(namespace))
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("url to list cloudspaces: %s\n", url)
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("access denied: you do not have permission to list cloudspaces in the organization '%s'", namespace)
	}

	var interm cloudSpaceListResponse
	if err := json.NewDecoder(resp.Body).Decode(&interm); err != nil {
		return nil, err
	}

	var finalList CloudSpaceList
	for _, cs := range interm.Items {
		finalList.Items = append(finalList.Items, CloudSpace{
			Name:              cs.Metadata.Name,
			OrgID:             cs.Metadata.Namespace,
			CreationTimestamp: cs.Metadata.CreationTimestamp,
			//	BidRequests:       cs.Spec.BidRequests,
			//Cloud:             cs.Spec.Cloud,
			Cni:               cs.Spec.Cni,
			DeploymentType:    cs.Spec.DeploymentType,
			GpuEnabled:        cs.Spec.GpuEnabled,
			KubernetesVersion: cs.Spec.KubernetesVersion,
			Region:            cs.Spec.Region,
			//Type:              cs.Spec.Type,
			PreemptionWebhookURL: cs.Spec.Webhook,
			APIServerEndpoint:    cs.Status.APIServerEndpoint,
			AssignedServers:      cs.Status.AssignedServers,
			//	Bids:              cs.Status.Bids,
			Health: cs.Status.Health,
		})
	}
	return &finalList, nil
}

// CreateCloudspace creates a new cloudspace in the given namespace.
func (c *RackspaceSpotClient) CreateCloudspace(ctx context.Context, cs CloudSpace) error {
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces", c.BaseURL, cs.OrgID)
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
			Namespace: cs.OrgID,
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
		fmt.Printf("%v\n", err)
		return err
	}
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodPost, url, body, c.authHeader())
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}
	return nil
}

// DeleteCloudspace deletes a cloudspace by name in the given namespace.
func (c *RackspaceSpotClient) DeleteCloudspace(ctx context.Context, namespace, name string) error {
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces/%s", c.BaseURL, namespace, name)
	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodDelete, url, nil, c.authHeader())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}
	return nil
}

// GetCloudspace retrieves a cloudspace by name in the given namespace.
func (c *RackspaceSpotClient) GetCloudspace(ctx context.Context, namespace, name string) (*CloudSpace, error) {
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces/%s", c.BaseURL, url.PathEscape(namespace), url.PathEscape(name))

	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("cloudspace %s not found", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get cloudspace: %s", resp.Status)
	}

	var interm cloudSpaceGetResponse
	if err := json.NewDecoder(resp.Body).Decode(&interm); err != nil {
		fmt.Printf("failed to decode response: %v\n", err)
		return nil, err
	}

	spotNodePools, err := c.ListSpotNodePools(ctx, namespace, interm.Metadata.Name)
	if err != nil {
		return nil, err
	}
	finalList := CloudSpace{
		Name:              interm.Metadata.Name,
		OrgID:             interm.Metadata.Namespace,
		CreationTimestamp: interm.Metadata.CreationTimestamp,
		//Cloud:             interm.Spec.Cloud,
		Cni:               interm.Spec.Cni,
		DeploymentType:    interm.Spec.DeploymentType,
		GpuEnabled:        interm.Spec.GpuEnabled,
		KubernetesVersion: interm.Spec.KubernetesVersion,
		Region:            interm.Spec.Region,
		//Type:              interm.Spec.Type,
		PreemptionWebhookURL: interm.Spec.Webhook,
		APIServerEndpoint:    interm.Status.APIServerEndpoint,
		AssignedServers:      interm.Status.AssignedServers,
		SpotNodepools:        spotNodePools,
		Health:               interm.Status.Health,
	}
	return &finalList, nil
}

// ListSpotNodePools retrieves all spot node pools in a namespace.
func (c *RackspaceSpotClient) ListSpotNodePools(ctx context.Context, namespace string, cloudspaceName string) ([]*SpotNodePool, error) {
	labelKey := "ngpc.rxt.io/cloudspace"
	labelSelector := fmt.Sprintf("%s=%s", labelKey, cloudspaceName)
	encodedSelector := url.QueryEscape(labelSelector)

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/spotnodepools?labelSelector=%s", c.BaseURL, namespace, encodedSelector)

	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodGet, url, nil, c.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", resp.Status)
	}
	var pool SpotNodePoolListResponse
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, err
	}

	var finalList []*SpotNodePool
	for _, item := range pool.Items {
		finalList = append(finalList, &SpotNodePool{
			Name:        item.Metadata.Name,
			Org:         item.Metadata.Namespace,
			Cloudspace:  item.Spec.CloudSpace,
			ServerClass: item.Spec.ServerClass,
			Desired:     item.Spec.Desired,
			BidPrice:    item.Spec.BidPrice + "$",
		})
	}
	return finalList, nil
}

// CreateSpotNodePool creates a new spot node pool in the given namespace.
func (c *RackspaceSpotClient) CreateSpotNodePool(ctx context.Context, pool SpotNodePool) error {
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/spotnodepools", c.BaseURL, pool.Org)

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
			Namespace: pool.Org,
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
		fmt.Errorf("%v\n", err)
		return err
	}
	fmt.Printf("body - %+v \n", string(body))

	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodPost, url, body, c.authHeader())

	b, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("spot node pool created response body : %s\n", string(b))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}
	fmt.Printf("spot node pool created: %s\n", string(b))
	return nil
}

// DeleteSpotNodePool deletes a spot node pool by name in the given namespace.
func (c *RackspaceSpotClient) DeleteSpotNodePool(ctx context.Context, namespace, name string) error {
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
func (c *RackspaceSpotClient) ListOnDemandNodePools(ctx context.Context, namespace string) ([]OnDemandNodePool, error) {
	// TODO: Implement API call to /api/namespaces/{namespace}/ondemandnodepools
	return nil, nil
}

// CreateOnDemandNodePool creates a new on-demand node pool in the given namespace.
func (c *RackspaceSpotClient) CreateOnDemandNodePool(ctx context.Context, pool OnDemandNodePool) error {
	// 	url := fmt.Sprintf("%s/api/namespaces/%s/ondemandnodepools", c.BaseURL, pool.Namespace)
	// 	body, err := json.Marshal(pool)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	resp, err := httpclient.DoRequest(ctx, c.HTTPClient, http.MethodPost, url, body, c.authHeader())
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	defer resp.Body.Close()
	// 	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
	// 		b, _ := ioutil.ReadAll(resp.Body)
	// 		return nil, fmt.Errorf("failed to create on-demand node pool: %s", string(b))
	// 	}
	// 	var created OnDemandNodePool
	// 	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
	// 		return nil, err
	// 	}
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
func (c *RackspaceSpotClient) ListServerClasses(ctx context.Context) (*ServerClassList, error) {
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
	var classes []ServerClass
	if err := json.NewDecoder(resp.Body).Decode(&classes); err != nil {
		return nil, err
	}
	return &ServerClassList{Items: classes}, nil
}

// GetServerClass retrieves a server class by name.
func (c *RackspaceSpotClient) GetServerClass(ctx context.Context, name string) (*ServerClass, error) {
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
	var class ServerClass
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
