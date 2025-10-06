// examples/cloudspace/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	v1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

func main() {
	// Initialize spot client

	spotClient, err := v1.NewSpotClient(&v1.Config{
		RefreshToken: "<YOUR_REFRESH-TOKEN>",
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	_, err = spotClient.Authenticate(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		log.Fatalf("Failed to authenticate: %v", err)
	}

	ctx := context.Background()
	listRegions(ctx, spotClient)

	fmt.Println("Let's Create a cloudspace using SDK...")
	createCloudspace(ctx, spotClient)
}

func createCloudspace(ctx context.Context, spotClient *v1.RackspaceSpotClient) {
	spotNodePool := v1.SpotNodePool{
		Name:        "sdk-spot-nodepool",
		Org:         "hooli",
		Cloudspace:  "sdk-cloudspace",
		ServerClass: "ch.vs1.large-dfw",
		Desired:     1,
		CustomAnnotations: map[string]string{
			"example.com/annotation": "value",
		},
		CustomLabels: map[string]string{
			"example.com/label": "value",
		},
		BidPrice: "$0.08",
	}
	err := spotClient.CreateCloudspace(ctx, v1.CloudSpace{
		Name:              "sdk-cloudspace",
		Org:               "hooli",
		KubernetesVersion: "1.31.1",
		CNI:               "calico",
		Region:            "us-east-iad-1",
		SpotNodepools: []*v1.SpotNodePool{
			&spotNodePool,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create cloudspace: %v", err)
	}
	fmt.Println("Successfully created cloudspace")
	time.Sleep(time.Second * 10)
	cloudspace, err := spotClient.GetCloudspace(ctx, "hooli", "sdk-cloudspace")
	if err != nil {
		log.Fatalf("Failed to get cloudspace: %v", err)
	}
	cloudspaceJSON, err := json.Marshal(cloudspace)
	if err != nil {
		log.Fatalf("Failed to marshal cloudspace: %v", err)
	}
	fmt.Printf("cloudspace: %s\n", cloudspaceJSON)
}

func listRegions(ctx context.Context, spotClient *v1.RackspaceSpotClient) {
	regions, err := spotClient.ListRegions(ctx)
	if err != nil {
		log.Fatalf("Failed to list regions: %v", err)
	}

	fmt.Println("Regions:")
	for _, region := range regions {
		fmt.Printf("- %s (%s)\n", region.Name, region.Name)
	}
}
