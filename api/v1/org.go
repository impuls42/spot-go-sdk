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
