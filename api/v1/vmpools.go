package rxtspot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListVMPools retrieves all VM pools for a given VM cloud space.
func (c *RackspaceSpotClient) ListVMPools(ctx context.Context, org string, vmCloudSpace string) ([]*VMPool, error) {
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

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmpools?labelSelector=ngpc.rxt.io/vmcloudspace=%s", c.BaseURL, orgID, vmCloudSpace)

	var interm vmPoolListResponse
	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		return nil, c.handleAPIError(err, "vm pools", "", "list")
	}

	var pools []*VMPool
	for _, item := range interm.Items {
		pools = append(pools, &VMPool{
			Name:              item.Metadata.Name,
			CreationTimestamp: item.Metadata.CreationTimestamp,
			Org:               org,
			VMCloudSpace:      item.Spec.VMCloudSpace,
			ServerClass:       item.Spec.ServerClass,
			Desired:           item.Spec.Desired,
			BidPrice:          item.Spec.BidPrice,
			PoolType:          item.Spec.PoolType,
			VMImage:           item.Spec.VMImage,
			VMUserData:        item.Spec.VMUserData,
			WonCount:          item.Status.WonCount,
			BidStatus:         item.Status.BidStatus,
		})
	}
	return pools, nil
}

// CreateVMPool creates a new VM pool in the given namespace.
func (c *RackspaceSpotClient) CreateVMPool(ctx context.Context, org string, pool VMPool) error {
	if err := ValidateOrgName(org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	if pool.VMCloudSpace == "" {
		return fmt.Errorf("vm cloudspace name is required")
	}
	if pool.ServerClass == "" {
		return fmt.Errorf("server class is required")
	}
	if pool.BidPrice == "" {
		return fmt.Errorf("bid price is required")
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return c.handleAPIError(err, "organization", org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}

	reqBody := VMPoolCreateRequestBody{
		APIVersion: "ngpc.rxt.io/v1",
		Kind:       "VMPool",
		Metadata: ObjectMeta{
			Name:      pool.Name,
			Namespace: orgID,
		},
	}
	reqBody.Spec.BidPrice = pool.BidPrice
	reqBody.Spec.Desired = pool.Desired
	reqBody.Spec.ServerClass = pool.ServerClass
	reqBody.Spec.VMCloudSpace = pool.VMCloudSpace
	if pool.PoolType != "" {
		reqBody.Spec.PoolType = pool.PoolType
	}
	if pool.VMImage != "" {
		reqBody.Spec.VMImage = pool.VMImage
	}
	if pool.VMUserData != "" {
		reqBody.Spec.VMUserData = pool.VMUserData
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return c.handleAPIError(err, "vm pool", pool.Name, "create")
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmpools", c.BaseURL, orgID)

	if err := c.doRequest(ctx, http.MethodPost, url, body, c.authHeader(), nil); err != nil {
		return c.handleAPIError(err, "vm pool", pool.Name, "create")
	}
	return nil
}

// GetVMPool retrieves a VM pool by name in the given namespace.
func (c *RackspaceSpotClient) GetVMPool(ctx context.Context, org, name string) (*VMPool, error) {
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

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmpools/%s", c.BaseURL, orgID, name)
	var interm vmPoolGetResponse
	err = c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm)
	if err != nil {
		return nil, c.handleAPIError(err, "vm pool", name, "get")
	}

	return &VMPool{
		Name:              interm.Metadata.Name,
		CreationTimestamp: interm.Metadata.CreationTimestamp,
		Org:               org,
		VMCloudSpace:      interm.Spec.VMCloudSpace,
		ServerClass:       interm.Spec.ServerClass,
		Desired:           interm.Spec.Desired,
		BidPrice:          interm.Spec.BidPrice,
		PoolType:          interm.Spec.PoolType,
		VMImage:           interm.Spec.VMImage,
		VMUserData:        interm.Spec.VMUserData,
		WonCount:          interm.Status.WonCount,
		BidStatus:         interm.Status.BidStatus,
	}, nil
}

// UpdateVMPool updates an existing VM pool.
func (c *RackspaceSpotClient) UpdateVMPool(ctx context.Context, org string, pool VMPool) error {
	if err := ValidateOrgName(org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	if pool.Name == "" {
		return fmt.Errorf("vm pool name is required")
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return c.handleAPIError(err, "organization", org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}

	var updateBody VMPoolUpdateRequestBody
	updateBody.Spec.Desired = pool.Desired
	updateBody.Spec.BidPrice = pool.BidPrice

	body, err := json.Marshal(updateBody)
	if err != nil {
		return c.handleAPIError(err, "vm pool", pool.Name, "update")
	}
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmpools/%s", c.BaseURL, orgID, pool.Name)

	if err := c.doRequest(ctx, http.MethodPatch, url, body, c.authHeader(), nil); err != nil {
		return c.handleAPIError(err, "vm pool", pool.Name, "update")
	}
	return nil
}

// DeleteVMPool deletes a VM pool by name in the given namespace.
func (c *RackspaceSpotClient) DeleteVMPool(ctx context.Context, org, name string) error {
	if err := ValidateOrgName(org); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}

	exists, orgID, err := c.getOrgIDIFExists(ctx, org)
	if err != nil {
		return c.handleAPIError(err, "organization", org, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", org)
	}

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/vmpools/%s", c.BaseURL, orgID, name)
	err = c.doRequest(ctx, http.MethodDelete, url, nil, c.authHeader(), nil)
	return c.handleAPIError(err, "vm pool", name, "delete")
}
