# Monzo Go Client

A simple, lightweight Go library for interacting with the Monzo API. This library provides a clean, idiomatic Go interface to all major Monzo API endpoints with **fully automated OAuth authentication**.

## Features

- **🚀 Automatic OAuth Flow**: Complete OAuth handled automatically - just provide your credentials!
- **💾 Token Persistence**: Tokens are saved to disk and automatically loaded on subsequent runs
- **🔄 Auto-Refresh**: Expired tokens are automatically refreshed without user intervention
- **📦 Complete API Coverage**: All major Monzo API endpoints supported
- **🪶 Lightweight**: Zero external dependencies, uses only the Go standard library
- **✅ Well-tested**: 64%+ test coverage with comprehensive unit tests
- **🏗️ Easy to Extend**: Clean architecture makes it simple to add new endpoints
- **🔒 Type-safe**: Strongly typed request/response structs
- **⏱️ Context Support**: All API calls support context for cancellation and timeouts

## Installation

```bash
go get github.com/esteanes/monzo/monzo
```

## Quick Start (Easiest Way)

The library handles **ALL OAuth complexity** for you. Just 3 lines of code!

```go
package main

import (
    "context"
    "log"

    "github.com/esteanes/monzo/monzo"
)

func main() {
    // That's it! This handles EVERYTHING:
    // ✓ Opens browser for OAuth
    // ✓ Handles callback automatically
    // ✓ Saves token to disk (~/.monzo/token.json)
    // ✓ Loads existing token on next run
    // ✓ Auto-refreshes when expired
    client, err := monzo.SimpleAuth("your_client_id", "your_client_secret")
    if err != nil {
        log.Fatal(err)
    }

    // Now just use the client - it handles everything!
    accounts, _ := client.Accounts.List(context.Background(), nil)
    for _, account := range accounts {
        balance, _ := client.Balance.Get(context.Background(), account.ID)
        println(account.Description, "£", balance.Balance/100)
    }
}
```

**First run**: Opens browser for OAuth authorization
**Subsequent runs**: Uses saved token automatically (no browser needed!)

## How Authentication Works

### The Simple Way (Recommended)

```go
client, err := monzo.SimpleAuth(clientID, clientSecret)
// That's it! Use client for all API calls
```

This single function:
1. Checks for existing token in `~/.monzo/token.json`
2. If found and valid, uses it immediately
3. If expired, automatically refreshes it
4. If not found, starts OAuth flow:
   - Starts local callback server on `http://localhost:8080`
   - Opens your browser to Monzo's authorization page
   - Waits for you to approve
   - Receives the callback automatically
   - Exchanges code for token
   - Saves token for future use
5. Returns a client that auto-refreshes tokens when they expire

## Advanced Authentication

### Custom OAuth Configuration

If you need more control, you can customize the OAuth flow:

```go
config := &monzo.OAuthConfig{
    ClientID:     "your_client_id",
    ClientSecret: "your_client_secret",
    RedirectURL:  "http://localhost:8080/",
}

flow := monzo.NewOAuthFlow(config)

// Authenticate with custom timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

client, err := flow.Authenticate(ctx)
if err != nil {
    log.Fatal(err)
}

// Get auto-refreshing client
autoRefreshClient := flow.NewAutoRefreshingClient()
```

### Custom Token Storage

By default, tokens are stored in `~/.monzo/token.json`. You can customize this:

```go
// Use custom token file location
store := monzo.NewFileTokenStore("/path/to/token.json")

client, err := monzo.AuthenticateWithStore(ctx, config, store)
```

Or implement your own storage (database, encrypted file, etc.):

```go
type MyTokenStore struct {
    // your implementation
}

func (s *MyTokenStore) Save(token *monzo.Token) error {
    // Save token to your storage
    return nil
}

func (s *MyTokenStore) Load() (*monzo.Token, error) {
    // Load token from your storage
    return token, nil
}

func (s *MyTokenStore) Delete() error {
    // Delete token from your storage
    return nil
}

// Use custom store
myStore := &MyTokenStore{}
client, err := monzo.AuthenticateWithStore(ctx, config, myStore)
```

### Manual OAuth (for advanced use cases)

If you need full control over the OAuth flow:

```go
client := monzo.NewClient()

// Step 1: Get authorization URL
authURL := client.Auth.GetAuthURL(
    "your_client_id",
    "http://localhost:8080/",
    "random_state_token",
)
// Direct user to authURL

// Step 2: Exchange code for token
tokenResp, err := client.Auth.ExchangeAuthorizationCode(
    context.Background(),
    "your_client_id",
    "your_client_secret",
    "http://localhost:8080/",
    "authorization_code",
)

client.SetAccessToken(tokenResp.AccessToken)

// Step 3: Manually refresh when needed
tokenResp, err = client.Auth.RefreshAccessToken(
    context.Background(),
    "your_client_id",
    "your_client_secret",
    tokenResp.RefreshToken,
)
```

## Usage Examples

### Accounts

```go
// List all accounts
accounts, err := client.Accounts.List(context.Background(), nil)

// Filter by account type
opts := &monzo.ListAccountsOptions{
    AccountType: monzo.AccountTypeUKRetail,
}
accounts, err := client.Accounts.List(context.Background(), opts)
```

### Balance

```go
balance, err := client.Balance.Get(context.Background(), "acc_00009237aqC8c5umZmrRdh")

fmt.Printf("Balance: £%.2f\n", float64(balance.Balance)/100)
fmt.Printf("Total Balance (including pots): £%.2f\n", float64(balance.TotalBalance)/100)
fmt.Printf("Spent today: £%.2f\n", float64(-balance.SpendToday)/100)
```

### Pots

```go
// List all pots
pots, err := client.Pots.List(context.Background(), "acc_00009237aqC8c5umZmrRdh")

// Deposit into a pot
pot, err := client.Pots.Deposit(
    context.Background(),
    "pot_0000778xxfgh4iu8z83nWb",
    "acc_00009237aqC8c5umZmrRdh",
    "unique_dedupe_id_1",
    10000, // £100.00 in pennies
)

// Withdraw from a pot
pot, err := client.Pots.Withdraw(
    context.Background(),
    "pot_0000778xxfgh4iu8z83nWb",
    "acc_00009237aqC8c5umZmrRdh",
    "unique_dedupe_id_2",
    5000, // £50.00 in pennies
)
```

### Transactions

```go
// List transactions
opts := &monzo.ListOptions{
    AccountID: "acc_00009237aqC8c5umZmrRdh",
    Limit:     50,
    Since:     "2024-01-01T00:00:00Z",
}
transactions, err := client.Transactions.List(context.Background(), opts)

// Get a single transaction with merchant details
transaction, err := client.Transactions.Get(
    context.Background(),
    "tx_00008zIcpb1TB4yeIFXMzx",
    []string{"merchant"}, // Expand merchant details
)

// Annotate a transaction with metadata
metadata := map[string]string{
    "notes":    "Business expense",
    "category": "travel",
}
transaction, err := client.Transactions.Annotate(
    context.Background(),
    "tx_00008zIcpb1TB4yeIFXMzx",
    metadata,
)
```

### Webhooks

```go
// Register a webhook
webhook, err := client.Webhooks.Register(
    context.Background(),
    "acc_00009237aqC8c5umZmrRdh",
    "https://your-app.com/webhooks/monzo",
)

// List webhooks
webhooks, err := client.Webhooks.List(
    context.Background(),
    "acc_00009237aqC8c5umZmrRdh",
)

// Delete a webhook
err := client.Webhooks.Delete(context.Background(), "webhook_000091yhhOmrXQaVZ1Irsv")
```

### Feed Items

```go
// Create a feed item
opts := &monzo.CreateFeedItemOptions{
    AccountID: "acc_00009237aqC8c5umZmrRdh",
    Type:      monzo.FeedItemTypeBasic,
    URL:       "https://your-app.com/details",
    Params: monzo.BasicFeedItemParams{
        Title:           "Payment Reminder",
        ImageURL:        "https://your-app.com/icon.png",
        Body:            "Your subscription is due soon",
        BackgroundColor: "#FCF1EE",
        TitleColor:      "#333333",
    },
}

err := client.FeedItems.Create(context.Background(), opts)
```

### Attachments

```go
// Step 1: Get upload URL
uploadResp, err := client.Attachments.Upload(
    context.Background(),
    "receipt.png",
    "image/png",
    12345, // file size in bytes
)

// Step 2: Upload file to uploadResp.UploadURL (using standard HTTP)
// ... upload logic ...

// Step 3: Register the attachment
attachment, err := client.Attachments.Register(
    context.Background(),
    "tx_00008zIcpb1TB4yeIFXMzx", // transaction ID
    uploadResp.FileURL,
    "image/png",
)

// Deregister an attachment
err := client.Attachments.Deregister(context.Background(), "attach_00009238aOAIvVqfb9LrZh")
```

### Receipts

```go
// Create a receipt
receipt := &monzo.Receipt{
    TransactionID: "tx_00008zIcpb1TB4yeIFXMzx",
    ExternalID:    "order-12345",
    Total:         1299, // £12.99 in pennies
    Currency:      "GBP",
    Items: []monzo.ReceiptItem{
        {
            Description: "Burger",
            Amount:      899,
            Currency:    "GBP",
            Quantity:    1,
        },
        {
            Description: "Fries",
            Amount:      400,
            Currency:    "GBP",
            Quantity:    1,
        },
    },
    Taxes: []monzo.ReceiptTax{
        {
            Description: "VAT",
            Amount:      100,
            Currency:    "GBP",
        },
    },
}

receiptID, err := client.Receipts.Create(context.Background(), receipt)

// Get a receipt
receipt, err := client.Receipts.Get(context.Background(), "order-12345")

// Delete a receipt
err := client.Receipts.Delete(context.Background(), "order-12345")
```

## Advanced Configuration

### Custom HTTP Client

```go
import (
    "net/http"
    "time"
)

httpClient := &http.Client{
    Timeout: 10 * time.Second,
    // Add custom transport, proxy settings, etc.
}

client := monzo.NewClient(
    monzo.WithHTTPClient(httpClient),
    monzo.WithAccessToken("your_token"),
)
```

### Custom Base URL (for testing)

```go
client := monzo.NewClient(
    monzo.WithBaseURL("https://test-api.example.com"),
    monzo.WithAccessToken("test_token"),
)
```

## Error Handling

The library returns typed errors that include HTTP status codes and API error messages:

```go
accounts, err := client.Accounts.List(context.Background(), nil)
if err != nil {
    if apiErr, ok := err.(*monzo.ErrorResponse); ok {
        fmt.Printf("API Error: %s (Status: %d)\n", apiErr.Message, apiErr.Response.StatusCode)

        // Check for specific OAuth errors
        if apiErr.ErrorCode == "invalid_token" {
            // Token expired or invalid - refresh token
        }
    } else {
        // Network or other error
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Architecture

The library is structured to make it easy to add new endpoints:

```
monzo/
├── client.go          # Core HTTP client and request handling
├── auth.go            # Authentication service
├── accounts.go        # Accounts service
├── balance.go         # Balance service
├── pots.go            # Pots service
├── transactions.go    # Transactions service
├── webhooks.go        # Webhooks service
├── feeditems.go       # Feed items service
├── attachments.go     # Attachments service
├── receipts.go        # Receipts service
└── *_test.go          # Comprehensive unit tests
```

### Adding a New Endpoint

1. Create a new service struct with a client reference
2. Implement the endpoint methods
3. Add the service to the `Client` struct
4. Initialize it in `NewClient()`
5. Write comprehensive tests

Example:

```go
// newfeature.go
type NewFeatureService struct {
    client *Client
}

func (s *NewFeatureService) Get(ctx context.Context, id string) (*Feature, error) {
    req, err := s.client.newRequest(ctx, "GET", "/new-feature/"+id, nil)
    if err != nil {
        return nil, err
    }

    var feature Feature
    if _, err := s.client.do(req, &feature); err != nil {
        return nil, err
    }

    return &feature, nil
}
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./monzo/...

# Run with coverage
go test -cover ./monzo/...

# Run with verbose output
go test -v ./monzo/...
```

Current test coverage: **83.2%**

## Requirements

- Go 1.16 or higher
- A Monzo developer account and API credentials

## Contributing

Contributions are welcome! Please ensure:

1. All new features include comprehensive unit tests
2. Code follows Go best practices and conventions
3. Tests pass with `go test ./...`
4. Code is formatted with `go fmt`

## Resources

- [Monzo API Documentation](https://docs.monzo.com/)
- [Monzo Developer Portal](https://developers.monzo.com/)
- [Monzo Developer Community](https://community.monzo.com/c/developers)

## License

This is an unofficial library and is not affiliated with or endorsed by Monzo Bank Limited.
