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

// GetPriceDetails retrieves the price details for a server class.
func (c *RackspaceSpotClient) GetPriceDetails(ctx context.Context) ([]*PriceDetails, error) {

	url := "https://ngpc-prod-public-data.s3.us-east-2.amazonaws.com/percentiles.json"
	var serverData ServerData
	if err := c.doRequest(ctx, http.MethodGet, url, nil, nil, &serverData); err != nil {
		return nil, err
	}
	var completePriceDetails []*PriceDetails

	for region, details := range serverData.Regions {
		for serverClassName, pricingDetails := range details.ServerClasses {
			completePriceDetails = append(completePriceDetails, &PriceDetails{
				ServerClassName: serverClassName,
				Region:          region,
				MarketPrice:     "$" + pricingDetails.MarketPrice,
				CPU:             pricingDetails.CPU,
				Memory:          pricingDetails.Memory,
				DisplayName:     pricingDetails.DisplayName,
				Category:        pricingDetails.Category,
			})
		}
	}
	return completePriceDetails, nil
}

func (c *RackspaceSpotClient) GetPriceDetailsForServerClass(ctx context.Context, serverClass string) (*PriceDetails, error) {

	url := "https://ngpc-prod-public-data.s3.us-east-2.amazonaws.com/percentiles.json"
	var serverData ServerData
	if err := c.doRequest(ctx, http.MethodGet, url, nil, nil, &serverData); err != nil {
		return nil, err
	}
	var priceDetails PriceDetails

	for region, details := range serverData.Regions {
		for serverClassName, pricingDetails := range details.ServerClasses {
			if serverClassName == serverClass {
				priceDetails = PriceDetails{
					ServerClassName: serverClassName,
					Region:          region,
					MarketPrice:     "$" + pricingDetails.MarketPrice,
					CPU:             pricingDetails.CPU,
					Memory:          pricingDetails.Memory,
					DisplayName:     pricingDetails.DisplayName,
					Category:        pricingDetails.Category,
				}
				return &priceDetails, nil

			}
		}
	}
	return nil, fmt.Errorf("server class '%s' not found", serverClass)
}

func (c *RackspaceSpotClient) GetPriceDetailsForRegion(ctx context.Context, regionName string) (*PriceDetails, error) {
	url := "https://ngpc-prod-public-data.s3.us-east-2.amazonaws.com/percentiles.json"
	var serverData ServerData
	if err := c.doRequest(ctx, http.MethodGet, url, nil, nil, &serverData); err != nil {
		return nil, err
	}
	var priceDetails PriceDetails

	for region, details := range serverData.Regions {
		if region == regionName {
			for serverClassName, pricingDetails := range details.ServerClasses {
				priceDetails = PriceDetails{
					ServerClassName: serverClassName,
					Region:          region,
					MarketPrice:     "$" + pricingDetails.MarketPrice,
					CPU:             pricingDetails.CPU,
					Memory:          pricingDetails.Memory,
					DisplayName:     pricingDetails.DisplayName,
					Category:        pricingDetails.Category,
				}
			}
		}
		return &priceDetails, nil
	}
	return nil, fmt.Errorf("region '%s' not found", regionName)
}

func (c *RackspaceSpotClient) GetMarketPriceForServerClass(ctx context.Context, serverClass string) (string, error) {
	url := "https://ngpc-prod-public-data.s3.us-east-2.amazonaws.com/percentiles.json"
	var serverData ServerData
	if err := c.doRequest(ctx, http.MethodGet, url, nil, nil, &serverData); err != nil {
		return "", err
	}

	for _, details := range serverData.Regions {
		for serverClassName, pricingDetails := range details.ServerClasses {
			if serverClassName == serverClass {
				return "$" + pricingDetails.MarketPrice, nil
			}
		}
	}
	return "", fmt.Errorf("server class '%s' not found", serverClass)
}

func (c *RackspaceSpotClient) GetMinimumBidPriceForServerClass(ctx context.Context, serverClass string) (string, error) {
	ServerClassDetails, _ := c.GetServerClass(ctx, serverClass)
	return ServerClassDetails.MinBidPricePerHour, nil
}
