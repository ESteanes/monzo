package monzo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	// DefaultBaseURL is the default base URL for the Monzo API
	DefaultBaseURL = "https://api.monzo.com"
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
)

// Client manages communication with the Monzo API
type Client struct {
	baseURL     string
	httpClient  *http.Client
	accessToken string
	cache       *cache.Cache

	// Services
	Auth         *AuthService
	Accounts     *AccountsService
	Balance      *BalanceService
	Pots         *PotsService
	Transactions *TransactionsService
	Webhooks     *WebhooksService
	FeedItems    *FeedItemsService
	Attachments  *AttachmentsService
	Receipts     *ReceiptsService
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithAccessToken sets the access token for authenticated requests
func WithAccessToken(token string) ClientOption {
	return func(c *Client) {
		c.accessToken = token
	}
}

// WithCache sets a custom cache for the client
func WithCache(customCache *cache.Cache) ClientOption {
	return func(c *Client) {
		c.cache = customCache
	}
}

// WithCacheTTL sets the default cache TTL and cleanup interval
func WithCacheTTL(defaultExpiration, cleanupInterval time.Duration) ClientOption {
	return func(c *Client) {
		c.cache = cache.New(defaultExpiration, cleanupInterval)
	}
}

// NewClient creates a new Monzo API client
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		// Default cache: 5 minute expiration, 10 minute cleanup interval
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Initialize services
	c.Auth = &AuthService{client: c}
	c.Accounts = &AccountsService{client: c}
	c.Balance = &BalanceService{client: c}
	c.Pots = &PotsService{client: c}
	c.Transactions = &TransactionsService{client: c}
	c.Webhooks = &WebhooksService{client: c}
	c.FeedItems = &FeedItemsService{client: c}
	c.Attachments = &AttachmentsService{client: c}
	c.Receipts = &ReceiptsService{client: c}

	return c
}

// SetAccessToken sets the access token for the client
func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

// GetCache returns the client's cache instance
func (c *Client) GetCache() *cache.Cache {
	return c.cache
}

// ClearCache clears all cached data
func (c *Client) ClearCache() {
	if c.cache != nil {
		c.cache.Flush()
	}
}

// CacheGet retrieves a value from the cache
func (c *Client) CacheGet(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

// CacheSet stores a value in the cache with the default expiration time
func (c *Client) CacheSet(key string, value interface{}) {
	c.cache.Set(key, value, cache.DefaultExpiration)
}

// CacheSetWithExpiration stores a value in the cache with a custom expiration time
func (c *Client) CacheSetWithExpiration(key string, value interface{}, expiration time.Duration) {
	c.cache.Set(key, value, expiration)
}

// CacheDelete removes a value from the cache
func (c *Client) CacheDelete(key string) {
	c.cache.Delete(key)
}

// CacheItemCount returns the number of items in the cache
func (c *Client) CacheItemCount() int {
	return c.cache.ItemCount()
}

// newRequest creates an authenticated API request
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	u := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case url.Values:
			reqBody = strings.NewReader(v.Encode())
		default:
			buf := new(bytes.Buffer)
			if err := json.NewEncoder(buf).Encode(body); err != nil {
				return nil, fmt.Errorf("failed to encode request body: %w", err)
			}
			reqBody = buf
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type based on body type
	if body != nil {
		switch body.(type) {
		case url.Values:
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		default:
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Add authorization header if access token is set
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	return req, nil
}

// do sends an API request and returns the API response
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return resp, err
	}

	if v != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return resp, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return resp, nil
}

// ErrorResponse represents an error response from the Monzo API
type ErrorResponse struct {
	Response   *http.Response
	Message    string `json:"message"`
	Code       string `json:"code"`
	ErrorCode  string `json:"error"` // OAuth errors
}

func (e *ErrorResponse) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("monzo: %s (OAuth error: %s)", e.Message, e.ErrorCode)
	}
	return fmt.Sprintf("monzo: %d %s - %s", e.Response.StatusCode, e.Response.Status, e.Message)
}

// checkResponse checks the API response for errors
func checkResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errResp := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		json.Unmarshal(data, errResp)
	}

	return errResp
}
