package monzo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TokenStore handles persistent storage of OAuth tokens
type TokenStore interface {
	Save(token *Token) error
	Load() (*Token, error)
	Delete() error
}

// FileTokenStore stores tokens in a JSON file
type FileTokenStore struct {
	filePath string
}

// storedToken represents the token format stored in files
type storedToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewFileTokenStore creates a new file-based token store
func NewFileTokenStore(filePath string) *FileTokenStore {
	return &FileTokenStore{
		filePath: filePath,
	}
}

// DefaultTokenStore creates a token store in the user's home directory
func DefaultTokenStore() (*FileTokenStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".monzo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	tokenPath := filepath.Join(configDir, "token.json")
	return NewFileTokenStore(tokenPath), nil
}

// Save writes the token to the file
func (s *FileTokenStore) Save(token *Token) error {
	token.mu.RLock()
	defer token.mu.RUnlock()

	stored := storedToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.ExpiresAt,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// Load reads the token from the file
func (s *FileTokenStore) Load() (*Token, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No token stored yet
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var stored storedToken
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &Token{
		AccessToken:  stored.AccessToken,
		RefreshToken: stored.RefreshToken,
		ExpiresAt:    stored.ExpiresAt,
	}, nil
}

// Delete removes the token file
func (s *FileTokenStore) Delete() error {
	if err := os.Remove(s.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}
	return nil
}

// AuthenticateWithStore performs OAuth authentication with token persistence
func AuthenticateWithStore(ctx context.Context, config *OAuthConfig, store TokenStore) (*AutoRefreshingClient, error) {
	// Try to load existing token
	token, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load token: %w", err)
	}

	flow := NewOAuthFlow(config)

	// If we have a valid token, use it
	if token != nil {
		flow.token = token

		// Check if token is expired and refresh if needed
		if token.IsExpired() {
			fmt.Println("Token expired, refreshing...")
			if err := flow.RefreshToken(ctx); err != nil {
				fmt.Println("Failed to refresh token, re-authenticating...")
				// Token refresh failed, need to re-authenticate
				token = nil
			} else {
				// Save refreshed token
				if err := store.Save(flow.token); err != nil {
					return nil, fmt.Errorf("failed to save refreshed token: %w", err)
				}
				fmt.Println("Token refreshed successfully!")
			}
		} else {
			fmt.Println("Using existing token...")
		}
	}

	// If no valid token, perform OAuth flow
	if token == nil || flow.token == nil {
		fmt.Println("No valid token found, starting OAuth flow...")
		if _, err := flow.Authenticate(ctx); err != nil {
			return nil, err
		}

		// Save the new token
		if err := store.Save(flow.token); err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	}

	return flow.NewAutoRefreshingClient(), nil
}

// SimpleAuth is the easiest way to authenticate - handles everything automatically
func SimpleAuth(clientID, clientSecret string) (*AutoRefreshingClient, error) {
	config := &OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/",
	}

	store, err := DefaultTokenStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create token store: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	return AuthenticateWithStore(ctx, config, store)
}
