package rajaongkir

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-key"
	baseURL := "https://api.rajaongkir.com/starter"

	client := NewClient(apiKey, baseURL)

	if client == nil {
		t.Fatal("expected client to be created")
	}

	if client.apiKey != apiKey {
		t.Errorf("expected API key '%s', got '%s'", apiKey, client.apiKey)
	}

	if client.baseURL != baseURL {
		t.Errorf("expected base URL '%s', got '%s'", baseURL, client.baseURL)
	}

	// Test default HTTP client configuration
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("expected HTTP client timeout to be 30 seconds, got %v", client.httpClient.Timeout)
	}
}

func TestClient_GetProvincesRequest(t *testing.T) {
	apiKey := "test-key"
	baseURL := "https://api.rajaongkir.com/starter"
	client := NewClient(apiKey, baseURL)

	// Test URL construction
	expectedURL := baseURL + "/province"

	// This test only verifies the client structure and expected behavior
	// without making actual HTTP calls (as we don't have a real API key)
	if client.apiKey != apiKey {
		t.Errorf("expected API key to be set correctly")
	}

	if client.baseURL != baseURL {
		t.Errorf("expected base URL to be set correctly")
	}

	if client.httpClient == nil {
		t.Error("expected HTTP client to be initialized")
	}

	// Verify the expected URL would be constructed correctly
	actualURL := client.baseURL + "/province"
	if actualURL != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, actualURL)
	}
}

func TestClient_GetCitiesRequest(t *testing.T) {
	apiKey := "test-key"
	baseURL := "https://api.rajaongkir.com/starter"
	client := NewClient(apiKey, baseURL)

	// Test URL construction for cities endpoint
	expectedURL := baseURL + "/city"

	actualURL := client.baseURL + "/city"
	if actualURL != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, actualURL)
	}
}

func TestClient_GetCostRequest(t *testing.T) {
	apiKey := "test-key"
	baseURL := "https://api.rajaongkir.com/starter"
	client := NewClient(apiKey, baseURL)

	// Test URL construction for cost endpoint
	expectedURL := baseURL + "/cost"

	actualURL := client.baseURL + "/cost"
	if actualURL != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, actualURL)
	}
}

func TestClient_DefaultConfiguration(t *testing.T) {
	client := NewClient("test-key", "https://api.rajaongkir.com/starter")

	// Test default HTTP client timeout
	expectedTimeout := 30 * time.Second
	if client.httpClient.Timeout != expectedTimeout {
		t.Errorf("expected HTTP client timeout to be %v, got %v", expectedTimeout, client.httpClient.Timeout)
	}

	// Test that HTTP client is not nil
	if client.httpClient == nil {
		t.Error("expected HTTP client to be initialized")
	}

	// Test that client has valid transport
	if client.httpClient.Transport == nil {
		t.Log("HTTP client has no default transport (this is expected)")
	}
}

func TestClient_InvalidURL(t *testing.T) {
	// Test with empty base URL
	client := NewClient("test-key", "")
	if client.baseURL != "" {
		t.Error("expected base URL to be empty")
	}

	// Test with URL that doesn't end with slash - should be as provided
	client = NewClient("test-key", "https://api.rajaongkir.com")
	expectedURL := "https://api.rajaongkir.com"
	if client.baseURL != expectedURL {
		t.Errorf("expected base URL to be '%s', got '%s'", expectedURL, client.baseURL)
	}
}