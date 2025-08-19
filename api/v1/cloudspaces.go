package rxtspot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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
	// fmt.Printf("url -----: %s\n", url)
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
