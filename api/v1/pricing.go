package rxtspot

import (
	"context"
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
	serverClasses, err := c.ListServerClasses(ctx, "")
	if err != nil {
		return nil, err
	}
	var completePriceDetails []*PriceDetails

	for _, item := range serverClasses.Items {
		completePriceDetails = append(completePriceDetails, &PriceDetails{
			ServerClassName: item.Name,
			Region:          item.Region,
			MarketPrice:     item.CurrentMarketPricePerHour,
			CPU:             item.Resources.CPU,
			Memory:          item.Resources.Memory,
			DisplayName:     item.Displayname,
			Category:        item.Category,
		})
	}
	return completePriceDetails, nil
}

func (c *RackspaceSpotClient) GetPriceDetailsForServerClass(ctx context.Context, serverClassName string) (*PriceDetails, error) {
	serverClass, err := c.GetServerClass(ctx, serverClassName)
	if err != nil {
		return nil, err
	}
	priceDetails := PriceDetails{
		ServerClassName: serverClass.Name,
		Region:          serverClass.Region,
		MarketPrice:     serverClass.CurrentMarketPricePerHour,
		CPU:             serverClass.Resources.CPU,
		Memory:          serverClass.Resources.Memory,
		DisplayName:     serverClass.Displayname,
		Category:        serverClass.Category,
	}
	return &priceDetails, nil

}

func (c *RackspaceSpotClient) GetPriceDetailsForRegion(ctx context.Context, regionName string) (*PriceDetails, error) {
	serverClassess, err := c.ListServerClasses(ctx, regionName)
	if err != nil {
		return nil, err
	}

	var priceDetails PriceDetails

	for _, item := range serverClassess.Items {
		if item.Region == regionName {
			priceDetails = PriceDetails{
				ServerClassName: item.Name,
				Region:          regionName,
				MarketPrice:     item.CurrentMarketPricePerHour,
				CPU:             item.Resources.CPU,
				Memory:          item.Resources.Memory,
				DisplayName:     item.Displayname,
				Category:        item.Category,
			}
		}
	}
	return &priceDetails, nil
}

func (c *RackspaceSpotClient) GetMarketPriceForServerClass(ctx context.Context, serverClassStatus *ServerClassStatus) string {
	marketPricePerHour := "N/A"

	if serverClassStatus != nil {
		marketPricePerHour = "$" + serverClassStatus.SpotPricing.MarketPricePerHour

	}
	return marketPricePerHour
}

func (c *RackspaceSpotClient) GetMinimumBidPriceForServerClass(ctx context.Context, serverClass *ServerClassSpec) string {
	var minBidPrice = "N/A"

	if serverClass != nil {
		minBidPrice = "$" + serverClass.MinBidPricePerHour

	}
	return minBidPrice
}
