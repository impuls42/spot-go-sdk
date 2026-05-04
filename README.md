# Rackspace Spot Go SDK (In Development)

This package provides an idiomatic Go SDK for interacting with the Rackspace Spot platform. It enables developers and DevOps teams to programmatically manage cloud resources such as kubernetes cloudspaces (Kubernetes clusters), Virtual Machines in form of spot node pools, and on-demand node pools. Refer to [VM_TESTING_GUIDE.md](VM_TESTING_GUIDE.md) for instructions on specifically testing Virtual Machine (VM) functionality.

**Versioned API structure:**
- All types and client logic for API v1 are in `api/v1/` (import as `v1`).
- This structure is similar to AWS SDKs and supports future API versions (e.g., `api/v2/`).


## Installation

### 1. Install the SDK

Clone this repository and use Go modules to import the SDK in your project:

```sh
git clone https://github.com/rackspace-spot/spot-go-sdk
cd spot-go-sdk
```

Or add to your Go project:

```go
import v1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
```

### 2. Authentication

You need a Rackspace Spot refresh token. Use your refresh token to create spotClient

```go
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

```

### 3. Example Usage

See [`examples/main.go`](examples/main.go) for a full example of Kubernetes Cloudspaces and [`examples/test_vm_full.go`](examples/test_vm_full.go) for a full example of Virtual Machines.

Here is a minimal usage snippet:

```go

package main

import (
 "context"
 "fmt"
 "log"

 v1 "github.com/rackerlabs/spot-go-sdk/rxtspot/api/v1"
)

func main() {

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

 regions, err := spotClient.ListRegions(ctx)
 if err != nil {
  log.Fatalf("Failed to list regions: %v", err)
 }

 fmt.Println("Regions:")
 for _, region := range regions {
  fmt.Printf("- %s (%s)\n", region.Name, region.Name)
 }
}


```


