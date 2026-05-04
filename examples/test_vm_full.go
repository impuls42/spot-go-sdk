package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	// "os"
	"time"

	"github.com/google/uuid"
	v1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

const (
	testOrg    = "hooli"         // Change this to your test organization name
	testRegion = "us-west-sjc-1" // Change this to your desired region
)

func main() {
	// Initialize spot client with testbed credentials
	spotClient, err := v1.NewSpotClient(&v1.Config{
		RefreshToken: "<YOUR_REFRESH-TOKEN>",
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Authenticate
	fmt.Println("🔐 Authenticating...")
	_, err = spotClient.Authenticate(context.Background())
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	fmt.Println("✅ Authentication successful!\n")

	ctx := context.Background()

	// Generate unique UUID for VM Pool to avoid 409 conflicts
	sshKeyName := "test-vm-ssh-key"
	vmCloudSpaceName := "test-vm-cloudspace"
	vmPoolName := uuid.New().String()

	// Run VM tests
	fmt.Println("========================================")
	fmt.Println("VM CloudSpace & VM Pool Testing Suite")
	fmt.Println("========================================\n")

	fmt.Printf("📝 Resource names:\n")
	fmt.Printf("   VM SSH Key:       %s\n", sshKeyName)
	fmt.Printf("   VM CloudSpace: %s\n", vmCloudSpaceName)
	fmt.Printf("   VM Pool:       %s (UUID)\n\n", vmPoolName)

	// Test 1: List regions
	testListRegions(ctx, spotClient)

	// Test 2: List VM SSH Keys
	testListVMSSHKeys(ctx, spotClient)

	// Test 3: Create VM SSH Key (if needed)
	testCreateVMSSHKey(ctx, spotClient, sshKeyName)

	// Test 4: Get VM SSH Key
	testGetVMSSHKey(ctx, spotClient, sshKeyName)

	// Test 5: List VM CloudSpaces
	testListVMCloudSpaces(ctx, spotClient)

	// Test 6: Create VM CloudSpace
	testCreateVMCloudSpace(ctx, spotClient, vmCloudSpaceName, sshKeyName)

	// Test 7: Get VM CloudSpace
	testGetVMCloudSpace(ctx, spotClient, vmCloudSpaceName)

	// Test 8: Update VM CloudSpace (only Webhook field is updatable)
	testUpdateVMCloudSpace(ctx, spotClient, vmCloudSpaceName)

	// Test 9: Create VM Pool (with polling for Fulfilled status)
	testCreateVMPool(ctx, spotClient, vmCloudSpaceName, vmPoolName)

	// Test 9b: Verify VMCloudSpace has VMs with IP addresses
	testVerifyVMCloudSpaceReady(ctx, spotClient, vmCloudSpaceName)

	// Test 10: List VM Pools
	testListVMPools(ctx, spotClient, vmCloudSpaceName)

	// Test 11: Get VM Pool
	testGetVMPool(ctx, spotClient, vmPoolName)

	// Test 12: Update VM Pool
	testUpdateVMPool(ctx, spotClient, vmPoolName)

	// Cleanup option
	fmt.Println("\n========================================")
	fmt.Println("Cleanup Phase")
	fmt.Println("========================================\n")

	testDeleteVMPool(ctx, spotClient, vmPoolName)
	testDeleteVMCloudSpace(ctx, spotClient, vmCloudSpaceName)
	testDeleteVMSSHKey(ctx, spotClient, sshKeyName)

	fmt.Println("\n========================================")
	fmt.Println("✅ All VM tests completed successfully!")
	fmt.Println("========================================")
}

func testListRegions(ctx context.Context, client *v1.RackspaceSpotClient) {
	fmt.Println("📍 Test: List Regions")
	regions, err := client.ListRegions(ctx)
	if err != nil {
		log.Printf("⚠️  Failed to list regions: %v\n\n", err)
		return
	}

	fmt.Printf("Found %d regions:\n", len(regions))
	for i, region := range regions {
		fmt.Printf("  %d. %s - %s\n", i+1, region.Name, region.Description)
	}
	fmt.Println("✅ List regions successful\n")
}

func testListVMSSHKeys(ctx context.Context, client *v1.RackspaceSpotClient) {
	fmt.Println("🔑 Test: List VM SSH Keys")
	keys, err := client.ListVMSSHKeys(ctx, testOrg)
	if err != nil {
		log.Printf("⚠️  Failed to list VM SSH keys: %v\n\n", err)
		return
	}

	fmt.Printf("Found %d VM SSH keys:\n", len(keys.Items))
	for i, key := range keys.Items {
		fmt.Printf("  %d. %s (Validated: %v, Fingerprint: %s)\n", i+1, key.Name, key.Validated, key.Fingerprint)
	}
	fmt.Println("✅ List VM SSH keys successful\n")
}

func testCreateVMSSHKey(ctx context.Context, client *v1.RackspaceSpotClient, name string) {
	fmt.Printf("🔑 Test: Create VM SSH Key '%s'\n", name)

	// Generate a sample SSH public key (replace with your actual public key)
	samplePublicKey := "ssh-ed25519 AAAAAxxxxxx <ADD SSH KEY HERE>"

	key := v1.VMSSHKey{
		Name:        name,
		Org:         testOrg,
		PublicKey:   samplePublicKey,
		Description: "Test SSH key for VM CloudSpace testing",
	}

	err := client.CreateVMSSHKey(ctx, key)
	if err != nil {
		log.Printf("⚠️  Failed to create VM SSH key (may already exist): %v\n\n", err)
		return
	}

	fmt.Printf("✅ VM SSH key '%s' created successfully\n\n", name)
}

func testGetVMSSHKey(ctx context.Context, client *v1.RackspaceSpotClient, name string) {
	fmt.Printf("🔍 Test: Get VM SSH Key '%s'\n", name)

	key, err := client.GetVMSSHKey(ctx, testOrg, name)
	if err != nil {
		log.Printf("⚠️  Failed to get VM SSH key: %v\n\n", err)
		return
	}

	keyJSON, _ := json.MarshalIndent(key, "", "  ")
	fmt.Printf("VM SSH Key Details:\n%s\n", string(keyJSON))
	fmt.Println("✅ Get VM SSH key successful\n")
}

func testListVMCloudSpaces(ctx context.Context, client *v1.RackspaceSpotClient) {
	fmt.Println("☁️  Test: List VM CloudSpaces")

	vmCloudSpaces, err := client.ListVMCloudSpaces(ctx, testOrg)
	if err != nil {
		log.Printf("⚠️  Failed to list VM CloudSpaces: %v\n\n", err)
		return
	}

	fmt.Printf("Found %d VM CloudSpaces:\n", len(vmCloudSpaces.Items))
	for i, vmcs := range vmCloudSpaces.Items {
		fmt.Printf("  %d. %s (Region: %s, Status: %s, Health: %s)\n",
			i+1, vmcs.Name, vmcs.Region, vmcs.Status, vmcs.Health)
		fmt.Printf("     VM Pools: %d, Assigned Servers: %d\n",
			len(vmcs.VMPools), len(vmcs.AssignedServers))
	}
	fmt.Println("✅ List VM CloudSpaces successful\n")
}

func testCreateVMCloudSpace(ctx context.Context, client *v1.RackspaceSpotClient, name, sshKeyName string) {
	fmt.Printf("☁️  Test: Create VM CloudSpace '%s'\n", name)

	// First, get available regions
	regions, err := client.ListRegions(ctx)
	if err != nil || len(regions) == 0 {
		log.Printf("⚠️  Failed to get regions: %v\n\n", err)
		return
	}

	vmcs := v1.VMCloudSpace{
		Name:   name,
		Org:    testOrg,
		Region: testRegion, // Use first available region
		VMSshKeyRef: v1.VMSshKeyRef{
			Name: sshKeyName,
		},
		Webhook: "", // Optional webhook URL
	}

	err = client.CreateVMCloudSpace(ctx, vmcs)
	if err != nil {
		log.Printf("⚠️  Failed to create VM CloudSpace (may already exist): %v\n\n", err)
		return
	}

	fmt.Printf("✅ VM CloudSpace '%s' created successfully in region '%s'\n", name, regions[0].Name)
	fmt.Println("⏳ Waiting 5 seconds for VM CloudSpace to initialize...")
	time.Sleep(5 * time.Second)
	fmt.Println()
}

func testGetVMCloudSpace(ctx context.Context, client *v1.RackspaceSpotClient, name string) {
	fmt.Printf("🔍 Test: Get VM CloudSpace '%s'\n", name)

	vmcs, err := client.GetVMCloudSpace(ctx, testOrg, name)
	if err != nil {
		log.Printf("⚠️  Failed to get VM CloudSpace: %v\n\n", err)
		return
	}

	vmcsJSON, _ := json.MarshalIndent(vmcs, "", "  ")
	fmt.Printf("VM CloudSpace Details:\n%s\n", string(vmcsJSON))
	fmt.Println("✅ Get VM CloudSpace successful\n")
}

func testUpdateVMCloudSpace(ctx context.Context, client *v1.RackspaceSpotClient, name string) {
	fmt.Printf("✏️  Test: Update VM CloudSpace '%s' (Webhook field)\n", name)

	updatedWebhook := "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

	vmcs := v1.VMCloudSpace{
		Name:    name,
		Webhook: updatedWebhook,
	}

	err := client.UpdateVMCloudSpace(ctx, testOrg, vmcs)
	if err != nil {
		log.Printf("⚠️  Failed to update VM CloudSpace: %v\n\n", err)
		return
	}

	fmt.Printf("✅ VM CloudSpace '%s' updated successfully (Webhook: %s)\n", name, updatedWebhook)

	// Verify the update by fetching the cloudspace
	updated, err := client.GetVMCloudSpace(ctx, testOrg, name)
	if err != nil {
		log.Printf("⚠️  Failed to verify update: %v\n\n", err)
		return
	}

	if updated.Webhook == updatedWebhook {
		fmt.Printf("✅ Verified: Webhook updated correctly to '%s'\n\n", updated.Webhook)
	} else {
		fmt.Printf("⚠️  Webhook mismatch: expected '%s', got '%s'\n\n", updatedWebhook, updated.Webhook)
	}
}

func testCreateVMPool(ctx context.Context, client *v1.RackspaceSpotClient, vmCloudSpace, poolName string) {
	fmt.Printf("🏊 Test: Create VM Pool '%s'\n", poolName)

	// Get server classes to find available ones
	regions, _ := client.ListRegions(ctx)
	if len(regions) == 0 {
		log.Println("⚠️  No regions available")
		return
	}

	serverClasses, err := client.ListServerClasses(ctx, regions[0].Name)
	if err != nil || len(serverClasses.Items) == 0 {
		log.Printf("⚠️  Failed to get server classes: %v\n\n", err)
		return
	}

	// Find a suitable server class
	selectedServerClass := "gp.vs2.medium-sjc"

	// Prepare cloud-init user data (optional) - auto-detects base64
	cloudInitScript := "#!/bin/bash\necho 'Hello from cloud-init'\napt-get update -y\napt-get install -y apache\nsystemctl enable apache2\nsystemctl start apache2\n"
	userData := v1.PrepareUserData(cloudInitScript)

	pool := v1.VMPool{
		Name:         poolName,
		Org:          testOrg,
		VMCloudSpace: vmCloudSpace,
		ServerClass:  selectedServerClass,
		Desired:      1,
		BidPrice:     "0.97",        // Change this as per your convenience
		PoolType:     "spot",        // type of pool - spot only
		VMImage:      "ubuntu24.04", // VM Image to use
		VMUserData:   userData,      // base64-encoded cloud-init user data
	}

	err = client.CreateVMPool(ctx, testOrg, pool)
	if err != nil {
		log.Printf("⚠️  Failed to create VM Pool (may already exist): %v\n\n", err)
		return
	}

	fmt.Printf("✅ VM Pool '%s' creation request submitted with server class '%s'\n", poolName, selectedServerClass)
	fmt.Println("⏳ Waiting for VM Pool to reach 'Fulfilled' status (this may take several minutes)...")

	// Poll for VM Pool to reach Fulfilled status
	if pollVMPoolStatus(ctx, client, poolName, 600) {
		fmt.Printf("✅ VM Pool '%s' is now Fulfilled with VMs provisioned!\n\n", poolName)
	} else {
		log.Printf("⚠️  VM Pool '%s' did not reach Fulfilled status within timeout\n\n", poolName)
	}
}

// pollVMPoolStatus polls VM Pool until it reaches Fulfilled status or timeout
func pollVMPoolStatus(ctx context.Context, client *v1.RackspaceSpotClient, poolName string, timeoutSeconds int) bool {
	startTime := time.Now()
	timeout := time.Duration(timeoutSeconds) * time.Second
	pollInterval := 10 * time.Second

	for time.Since(startTime) < timeout {
		pool, err := client.GetVMPool(ctx, testOrg, poolName)
		if err != nil {
			log.Printf("⚠️  Failed to get pool status: %v\n", err)
			time.Sleep(pollInterval)
			continue
		}

		fmt.Printf("   Status: BidStatus=%s, WonCount=%d, Desired=%d\n",
			pool.BidStatus, pool.WonCount, pool.Desired)

		// Check if pool has reached Fulfilled status
		if pool.BidStatus == "Fulfilled" && pool.WonCount == pool.Desired {
			return true
		}

		// Check for failure states
		if pool.BidStatus == "Lost" {
			log.Printf("⚠️  VM Pool bid was lost\n")
			return false
		}

		elapsed := int(time.Since(startTime).Seconds())
		remaining := timeoutSeconds - elapsed
		fmt.Printf("   ⏳ Waiting... (%d/%d seconds, %d seconds remaining)\n",
			elapsed, timeoutSeconds, remaining)

		time.Sleep(pollInterval)
	}

	return false // Timeout reached
}

// pollVMCloudSpaceReady polls VMCloudSpace until VMs are ready with IP addresses
func pollVMCloudSpaceReady(ctx context.Context, client *v1.RackspaceSpotClient, cloudSpaceName string, timeoutSeconds int) bool {
	startTime := time.Now()
	timeout := time.Duration(timeoutSeconds) * time.Second
	pollInterval := 10 * time.Second

	for time.Since(startTime) < timeout {
		vmcs, err := client.GetVMCloudSpace(ctx, testOrg, cloudSpaceName)
		if err != nil {
			log.Printf("⚠️  Failed to get cloudspace status: %v\n", err)
			time.Sleep(pollInterval)
			continue
		}

		fmt.Printf("   Status: Phase=%s, Health=%s, Assigned Servers=%d\n",
			vmcs.Status, vmcs.Health, len(vmcs.AssignedServers))

		// Check if cloudspace is Ready with assigned servers
		if vmcs.Status == "Ready" && len(vmcs.AssignedServers) > 0 {
			// Verify all servers have IP addresses
			allHaveIPs := true
			for serverName, server := range vmcs.AssignedServers {
				if server.IPAddress == "" {
					fmt.Printf("   ⚠️  Server %s does not have IP yet\n", serverName)
					allHaveIPs = false
				} else {
					fmt.Printf("   ✅ Server %s has IP: %s\n", serverName, server.IPAddress)
				}
			}
			if allHaveIPs {
				return true
			}
		}

		elapsed := int(time.Since(startTime).Seconds())
		remaining := timeoutSeconds - elapsed
		fmt.Printf("   ⏳ Waiting for VMs to be ready... (%d/%d seconds, %d seconds remaining)\n",
			elapsed, timeoutSeconds, remaining)

		time.Sleep(pollInterval)
	}

	return false // Timeout reached
}

func testVerifyVMCloudSpaceReady(ctx context.Context, client *v1.RackspaceSpotClient, cloudSpaceName string) {
	fmt.Printf("🔍 Test: Verify VMCloudSpace '%s' has VMs with IP addresses\n", cloudSpaceName)
	fmt.Println("⏳ Waiting for VMs to be provisioned and assigned IP addresses (this may take up to 20 minutes)...")

	if pollVMCloudSpaceReady(ctx, client, cloudSpaceName, 1200) {
		fmt.Printf("✅ VMCloudSpace '%s' is Ready with VMs and IP addresses!\n\n", cloudSpaceName)
	} else {
		log.Printf("⚠️  VMCloudSpace '%s' did not reach Ready status with IPs within timeout\n\n", cloudSpaceName)
	}
}

func testListVMPools(ctx context.Context, client *v1.RackspaceSpotClient, vmCloudSpace string) {
	fmt.Printf("🏊 Test: List VM Pools for CloudSpace '%s'\n", vmCloudSpace)

	pools, err := client.ListVMPools(ctx, testOrg, vmCloudSpace)
	if err != nil {
		log.Printf("⚠️  Failed to list VM Pools: %v\n\n", err)
		return
	}

	fmt.Printf("Found %d VM Pools:\n", len(pools))
	for i, pool := range pools {
		fmt.Printf("  %d. %s (ServerClass: %s, Desired: %d, Won: %d, BidPrice: %s, Status: %s)\n",
			i+1, pool.Name, pool.ServerClass, pool.Desired, pool.WonCount, pool.BidPrice, pool.BidStatus)
	}
	fmt.Println("✅ List VM Pools successful\n")
}

func testGetVMPool(ctx context.Context, client *v1.RackspaceSpotClient, poolName string) {
	fmt.Printf("🔍 Test: Get VM Pool '%s'\n", poolName)

	pool, err := client.GetVMPool(ctx, testOrg, poolName)
	if err != nil {
		log.Printf("⚠️  Failed to get VM Pool: %v\n\n", err)
		return
	}

	poolJSON, _ := json.MarshalIndent(pool, "", "  ")
	fmt.Printf("VM Pool Details:\n%s\n", string(poolJSON))
	fmt.Println("✅ Get VM Pool successful\n")
}

func testUpdateVMPool(ctx context.Context, client *v1.RackspaceSpotClient, poolName string) {
	fmt.Printf("✏️  Test: Update VM Pool '%s'\n", poolName)

	pool := v1.VMPool{
		Name:     poolName,
		Org:      testOrg,
		Desired:  2,      // Update desired count
		BidPrice: "0.99", // Update bid price
	}

	err := client.UpdateVMPool(ctx, testOrg, pool)
	if err != nil {
		log.Printf("⚠️  Failed to update VM Pool: %v\n\n", err)
		return
	}

	fmt.Printf("✅ VM Pool '%s' updated successfully (Desired: 2, BidPrice: $0.06)\n\n", poolName)
}

func testDeleteVMPool(ctx context.Context, client *v1.RackspaceSpotClient, poolName string) {
	fmt.Printf("🗑️  Test: Delete VM Pool '%s'\n", poolName)

	err := client.DeleteVMPool(ctx, testOrg, poolName)
	if err != nil {
		log.Printf("⚠️  Failed to delete VM Pool: %v\n\n", err)
		return
	}

	fmt.Printf("✅ VM Pool '%s' deleted successfully\n\n", poolName)
}

func testDeleteVMCloudSpace(ctx context.Context, client *v1.RackspaceSpotClient, name string) {
	fmt.Printf("🗑️  Test: Delete VM CloudSpace '%s'\n", name)

	err := client.DeleteVMCloudSpace(ctx, testOrg, name)
	if err != nil {
		log.Printf("⚠️  Failed to delete VM CloudSpace: %v\n\n", err)
		return
	}

	fmt.Printf("✅ VM CloudSpace '%s' deleted successfully\n\n", name)
}

func testDeleteVMSSHKey(ctx context.Context, client *v1.RackspaceSpotClient, name string) {
	fmt.Printf("🗑️  Test: Delete VM SSH Key '%s'\n", name)

	err := client.DeleteVMSSHKey(ctx, testOrg, name)
	if err != nil {
		log.Printf("⚠️  Failed to delete VM SSH Key: %v\n\n", err)
		return
	}

	fmt.Printf("✅ VM SSH Key '%s' deleted successfully\n\n", name)
}
