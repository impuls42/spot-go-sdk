package rxtspot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func newTestClient(serverURL string) *RackspaceSpotClient {
	return &RackspaceSpotClient{
		BaseURL: serverURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Token:        "test-token",
		RefreshToken: "test-refresh",
		RetryConfig: RetryConfig{
			MaxRetries:      1,
			InitialInterval: 1 * time.Millisecond,
		},
	}
}

func orgListHandler(orgName, orgID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"organizations":[{"name":"%s","id":"%s"}]}`, orgName, orgID)
	}
}

func cloudspaceResponseJSON(name, namespace string) string {
	return fmt.Sprintf(`{
		"metadata": {"name": "%s", "namespace": "%s", "creationTimestamp": "2025-01-01T00:00:00Z"},
		"spec": {
			"bidRequests": [],
			"cloud": "default",
			"cni": "calico",
			"deploymentType": "gen2",
			"gpuEnabled": true,
			"HAControlPlane": true,
			"kubernetesVersion": "1.31",
			"region": "us-east-iad-1",
			"type": "",
			"webhook": "https://example.com/hook"
		},
		"status": {
			"APIServerEndpoint": "https://1.2.3.4",
			"assignedServers": {},
			"phase": "Ready",
			"reason": ""
		}
	}`, name, namespace)
}

func emptyNodePoolListJSON() string {
	return `{"apiVersion":"ngpc.rxt.io/v1","kind":"List","items":[],"metadata":{"continue":"","resourceVersion":"1"}}`
}

func TestUpdateCloudspace_NoMutableFields(t *testing.T) {
	client := newTestClient("")
	opts := CloudSpaceUpdateOptions{Name: "test-cs"}

	_, err := client.UpdateCloudspace(context.Background(), "test-org", opts)
	if err == nil {
		t.Fatal("expected error when no mutable fields are set, got nil")
	}
}

func TestUpdateCloudspace_InvalidOrg(t *testing.T) {
	client := newTestClient("")
	opts := CloudSpaceUpdateOptions{Name: "test-cs", KubernetesVersion: StringPtr("1.31")}

	_, err := client.UpdateCloudspace(context.Background(), "", opts)
	if err == nil {
		t.Fatal("expected error for invalid org, got nil")
	}
}

func TestUpdateCloudspace_PatchBodyShape(t *testing.T) {
	var receivedBody json.RawMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/apis/auth.ngpc.rxt.io/v1/organizations":
			orgListHandler("test-org", "org-123")(w, r)
			return
		case r.URL.Path == "/apis/ngpc.rxt.io/v1/namespaces/org-123/cloudspaces/test-cs" && r.Method == http.MethodPatch:
			if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, cloudspaceResponseJSON("test-cs", "org-123"))
			return
		case r.URL.Path == "/apis/ngpc.rxt.io/v1/namespaces/org-123/spotnodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyNodePoolListJSON())
			return
		case r.URL.Path == "/apis/ngpc.rxt.io/v1/namespaces/org-123/ondemandnodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyNodePoolListJSON())
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	gpuEnabled := true
	haControlPlane := true
	opts := CloudSpaceUpdateOptions{
		Name:              "test-cs",
		KubernetesVersion: StringPtr("1.31"),
		GpuEnabled:        &gpuEnabled,
		HAControlPlane:    &haControlPlane,
	}

	result, err := client.UpdateCloudspace(context.Background(), "test-org", opts)
	if err != nil {
		t.Fatalf("UpdateCloudspace returned error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(receivedBody, &parsed); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}

	spec, ok := parsed["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("request body missing 'spec' key")
	}

	if v := spec["kubernetesVersion"]; v != "1.31" {
		t.Errorf("expected kubernetesVersion=1.31 in patch, got %v", v)
	}
	if v := spec["gpuEnabled"]; v != true {
		t.Errorf("expected gpuEnabled=true in patch, got %v", v)
	}
	if v := spec["HAControlPlane"]; v != true {
		t.Errorf("expected HAControlPlane=true in patch, got %v", v)
	}

	if _, exists := spec["deploymentType"]; exists {
		t.Error("deploymentType should not be in patch body (immutable)")
	}
	if _, exists := spec["region"]; exists {
		t.Error("region should not be in patch body (immutable)")
	}
	if _, exists := spec["cloud"]; exists {
		t.Error("cloud should not be in patch body (immutable)")
	}

	if result == nil {
		t.Fatal("expected non-nil CloudSpace result")
	}
	if result.Name != "test-cs" {
		t.Errorf("expected Name=test-cs, got %s", result.Name)
	}
	if result.GpuEnabled != true {
		t.Error("expected GpuEnabled=true in response")
	}
	if result.HAControlPlane != true {
		t.Error("expected HAControlPlane=true in response")
	}
}

func TestUpdateCloudspace_OnlyWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/apis/auth.ngpc.rxt.io/v1/organizations":
			orgListHandler("test-org", "org-123")(w, r)
			return
		case r.URL.Path == "/apis/ngpc.rxt.io/v1/namespaces/org-123/cloudspaces/test-cs" && r.Method == http.MethodPatch:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)

			spec, ok := body["spec"].(map[string]interface{})
			if ok {
				if _, exists := spec["gpuEnabled"]; exists {
					t.Error("gpuEnabled should be omitted when not set (nil)")
				}
				if _, exists := spec["HAControlPlane"]; exists {
					t.Error("HAControlPlane should be omitted when not set (nil)")
				}
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, cloudspaceResponseJSON("test-cs", "org-123"))
			return
		case r.URL.Path == "/apis/ngpc.rxt.io/v1/namespaces/org-123/spotnodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyNodePoolListJSON())
			return
		case r.URL.Path == "/apis/ngpc.rxt.io/v1/namespaces/org-123/ondemandnodepools":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, emptyNodePoolListJSON())
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	opts := CloudSpaceUpdateOptions{
		Name:                 "test-cs",
		PreemptionWebhookURL: StringPtr("https://example.com/hook"),
	}

	result, err := client.UpdateCloudspace(context.Background(), "test-org", opts)
	if err != nil {
		t.Fatalf("UpdateCloudspace returned error: %v", err)
	}
	if result.PreemptionWebhookURL != "https://example.com/hook" {
		t.Errorf("expected webhook in response, got %s", result.PreemptionWebhookURL)
	}
}

func TestBoolPtr(t *testing.T) {
	val := BoolPtr(true)
	if *val != true {
		t.Error("expected true")
	}
	val2 := BoolPtr(false)
	if *val2 != false {
		t.Error("expected false")
	}
}

func TestStringPtr(t *testing.T) {
	p := StringPtr("hello")
	if p == nil || *p != "hello" {
		t.Error("expected non-nil pointer to 'hello'")
	}
}

func TestCloudSpaceFromResponse(t *testing.T) {
	resp := &cloudSpaceGetResponse{
		Metadata: struct {
			Name              string    `json:"name"`
			Namespace         string    `json:"namespace"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
		}{Name: "my-cs", Namespace: "org-123"},
		Spec: struct {
			BidRequests       []string `json:"bidRequests"`
			Cloud             string   `json:"cloud"`
			CNI               string   `json:"cni"`
			DeploymentType    string   `json:"deploymentType"`
			GpuEnabled        bool     `json:"gpuEnabled"`
			HAControlPlane    bool     `json:"HAControlPlane"`
			KubernetesVersion string   `json:"kubernetesVersion"`
			Region            string   `json:"region"`
			Type              string   `json:"type"`
			Webhook           string   `json:"webhook"`
		}{
			GpuEnabled:        true,
			HAControlPlane:    true,
			KubernetesVersion: "1.31",
			CNI:               "calico",
			Region:            "us-east-iad-1",
		},
		Status: struct {
			APIServerEndpoint        string                    `json:"APIServerEndpoint"`
			AssignedServers          map[string]AssignedServer `json:"assignedServers"`
			Bids                     map[string]Bid            `json:"bids"`
			CloudspaceClassName      string                    `json:"cloudspaceClassName"`
			CurrentKubernetesVersion string                    `json:"currentKubernetesVersion"`
			FirstReadyTimestamp      time.Time                 `json:"firstReadyTimestamp"`
			Health                   string                    `json:"health"`
			Reason                   string                    `json:"reason"`
			Phase                    string                    `json:"phase"`
		}{
			APIServerEndpoint: "https://1.2.3.4",
			Phase:             "Ready",
		},
	}

	cs := cloudSpaceFromResponse("test-org", resp, nil, nil)

	if cs.Name != "my-cs" {
		t.Errorf("expected Name=my-cs, got %s", cs.Name)
	}
	if !cs.GpuEnabled {
		t.Error("expected GpuEnabled=true")
	}
	if !cs.HAControlPlane {
		t.Error("expected HAControlPlane=true")
	}
	if cs.KubernetesVersion != "1.31" {
		t.Errorf("expected KubernetesVersion=1.31, got %s", cs.KubernetesVersion)
	}
}

func TestUpdateSpotNodePool_AutoscalingNil(t *testing.T) {
	var receivedBody json.RawMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/apis/auth.ngpc.rxt.io/v1/organizations":
			orgListHandler("test-org", "org-123")(w, r)
			return
		default:
			json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{}`)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	opts := SpotNodePoolUpdateOptions{
		Name:    "test-pool",
		Desired: IntPtr(3),
	}

	err := client.UpdateSpotNodePool(context.Background(), "test-org", opts)
	if err != nil {
		t.Fatalf("UpdateSpotNodePool returned error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(receivedBody, &parsed)

	spec := parsed["spec"].(map[string]interface{})
	if _, exists := spec["autoscaling"]; exists {
		t.Error("autoscaling should be omitted when nil (no clobber)")
	}
}

func TestUpdateSpotNodePool_AutoscalingSet(t *testing.T) {
	var receivedBody json.RawMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/apis/auth.ngpc.rxt.io/v1/organizations":
			orgListHandler("test-org", "org-123")(w, r)
			return
		default:
			json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{}`)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	opts := SpotNodePoolUpdateOptions{
		Name:    "test-pool",
		Desired: IntPtr(3),
		Autoscaling: &Autoscaling{
			Enabled:  true,
			MinNodes: int64(0),
			MaxNodes: int64(10),
		},
	}

	err := client.UpdateSpotNodePool(context.Background(), "test-org", opts)
	if err != nil {
		t.Fatalf("UpdateSpotNodePool returned error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(receivedBody, &parsed)

	spec := parsed["spec"].(map[string]interface{})
	autoscaling, ok := spec["autoscaling"].(map[string]interface{})
	if !ok {
		t.Fatal("expected autoscaling to be present in spec")
	}
	if autoscaling["enabled"] != true {
		t.Error("expected autoscaling.enabled=true")
	}
	if v, ok := autoscaling["minNodes"].(float64); !ok || v != 0 {
		t.Error("expected autoscaling.minNodes=0 (zero should be preserved)")
	}
}

func TestUpdateOnDemandNodePool_AutoscalingNil(t *testing.T) {
	var receivedBody json.RawMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/apis/auth.ngpc.rxt.io/v1/organizations":
			orgListHandler("test-org", "org-123")(w, r)
			return
		default:
			json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{}`)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	opts := OnDemandNodePoolUpdateOptions{
		Name:    "test-pool",
		Desired: IntPtr(2),
	}

	err := client.UpdateOnDemandNodePool(context.Background(), "test-org", opts)
	if err != nil {
		t.Fatalf("UpdateOnDemandNodePool returned error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(receivedBody, &parsed)

	spec := parsed["spec"].(map[string]interface{})
	if _, exists := spec["autoscaling"]; exists {
		t.Error("autoscaling should be omitted when nil (no clobber)")
	}
}

func TestAutoscalingPointerFields(t *testing.T) {
	enabled := true
	min := int64(1)
	max := int64(5)
	a := &Autoscaling{Enabled: enabled, MinNodes: min, MaxNodes: max}

	if !a.Enabled {
		t.Error("expected Enabled=true")
	}
	if a.MinNodes != 1 {
		t.Errorf("expected MinNodes=1, got %d", a.MinNodes)
	}
	if a.MaxNodes != 5 {
		t.Errorf("expected MaxNodes=5, got %d", a.MaxNodes)
	}
}

func TestAutoscalingTypeConsistency(t *testing.T) {
	typ := reflect.TypeOf(Autoscaling{})

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		switch f.Name {
		case "Enabled":
			if f.Type.Kind() != reflect.Bool {
				t.Errorf("Autoscaling.%s: expected bool, got %v", f.Name, f.Type)
			}
		case "MinNodes", "MaxNodes":
			if f.Type.Kind() != reflect.Int64 {
				t.Errorf("Autoscaling.%s: expected int64, got %v", f.Name, f.Type)
			}
		}
	}
}

func TestCreateSpotNodePool_NilAutoscaling(t *testing.T) {
	client := newTestClient("")
	pool := SpotNodePool{
		Name:        "test-pool",
		Org:         "test-org",
		Cloudspace:  "test-cs",
		ServerClass: "test-sc",
		Desired:     1,
		BidPrice:    "$0.08",
		Autoscaling: nil,
	}
	err := client.CreateSpotNodePool(context.Background(), "test-org", pool)
	if err == nil {
		t.Fatal("expected error when Autoscaling is nil, got nil")
	}
	if err.Error() != "autoscaling configuration is required for spot node pool creation" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCreateSpotNodePool_ZeroDesired(t *testing.T) {
	client := newTestClient("")
	pool := SpotNodePool{
		Name:        "test-pool",
		Org:         "test-org",
		Cloudspace:  "test-cs",
		ServerClass: "test-sc",
		Desired:     0,
		BidPrice:    "$0.08",
		Autoscaling: &Autoscaling{Enabled: true, MinNodes: 1, MaxNodes: 5},
	}
	err := client.CreateSpotNodePool(context.Background(), "test-org", pool)
	if err == nil {
		t.Fatal("expected error when Desired is 0, got nil")
	}
	if err.Error() != "desired count must be positive for spot node pool creation" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCreateOnDemandNodePool_NilAutoscaling(t *testing.T) {
	client := newTestClient("")
	pool := OnDemandNodePool{
		Name:        "test-pool",
		Org:         "test-org",
		Cloudspace:  "test-cs",
		ServerClass: "test-sc",
		Desired:     1,
		Autoscaling: nil,
	}
	err := client.CreateOnDemandNodePool(context.Background(), "test-org", pool)
	if err == nil {
		t.Fatal("expected error when Autoscaling is nil, got nil")
	}
	if err.Error() != "autoscaling configuration is required for on-demand node pool creation" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCreateOnDemandNodePool_ZeroDesired(t *testing.T) {
	client := newTestClient("")
	pool := OnDemandNodePool{
		Name:        "test-pool",
		Org:         "test-org",
		Cloudspace:  "test-cs",
		ServerClass: "test-sc",
		Desired:     0,
		Autoscaling: &Autoscaling{Enabled: true, MinNodes: 1, MaxNodes: 5},
	}
	err := client.CreateOnDemandNodePool(context.Background(), "test-org", pool)
	if err == nil {
		t.Fatal("expected error when Desired is 0, got nil")
	}
	if err.Error() != "desired count must be positive for on-demand node pool creation" {
		t.Errorf("unexpected error message: %v", err)
	}
}
