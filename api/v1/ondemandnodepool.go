package rxtspot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

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
