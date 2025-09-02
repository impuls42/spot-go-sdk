package rxtspot

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

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
		return nil, c.handleAPIError(err, "organization", "", "list")
	}
	return response.Organizations, nil
}

func (c *RackspaceSpotClient) DeleteOrganization(ctx context.Context, orgName string) error {
	if err := ValidateOrgName(orgName); err != nil {
		return fmt.Errorf("invalid organization name: %w", err)
	}
	exists, orgID, err := c.getOrgIDIFExistsWithoutNormalizing(ctx, orgName)
	if err != nil {
		return c.handleAPIError(err, "organization", orgName, "find")
	}
	if !exists {
		return fmt.Errorf("organization '%s' not found", orgName)
	}
	url := fmt.Sprintf("%s/apis/auth.ngpc.rxt.io/v1/organizations/%s", c.BaseURL, orgID) // Correct URL
	fmt.Printf("url: %s\n", url)
	err = c.doRequest(ctx, http.MethodDelete, url, nil, c.authHeader(), nil)
	return c.handleAPIError(err, "organization", orgName, "delete")
}

func (c *RackspaceSpotClient) getOrgIDIFExists(ctx context.Context, orgName string) (bool, string, error) {
	url := fmt.Sprintf("%s/apis/auth.ngpc.rxt.io/v1/organizations", c.BaseURL) // Correct URL

	// Structure for decoding
	var response struct {
		Organizations []Organization `json:"organizations"`
	}

	// Pass &response to doRequest so it decodes automatically
	err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &response)
	if err != nil {
		return false, "", c.handleAPIError(err, "organization", orgName, "find")
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

func (c *RackspaceSpotClient) getOrgIDIFExistsWithoutNormalizing(ctx context.Context, orgName string) (bool, string, error) {
	url := fmt.Sprintf("%s/apis/auth.ngpc.rxt.io/v1/organizations", c.BaseURL) // Correct URL

	// Structure for decoding
	var response struct {
		Organizations []Organization `json:"organizations"`
	}

	// Pass &response to doRequest so it decodes automatically
	err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &response)
	if err != nil {
		return false, "", c.handleAPIError(err, "organization", orgName, "find")
	}

	for _, org := range response.Organizations {
		if org.Name == orgName {
			return true, org.ID, nil
		}
	}
	return false, "", nil
}
