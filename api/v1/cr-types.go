package rxtspot

import "time"

// // Organization represents a Rackspace Spot organization (namespace).
// type Organization struct {
// 	Name      string `json:"name"`
// 	Namespace string `json:"namespace"`
// 	ID        string `json:"id"`
// }

// // Cloudspace represents a Kubernetes cluster in Rackspace Spot.
// type Cloudspace struct {
// 	Name              string `json:"name"`
// 	Namespace         string `json:"namespace"`
// 	Region            string `json:"region"`
// 	KubernetesVersion string `json:"kubernetes_version"`
// }

// // SpotNodePool represents a spot node pool in a cloudspace.
// type SpotNodePool struct {
// 	Name        string `json:"name"`
// 	Namespace   string `json:"namespace"`
// 	Cloudspace  string `json:"cloudspace"`
// 	ServerClass string `json:"server_class"`
// 	Desired     int    `json:"desired"`
// 	BidPrice    string `json:"bid_price"`
// }

// // OnDemandNodePool represents an on-demand node pool in a cloudspace.
// type OnDemandNodePool struct {
// 	Name        string `json:"name"`
// 	Namespace   string `json:"namespace"`
// 	Cloudspace  string `json:"cloudspace"`
// 	ServerClass string `json:"server_class"`
// 	Desired     int    `json:"desired"`
// }

// // Region represents a cloud region.
// type Region struct {
// 	Name        string `json:"name"`
// 	Description string `json:"description,omitempty"`
// }

// // ServerClassInfo represents a server class.
// type ServerClassInfo struct {
// 	Name        string `json:"name"`
// 	Description string `json:"description,omitempty"`
// }

// // PriceHistory represents the price history for a server class.
// type PriceHistory struct {
// 	History []PriceEntry `json:"history"`
// }

// // PriceEntry represents a single price point in the price history.
// type PriceEntry struct {
// 	Timestamp string  `json:"timestamp"`
// 	Price     float64 `json:"price"`
// }

type cloudSpaceGetResponse struct {
	Metadata struct {
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
	} `json:"metadata"`
	Spec struct {
		BidRequests       []string `json:"bidRequests"`
		Cloud             string   `json:"cloud"`
		Cni               string   `json:"cni"`
		DeploymentType    string   `json:"deploymentType"`
		GpuEnabled        bool     `json:"gpuEnabled"`
		KubernetesVersion string   `json:"kubernetesVersion"`
		Region            string   `json:"region"`
		Type              string   `json:"type"`
		Webhook           string   `json:"webhook"`
	} `json:"spec"`
	Status struct {
		APIServerEndpoint        string                    `json:"APIServerEndpoint"`
		AssignedServers          map[string]AssignedServer `json:"assignedServers"`
		Bids                     map[string]Bid            `json:"bids"`
		CloudspaceClassName      string                    `json:"cloudspaceClassName"`
		CurrentKubernetesVersion string                    `json:"currentKubernetesVersion"`
		FirstReadyTimestamp      time.Time                 `json:"firstReadyTimestamp"`
		Health                   string                    `json:"health"`
	} `json:"status"`
}

type Bid struct {
	BidName  string `json:"bidName"`
	Type     string `json:"type"`
	WonCount int    `json:"wonCount"`
}

type cloudSpaceListResponse struct {
	Items []cloudSpaceGetResponse `json:"items"`
}

type CloudSpaceCreateRequestBody struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name        string `json:"name"`
		Namespace   string `json:"namespace"`
		Annotations struct {
		} `json:"annotations"`
	} `json:"metadata"`
	Spec struct {
		DeploymentType    string `json:"deploymentType"`
		Cloud             string `json:"cloud"`
		Region            string `json:"region"`
		Webhook           string `json:"webhook"`
		Cni               string `json:"cni"`
		KubernetesVersion string `json:"kubernetesVersion"`
		HAControlPlane    bool   `json:"HAControlPlane"`
		GpuEnabled        bool   `json:"gpuEnabled"`
	} `json:"spec"`
}

type SpotNodePoolCreateRequestBody struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Labels    struct {
			NgpcRxtIoCloudspace string `json:"ngpc.rxt.io/cloudspace"`
		} `json:"labels"`
	} `json:"metadata"`
	Spec struct {
		ServerClass string `json:"serverClass"`
		Desired     int    `json:"desired"`
		BidPrice    string `json:"bidPrice"`
		CloudSpace  string `json:"cloudSpace"`
		Autoscaling struct {
			Enabled  bool  `json:"enabled"`
			MinNodes int64 `json:"minNodes"`
			MaxNodes int64 `json:"maxNodes"`
		} `json:"autoscaling"`
	} `json:"spec"`
}

type SpotNodePoolListResponse struct {
	APIVersion string `json:"apiVersion"`
	Items      []struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Name              string    `json:"name"`
			Namespace         string    `json:"namespace"`
		} `json:"metadata"`
		Spec struct {
			Autoscaling struct {
				Enabled  bool `json:"enabled"`
				MaxNodes int  `json:"maxNodes"`
				MinNodes int  `json:"minNodes"`
			} `json:"autoscaling"`
			BidPrice          string `json:"bidPrice"`
			CloudSpace        string `json:"cloudSpace"`
			CustomAnnotations struct {
			} `json:"customAnnotations"`
			CustomLabels struct {
			} `json:"customLabels"`
			CustomTaints []any  `json:"customTaints"`
			Desired      int    `json:"desired"`
			ServerClass  string `json:"serverClass"`
		} `json:"spec"`
		Status struct {
			BidStatus            string `json:"bidStatus"`
			CustomMetadataStatus struct {
			} `json:"customMetadataStatus"`
			WonCount int `json:"wonCount"`
		} `json:"status"`
	} `json:"items"`
	Kind     string `json:"kind"`
	Metadata struct {
		Continue        string `json:"continue"`
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
}
