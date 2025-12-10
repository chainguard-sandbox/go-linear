package linear

import (
	"context"
	"sync"
)

// CredentialProvider provides dynamic API credentials.
// Useful for credential rotation, secret managers, or token refresh.
//
// Example with AWS Secrets Manager:
//
//	type SecretsManagerProvider struct {
//	    secretName string
//	    client     *secretsmanager.SecretsManager
//	}
//
//	func (p *SecretsManagerProvider) GetCredential(ctx context.Context) (string, error) {
//	    result, err := p.client.GetSecretValue(&secretsmanager.GetSecretValueInput{
//	        SecretId: aws.String(p.secretName),
//	    })
//	    if err != nil {
//	        return "", err
//	    }
//	    return *result.SecretString, nil
//	}
//
//	provider := &SecretsManagerProvider{secretName: "linear-api-key"}
//	client, _ := linear.NewClient("", linear.WithCredentialProvider(provider))
type CredentialProvider interface {
	// GetCredential returns the current API credential.
	// Called on client creation and on 401 errors for refresh.
	GetCredential(ctx context.Context) (string, error)
}

// staticCredentialProvider implements CredentialProvider with a static API key.
type staticCredentialProvider struct {
	apiKey string
}

func (p *staticCredentialProvider) GetCredential(ctx context.Context) (string, error) {
	return p.apiKey, nil
}

// credentialCache caches credentials from a provider.
type credentialCache struct {
	mu       sync.RWMutex
	provider CredentialProvider
	cached   string
}

func newCredentialCache(provider CredentialProvider) *credentialCache {
	return &credentialCache{
		provider: provider,
	}
}

// Get returns the cached credential or fetches a fresh one.
func (c *credentialCache) Get(ctx context.Context) (string, error) {
	c.mu.RLock()
	if c.cached != "" {
		cached := c.cached
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	return c.Refresh(ctx)
}

// Refresh fetches a fresh credential from the provider and caches it.
func (c *credentialCache) Refresh(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cred, err := c.provider.GetCredential(ctx)
	if err != nil {
		return "", err
	}

	c.cached = cred
	return cred, nil
}
