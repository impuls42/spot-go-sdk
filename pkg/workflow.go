package internal

// // Authenticate authenticates the client with the Rackspace Spot API.
// func Authenticate(ctx context.Context, client *v1.RackspaceSpotClient) error {
// 	return client.Authenticate(ctx)
// }

// // SelectNamespace selects the first available namespace (organization).
// func SelectNamespace(ctx context.Context, client *v1.RackspaceSpotClient) (string, error) {
// 	orgs, err := client.ListOrganizations(ctx)
// 	if err != nil {
// 		return "", err
// 	}
// 	if len(orgs) == 0 {
// 		return "", fmt.Errorf("no organizations found")
// 	}
// 	fmt.Printf("Organizations: %+v\n", orgs)
// 	fmt.Printf("Using namespace: %s\n", orgs[0].Namespace)
// 	return orgs[0].Namespace, nil
// }

// // SelectRegion selects the first available region.
// func SelectRegion(ctx context.Context, client *v1.RackspaceSpotClient) (string, error) {
// 	regions, err := client.ListRegions(ctx)
// 	if err != nil {
// 		return "", err
// 	}
// 	if len(regions) == 0 {
// 		return "", fmt.Errorf("no regions found")
// 	}
// 	fmt.Printf("Regions: %+v\n", regions)
// 	return regions[0].Name, nil
// }

// // SelectServerClass selects the first available server class.
// func SelectServerClass(ctx context.Context, client *v1.RackspaceSpotClient) (string, error) {
// 	classes, err := client.ListServerClasses(ctx)
// 	if err != nil {
// 		return "", err
// 	}
// 	if len(classes) == 0 {
// 		return "", fmt.Errorf("no server classes found")
// 	}
// 	fmt.Printf("Server Classes: %+v\n", classes)
// 	return classes[0].Name, nil
// }
