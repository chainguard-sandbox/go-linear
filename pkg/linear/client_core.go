package linear

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Client manages communication with the Linear API.
type Client struct {
	gqlClient          intgraphql.LinearGraphQLClient
	config             *ClientConfig
	credentialProvider *credentialCache
}

// NewClient creates a new Linear API client for GraphQL operations.
//
// Parameters:
//   - apiKey: Linear API key from https://linear.app/settings/account/security
//     Format: "lin_api_xxx" for API keys, or OAuth token (Bearer prefix auto-added)
//     Required: yes
//   - opts: Optional configuration via functional options
//     Available: WithTimeout, WithHTTPClient, WithBaseURL, WithUserAgent
//
// Returns:
//   - *Client: Configured client ready for API calls
//   - error: Non-nil if apiKey is empty
//
// Authentication:
//   - API keys: Sent as-is in Authorization header
//   - OAuth tokens: "Bearer " prefix added automatically if not present
//   - Validates on first API call, not during NewClient
//
// Related: [WithTimeout], [WithHTTPClient], [Viewer]
//
// Example:
//
//	client, err := linear.NewClient("lin_api_xxx")
//	if err != nil {
//	    return err
//	}
//
// Example with options:
//
//	client, err := linear.NewClient(apiKey,
//	    linear.WithTimeout(60*time.Second),
//	    linear.WithBaseURL("https://api.linear.app/graphql"),
//	)
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	// Create default configs (apiKey can be empty if WithCredentialProvider is used)
	config := NewDefaultClientConfig(apiKey)
	config.Transport = NewDefaultTransportConfig()

	// Create static credential provider by default (empty apiKey handled)
	credProvider := newCredentialCache(&staticCredentialProvider{apiKey: apiKey})

	// Initialize client with config
	c := &Client{
		config:             config,
		credentialProvider: credProvider,
	}

	// Apply options (they mutate config through client reference)
	for _, opt := range opts {
		opt(c)
	}

	// Validate that we have a way to authenticate
	if apiKey == "" && c.config.CredentialProvider == nil {
		return nil, fmt.Errorf("apiKey is required (or use WithCredentialProvider)")
	}

	// Build and assign transport (conditional wrapping)
	c.config.HTTPClient.Transport = buildTransport(c.config)

	// Create GraphQL client with auth interceptor
	authInterceptor := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res any, next clientv2.RequestInterceptorFunc) error {
		// Get current credential (supports rotation)
		authValue, err := c.credentialProvider.Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to get credential: %w", err)
		}

		// Normalize authorization header
		if len(authValue) > 7 && authValue[:7] != "Bearer " {
			if len(authValue) > 8 && authValue[:8] != "lin_api_" {
				authValue = "Bearer " + authValue
			}
		}
		req.Header.Set("Authorization", authValue)
		req.Header.Set("User-Agent", c.config.UserAgent)

		// Execute request
		err = next(ctx, req, gqlInfo, res)

		// On 401, refresh credentials and retry once
		if err != nil && isAuthError(err) {
			if _, refreshErr := c.credentialProvider.Refresh(ctx); refreshErr == nil {
				// Retry with fresh credential
				newCred, _ := c.credentialProvider.Get(ctx)
				if len(newCred) > 7 && newCred[:7] != "Bearer " {
					if len(newCred) > 8 && newCred[:8] != "lin_api_" {
						newCred = "Bearer " + newCred
					}
				}
				req.Header.Set("Authorization", newCred)
				return next(ctx, req, gqlInfo, res)
			}
		}

		return err
	}

	c.gqlClient = intgraphql.NewClient(
		c.config.HTTPClient,
		c.config.BaseURL,
		nil,
		authInterceptor,
	)

	return c, nil
}

// Close closes idle connections and cleans up resources.
// It's safe to call Close multiple times.
func (c *Client) Close() error {
	if c.config.HTTPClient != nil {
		c.config.HTTPClient.CloseIdleConnections()
	}
	return nil
}

// isAuthError checks if an error is an authentication error (401).
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains 401 or authentication keywords
	errStr := err.Error()
	return contains(errStr, "401") || contains(errStr, "authentication") || contains(errStr, "unauthorized")
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
