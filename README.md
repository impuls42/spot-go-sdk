# Rackspace Spot Go SDK (In Development)

This package provides an idiomatic Go SDK for interacting with the Rackspace Spot platform. It enables developers and DevOps teams to programmatically manage cloud resources such as cloudspaces (Kubernetes clusters), spot node pools, and on-demand node pools.

**Versioned API structure:**
- All types and client logic for API v1 are in `api/v1/` (import as `v1`).
- This structure is similar to AWS SDKs and supports future API versions (e.g., `api/v2/`).

## Features (Planned)
- Authenticate with Rackspace Spot using OAuth2 refresh tokens
- Create, list, and delete cloudspaces
- Manage spot and on-demand node pools
- Query available regions, server classes, and price history
- Example CLI for resource management
- Comprehensive documentation and usage examples

## Roadmap
1. Core SDK: Authentication, cloudspace management
2. Node pool management (spot/on-demand)
3. Utility methods (regions, server classes, price history)
4. Example CLI tool
5. Tests and documentation


## Installation

### 1. Install the SDK

Clone this repository and use Go modules to import the SDK in your project:

```sh
git clone https://github.com/rackerlabs/spot-go-sdk.git
cd spot-go-sdk/rxtspot
```

Or add to your Go project:

```go
import v1 "github.com/rackerlabs/spot-go-sdk/rxtspot/api/v1"
```

### 2. Authentication

You need a Rackspace Spot refresh token. Set it as an environment variable:

```sh
export SPOT_REFRESH_TOKEN=your_refresh_token_here
```

### 3. Example Usage

See [`examples/main.go`](examples/main.go) for a full example. Here is a minimal usage snippet:

```go
package main

import (
    "context"
    "fmt"
    "os"
    v1 "github.com/rackerlabs/spot-go-sdk/rxtspot/api/v1"
)

func main() {
    refreshToken := os.Getenv("SPOT_REFRESH_TOKEN")
    client := v1.NewClient(refreshToken)
    if err := client.Authenticate(context.Background()); err != nil {
        panic(err)
    }
    orgs, err := client.ListOrganizations(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Println("Organizations:", orgs)
}
```

### 4. Run the Example

```sh
cd examples
export SPOT_REFRESH_TOKEN=your_refresh_token_here
go run main.go
```

This will demonstrate authentication and CRUD operations for all major objects.

---

_See the SDK source and examples for more advanced usage and integration._ 