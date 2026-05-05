package rxtspot

import (
	"context"
	"fmt"
	"net/http"
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

func (c *RackspaceSpotClient) getOrgIDIFExists(ctx context.Context, orgNameOrID string) (bool, string, error) {
	url := fmt.Sprintf("%s/apis/auth.ngpc.rxt.io/v1/organizations", c.BaseURL)

	var response struct {
		Organizations []Organization `json:"organizations"`
	}

	err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &response)
	if err != nil {
		return false, "", c.handleAPIError(err, "organization", orgNameOrID, "find")
	}

	for _, org := range response.Organizations {
		// Try matching by org name (preferred) or org ID (fallback)
		if org.Name == orgNameOrID || org.ID == orgNameOrID {
			return true, org.ID, nil
		}
	}
	return false, "", nil
}

func (c *RackspaceSpotClient) GetOrgID(ctx context.Context, orgNameOrID string) (bool, string, error) {
	return c.getOrgIDIFExists(ctx, orgNameOrID)
}
