package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "github.com/rackerlabs/spot-sdk/rxtspot/api/v1"
)

// Authenticate authenticates the client with the Rackspace Spot API.
func Authenticate(ctx context.Context, client *v1.RackspaceSpotClient) error {
	return client.Authenticate(ctx)
}

// SelectNamespace selects the first available namespace (organization).
func SelectNamespace(ctx context.Context, client *v1.RackspaceSpotClient) (string, error) {
	orgs, err := client.ListOrganizations(ctx)
	if err != nil {
		return "", err
	}
	if len(orgs) == 0 {
		return "", fmt.Errorf("no organizations found")
	}
	fmt.Printf("Organizations: %+v\n", orgs)
	fmt.Printf("Using namespace: %s\n", orgs[0].Namespace)
	return orgs[0].Namespace, nil
}

// SelectRegion selects the first available region.
func SelectRegion(ctx context.Context, client *v1.RackspaceSpotClient) (string, error) {
	regions, err := client.ListRegions(ctx)
	if err != nil {
		return "", err
	}
	if len(regions) == 0 {
		return "", fmt.Errorf("no regions found")
	}
	fmt.Printf("Regions: %+v\n", regions)
	return regions[0].Name, nil
}

// SelectServerClass selects the first available server class.
func SelectServerClass(ctx context.Context, client *v1.RackspaceSpotClient) (string, error) {
	classes, err := client.ListServerClasses(ctx)
	if err != nil {
		return "", err
	}
	if len(classes) == 0 {
		return "", fmt.Errorf("no server classes found")
	}
	fmt.Printf("Server Classes: %+v\n", classes)
	return classes[0].Name, nil
}

// CreateAndManageCloudspace creates, gets, and lists a cloudspace.
func CreateAndManageCloudspace(ctx context.Context, client *v1.RackspaceSpotClient, namespace, region, name string) error {
	cloudspace := v1.Cloudspace{
		Name:              name,
		Namespace:         namespace,
		Region:            region,
		KubernetesVersion: "v1.31.1",
	}
	createdCS, err := client.CreateCloudspace(ctx, cloudspace)
	if err != nil {
		return fmt.Errorf("CreateCloudspace failed: %w", err)
	}
	fmt.Printf("Created Cloudspace: %+v\n", createdCS)

	gotCS, err := client.GetCloudspace(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("GetCloudspace failed: %w", err)
	}
	fmt.Printf("Got Cloudspace: %+v\n", gotCS)

	cloudspaces, err := client.ListCloudspaces(ctx, namespace)
	if err != nil {
		return fmt.Errorf("ListCloudspaces failed: %w", err)
	}
	fmt.Printf("Cloudspaces: %+v\n", cloudspaces)
	return nil
}

// CreateAndManageNodePools creates, gets, and lists spot and on-demand node pools.
func CreateAndManageNodePools(ctx context.Context, client *v1.RackspaceSpotClient, namespace, cloudspace, serverClass, spotPoolName, onDemandPoolName string) error {
	spotPool := v1.SpotNodePool{
		Name:        spotPoolName,
		Namespace:   namespace,
		Cloudspace:  cloudspace,
		ServerClass: serverClass,
		Desired:     1,
		BidPrice:    "0.5",
	}
	createdSpot, err := client.CreateSpotNodePool(ctx, spotPool)
	if err != nil {
		return fmt.Errorf("CreateSpotNodePool failed: %w", err)
	}
	fmt.Printf("Created SpotNodePool: %+v\n", createdSpot)

	gotSpot, err := client.GetSpotNodePool(ctx, namespace, spotPoolName)
	if err != nil {
		return fmt.Errorf("GetSpotNodePool failed: %w", err)
	}
	fmt.Printf("Got SpotNodePool: %+v\n", gotSpot)

	spotPools, err := client.ListSpotNodePools(ctx, namespace)
	if err != nil {
		return fmt.Errorf("ListSpotNodePools failed: %w", err)
	}
	fmt.Printf("SpotNodePools: %+v\n", spotPools)

	onDemandPool := v1.OnDemandNodePool{
		Name:        onDemandPoolName,
		Namespace:   namespace,
		Cloudspace:  cloudspace,
		ServerClass: serverClass,
		Desired:     1,
	}
	createdOD, err := client.CreateOnDemandNodePool(ctx, onDemandPool)
	if err != nil {
		return fmt.Errorf("CreateOnDemandNodePool failed: %w", err)
	}
	fmt.Printf("Created OnDemandNodePool: %+v\n", createdOD)

	gotOD, err := client.GetOnDemandNodePool(ctx, namespace, onDemandPoolName)
	if err != nil {
		return fmt.Errorf("GetOnDemandNodePool failed: %w", err)
	}
	fmt.Printf("Got OnDemandNodePool: %+v\n", gotOD)

	onDemandPools, err := client.ListOnDemandNodePools(ctx, namespace)
	if err != nil {
		return fmt.Errorf("ListOnDemandNodePools failed: %w", err)
	}
	fmt.Printf("OnDemandNodePools: %+v\n", onDemandPools)
	return nil
}

// ShowRegionAndClassInfo prints region, server class, and price history info.
func ShowRegionAndClassInfo(ctx context.Context, client *v1.RackspaceSpotClient, region, serverClass string) {
	gotRegion, err := client.GetRegion(ctx, region)
	if err != nil {
		log.Printf("GetRegion failed: %v", err)
	} else {
		fmt.Printf("Got Region: %+v\n", gotRegion)
	}
	gotClass, err := client.GetServerClass(ctx, serverClass)
	if err != nil {
		log.Printf("GetServerClass failed: %v", err)
	} else {
		fmt.Printf("Got ServerClass: %+v\n", gotClass)
	}
	priceHistory, err := client.GetPriceHistory(ctx, serverClass)
	if err != nil {
		log.Printf("GetPriceHistory failed: %v", err)
	} else {
		fmt.Printf("PriceHistory: %+v\n", priceHistory)
	}
}

// CleanupResources deletes the created resources.
func CleanupResources(ctx context.Context, client *v1.RackspaceSpotClient, namespace, cloudspace, spotPool, onDemandPool string) {
	time.Sleep(2 * time.Second)
	if err := client.DeleteSpotNodePool(ctx, namespace, spotPool); err != nil {
		log.Printf("DeleteSpotNodePool failed: %v", err)
	} else {
		fmt.Println("Deleted SpotNodePool.")
	}
	time.Sleep(2 * time.Second)
	if err := client.DeleteOnDemandNodePool(ctx, namespace, onDemandPool); err != nil {
		log.Printf("DeleteOnDemandNodePool failed: %v", err)
	} else {
		fmt.Println("Deleted OnDemandNodePool.")
	}
	time.Sleep(2 * time.Second)
	if err := client.DeleteCloudspace(ctx, namespace, cloudspace); err != nil {
		log.Printf("DeleteCloudspace failed: %v", err)
	} else {
		fmt.Println("Deleted Cloudspace.")
	}
}
