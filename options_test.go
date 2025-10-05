package mistral

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithBaseURL(t *testing.T) {
	customURL := "https://custom.api.example.com"
	client := NewClient("test-api-key", WithBaseURL(customURL))

	assert.Equal(t, customURL, client.baseURL, "baseURL should be set to custom value")
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	client := NewClient("test-api-key", WithHTTPClient(customClient))

	assert.Equal(t, customClient, client.httpClient, "httpClient should be set to custom client")
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout, "custom client timeout should be preserved")
}

func TestWithTimeout(t *testing.T) {
	customTimeout := 120 * time.Second
	client := NewClient("test-api-key", WithTimeout(customTimeout))

	assert.Equal(t, customTimeout, client.httpClient.Timeout, "httpClient timeout should be set to custom value")
}

func TestMultipleOptions(t *testing.T) {
	customURL := "https://custom.api.example.com"
	customTimeout := 90 * time.Second

	client := NewClient("test-api-key",
		WithBaseURL(customURL),
		WithTimeout(customTimeout),
	)

	assert.Equal(t, customURL, client.baseURL, "baseURL should be set to custom value")
	assert.Equal(t, customTimeout, client.httpClient.Timeout, "timeout should be set to custom value")
}

func TestDefaultValues(t *testing.T) {
	client := NewClient("test-api-key")

	assert.Equal(t, defaultBaseURL, client.baseURL, "baseURL should be set to default value")
	assert.Equal(t, defaultTimeout, client.httpClient.Timeout, "timeout should be set to default value")
	assert.NotNil(t, client.httpClient, "httpClient should not be nil")
	assert.Equal(t, "test-api-key", client.apiKey, "apiKey should be set correctly")
}

func TestOptionChaining(t *testing.T) {
	// Test that options are applied in order
	firstTimeout := 100 * time.Second
	secondTimeout := 200 * time.Second

	client := NewClient("test-api-key",
		WithTimeout(firstTimeout),
		WithTimeout(secondTimeout), // This should override the first one
	)

	assert.Equal(t, secondTimeout, client.httpClient.Timeout, "last timeout option should win")
}

func TestWithHTTPClientAndTimeout(t *testing.T) {
	// Test that WithTimeout works even when a custom HTTP client is set
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	newTimeout := 60 * time.Second

	client := NewClient("test-api-key",
		WithHTTPClient(customClient),
		WithTimeout(newTimeout),
	)

	assert.Equal(t, customClient, client.httpClient, "httpClient should be the custom client")
	assert.Equal(t, newTimeout, client.httpClient.Timeout, "timeout should be updated by WithTimeout")
}
