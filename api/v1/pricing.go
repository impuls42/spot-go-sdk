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

	var serverData ServerData
	if err := c.doRequest(ctx, http.MethodGet, PriceDetailsURL, nil, nil, &serverData); err != nil {
		return nil, c.handleAPIError(err, "server class", "", "get price details")
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

	var serverData ServerData
	if err := c.doRequest(ctx, http.MethodGet, PriceDetailsURL, nil, nil, &serverData); err != nil {
		return nil, c.handleAPIError(err, "server class", serverClass, "get price details")
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
	var serverData ServerData
	if err := c.doRequest(ctx, http.MethodGet, PriceDetailsURL, nil, nil, &serverData); err != nil {
		return nil, c.handleAPIError(err, "region", regionName, "get price details")
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
	var serverData ServerData
	if err := c.doRequest(ctx, http.MethodGet, PriceDetailsURL, nil, nil, &serverData); err != nil {
		return "", c.handleAPIError(err, "server class", serverClass, "get market price")
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
