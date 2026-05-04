package rxtspot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListVMSSHKeys retrieves all VM SSH keys in a namespace.
func (c *RackspaceSpotClient) ListVMSSHKeys(ctx context.Context, org string) (*VMSSHKeyList, error) {
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

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmsshkeys", c.BaseURL, orgID)

	var interm vmSSHKeyListResponse
	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		return nil, c.handleAPIError(err, "vm ssh keys", "", "list")
	}

	var finalList VMSSHKeyList
	for _, item := range interm.Items {
		finalList.Items = append(finalList.Items, VMSSHKey{
			Name:              item.Metadata.Name,
			CreationTimestamp: item.Metadata.CreationTimestamp,
			Org:               org,
			PublicKey:         item.Spec.PublicKey,
			Description:       item.Spec.Description,
			Fingerprint:       item.Status.Fingerprint,
			Validated:         item.Status.Validated,
		})
	}
	return &finalList, nil
}

// CreateVMSSHKey creates a new VM SSH key in the given namespace.
func (c *RackspaceSpotClient) CreateVMSSHKey(ctx context.Context, key VMSSHKey) error {
	if err := ValidateOrgName(key.Org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(key.Name); err != nil {
		return fmt.Errorf("invalid vm ssh key name: %w", err)
	}
	if key.PublicKey == "" {
		return fmt.Errorf("public key is required")
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, key.Org)
	if err != nil {
		return c.handleAPIError(err, "organization", key.Org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", key.Org)
	}

	reqBody := VMSSHKeyCreateRequestBody{
		APIVersion: "ngpc.rxt.io/v1",
		Kind:       "VMSSHKey",
	}
	reqBody.Metadata.Name = key.Name
	reqBody.Metadata.Namespace = orgID
	reqBody.Spec.PublicKey = key.PublicKey
	reqBody.Spec.Description = key.Description

	body, err := json.Marshal(reqBody)
	if err != nil {
		return c.handleAPIError(err, "vm ssh key", key.Name, "create")
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmsshkeys", c.BaseURL, orgID)

	if err := c.doRequest(ctx, http.MethodPost, url, body, c.authHeader(), nil); err != nil {
		return c.handleAPIError(err, "vm ssh key", key.Name, "create")
	}
	return nil
}

// GetVMSSHKey retrieves a VM SSH key by name in the given namespace.
func (c *RackspaceSpotClient) GetVMSSHKey(ctx context.Context, org, name string) (*VMSSHKey, error) {
	if err := ValidateOrgName(org); err != nil {
		return nil, fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(name); err != nil {
		return nil, fmt.Errorf("invalid vm ssh key name: %w", err)
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("organization '%s' not found", org)
	}

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmsshkeys/%s", c.BaseURL, orgID, name)
	var interm vmSSHKeyGetResponse
	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		return nil, c.handleAPIError(err, "vm ssh key", name, "get")
	}

	return &VMSSHKey{
		Name:              interm.Metadata.Name,
		CreationTimestamp: interm.Metadata.CreationTimestamp,
		Org:               org,
		PublicKey:         interm.Spec.PublicKey,
		Description:       interm.Spec.Description,
		Fingerprint:       interm.Status.Fingerprint,
		Validated:         interm.Status.Validated,
	}, nil
}

// DeleteVMSSHKey deletes a VM SSH key by name in the given namespace.
func (c *RackspaceSpotClient) DeleteVMSSHKey(ctx context.Context, org, name string) error {
	if err := ValidateOrgName(org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	if err := ValidateResourceName(name); err != nil {
		return fmt.Errorf("invalid vm ssh key name: %w", err)
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return c.handleAPIError(err, "organization", org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmsshkeys/%s", c.BaseURL, orgID, name)
	err = c.doRequest(ctx, http.MethodDelete, url, nil, c.authHeader(), nil)
	return c.handleAPIError(err, "vm ssh key", name, "delete")
}
