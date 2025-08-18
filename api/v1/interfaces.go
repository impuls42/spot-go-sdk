package rxtspot

import "context"

// OrganizationAPI defines organization-related methods.
type OrganizationAPI interface {
	ListOrganizations(ctx context.Context) ([]Organization, error)
}

// CloudspaceAPI defines cloudspace-related methods.
type CloudspaceAPI interface {
	ListCloudspaces(ctx context.Context, org string) (*CloudSpaceList, error)
	CreateCloudspace(ctx context.Context, cs CloudSpace) error
	GetCloudspace(ctx context.Context, org, name string) (*CloudSpace, error)
	DeleteCloudspace(ctx context.Context, org, name string) error
	GetCloudspaceConfig(ctx context.Context, org, name string) (string, error)
}

// SpotNodePoolAPI defines spot node pool methods.
type SpotNodePoolAPI interface {
	ListSpotNodePools(ctx context.Context, org string, cloudspace string) ([]*SpotNodePool, error)
	CreateSpotNodePool(ctx context.Context, org string, pool SpotNodePool) error
	GetSpotNodePool(ctx context.Context, org, name string) (*SpotNodePool, error)
	DeleteSpotNodePool(ctx context.Context, org, name string) error
}

// OnDemandNodePoolAPI defines on-demand node pool methods.
type OnDemandNodePoolAPI interface {
	ListOnDemandNodePools(ctx context.Context, org string, cloudspace string) ([]*OnDemandNodePool, error)
	CreateOnDemandNodePool(ctx context.Context, org string, pool OnDemandNodePool) error
	GetOnDemandNodePool(ctx context.Context, org, name string) (*OnDemandNodePool, error)
	DeleteOnDemandNodePool(ctx context.Context, org, name string) error
}

// UtilityAPI defines utility methods for regions, server classes, price history, etc.
type UtilityAPI interface {
	ListRegions(ctx context.Context) ([]Region, error)
	GetRegion(ctx context.Context, name string) (*Region, error)
	ListServerClasses(ctx context.Context) (*ServerClassList, error)
	GetServerClass(ctx context.Context, name string) (*ServerClass, error)
	GetPriceHistory(ctx context.Context, serverClass string) (*PriceHistory, error)
}

// SpotAPI defines the complete interface for the Rackspace Spot SDK client.
// It embeds all the specific APIs to provide a unified interface.
type SpotAPI interface {
	Authenticate(ctx context.Context) error
	OrganizationAPI
	CloudspaceAPI
	SpotNodePoolAPI
	OnDemandNodePoolAPI
	UtilityAPI
}
