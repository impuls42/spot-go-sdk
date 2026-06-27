package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	// "os"

	v1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("VM SDK Quick Test")
	fmt.Println("========================================\n")

	// Initialize spot client with testbed credentials
	// Note: SDK appends /oauth/token to OAuthURL, so use base domain only
	spotClient, err := v1.NewSpotClient(&v1.Config{
		RefreshToken: "<YOUR_REFRESH-TOKEN>",
	})
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Authenticate
	fmt.Println("🔐 Authenticating...")
	_, err = spotClient.Authenticate(ctx)
	if err != nil {
		log.Fatalf("❌ Authentication failed: %v", err)
	}
	fmt.Println("✅ Authentication successful!\n")

	// Test 1: List Regions
	fmt.Println("📍 Listing available regions...")
	regions, err := spotClient.ListRegions(ctx)
	if err != nil {
		log.Printf("❌ Failed to list regions: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d regions:\n", len(regions))
		for i, region := range regions {
			fmt.Printf("   %d. %s - %s\n", i+1, region.Name, region.Description)
		}
	}
	fmt.Println()

	// Test 2: List VM SSH Keys
	fmt.Println("🔑 Listing VM SSH Keys...")
	keys, err := spotClient.ListVMSSHKeys(ctx, "hooli")
	if err != nil {
		log.Printf("❌ Failed to list VM SSH keys: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d VM SSH keys:\n", len(keys.Items))
		for i, key := range keys.Items {
			fmt.Printf("   %d. %s (Validated: %v)\n", i+1, key.Name, key.Validated)
		}
	}
	fmt.Println()

	// Test 3: List VM CloudSpaces
	fmt.Println("☁️  Listing VM CloudSpaces...")
	vmCloudSpaces, err := spotClient.ListVMCloudSpaces(ctx, "hooli")
	if err != nil {
		log.Printf("❌ Failed to list VM CloudSpaces: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d VM CloudSpaces:\n", len(vmCloudSpaces.Items))
		for i, vmcs := range vmCloudSpaces.Items {
			fmt.Printf("   %d. %s (Region: %s, Status: %s)\n",
				i+1, vmcs.Name, vmcs.Region, vmcs.Status)
			fmt.Printf("      - VM Pools: %d\n", len(vmcs.VMPools))
			fmt.Printf("      - Assigned Servers: %d\n", len(vmcs.AssignedServers))

			// Show VM Pools details
			if len(vmcs.VMPools) > 0 {
				fmt.Println("      VM Pools:")
				for j, pool := range vmcs.VMPools {
					fmt.Printf("        %d. %s (Desired: %d, Won: %d, BidPrice: %s)\n",
						j+1, pool.Name, pool.Desired, pool.WonCount, pool.BidPrice)
				}
			}
		}
	}
	fmt.Println()

	// Test 4: If VM CloudSpaces exist, get details of the first one
	if vmCloudSpaces != nil && len(vmCloudSpaces.Items) > 0 {
		firstVMCS := vmCloudSpaces.Items[0].Name
		fmt.Printf("🔍 Getting details of VM CloudSpace '%s'...\n", firstVMCS)
		vmcs, err := spotClient.GetVMCloudSpace(ctx, "hooli", firstVMCS)
		if err != nil {
			log.Printf("❌ Failed to get VM CloudSpace: %v\n", err)
		} else {
			vmcsJSON, _ := json.MarshalIndent(vmcs, "", "  ")
			fmt.Printf("✅ VM CloudSpace Details:\n%s\n", string(vmcsJSON))
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("✅ VM SDK test completed!")
	fmt.Println("========================================")
}
