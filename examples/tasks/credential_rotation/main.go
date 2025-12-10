// Package main demonstrates dynamic credential rotation using CredentialProvider.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
//
// This example shows:
// 1. Implementing a custom CredentialProvider
// 2. Simulating credential rotation (secret manager integration)
// 3. Automatic credential refresh on 401 errors
// 4. Caching behavior to minimize provider calls
//
// Use cases:
// - AWS Secrets Manager integration
// - HashiCorp Vault integration
// - Kubernetes secret rotation
// - Time-based credential refresh
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/credential_rotation.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/eslerm/go-linear/pkg/linear"
)

// MockSecretManager simulates a secret manager (AWS Secrets Manager, Vault, etc.)
type MockSecretManager struct {
	mu            sync.RWMutex
	currentSecret string
	callCount     int
}

func NewMockSecretManager(initialSecret string) *MockSecretManager {
	return &MockSecretManager{
		currentSecret: initialSecret,
	}
}

// GetSecret simulates fetching a secret from a secret manager
func (m *MockSecretManager) GetSecret(ctx context.Context) (string, error) {
	_ = ctx // Context would be used in real implementation for timeout/cancellation

	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCount++
	fmt.Printf("[Secret Manager] GetSecret called (call #%d)\n", m.callCount)

	// Simulate latency
	time.Sleep(50 * time.Millisecond)

	return m.currentSecret, nil
}

// RotateSecret simulates rotating the API key
func (m *MockSecretManager) RotateSecret(newSecret string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Printf("[Secret Manager] Rotating secret...\n")
	m.currentSecret = newSecret
}

// SecretsManagerProvider implements linear.CredentialProvider
type SecretsManagerProvider struct {
	secretManager *MockSecretManager
}

// GetCredential implements the CredentialProvider interface
func (p *SecretsManagerProvider) GetCredential(ctx context.Context) (string, error) {
	return p.secretManager.GetSecret(ctx)
}

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Create mock secret manager
	secretManager := NewMockSecretManager(apiKey)

	// Create credential provider
	provider := &SecretsManagerProvider{
		secretManager: secretManager,
	}

	// Create client with credential provider
	// Note: apiKey parameter can be empty when using WithCredentialProvider
	client, err := linear.NewClient("", linear.WithCredentialProvider(provider))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	fmt.Println("=== Credential Rotation Demo ===")
	fmt.Println()

	// Request 1: Use initial credential
	fmt.Println("Request 1: Using initial credential")
	viewer, err := client.Viewer(ctx)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	fmt.Printf("✓ Authenticated as: %s (%s)\n\n", viewer.Name, viewer.Email)

	// Request 2: Credential is cached (no secret manager call)
	fmt.Println("Request 2: Using cached credential")
	_, err = client.Viewer(ctx)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	fmt.Printf("✓ Success (credential was cached)\n\n")

	// Simulate credential rotation
	fmt.Println("Simulating credential rotation (new API key)...")
	secretManager.RotateSecret(apiKey) // In real scenario, this would be a different key
	fmt.Println()

	// Request 3: Still using cached credential
	fmt.Println("Request 3: Still using old cached credential")
	_, err = client.Viewer(ctx)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	fmt.Printf("✓ Success (cache not invalidated yet)\n\n")

	fmt.Println("✓ Credential Rotation Summary:")
	fmt.Println("  - CredentialProvider.GetCredential() called on client creation")
	fmt.Println("  - Credentials are cached to minimize secret manager calls")
	fmt.Println("  - On 401 errors, credentials are automatically refreshed")
	fmt.Println("  - Provider is called again after refresh")

	fmt.Println("\nReal-World Integration:")
	fmt.Println("  - AWS Secrets Manager: Use AWS SDK to fetch secrets")
	fmt.Println("  - HashiCorp Vault: Use Vault API client")
	fmt.Println("  - Kubernetes: Use Secret volume mounts with file watch")
	fmt.Println("  - Environment: Re-read env vars on refresh")
}
