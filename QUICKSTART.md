# Monzo Go Client - Quick Start Guide

Get started with the Monzo API in under 5 minutes!

## Step 1: Get Your Credentials

1. Go to [Monzo Developer Portal](https://developers.monzo.com/)
2. Create a new OAuth client
3. Set the redirect URL to: `http://localhost:8080/`
4. Note down your `Client ID` and `Client Secret`

## Step 2: Install the Library

```bash
go get github.com/esteanes/monzo-client/monzo
```

## Step 3: Write Your First Program

Create `main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/esteanes/monzo-client/monzo"
)

func main() {
	// Replace with your credentials
	client, err := monzo.SimpleAuth(
		"your_client_id_here",
		"your_client_secret_here",
	)
	if err != nil {
		log.Fatal(err)
	}

	// List your accounts
	accounts, err := client.Accounts.List(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Your Monzo Accounts:")
	for _, account := range accounts {
		fmt.Printf("\n%s\n", account.Description)

		balance, _ := client.Balance.Get(context.Background(), account.ID)
		fmt.Printf("Balance: £%.2f\n", float64(balance.Balance)/100)
	}
}
```

## Step 4: Run It!

```bash
go run main.go
```

**What happens:**

1. **First time**: A browser window opens for OAuth authorization
   - Log in to Monzo
   - Approve the access request
   - Return to terminal - you're authenticated!
   - Token saved to `~/.monzo/token.json`

2. **Every time after**: Uses the saved token automatically!
   - No browser needed
   - Token auto-refreshes when expired
   - Just works!

## What Can You Do?

### Check Your Balance

```go
balance, _ := client.Balance.Get(ctx, accountID)
fmt.Printf("£%.2f\n", float64(balance.Balance)/100)
```

### List Recent Transactions

```go
txs, _ := client.Transactions.List(ctx, &monzo.ListOptions{
	AccountID: accountID,
	Limit:     10,
})

for _, tx := range txs {
	fmt.Printf("%s: %s (£%.2f)\n",
		tx.Created[:10],
		tx.Description,
		float64(tx.Amount)/100)
}
```

### Move Money to a Pot

```go
pot, _ := client.Pots.Deposit(ctx, potID, accountID, "unique-id", 5000)
fmt.Printf("New pot balance: £%.2f\n", float64(pot.Balance)/100)
```

### Create a Feed Item

```go
client.FeedItems.Create(ctx, &monzo.CreateFeedItemOptions{
	AccountID: accountID,
	Params: monzo.BasicFeedItemParams{
		Title:    "Hello from Go!",
		ImageURL: "https://example.com/image.png",
		Body:     "This is a custom feed item",
	},
})
```

### Register a Webhook

```go
webhook, _ := client.Webhooks.Register(ctx,
	accountID,
	"https://your-app.com/webhook")
fmt.Printf("Webhook registered: %s\n", webhook.ID)
```

## Environment Variables (Recommended)

Instead of hardcoding credentials:

```bash
export MONZO_CLIENT_ID="your_client_id"
export MONZO_CLIENT_SECRET="your_client_secret"
```

Then in your code:

```go
import "os"

client, err := monzo.SimpleAuth(
	os.Getenv("MONZO_CLIENT_ID"),
	os.Getenv("MONZO_CLIENT_SECRET"),
)
```

## Common Patterns

### Error Handling

```go
accounts, err := client.Accounts.List(ctx, nil)
if err != nil {
	if apiErr, ok := err.(*monzo.ErrorResponse); ok {
		fmt.Printf("API Error: %s (Status: %d)\n",
			apiErr.Message,
			apiErr.Response.StatusCode)
	} else {
		log.Fatal(err)
	}
	return
}
```

### Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

accounts, err := client.Accounts.List(ctx, nil)
```

### Pagination

```go
opts := &monzo.ListOptions{
	AccountID: accountID,
	Limit:     50,
	Since:     "2024-01-01T00:00:00Z",
}
transactions, _ := client.Transactions.List(ctx, opts)
```

## Troubleshooting

### "Token expired" or "Unauthorized"

The library handles this automatically, but if you see these errors:

1. Delete the token file: `rm ~/.monzo/token.json`
2. Run your program again - it will re-authenticate

### "Address already in use"

The OAuth server runs on port 8080. If it's in use:

```go
config := &monzo.OAuthConfig{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	RedirectURL:  "http://localhost:9090/", // Change port
}
// Update your Monzo OAuth client redirect URL to match!
```

### No browser opens automatically

Copy the URL from the terminal output and open it manually in your browser.

## Next Steps

- Read the [full README](README.md) for all features
- Check [examples/](examples/) for more code samples
- View the [Monzo API docs](https://docs.monzo.com/) for endpoint details

## Getting Help

- Check the [examples](examples/) directory
- Read the [Monzo API documentation](https://docs.monzo.com/)
- Visit the [Monzo Developer Community](https://community.monzo.com/c/developers)

Happy coding! 🚀
