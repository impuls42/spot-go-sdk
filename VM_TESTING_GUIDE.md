# Rackspace Spot Go SDK for Virtual Machines (VM)

This guide explains how to test the Virtual Machines (VM) functionality in the Rackspace Spot spot-go-sdk.

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

### 2. Credentials for authentication
- **RefreshToken**: `<YOUR_REFRESH_TOKEN_HERE>`
- **Organization**: `spot-org` (update with you organization name)

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

## Testing Virtual Machines (VM) Functionality

### Option 1: Quick Simple Test
This test lists existing resources without creating anything:

```bash
cd examples
go run test_vm_simple.go
```

**What it tests:**
- Authentication
- List Regions
- List VM SSH Keys
- List VM CloudSpaces
- Get VM CloudSpace details (if any exist)

### Option 2: Comprehensive Test
This test creates, updates, and optionally deletes VM resources:

```bash
cd examples
go run test_vm_full.go
```

**What it tests:**
- All CRUD operations for VM SSH Keys
- All CRUD operations for VM CloudSpaces
- All CRUD operations for VM Pools
- Integration between components

## VM API Endpoints Tested

### VM SSH Keys
- ✅ `ListVMSSHKeys(ctx, org)` - List all VM SSH keys
- ✅ `CreateVMSSHKey(ctx, key)` - Create a new VM SSH key
- ✅ `GetVMSSHKey(ctx, org, name)` - Get specific VM SSH key
- ✅ `DeleteVMSSHKey(ctx, org, name)` - Delete a VM SSH key

### VM CloudSpaces
- ✅ `ListVMCloudSpaces(ctx, org)` - List all VM CloudSpaces
- ✅ `CreateVMCloudSpace(ctx, vmcs)` - Create a new VM CloudSpace
- ✅ `GetVMCloudSpace(ctx, org, name)` - Get specific VM CloudSpace
- ✅ `UpdateVMCloudSpace(ctx, org, vmcs)` - Update specific VM CloudSpace
- ✅ `DeleteVMCloudSpace(ctx, org, name)` - Delete a VM CloudSpace

### VM Pools
- ✅ `ListVMPools(ctx, org, vmCloudSpace)` - List VM Pools in a CloudSpace
- ✅ `CreateVMPool(ctx, org, pool)` - Create a new VM Pool
- ✅ `GetVMPool(ctx, org, name)` - Get specific VM Pool
- ✅ `UpdateVMPool(ctx, org, pool)` - Update VM Pool (desired count, bid price)
- ✅ `DeleteVMPool(ctx, org, name)` - Delete a VM Pool

## Usage in Your Own Code

```go
package main

import (
    "context"
    "log"
    
    v1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

func main() {
    // Initialize client with testbed
    client, err := v1.NewSpotClient(&v1.Config{
        RefreshToken: "<YOUR_REFRESH_TOKEN_HERE>",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Authenticate
    ctx := context.Background()
    _, err = client.Authenticate(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create VM SSH Key
    err = client.CreateVMSSHKey(ctx, v1.VMSSHKey{
        Name:        "my-ssh-key",
        Org:         "default",
        PublicKey:   "ssh-rsa AAAA...",
        Description: "My SSH key",
    })
    
    // Create VM CloudSpace
    err = client.CreateVMCloudSpace(ctx, v1.VMCloudSpace{
        Name:         "my-vm-cloudspace",
        Org:          "default",
        Region:       "us-west-sjc-1",
        VMSSHKeyName: "my-ssh-key",
    })
    
    // Create VM Pool
    err = client.CreateVMPool(ctx, "default", v1.VMPool{
        Name:         "my-vm-pool",
        VMCloudSpace: "my-vm-cloudspace",
        ServerClass:  "gp.vs2.medium-sjc",
        Desired:      2,
        BidPrice:     "0.55",
        PoolType:     "spot",
        VMUserData:   "#!/bin/bash\necho 'Hello World' > /tmp/hello.txt",
    })
    
    // List VM CloudSpaces
    vmcs, err := client.ListVMCloudSpaces(ctx, "default")
    if err != nil {
        log.Fatal(err)
    }
    
    for _, vmCloudSpace := range vmcs.Items {
        log.Printf("VM CloudSpace: %s (Status: %s)", 
            vmCloudSpace.Name, vmCloudSpace.Status)
        
        // Access VM Pools
        for _, pool := range vmCloudSpace.VMPools {
            log.Printf("  - Pool: %s (Desired: %d, Won: %d)", 
                pool.Name, pool.Desired, pool.WonCount)
        }
    }
}
```

## Troubleshooting

### Authentication Errors
- Verify the refresh token is correct

### Resource Creation Failures
- Check if organization exists
- Verify SSH key exists before creating VM CloudSpace
- Ensure region and server class are valid
- Check if resource already exists

### API Errors
- Check HTTP status codes in error messages
- Verify request body format matches API spec

## File Structure

```
spot-go-sdk/
├── api/v1/
│   ├── client.go           # Main client and config
│   ├── auth.go             # Authentication
│   ├── vmcloudspaces.go    # VM CloudSpace operations
│   ├── vmpools.go          # VM Pool operations
│   ├── vmsshkeys.go        # VM SSH Key operations
│   ├── types.go            # Public types
│   └── backend-types.go    # Internal API types
├── examples/
│   ├── vm_simple_test.go   # Quick read-only test
│   └── vm_test.go          # Comprehensive CRUD test
└── go.mod                  # Dependencies
```

## Next Steps

1. Run the simple test to verify connectivity
2. Run comprehensive test to create/update/delete resources
3. Integrate SDK into your application

## Notes

- Clean up test resources when done
- VM CloudSpaces require VM SSH Keys to be created first
- VM Pools require VM CloudSpaces to be created first
