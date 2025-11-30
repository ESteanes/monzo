# Monzo Go Client

A simple, lightweight Go library for interacting with the Monzo API. This library provides a clean, idiomatic Go interface to all major Monzo API endpoints with extensive test coverage.

## Features

- **Complete API Coverage**: Supports all major Monzo API endpoints including accounts, transactions, pots, webhooks, and more
- **Lightweight**: Minimal dependencies, uses only the Go standard library
- **Well-tested**: 83%+ test coverage with comprehensive unit tests
- **Easy to Extend**: Clean architecture makes it simple to add new endpoints
- **Type-safe**: Strongly typed request/response structs
- **Context Support**: All API calls support context for cancellation and timeouts

## Installation

```bash
go get github.com/esteanes/monzo-client/monzo
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/esteanes/monzo-client/monzo"
)

func main() {
    // Create a new client with your access token
    client := monzo.NewClient(
        monzo.WithAccessToken("your_access_token"),
    )

    // List all accounts
    accounts, err := client.Accounts.List(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }

    for _, account := range accounts {
        fmt.Printf("Account: %s (%s)\n", account.Description, account.ID)
    }

    // Get balance for the first account
    if len(accounts) > 0 {
        balance, err := client.Balance.Get(context.Background(), accounts[0].ID)
        if err != nil {
            log.Fatal(err)
        }

        fmt.Printf("Balance: £%.2f\n", float64(balance.Balance)/100)
    }
}
```

## Authentication

The Monzo API uses OAuth 2.0 for authentication. This library provides helpers for the OAuth flow:

### Step 1: Get Authorization URL

```go
client := monzo.NewClient()

authURL := client.Auth.GetAuthURL(
    "your_client_id",
    "https://your-app.com/callback",
    "random_state_token",
)

// Redirect user to authURL
```

### Step 2: Exchange Authorization Code

```go
tokenResp, err := client.Auth.ExchangeAuthorizationCode(
    context.Background(),
    "your_client_id",
    "your_client_secret",
    "https://your-app.com/callback",
    "authorization_code_from_callback",
)

if err != nil {
    log.Fatal(err)
}

// Store tokenResp.AccessToken and tokenResp.RefreshToken securely
client.SetAccessToken(tokenResp.AccessToken)
```

### Step 3: Refresh Access Token (when expired)

```go
tokenResp, err := client.Auth.RefreshAccessToken(
    context.Background(),
    "your_client_id",
    "your_client_secret",
    "your_refresh_token",
)

if err != nil {
    log.Fatal(err)
}

client.SetAccessToken(tokenResp.AccessToken)
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
