package rxtspot

import (
	"context"
	"fmt"
	"net/http"
)

type ServerData struct {
	Regions map[string]RegionPricingDetails `json:"regions"`
}

type RegionPricingDetails struct {
	Generation    string                               `json:"generation"`
	ServerClasses map[string]ServerClassPricingDetails `json:"serverclasses"`
}

type ServerClassPricingDetails struct {
	Percentile20 float64 `json:"20_percentile"`
	Percentile50 float64 `json:"50_percentile"`
	Percentile80 float64 `json:"80_percentile"`
	MarketPrice  string  `json:"market_price"`
	CPU          string  `json:"cpu"`
	Memory       string  `json:"memory"`
	DisplayName  string  `json:"display_name"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
}

var PriceDetailsURL = GetPriceDetailsURL()

// GetPriceDetails retrieves the price details for a server class.
func (c *RackspaceSpotClient) GetPriceDetails(ctx context.Context) ([]*PriceDetails, error) {

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/serverclasses", c.BaseURL)
	var interm ListServerClassesResponse
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm); err != nil {
		return nil, c.handleAPIError(err, "server class", "", "list")
	}
	var completePriceDetails []*PriceDetails

	for _, item := range interm.Items {
		completePriceDetails = append(completePriceDetails, &PriceDetails{
			ServerClassName: item.Metadata.Name,
			Region:          item.Spec.Region,
			MarketPrice:     "$" + item.Status.SpotPricing.MarketPricePerHour,
			CPU:             item.Spec.Resources.CPU,
			Memory:          item.Spec.Resources.Memory,
			DisplayName:     item.Spec.DisplayName,
			Category:        item.Spec.Category,
		})
	}

	return completePriceDetails, nil
}

func (c *RackspaceSpotClient) GetPriceDetailsForServerClass(ctx context.Context, serverClassName string) (*PriceDetails, error) {

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/serverclasses", c.BaseURL)
	var interm ListServerClassesResponse
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm); err != nil {
		return nil, c.handleAPIError(err, "server class", "", "list")
	}
	var priceDetails PriceDetails

	for _, item := range interm.Items {
		if item.Metadata.Name == serverClassName {
			priceDetails = PriceDetails{
				ServerClassName: item.Metadata.Name,
				Region:          item.Spec.Region,
				MarketPrice:     "$" + item.Status.SpotPricing.MarketPricePerHour,
				CPU:             item.Spec.Resources.CPU,
				Memory:          item.Spec.Resources.Memory,
				DisplayName:     item.Spec.DisplayName,
				Category:        item.Spec.Category,
			}
			return &priceDetails, nil
		}
	}
	return nil, fmt.Errorf("server class '%s' not found", serverClassName)
}

func (c *RackspaceSpotClient) GetPriceDetailsForRegion(ctx context.Context, regionName string) (*PriceDetails, error) {
	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/serverclasses", c.BaseURL)
	var interm ListServerClassesResponse
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm); err != nil {
		return nil, c.handleAPIError(err, "server class", "", "list")
	}
	var priceDetails PriceDetails

	for _, item := range interm.Items {
		if item.Spec.Region == regionName {
			priceDetails = PriceDetails{
				ServerClassName: item.Metadata.Name,
				Region:          regionName,
				MarketPrice:     "$" + item.Status.SpotPricing.MarketPricePerHour,
				CPU:             item.Spec.Resources.CPU,
				Memory:          item.Spec.Resources.Memory,
				DisplayName:     item.Spec.DisplayName,
				Category:        item.Spec.Category,
			}
		}
	}
	return &priceDetails, nil
}

func (c *RackspaceSpotClient) GetMarketPriceForServerClass(ctx context.Context, serverClassName string) (string, error) {
	marketPricePerHour := "N/A"

	url := fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/serverclasses", c.BaseURL)
	var interm ListServerClassesResponse
	if err := c.doRequest(ctx, http.MethodGet, url, nil, c.authHeader(), &interm); err != nil {
		return "", c.handleAPIError(err, "server class", "", "list")
	}
	for _, item := range interm.Items {
		if item.Metadata.Name == serverClassName {
			marketPricePerHour = "$" + item.Status.SpotPricing.MarketPricePerHour
			return marketPricePerHour, nil
		}
	}
	return "", fmt.Errorf("server class '%s' not found", serverClassName)
}

func (c *RackspaceSpotClient) GetMinimumBidPriceForServerClass(ctx context.Context, serverClass string) (string, error) {
	ServerClassDetails, _ := c.GetServerClass(ctx, serverClass)
	return ServerClassDetails.MinBidPricePerHour, nil
}
