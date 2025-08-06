package rxtspot

// Organization represents a Rackspace Spot organization (namespace).
type Organization struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	ID        string `json:"id"`
}

// Cloudspace represents a Kubernetes cluster in Rackspace Spot.
type Cloudspace struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	Region            string `json:"region"`
	KubernetesVersion string `json:"kubernetes_version"`
}

// SpotNodePool represents a spot node pool in a cloudspace.
type SpotNodePool struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Cloudspace  string `json:"cloudspace"`
	ServerClass string `json:"server_class"`
	Desired     int    `json:"desired"`
	BidPrice    string `json:"bid_price"`
}

// OnDemandNodePool represents an on-demand node pool in a cloudspace.
type OnDemandNodePool struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Cloudspace  string `json:"cloudspace"`
	ServerClass string `json:"server_class"`
	Desired     int    `json:"desired"`
}

// Region represents a cloud region.
type Region struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ServerClassInfo represents a server class.
type ServerClassInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// PriceHistory represents the price history for a server class.
type PriceHistory struct {
	History []PriceEntry `json:"history"`
}

// PriceEntry represents a single price point in the price history.
type PriceEntry struct {
	Timestamp string  `json:"timestamp"`
	Price     float64 `json:"price"`
}
