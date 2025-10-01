package rxtspot

import "os"

func GetPriceDetailsURL() string {
	if os.Getenv("NGPC_PRICE_DETAILS_URL") != "" {
		return os.Getenv("NGPC_PRICE_DETAILS_URL")
	}
	return "https://ngpc-prod-public-data.s3.us-east-2.amazonaws.com/percentiles.json"
}
