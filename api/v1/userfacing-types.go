package rxtspot

import "time"

type CloudSpaceList struct {
	Items []CloudSpace `json:"cloudspaces"`
}

type CloudSpace struct {
	Name                 string                    `json:"name"`
	Org                  string                    `json:"org"`
	CreationTimestamp    time.Time                 `json:"creationTimestamp,omitempty"`
	Cni                  string                    `json:"cni,omitempty"`
	DeploymentType       string                    `json:"deploymentType,omitempty"`
	GpuEnabled           bool                      `json:"gpuEnabled,omitempty"`
	KubernetesVersion    string                    `json:"kubernetesVersion,omitempty"`
	Region               string                    `json:"region,omitempty"`
	PreemptionWebhookURL string                    `json:"preemptionWebhookURL,omitempty"`
	APIServerEndpoint    string                    `json:"apiserverEndpoint,omitempty"`
	AssignedServers      map[string]AssignedServer `json:"assignedServers,omitempty"`
	SpotNodepools        []*SpotNodePool           `json:"spotNodepools,omitempty"`
	OnDemandNodePools    []*OnDemandNodePool       `json:"ondemandnodepools,omitempty"`
	Health               string                    `json:"status,omitempty"`
}

type AssignedServer struct {
	IP              string `json:"IP"`
	ClusterRole     string `json:"clusterRole"`
	ServerClassName string `json:"serverClassName"`
	State           string `json:"state"`
}

type SpotNodePoolList struct {
	Items []SpotNodePool `json:"spotnodepools"`
}

type SpotNodePool struct {
	Name              string            `json:"name"`
	Org               string            `json:"org,omitempty"`
	Cloudspace        string            `json:"cloudspace,omitempty"`
	ServerClass       string            `json:"server_class,omitempty"`
	Desired           int               `json:"desired,omitempty"`
	CustomAnnotations map[string]string `json:"customAnnotations,omitempty"`
	CustomLabels      map[string]string `json:"customLabels,omitempty"`
	CustomTaints      map[string]string `json:"customTaints,omitempty"`
	Autoscaling       struct {
		Enabled  bool  `json:"enabled"`
		MinNodes int64 `json:"minNodes"`
		MaxNodes int64 `json:"maxNodes"`
	} `json:"autoscaling"`
	BidPrice string `json:"bid_price,omitempty"`
}

type OnDemandNodePoolList struct {
	Items []OnDemandNodePool `json:"ondemandnodepools"`
}

type OnDemandNodePool struct {
	Name              string            `json:"name"`
	Org               string            `json:"org,omitempty"`
	Cloudspace        string            `json:"cloudspace,omitempty"`
	ServerClass       string            `json:"server_class,omitempty"`
	Desired           int               `json:"desired,omitempty"`
	CustomAnnotations map[string]string `json:"customAnnotations,omitempty"`
	CustomLabels      map[string]string `json:"customLabels,omitempty"`
	CustomTaints      map[string]string `json:"customTaints,omitempty"`
	Autoscaling       struct {
		Enabled  bool `json:"enabled"`
		MinNodes int  `json:"minNodes"`
		MaxNodes int  `json:"maxNodes"`
	} `json:"autoscaling"`
}

type OrganizationList struct {
	Items []Organization `json:"organizations"`
}

type Organization struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type RegionList struct {
	Items []Region `json:"regions"`
}

type Region struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ServerClassList struct {
	Items []ServerClass `json:"serverclasses"`
}

type ServerClass struct {
	Name                      string   `json:"name"`
	Category                  string   `json:"category,omitempty"`
	Displayname               string   `json:"displayname,omitempty"`
	Region                    string   `json:"region,omitempty"`
	MinBidPricePerHour        string   `json:"min_bid_price_per_hour,omitempty"`
	CurrentMarketPricePerHour string   `json:"current_market_price_per_hour,omitempty"`
	Resources                 Resource `json:"resources,omitempty"`
}

type Resource struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	GPU    string `json:"gpu,omitempty"`
}

type PriceDetails struct {
	ServerClassName string `json:"serverclassname"`
	DisplayName     string `json:"display_name"`
	Category        string `json:"category"`
	Region          string `json:"region"`
	MarketPrice     string `json:"current_market_price"`
	CPU             string `json:"cpu"`
	Memory          string `json:"memory"`
}
