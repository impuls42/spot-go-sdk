package rxtspot

import (
	"context"
	"fmt"
	"net/http"
)

// ListServerClasses retrieves all available server classes.
func (c *RackspaceSpotClient) ListServerClasses(ctx context.Context, region string) (*ServerClassList, error) {
	if region != "" {
		exists, err := c.checkIfRegionExists(ctx, region)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("region '%s' not found", region)
		}
	}

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/serverclasses", c.BaseURL)

	var interm ListServerClassesResponse
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm); err != nil {
		return nil, err
	}
	var serverclasses []ServerClass
	if region != "" {
		for _, item := range interm.Items {
			if item.Spec.Region == region {
				marketPrice, err := c.GetMarketPriceForServerClass(ctx, item.Metadata.Name)
				if err != nil {
					return nil, err
				}
				serverclasses = append(serverclasses, ServerClass{
					Name:                      item.Metadata.Name,
					Category:                  item.Spec.Category,
					Region:                    item.Spec.Region,
					CurrentMarketPricePerHour: marketPrice,
					MinBidPricePerHour:        "$" + item.Spec.MinBidPricePerHour,
					Resources: Resource{
						CPU:    item.Spec.Resources.CPU,
						Memory: item.Spec.Resources.Memory,
					},
				})
			}
		}
		return &ServerClassList{Items: serverclasses}, nil
	}

	for _, item := range interm.Items {
		marketPrice, _ := c.GetMarketPriceForServerClass(ctx, item.Metadata.Name)
		serverclasses = append(serverclasses, ServerClass{
			Name:                      item.Metadata.Name,
			Category:                  item.Spec.Category,
			Region:                    item.Spec.Region,
			CurrentMarketPricePerHour: marketPrice,
			MinBidPricePerHour:        "$" + item.Spec.MinBidPricePerHour,
			Resources: Resource{
				CPU:    item.Spec.Resources.CPU,
				Memory: item.Spec.Resources.Memory,
			},
		})
	}
	return &ServerClassList{Items: serverclasses}, nil
}

// GetServerClass retrieves a server class by name.
func (c *RackspaceSpotClient) GetServerClass(ctx context.Context, name string) (*ServerClass, error) {
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/serverclasses", c.BaseURL)

	var interm ListServerClassesResponse
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm); err != nil {
		return nil, err
	}
	marketPrice, err := c.GetMarketPriceForServerClass(ctx, name)
	if err != nil {
		return nil, err
	}

	var serverclass ServerClass

	for _, item := range interm.Items {
		if item.Metadata.Name == name {
			serverclass = ServerClass{
				Name:                      item.Metadata.Name,
				Category:                  item.Spec.Category,
				Region:                    item.Spec.Region,
				MinBidPricePerHour:        "$" + item.Spec.MinBidPricePerHour,
				CurrentMarketPricePerHour: marketPrice,
				Resources: Resource{
					CPU:    item.Spec.Resources.CPU,
					Memory: item.Spec.Resources.Memory,
				},
			}
			return &serverclass, nil
		}
	}
	return nil, fmt.Errorf("server class '%s' not found", name)
}
