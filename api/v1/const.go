package rxtspot

import "os"

// PricingURL is the default URL for percentile pricing data.
const PricingURL = "https://ngpc-prod-public-data.s3.us-east-2.amazonaws.com/percentiles.json"

func GetPriceDetailsURL() string {
	if val, exists := os.LookupEnv("NGPC_PRICE_DETAILS_URL"); exists && val != "" {
		return val
	}
	return PricingURL
}
