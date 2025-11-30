package monzo

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAuthService_GetAuthURL(t *testing.T) {
	client := NewClient()

	t.Run("generates auth URL without state", func(t *testing.T) {
		clientID := "test_client_id"
		redirectURI := "https://example.com/callback"

		url := client.Auth.GetAuthURL(clientID, redirectURI, "")

		if !strings.Contains(url, "client_id=test_client_id") {
			t.Error("expected URL to contain client_id parameter")
		}

		if !strings.Contains(url, "redirect_uri=https") {
			t.Error("expected URL to contain redirect_uri parameter")
		}

		if !strings.Contains(url, "response_type=code") {
			t.Error("expected URL to contain response_type parameter")
		}

		if !strings.HasPrefix(url, AuthURL) {
			t.Errorf("expected URL to start with %s", AuthURL)
		}
	})

	t.Run("generates auth URL with state", func(t *testing.T) {
		clientID := "test_client_id"
		redirectURI := "https://example.com/callback"
		state := "random_state_token"

		url := client.Auth.GetAuthURL(clientID, redirectURI, state)

		if !strings.Contains(url, "state=random_state_token") {
			t.Error("expected URL to contain state parameter")
		}
	})
}

func TestAuthService_ExchangeAuthorizationCode(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != TokenEndpoint {
			t.Errorf("expected path %s, got %s", TokenEndpoint, r.URL.Path)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		if r.FormValue("grant_type") != "authorization_code" {
			t.Error("expected grant_type to be authorization_code")
		}

		resp := TokenResponse{
			AccessToken:  "access_token_123",
			ClientID:     "client_id_123",
			ExpiresIn:    21600,
			RefreshToken: "refresh_token_123",
			TokenType:    "Bearer",
			UserID:       "user_id_123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server, client := setupTestServer(handler)
	defer server.Close()

	tokenResp, err := client.Auth.ExchangeAuthorizationCode(
		context.Background(),
		"client_id",
		"client_secret",
		"https://example.com/callback",
		"auth_code",
	)

	if err != nil {
		t.Fatalf("ExchangeAuthorizationCode() error = %v", err)
	}

	if tokenResp.AccessToken != "access_token_123" {
		t.Errorf("expected access_token access_token_123, got %s", tokenResp.AccessToken)
	}

	if tokenResp.RefreshToken != "refresh_token_123" {
		t.Errorf("expected refresh_token refresh_token_123, got %s", tokenResp.RefreshToken)
	}

	if tokenResp.ExpiresIn != 21600 {
		t.Errorf("expected expires_in 21600, got %d", tokenResp.ExpiresIn)
	}
}

func TestAuthService_RefreshAccessToken(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		if r.FormValue("grant_type") != "refresh_token" {
			t.Error("expected grant_type to be refresh_token")
		}

		resp := TokenResponse{
			AccessToken:  "new_access_token",
			ClientID:     "client_id_123",
			ExpiresIn:    21600,
			RefreshToken: "new_refresh_token",
			TokenType:    "Bearer",
			UserID:       "user_id_123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server, client := setupTestServer(handler)
	defer server.Close()

	tokenResp, err := client.Auth.RefreshAccessToken(
		context.Background(),
		"client_id",
		"client_secret",
		"old_refresh_token",
	)

	if err != nil {
		t.Fatalf("RefreshAccessToken() error = %v", err)
	}

	if tokenResp.AccessToken != "new_access_token" {
		t.Errorf("expected access_token new_access_token, got %s", tokenResp.AccessToken)
	}

	if tokenResp.RefreshToken != "new_refresh_token" {
		t.Errorf("expected refresh_token new_refresh_token, got %s", tokenResp.RefreshToken)
	}
}

func TestAuthService_Logout(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != LogoutEndpoint {
			t.Errorf("expected path %s, got %s", LogoutEndpoint, r.URL.Path)
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Error("expected Authorization header with Bearer token")
		}

		w.WriteHeader(http.StatusOK)
	}

	server, client := setupTestServer(handler)
	defer server.Close()

	err := client.Auth.Logout(context.Background())

	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
}

func TestAuthService_WhoAmI(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		if r.URL.Path != WhoamiEndpoint {
			t.Errorf("expected path %s, got %s", WhoamiEndpoint, r.URL.Path)
		}

		resp := WhoAmIResponse{
			Authenticated: true,
			ClientID:      "client_id_123",
			UserID:        "user_id_123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server, client := setupTestServer(handler)
	defer server.Close()

	whoAmIResp, err := client.Auth.WhoAmI(context.Background())

	if err != nil {
		t.Fatalf("WhoAmI() error = %v", err)
	}

	if !whoAmIResp.Authenticated {
		t.Error("expected authenticated to be true")
	}

	if whoAmIResp.ClientID != "client_id_123" {
		t.Errorf("expected client_id client_id_123, got %s", whoAmIResp.ClientID)
	}

	if whoAmIResp.UserID != "user_id_123" {
		t.Errorf("expected user_id user_id_123, got %s", whoAmIResp.UserID)
	}
}
