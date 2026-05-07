package rxtspot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListVMCloudSpaces retrieves all VM cloud spaces in a namespace.
func (c *RackspaceSpotClient) ListVMCloudSpaces(ctx context.Context, org string) (*VMCloudSpaceList, error) {
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
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmcloudspaces", c.BaseURL, orgID)

	var interm vmCloudSpaceListResponse
	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		return nil, c.handleAPIError(err, "vm cloudspaces", "", "list")
	}

	var finalList VMCloudSpaceList
	for _, vmcs := range interm.Items {
		// Fetch VM pools for this VM cloud space
		vmPools, err := c.ListVMPools(ctx, org, vmcs.Metadata.Name)
		if err != nil {
			return nil, c.handleAPIError(err, "vm pools", "", "list for vm cloudspace "+vmcs.Metadata.Name)
		}

		finalList.Items = append(finalList.Items, VMCloudSpace{
			Name:              vmcs.Metadata.Name,
			Org:               org,
			CreationTimestamp: vmcs.Metadata.CreationTimestamp,
			Region:            vmcs.Spec.Region,
			Webhook:           vmcs.Spec.Webhook,
			VMSshKeyRef: VMSshKeyRef{
				Name:      vmcs.Spec.VMSshKeyRef.Name,
				Namespace: vmcs.Spec.VMSshKeyRef.Namespace,
			},
			VMPools:         vmPools,
			AssignedServers: vmcs.Status.AssignedServers,
			Status:          vmcs.Status.Phase,
			Message:         vmcs.Status.Reason,
			Health:          vmcs.Status.Health,
		})
	}
	return &finalList, nil
}

// CreateVMCloudSpace creates a new VM cloud space in the given namespace.
func (c *RackspaceSpotClient) CreateVMCloudSpace(ctx context.Context, vmcs VMCloudSpace) error {
	if err := ValidateOrgName(vmcs.Org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(vmcs.Name); err != nil {
		return fmt.Errorf("invalid vm cloudspace name: %w", err)
	}
	if vmcs.Region == "" {
		return fmt.Errorf("region is required")
	}
	if vmcs.VMSshKeyRef.Name == "" {
		return fmt.Errorf("vm ssh key name is required")
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, vmcs.Org)
	if err != nil {
		return c.handleAPIError(err, "organization", vmcs.Org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", vmcs.Org)
	}

	reqBody := VMCloudSpaceCreateRequestBody{
		APIVersion: "ngpc.rxt.io/v1",
		Kind:       "VMCloudSpace",
	}
	reqBody.Metadata.Name = vmcs.Name
	reqBody.Metadata.Namespace = orgID
	reqBody.Spec.Region = vmcs.Region
	reqBody.Spec.Webhook = vmcs.Webhook
	reqBody.Spec.VMSshKeyRef.Name = vmcs.VMSshKeyRef.Name
	if vmcs.VMSshKeyRef.Namespace != "" {
		reqBody.Spec.VMSshKeyRef.Namespace = vmcs.VMSshKeyRef.Namespace
	} else {
		reqBody.Spec.VMSshKeyRef.Namespace = orgID
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return c.handleAPIError(err, "vm cloudspace", vmcs.Name, "create")
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmcloudspaces", c.BaseURL, orgID)

	if err := c.doRequest(ctx, http.MethodPost, url, body, c.authHeader(), nil); err != nil {
		return c.handleAPIError(err, "vm cloudspace", vmcs.Name, "create")
	}
	return nil
}

// GetVMCloudSpace retrieves a VM cloud space by name in the given namespace.
func (c *RackspaceSpotClient) GetVMCloudSpace(ctx context.Context, org, name string) (*VMCloudSpace, error) {
	if err := ValidateOrgName(org); err != nil {
		return nil, fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(name); err != nil {
		return nil, fmt.Errorf("invalid vm cloudspace name: %w", err)
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", org)
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmcloudspaces/%s", c.BaseURL, orgID, name)
	var interm vmCloudSpaceGetResponse

	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		return nil, c.handleAPIError(err, "vm cloudspace", name, "get")
	}

	// Fetch VM pools for this VM cloud space
	vmPools, err := c.ListVMPools(ctx, org, interm.Metadata.Name)
	if err != nil {
		return nil, c.handleAPIError(err, "vm pools", name, "list for vm cloudspace "+interm.Metadata.Name)
	}

	result := &VMCloudSpace{
		Name:              interm.Metadata.Name,
		Org:               org,
		CreationTimestamp: interm.Metadata.CreationTimestamp,
		Region:            interm.Spec.Region,
		Webhook:           interm.Spec.Webhook,
		VMSshKeyRef: VMSshKeyRef{
			Name:      interm.Spec.VMSshKeyRef.Name,
			Namespace: interm.Spec.VMSshKeyRef.Namespace,
		},
		VMPools:         vmPools,
		AssignedServers: interm.Status.AssignedServers,
		Status:          interm.Status.Phase,
		Message:         interm.Status.Reason,
		Health:          interm.Status.Health,
	}

	return result, nil
}

// UpdateVMCloudSpace updates an existing VM cloud space.
// Only the Webhook field is allowed to be modified.
func (c *RackspaceSpotClient) UpdateVMCloudSpace(ctx context.Context, org string, vmcs VMCloudSpace) error {
	if err := ValidateOrgName(org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	if vmcs.Name == "" {
		return fmt.Errorf("vm cloudspace name is required")
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return c.handleAPIError(err, "organization", org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}

	var updateBody VMCloudSpaceUpdateRequestBody
	updateBody.Spec.Webhook = vmcs.Webhook

	body, err := json.Marshal(updateBody)
	if err != nil {
		return c.handleAPIError(err, "vm cloudspace", vmcs.Name, "update")
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmcloudspaces/%s", c.BaseURL, orgID, vmcs.Name)

	if err := c.doRequest(ctx, http.MethodPatch, url, body, c.authHeader(), nil); err != nil {
		return c.handleAPIError(err, "vm cloudspace", vmcs.Name, "update")
	}
	return nil
}

// DeleteVMCloudSpace deletes a VM cloud space by name in the given namespace.
func (c *RackspaceSpotClient) DeleteVMCloudSpace(ctx context.Context, org, name string) error {
	if err := ValidateOrgName(org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(name); err != nil {
		return fmt.Errorf("invalid vm cloudspace name: %w", err)
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return c.handleAPIError(err, "organization", org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmcloudspaces/%s", c.BaseURL, orgID, name)
	err = c.doRequest(ctx, http.MethodDelete, url, nil, c.authHeader(), nil)
	return c.handleAPIError(err, "vm cloudspace", name, "delete")
}
