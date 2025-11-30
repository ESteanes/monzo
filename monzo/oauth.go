package monzo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// OAuthConfig holds the configuration for OAuth authentication
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Token represents an OAuth token with automatic refresh capability
type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	mu           sync.RWMutex
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return time.Now().After(t.ExpiresAt)
}

// OAuthFlow manages the complete OAuth authentication flow
type OAuthFlow struct {
	config     *OAuthConfig
	client     *Client
	token      *Token
	callbackCh chan string
	errorCh    chan error
	server     *http.Server
}

// NewOAuthFlow creates a new OAuth flow manager
func NewOAuthFlow(config *OAuthConfig) *OAuthFlow {
	return &OAuthFlow{
		config:     config,
		client:     NewClient(),
		callbackCh: make(chan string, 1),
		errorCh:    make(chan error, 1),
	}
}

// generateState generates a random state token for CSRF protection
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Authenticate performs the complete OAuth flow and returns an authenticated client
func (o *OAuthFlow) Authenticate(ctx context.Context) (*Client, error) {
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Start local callback server
	if err := o.startCallbackServer(); err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer o.stopCallbackServer()

	// Generate and display authorization URL
	authURL := o.client.Auth.GetAuthURL(o.config.ClientID, o.config.RedirectURL, state)
	fmt.Println("\n=================================")
	fmt.Println("Monzo OAuth Authentication")
	fmt.Println("=================================")
	fmt.Println("\nPlease visit this URL to authorize the application:")
	fmt.Println("\n" + authURL)
	fmt.Println("\nWaiting for authorization...")

	// Wait for callback or timeout
	select {
	case code := <-o.callbackCh:
		// Exchange code for token
		tokenResp, err := o.client.Auth.ExchangeAuthorizationCode(
			ctx,
			o.config.ClientID,
			o.config.ClientSecret,
			o.config.RedirectURL,
			code,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
		}

		// Create token with expiration
		o.token = &Token{
			AccessToken:  tokenResp.AccessToken,
			RefreshToken: tokenResp.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		}

		// Create authenticated client
		client := NewClient(
			WithAccessToken(o.token.AccessToken),
		)

		fmt.Println("\n✓ Authentication successful!")
		return client, nil

	case err := <-o.errorCh:
		return nil, fmt.Errorf("authentication failed: %w", err)

	case <-ctx.Done():
		return nil, fmt.Errorf("authentication timed out: %w", ctx.Err())
	}
}

// startCallbackServer starts a local HTTP server to handle the OAuth callback
func (o *OAuthFlow) startCallbackServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", o.handleCallback)

	o.server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := o.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			o.errorCh <- err
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

// stopCallbackServer stops the callback server
func (o *OAuthFlow) stopCallbackServer() {
	if o.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		o.server.Shutdown(ctx)
	}
}

// handleCallback handles the OAuth callback request
func (o *OAuthFlow) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Extract code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		errMsg := r.URL.Query().Get("error")
		if errMsg == "" {
			errMsg = "no authorization code received"
		}
		o.errorCh <- fmt.Errorf("authorization failed: %s", errMsg)
		http.Error(w, "Authorization failed", http.StatusBadRequest)
		return
	}

	if state == "" {
		o.errorCh <- fmt.Errorf("no state parameter received")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Send success response to browser
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Authentication Successful</title>
			<style>
				body {
					font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
					margin: 0;
					background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
				}
				.container {
					background: white;
					padding: 40px;
					border-radius: 10px;
					box-shadow: 0 10px 40px rgba(0,0,0,0.1);
					text-align: center;
					max-width: 400px;
				}
				.success-icon {
					font-size: 64px;
					margin-bottom: 20px;
				}
				h1 {
					color: #333;
					margin: 0 0 10px 0;
				}
				p {
					color: #666;
					margin: 0;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="success-icon">✓</div>
				<h1>Authentication Successful!</h1>
				<p>You can close this window and return to your application.</p>
			</div>
		</body>
		</html>
	`)

	// Send code through channel
	o.callbackCh <- code
}

// GetToken returns the current token
func (o *OAuthFlow) GetToken() *Token {
	return o.token
}

// RefreshToken refreshes the access token using the refresh token
func (o *OAuthFlow) RefreshToken(ctx context.Context) error {
	if o.token == nil || o.token.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	o.token.mu.Lock()
	defer o.token.mu.Unlock()

	tokenResp, err := o.client.Auth.RefreshAccessToken(
		ctx,
		o.config.ClientID,
		o.config.ClientSecret,
		o.token.RefreshToken,
	)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	o.token.AccessToken = tokenResp.AccessToken
	o.token.RefreshToken = tokenResp.RefreshToken
	o.token.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// AutoRefreshingClient creates a client that automatically refreshes tokens
type AutoRefreshingClient struct {
	*Client
	oauthFlow *OAuthFlow
	mu        sync.RWMutex
}

// NewAutoRefreshingClient creates a client that automatically handles token refresh
func (o *OAuthFlow) NewAutoRefreshingClient() *AutoRefreshingClient {
	if o.token == nil {
		return nil
	}

	return &AutoRefreshingClient{
		Client:    NewClient(WithAccessToken(o.token.AccessToken)),
		oauthFlow: o,
	}
}

// ensureValidToken checks if token is valid and refreshes if needed
func (c *AutoRefreshingClient) ensureValidToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.oauthFlow.token.IsExpired() {
		if err := c.oauthFlow.RefreshToken(ctx); err != nil {
			return err
		}
		c.SetAccessToken(c.oauthFlow.token.AccessToken)
	}

	return nil
}

// Override do method to ensure valid token before each request
func (c *AutoRefreshingClient) do(req *http.Request, v interface{}) (*http.Response, error) {
	if err := c.ensureValidToken(req.Context()); err != nil {
		return nil, fmt.Errorf("failed to ensure valid token: %w", err)
	}

	// Update authorization header with current token
	c.mu.RLock()
	token := c.oauthFlow.token.AccessToken
	c.mu.RUnlock()

	req.Header.Set("Authorization", "Bearer "+token)

	return c.Client.do(req, v)
}
