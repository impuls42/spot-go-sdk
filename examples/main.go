package main

import (
	"context"
	"log"
	"os"

	v1 "github.com/rackerlabs/spot-sdk/rxtspot/api/v1"
	"github.com/rackerlabs/spot-sdk/rxtspot/examples/internal"
)

func main() {
	ctx := context.Background()
	refreshToken := os.Getenv("SPOT_REFRESH_TOKEN")
	if refreshToken == "" {
		log.Fatal("SPOT_REFRESH_TOKEN environment variable is required")
	}

	client := v1.NewClient(refreshToken)
	if err := internal.Authenticate(ctx, client); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	log.Println("Authenticated successfully.")

	namespace, err := internal.SelectNamespace(ctx, client)
	if err != nil {
		log.Fatalf("Namespace selection failed: %v", err)
	}
	region, err := internal.SelectRegion(ctx, client)
	if err != nil {
		log.Fatalf("Region selection failed: %v", err)
	}
	serverClass, err := internal.SelectServerClass(ctx, client)
	if err != nil {
		log.Fatalf("Server class selection failed: %v", err)
	}

	cloudspaceName := "example-cloudspace"
	spotPoolName := "example-spot-pool"
	onDemandPoolName := "example-ondemand-pool"

	if err := internal.CreateAndManageCloudspace(ctx, client, namespace, region, cloudspaceName); err != nil {
		log.Fatalf("Cloudspace management failed: %v", err)
	}
	if err := internal.CreateAndManageNodePools(ctx, client, namespace, cloudspaceName, serverClass, spotPoolName, onDemandPoolName); err != nil {
		log.Fatalf("Node pool management failed: %v", err)
	}

	internal.ShowRegionAndClassInfo(ctx, client, region, serverClass)
	defer internal.CleanupResources(ctx, client, namespace, cloudspaceName, spotPoolName, onDemandPoolName)
}
