package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	v1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spot-go-sdk/api/v1/mocks"

	// "github.com/rackspace-spot/spot-go-sdk/api/v1/mocks" // Import your mocks package
	"go.uber.org/mock/gomock"
)

// TestListRegions tests the ListRegions function with a mocked SpotClient.
func TestListRegions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock ListRegions method
	mockSpotAPI.EXPECT().ListRegions(gomock.Any()).Return([]v1.Region{
		{Name: "us-east-1", Description: "US East"},
		{Name: "us-west-1", Description: "US West"},
	}, nil)

	// Call the listRegions function (assuming you have a function that uses the SpotClient interface)
	regions, err := mockSpotAPI.ListRegions(context.Background())
	if err != nil {
		t.Fatalf("Failed to list regions: %v", err)
	}

	// Check the result
	if len(regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(regions))
	}
	if regions[0].Name != "us-east-1" {
		t.Errorf("Expected region name us-east-1, got %s", regions[0].Name)
	}
	if regions[1].Name != "us-west-1" {
		t.Errorf("Expected region name us-west-1, got %s", regions[1].Name)
	}
}

// TestCreateCloudspace tests the CreateCloudspace function with a mocked SpotClient.
func TestCreateCloudspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock CreateCloudspace method
	mockSpotAPI.EXPECT().CreateCloudspace(gomock.Any(), gomock.Any()).Return(nil)

	spotNodePool := v1.SpotNodePool{
		Name:        "test-sdk-spot-nodepool",
		Org:         "test-sdk-org",
		Cloudspace:  "test-sdk-cloudspace",
		ServerClass: "ch.vs1.large-dfw",
		Desired:     1,
		CustomAnnotations: map[string]string{
			"example.com/annotation": "value",
		},
		CustomLabels: map[string]string{
			"example.com/label": "value",
		},
		BidPrice: "$0.08",
	}

	// Call the CreateCloudspace function
	err := mockSpotAPI.CreateCloudspace(context.Background(), v1.CloudSpace{
		Name:              "test-sdk-cloudspace",
		Org:               "test-sdk-org",
		KubernetesVersion: "1.31.1",
		CNI:               "calico",
		Region:            "us-east-iad-1",
		SpotNodepools: []*v1.SpotNodePool{
			&spotNodePool,
		},
	})

	if err != nil {
		t.Fatalf("Failed to create cloudspace: %v", err)
	}
}

// TestGetCloudspace tests the GetCloudspace function with a mocked SpotClient.
func TestGetCloudspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock GetCloudspace method
	mockSpotAPI.EXPECT().GetCloudspace(gomock.Any(), "hooli", "sdk-cloudspace").Return(&v1.CloudSpace{
		Name:              "sdk-cloudspace",
		Org:               "hooli",
		KubernetesVersion: "1.31.1",
		CNI:               "calico",
		Region:            "us-east-iad-1",
	}, nil)

	// Call the GetCloudspace function
	cloudspace, err := mockSpotAPI.GetCloudspace(context.Background(), "hooli", "sdk-cloudspace")
	if err != nil {
		t.Fatalf("Failed to get cloudspace: %v", err)
	}

	// Check the result
	if cloudspace.Name != "sdk-cloudspace" {
		t.Errorf("Expected cloudspace name sdk-cloudspace, got %s", cloudspace.Name)
	}
	if cloudspace.Org != "hooli" {
		t.Errorf("Expected cloudspace org hooli, got %s", cloudspace.Org)
	}
}

// TestAuthentication tests the Authentication function with a mocked SpotClient.
func TestAuthentication(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock Authenticate method
	mockSpotAPI.EXPECT().Authenticate(gomock.Any()).Return("test_access_token", nil)

	// Call the Authenticate function
	token, err := mockSpotAPI.Authenticate(context.Background())
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	// Check the token
	if token != "test_access_token" {
		t.Errorf("Expected token test_access_token, got %s", token)
	}
}

// TestListRegions_Error tests the ListRegions function with a mocked SpotClient that returns an error.
func TestListRegions_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock ListRegions method to return an error
	mockSpotAPI.EXPECT().ListRegions(gomock.Any()).Return(nil, fmt.Errorf("Internal Server Error"))

	// Call the listRegions function
	_, err := mockSpotAPI.ListRegions(context.Background())
	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	if !strings.Contains(err.Error(), "Internal Server Error") {
		t.Errorf("Expected error message to contain 'Internal Server Error', got '%s'", err.Error())
	}
}

// TestCreateCloudspace_Error tests the CreateCloudspace function with a mocked SpotClient that returns an error.
func TestCreateCloudspace_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock CreateCloudspace method to return an error
	mockSpotAPI.EXPECT().CreateCloudspace(gomock.Any(), gomock.Any()).Return(fmt.Errorf("Internal Server Error"))

	spotNodePool := v1.SpotNodePool{
		Name:        "sdk-spot-nodepool",
		Org:         "hooli",
		Cloudspace:  "sdk-cloudspace",
		ServerClass: "ch.vs1.large-dfw",
		Desired:     1,
		CustomAnnotations: map[string]string{
			"example.com/annotation": "value",
		},
		CustomLabels: map[string]string{
			"example.com/label": "value",
		},
		BidPrice: "$0.08",
	}

	// Call the CreateCloudspace function
	err := mockSpotAPI.CreateCloudspace(context.Background(), v1.CloudSpace{
		Name:              "sdk-cloudspace",
		Org:               "hooli",
		KubernetesVersion: "1.31.1",
		CNI:               "calico",
		Region:            "us-east-iad-1",
		SpotNodepools: []*v1.SpotNodePool{
			&spotNodePool,
		},
	})

	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	if !strings.Contains(err.Error(), "Internal Server Error") {
		t.Errorf("Expected error message to contain 'Internal Server Error', got '%s'", err.Error())
	}
}

// TestGetCloudspace_Error tests the GetCloudspace function with a mocked SpotClient that returns an error.
func TestGetCloudspace_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock GetCloudspace method to return an error
	mockSpotAPI.EXPECT().GetCloudspace(gomock.Any(), "hooli", "sdk-cloudspace").Return(nil, fmt.Errorf("Internal Server Error"))

	// Call the GetCloudspace function
	_, err := mockSpotAPI.GetCloudspace(context.Background(), "hooli", "sdk-cloudspace")
	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	if !strings.Contains(err.Error(), "Internal Server Error") {
		t.Errorf("Expected error message to contain 'Internal Server Error', got '%s'", err.Error())
	}
}

// TestAuthentication_Error tests the Authentication function with a mocked SpotClient that returns an error.
func TestAuthentication_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock Authenticate method to return an error
	mockSpotAPI.EXPECT().Authenticate(gomock.Any()).Return("", fmt.Errorf("Internal Server Error"))

	// Call the Authenticate function
	_, err := mockSpotAPI.Authenticate(context.Background())
	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	if !strings.Contains(err.Error(), "Internal Server Error") {
		t.Errorf("Expected error message to contain 'Internal Server Error', got '%s'", err.Error())
	}
}

// TestListRegions_EmptyResponse tests the ListRegions function with a mocked SpotClient that returns an empty response.
func TestListRegions_EmptyResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock ListRegions method to return an empty response
	mockSpotAPI.EXPECT().ListRegions(gomock.Any()).Return([]v1.Region{}, nil)

	// Call the listRegions function
	regions, err := mockSpotAPI.ListRegions(context.Background())
	if err != nil {
		t.Fatalf("Failed to list regions: %v", err)
	}

	// Check the result
	if len(regions) != 0 {
		t.Errorf("Expected 0 regions, got %d", len(regions))
	}
}

// TestGetCloudspace_NotFound tests the GetCloudspace function with a mocked SpotClient that returns a 404 Not Found error.
func TestGetCloudspace_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock GetCloudspace method to return an error
	mockSpotAPI.EXPECT().GetCloudspace(gomock.Any(), "hooli", "sdk-cloudspace").Return(nil, fmt.Errorf("Not Found"))

	// Call the GetCloudspace function
	_, err := mockSpotAPI.GetCloudspace(context.Background(), "hooli", "sdk-cloudspace")
	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	if !strings.Contains(err.Error(), "Not Found") {
		t.Errorf("Expected error message to contain 'Not Found', got '%s'", err.Error())
	}
}

// TestCreateCloudspace_InvalidInput tests the CreateCloudspace function with a mocked SpotClient that returns a 400 Bad Request error.
func TestCreateCloudspace_InvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock CreateCloudspace method to return an error
	mockSpotAPI.EXPECT().CreateCloudspace(gomock.Any(), gomock.Any()).Return(fmt.Errorf("Bad Request"))

	// Call the CreateCloudspace function with invalid input (e.g., empty name)
	err := mockSpotAPI.CreateCloudspace(context.Background(), v1.CloudSpace{})
	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	if !strings.Contains(err.Error(), "Bad Request") {
		t.Errorf("Expected error message to contain 'Bad Request', got '%s'", err.Error())
	}
}

// TestAuthentication_InvalidCredentials tests the Authentication function with a mocked SpotClient that returns a 401 Unauthorized error.
func TestAuthentication_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotAPI := mocks.NewMockSpotAPI(ctrl)

	// Define the behavior of the mock Authenticate method to return an error
	mockSpotAPI.EXPECT().Authenticate(gomock.Any()).Return("", fmt.Errorf("Unauthorized"))

	// Call the Authenticate function
	_, err := mockSpotAPI.Authenticate(context.Background())
	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	if !strings.Contains(err.Error(), "Unauthorized") {
		t.Errorf("Expected error message to contain 'Unauthorized', got '%s'", err.Error())
	}
}
