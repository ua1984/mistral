package mistral

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the Client.
// Options allow you to customize client behavior such as the base URL,
// HTTP client, and request timeout. Pass options to NewClient to configure
// the client during initialization.
type Option func(*Client)

// WithBaseURL sets a custom base URL for the Mistral API.
// Use this option if you need to use a different API endpoint, such as a proxy
// or a custom deployment of the Mistral API.
//
// Parameters:
//   - baseURL: The base URL to use (e.g., "https://api.custom-domain.com").
//     Do not include a trailing slash
//
// Returns:
//   - An Option that configures the client's base URL
//
// Example:
//
//	client := mistral.NewClient(
//	    "your-api-key",
//	    mistral.WithBaseURL("https://api.custom-domain.com"),
//	)
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client for making API requests.
// Use this option when you need fine-grained control over HTTP behavior,
// such as custom TLS configuration, transport settings, or connection pooling.
//
// Parameters:
//   - httpClient: A configured *http.Client to use for all API requests.
//     The client should have appropriate timeout and transport settings
//
// Returns:
//   - An Option that configures the client's HTTP client
//
// Example:
//
//	customHTTPClient := &http.Client{
//	    Timeout: 90 * time.Second,
//	    Transport: &http.Transport{
//	        MaxIdleConns: 100,
//	    },
//	}
//	client := mistral.NewClient(
//	    "your-api-key",
//	    mistral.WithHTTPClient(customHTTPClient),
//	)
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the timeout for all HTTP requests made by the client.
// This is a convenience option that configures the timeout on the client's
// HTTP client. The timeout applies to the entire request-response cycle,
// including connection time, request sending, and response reading.
//
// Parameters:
//   - timeout: The maximum duration to wait for a request to complete.
//     Choose based on your expected response times. Streaming requests may need longer timeouts
//
// Returns:
//   - An Option that configures the client's request timeout
//
// Example:
//
//	client := mistral.NewClient(
//	    "your-api-key",
//	    mistral.WithTimeout(30 * time.Second),
//	)
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}
