package rxtspot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ListCloudspaces retrieves all cloudspaces in a namespace.
func (c *RackspaceSpotClient) ListCloudspaces(ctx context.Context, org string) (*CloudSpaceList, error) {
	if err := ValidateOrgName(org); err != nil {
		return nil, fmt.Errorf("invalid organization name: %w", err)
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return nil, c.handleAPIError(err, "organization", org, "find")
	}
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", org)
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces", c.BaseURL, orgID)

	// Pass &interm to be populated by doRequest JSON decoding
	var interm cloudSpaceListResponse
	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		return nil, c.handleAPIError(err, "cloudspaces", "", "list")
	}
	var finalList CloudSpaceList
	for _, cs := range interm.Items {
		// Spot node pools
		spotNodePools, err := c.ListSpotNodePools(ctx, org, cs.Metadata.Name)
		if err != nil {
			return nil, c.handleAPIError(err, "spot node pools", "", "list for cloudspace "+cs.Metadata.Name)
		}

		// OnDemand node pools
		onDemandNodePools, err := c.ListOnDemandNodePools(ctx, org, cs.Metadata.Name)
		if err != nil {
			return nil, c.handleAPIError(err, "on-demand node pools", "", "list for cloudspace "+cs.Metadata.Name)
		}
		finalList.Items = append(finalList.Items, cloudSpaceFromResponse(org, &cs, spotNodePools, onDemandNodePools))
	}
	return &finalList, nil
}

// CreateCloudspace creates a new cloudspace in the given namespace.
func (c *RackspaceSpotClient) CreateCloudspace(ctx context.Context, cs CloudSpace) error {
	if err := ValidateOrgName(cs.Org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(cs.Name); err != nil {
		return fmt.Errorf("invalid cloudspace name: %w", err)
	}
	if cs.Region == "" {
		return fmt.Errorf("region is required")
	}
	if cs.KubernetesVersion == "" {
		return fmt.Errorf("kubernetes version is required")
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, cs.Org)
	if err != nil {
		return c.handleAPIError(err, "organization", cs.Org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", cs.Org)
	}

	gpuEnabled := BoolPtr(false)
	if cs.GpuEnabled != nil {
		gpuEnabled = cs.GpuEnabled
	}
	haControlPlane := false
	if cs.HAControlPlane != nil {
		haControlPlane = *cs.HAControlPlane
	}

	cloudspaceCreateRequestBody := CloudSpaceCreateRequestBody{
		APIVersion: "ngpc.rxt.io/v1",
		Kind:       "CloudSpace",
		Metadata: struct {
			Name        string            `json:"name"`
			Namespace   string            `json:"namespace"`
			Annotations map[string]string `json:"annotations"`
		}{
			Name:        cs.Name,
			Namespace:   orgID,
			Annotations: map[string]string{},
		},
		Spec: struct {
			DeploymentType    string `json:"deploymentType"`
			Cloud             string `json:"cloud"`
			Region            string `json:"region"`
			Webhook           string `json:"webhook"`
			CNI               string `json:"cni"`
			KubernetesVersion string `json:"kubernetesVersion"`
			HAControlPlane    bool   `json:"HAControlPlane"`
			GpuEnabled        bool   `json:"gpuEnabled"`
		}{
			DeploymentType:    "gen2",
			Cloud:             "default",
			Region:            cs.Region,
			Webhook:           cs.PreemptionWebhookURL,
			CNI:               cs.CNI,
			KubernetesVersion: cs.KubernetesVersion,
			HAControlPlane:    haControlPlane,
			GpuEnabled:        *gpuEnabled,
		},
	}

	body, err := json.Marshal(cloudspaceCreateRequestBody)
	if err != nil {
		return c.handleAPIError(err, "cloudspace", cs.Name, "create")
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces", c.BaseURL, orgID)

	if err := c.doRequest(ctx, http.MethodPost, url, body, c.authHeader(), nil); err != nil {
		return c.handleAPIError(err, "cloudspace", cs.Name, "create")
	}
	return nil
}

// DeleteCloudspace deletes a cloudspace by name in the given namespace.
func (c *RackspaceSpotClient) DeleteCloudspace(ctx context.Context, org, name string) error {
	if err := ValidateOrgName(org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(name); err != nil {
		return fmt.Errorf("invalid cloudspace name: %w", err)
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return c.handleAPIError(err, "organization", org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces/%s", c.BaseURL, orgID, name)
	err = c.doRequest(ctx, http.MethodDelete, url, nil, c.authHeader(), nil)
	return c.handleAPIError(err, "cloudspace", name, "delete")

}

// GetCloudspace retrieves a cloudspace by name in the given namespace.
func (c *RackspaceSpotClient) GetCloudspace(ctx context.Context, org, name string) (*CloudSpace, error) {
	if err := ValidateOrgName(org); err != nil {
		return nil, fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(name); err != nil {
		return nil, fmt.Errorf("invalid cloudspace name: %w", err)
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", org)
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces/%s", c.BaseURL, orgID, name)
	var interm cloudSpaceGetResponse

	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		return nil, c.handleAPIError(err, "cloudspace", name, "get")
	}

	spotNodePools, err := c.ListSpotNodePools(ctx, org, interm.Metadata.Name)
	if err != nil {
		return nil, c.handleAPIError(err, "spot node pool", name, "list for cloudspace "+interm.Metadata.Name)
	}

	onDemandNodePools, err := c.ListOnDemandNodePools(ctx, org, interm.Metadata.Name)
	if err != nil {
		return nil, c.handleAPIError(err, "on-demand node pool", name, "list for cloudspace "+interm.Metadata.Name)
	}

	result := cloudSpaceFromResponse(org, &interm, spotNodePools, onDemandNodePools)
	return &result, nil
}

// UpdateCloudspace updates mutable fields on an existing cloudspace.
// Only the following fields are mutable via PATCH:
// kubernetesVersion, webhook (preemptionWebhookURL), cni,
// HAControlPlane, and gpuEnabled. At least one mutable field must be set.
func (c *RackspaceSpotClient) UpdateCloudspace(ctx context.Context, org string, cs CloudSpace) (*CloudSpace, error) {
	if err := ValidateOrgName(org); err != nil {
		return nil, fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(cs.Name); err != nil {
		return nil, fmt.Errorf("invalid cloudspace name: %w", err)
	}

	spec := cloudspaceUpdateSpec{
		KubernetesVersion: stringPtr(cs.KubernetesVersion),
		Webhook:           stringPtr(cs.PreemptionWebhookURL),
		CNI:               stringPtr(cs.CNI),
		HAControlPlane:    cs.HAControlPlane,
		GpuEnabled:        cs.GpuEnabled,
	}
	if spec == (cloudspaceUpdateSpec{}) {
		return nil, errors.New("update requires at least one mutable field: kubernetesVersion, preemptionWebhookURL, cni, HAControlPlane, or gpuEnabled")
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("invalid organization: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", org)
	}

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces/%s", c.BaseURL, orgID, cs.Name)
	updateBody := cloudspaceUpdateRequestBody{Spec: spec}

	body, err := json.Marshal(updateBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update body: %w", err)
	}

	var resp cloudSpaceGetResponse
	err = c.doRequest(ctx, http.MethodPatch, url, body, c.authHeader(), &resp)
	if err != nil {
		return nil, c.handleAPIError(err, "cloudspace", cs.Name, "update")
	}

	spotNodePools, err := c.ListSpotNodePools(ctx, org, resp.Metadata.Name)
	if err != nil {
		return nil, c.handleAPIError(err, "spot node pool", cs.Name, "list for cloudspace "+resp.Metadata.Name)
	}
	onDemandNodePools, err := c.ListOnDemandNodePools(ctx, org, resp.Metadata.Name)
	if err != nil {
		return nil, c.handleAPIError(err, "on-demand node pool", cs.Name, "list for cloudspace "+resp.Metadata.Name)
	}

	updated := cloudSpaceFromResponse(org, &resp, spotNodePools, onDemandNodePools)
	return &updated, nil
}

func cloudSpaceFromResponse(org string, resp *cloudSpaceGetResponse, spotNodePools []*SpotNodePool, onDemandNodePools []*OnDemandNodePool) CloudSpace {
	return CloudSpace{
		Name:                 resp.Metadata.Name,
		Org:                  org,
		CreationTimestamp:    resp.Metadata.CreationTimestamp,
		CNI:                  resp.Spec.CNI,
		DeploymentType:       resp.Spec.DeploymentType,
		GpuEnabled:           BoolPtr(resp.Spec.GpuEnabled),
		HAControlPlane:       BoolPtr(resp.Spec.HAControlPlane),
		KubernetesVersion:    resp.Spec.KubernetesVersion,
		Region:               resp.Spec.Region,
		PreemptionWebhookURL: resp.Spec.Webhook,
		APIServerEndpoint:    resp.Status.APIServerEndpoint,
		AssignedServers:      resp.Status.AssignedServers,
		SpotNodepools:        spotNodePools,
		OnDemandNodePools:    onDemandNodePools,
		Status:               resp.Status.Phase,
		Message:              resp.Status.Reason,
	}
}

func BoolPtr(b bool) *bool { return &b }

func IntPtr(i int) *int { return &i }

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (c *RackspaceSpotClient) GetCloudspaceConfig(ctx context.Context, namespace, name string) (string, error) {
	if err := ValidateOrgName(namespace); err != nil {
		return "", fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(name); err != nil {
		return "", fmt.Errorf("invalid cloudspace name: %w", err)
	}
	if c.RefreshToken == "" {
		return "", fmt.Errorf("refresh token is required")
	}
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
		return "", c.handleAPIError(err, "cloudspace", name, "get kubeconfig")
	}
	var kubeConfigResponse KubeConfigResponse
	if err := c.doRequest(ctx, http.MethodPost, url, jsonBody, c.authHeader(), &kubeConfigResponse); err != nil {
		return "", c.handleAPIError(err, "cloudspace", name, "get kubeconfig")
	}
	return kubeConfigResponse.Data.Kubeconfig, nil
}
