package monzo

import (
	"context"
	"fmt"
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
	storedToken StoredToken
}

// storedToken represents the token format stored in files
type StoredToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// NewTokenStore creates a new file-based token store
func NewTokenStore() *FileTokenStore {
	return &FileTokenStore{}
}

func DefaultTokenStore() (*FileTokenStore, error) {
	return NewTokenStore(), nil
}

// Save writes the token to the file
func (s *FileTokenStore) Save(token *Token) error {
	s.storedToken.AccessToken = token.AccessToken
	s.storedToken.ExpiresAt = token.ExpiresAt
	s.storedToken.RefreshToken = token.RefreshToken
	return nil
}

// Load reads the token from the file
func (s *FileTokenStore) Load() (*Token, error) {
	return &Token{
		AccessToken:  s.storedToken.AccessToken,
		RefreshToken: s.storedToken.RefreshToken,
		ExpiresAt:    s.storedToken.ExpiresAt,
	}, nil
}

// Delete removes the token file
func (s *FileTokenStore) Delete() error {
	s.storedToken.AccessToken = ""
	s.storedToken.RefreshToken = ""
	s.storedToken.ExpiresAt = time.Unix(0, 0)
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
