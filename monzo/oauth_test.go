package monzo

import (
	"context"
	"testing"
	"time"
)

func TestToken_IsExpired(t *testing.T) {
	t.Run("token is expired", func(t *testing.T) {
		token := &Token{
			AccessToken:  "test",
			RefreshToken: "test",
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		}

		if !token.IsExpired() {
			t.Error("expected token to be expired")
		}
	})

	t.Run("token is not expired", func(t *testing.T) {
		token := &Token{
			AccessToken:  "test",
			RefreshToken: "test",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}

		if token.IsExpired() {
			t.Error("expected token to not be expired")
		}
	})

	t.Run("token expires soon", func(t *testing.T) {
		token := &Token{
			AccessToken:  "test",
			RefreshToken: "test",
			ExpiresAt:    time.Now().Add(30 * time.Second),
		}

		if token.IsExpired() {
			t.Error("expected token to not be expired yet")
		}
	})
}

func TestNewOAuthFlow(t *testing.T) {
	config := &OAuthConfig{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		RedirectURL:  "http://localhost:8080/",
	}

	flow := NewOAuthFlow(config)

	if flow == nil {
		t.Fatal("expected flow to be created")
	}

	if flow.config.ClientID != "test_client_id" {
		t.Errorf("expected client_id test_client_id, got %s", flow.config.ClientID)
	}

	if flow.client == nil {
		t.Error("expected client to be initialized")
	}

	if flow.callbackCh == nil {
		t.Error("expected callbackCh to be initialized")
	}

	if flow.errorCh == nil {
		t.Error("expected errorCh to be initialized")
	}
}

func TestGenerateState(t *testing.T) {
	state1, err := generateState()
	if err != nil {
		t.Fatalf("generateState() error = %v", err)
	}

	if state1 == "" {
		t.Error("expected non-empty state")
	}

	state2, err := generateState()
	if err != nil {
		t.Fatalf("generateState() error = %v", err)
	}

	if state1 == state2 {
		t.Error("expected different states on each call")
	}

	// Check length - should be base64 encoded 32 bytes
	if len(state1) < 40 {
		t.Error("expected state to be at least 40 characters")
	}
}

func TestOAuthFlow_GetToken(t *testing.T) {
	config := &OAuthConfig{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		RedirectURL:  "http://localhost:8080/",
	}

	flow := NewOAuthFlow(config)

	if flow.GetToken() != nil {
		t.Error("expected nil token initially")
	}

	flow.token = &Token{
		AccessToken:  "test_access",
		RefreshToken: "test_refresh",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	token := flow.GetToken()
	if token == nil {
		t.Fatal("expected token to be returned")
	}

	if token.AccessToken != "test_access" {
		t.Errorf("expected access_token test_access, got %s", token.AccessToken)
	}
}

func TestOAuthFlow_RefreshToken_NoToken(t *testing.T) {
	config := &OAuthConfig{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		RedirectURL:  "http://localhost:8080/",
	}

	flow := NewOAuthFlow(config)

	err := flow.RefreshToken(context.Background())
	if err == nil {
		t.Error("expected error when no token available")
	}
}

func TestNewAutoRefreshingClient(t *testing.T) {
	config := &OAuthConfig{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		RedirectURL:  "http://localhost:8080/",
	}

	flow := NewOAuthFlow(config)

	// Should return nil when no token
	client := flow.NewAutoRefreshingClient()
	if client != nil {
		t.Error("expected nil client when no token")
	}

	// Set token
	flow.token = &Token{
		AccessToken:  "test_access",
		RefreshToken: "test_refresh",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	client = flow.NewAutoRefreshingClient()
	if client == nil {
		t.Fatal("expected client to be created")
	}

	if client.Client == nil {
		t.Error("expected underlying client to be initialized")
	}

	if client.oauthFlow != flow {
		t.Error("expected oauth flow to be set")
	}
}
