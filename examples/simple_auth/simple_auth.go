package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/esteanes/monzo-client/monzo"
)

func main() {
	// Get credentials from environment variables
	clientID := os.Getenv("MONZO_CLIENT_ID")
	clientSecret := os.Getenv("MONZO_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Please set MONZO_CLIENT_ID and MONZO_CLIENT_SECRET environment variables")
	}

	// That's it! SimpleAuth handles everything:
	// - Opens browser for OAuth
	// - Handles callback
	// - Saves token to disk
	// - Loads existing token on next run
	// - Auto-refreshes when expired
	client, err := monzo.SimpleAuth(clientID, clientSecret)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	ctx := context.Background()

	// Now just use the client - it handles token refresh automatically!
	fmt.Println("\n=== Your Accounts ===")
	accounts, err := client.Accounts.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list accounts: %v", err)
	}

	for _, account := range accounts {
		fmt.Printf("\n%s (%s)\n", account.Description, account.ID)

		balance, err := client.Balance.Get(ctx, account.ID)
		if err != nil {
			log.Printf("Failed to get balance: %v", err)
			continue
		}

		fmt.Printf("Balance: £%.2f\n", float64(balance.Balance)/100)
		fmt.Printf("Spent today: £%.2f\n", float64(-balance.SpendToday)/100)
	}
}
