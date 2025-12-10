package linear

import (
	"context"
	"errors"
	"testing"
)

type mockCredentialProvider struct {
	credentials []string
	callCount   int
}

func (m *mockCredentialProvider) GetCredential(ctx context.Context) (string, error) {
	if m.callCount >= len(m.credentials) {
		return "", errors.New("no more credentials")
	}
	cred := m.credentials[m.callCount]
	m.callCount++
	return cred, nil
}

func TestCredentialCache_Get(t *testing.T) {
	provider := &mockCredentialProvider{
		credentials: []string{"key1", "key2"},
	}
	cache := newCredentialCache(provider)

	// First call fetches from provider
	cred1, err := cache.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if cred1 != "key1" {
		t.Errorf("Get() = %q, want %q", cred1, "key1")
	}

	// Second call returns cached value (provider not called)
	cred2, err := cache.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if cred2 != "key1" {
		t.Errorf("Get() = %q, want %q (cached)", cred2, "key1")
	}

	if provider.callCount != 1 {
		t.Errorf("Provider called %d times, want 1 (caching should prevent second call)", provider.callCount)
	}
}

func TestCredentialCache_Refresh(t *testing.T) {
	provider := &mockCredentialProvider{
		credentials: []string{"key1", "key2", "key3"},
	}
	cache := newCredentialCache(provider)

	// Initial fetch
	cred1, _ := cache.Get(context.Background())
	if cred1 != "key1" {
		t.Errorf("Initial Get() = %q, want %q", cred1, "key1")
	}

	// Refresh gets new credential
	cred2, err := cache.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if cred2 != "key2" {
		t.Errorf("Refresh() = %q, want %q", cred2, "key2")
	}

	// Get returns refreshed credential
	cred3, _ := cache.Get(context.Background())
	if cred3 != "key2" {
		t.Errorf("Get() after Refresh() = %q, want %q", cred3, "key2")
	}
}

func TestStaticCredentialProvider(t *testing.T) {
	provider := &staticCredentialProvider{apiKey: "test-key"}

	cred, err := provider.GetCredential(context.Background())
	if err != nil {
		t.Fatalf("GetCredential() error = %v", err)
	}
	if cred != "test-key" {
		t.Errorf("GetCredential() = %q, want %q", cred, "test-key")
	}

	// Should return same value on subsequent calls
	cred2, _ := provider.GetCredential(context.Background())
	if cred2 != "test-key" {
		t.Errorf("GetCredential() = %q, want consistent value", cred2)
	}
}
