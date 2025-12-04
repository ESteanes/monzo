package monzo

import (
	"context"
	"fmt"
	"net/url"
)

const (
	// AuthURL is the Monzo OAuth authorization URL
	AuthURL = "https://auth.monzo.com/"
	// TokenEndpoint is the token exchange endpoint
	TokenEndpoint = "/oauth2/token"
	// LogoutEndpoint is the logout endpoint
	LogoutEndpoint = "/oauth2/logout"
	// WhoamiEndpoint is the endpoint to get info about the access token
	WhoamiEndpoint = "/ping/whoami"
)

// AuthService handles authentication with the Monzo API
type AuthService struct {
	client *Client
}

// TokenResponse represents the response from the token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ClientID     string `json:"client_id"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id"`
}

// WhoAmIResponse represents the response from the whoami endpoint
type WhoAmIResponse struct {
	Authenticated bool   `json:"authenticated"`
	ClientID      string `json:"client_id"`
	UserID        string `json:"user_id"`
}

// GetAuthURL generates the OAuth authorization URL
func (s *AuthService) GetAuthURL(clientID, redirectURI, state string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)

	return AuthURL + "?" + params.Encode()
}

// ExchangeAuthorizationCode exchanges an authorization code for an access token
func (s *AuthService) ExchangeAuthorizationCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("code", code)

	req, err := s.client.newRequest(ctx, "POST", TokenEndpoint, data)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	if _, err := s.client.do(req, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	return &tokenResp, nil
}

// RefreshAccessToken refreshes an access token using a refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)

	req, err := s.client.newRequest(ctx, "POST", TokenEndpoint, data)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	if _, err := s.client.do(req, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	return &tokenResp, nil
}

// Logout invalidates the access token
func (s *AuthService) Logout(ctx context.Context) error {
	req, err := s.client.newRequest(ctx, "POST", LogoutEndpoint, nil)
	if err != nil {
		return err
	}

	if _, err := s.client.do(req, nil); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}

// WhoAmI returns information about the current access token
func (s *AuthService) WhoAmI(ctx context.Context) (*WhoAmIResponse, error) {
	req, err := s.client.newRequest(ctx, "GET", WhoamiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp WhoAmIResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get whoami info: %w", err)
	}

	return &resp, nil
}
