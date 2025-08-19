package rxtspot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

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
