package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/esteanes/monzo/monzo"
)

func main() {
	// Get access token from environment variable
	accessToken := os.Getenv("MONZO_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("MONZO_ACCESS_TOKEN environment variable is required")
	}

	// Create a new Monzo client
	client := monzo.NewClient(
		monzo.WithAccessToken(accessToken),
	)

	ctx := context.Background()

	// Verify authentication
	whoami, err := client.Auth.WhoAmI(ctx)
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	fmt.Printf("Authenticated as user: %s\n\n", whoami.UserID)

	// List all accounts
	fmt.Println("=== Accounts ===")
	accounts, err := client.Accounts.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list accounts: %v", err)
	}

	for _, account := range accounts {
		fmt.Printf("Account: %s\n", account.Description)
		fmt.Printf("  ID: %s\n", account.ID)
		fmt.Printf("  Created: %s\n\n", account.Created)

		// Get balance for this account
		balance, err := client.Balance.Get(ctx, account.ID)
		if err != nil {
			log.Printf("Failed to get balance: %v", err)
			continue
		}

		fmt.Printf("  Balance: £%.2f\n", float64(balance.Balance)/100)
		fmt.Printf("  Total Balance (with pots): £%.2f\n", float64(balance.TotalBalance)/100)
		fmt.Printf("  Spent today: £%.2f\n", float64(-balance.SpendToday)/100)
		fmt.Printf("  Currency: %s\n\n", balance.Currency)

		// List pots for this account
		fmt.Println("  === Pots ===")
		pots, err := client.Pots.List(ctx, account.ID)
		if err != nil {
			log.Printf("Failed to list pots: %v", err)
		} else {
			if len(pots) == 0 {
				fmt.Println("  No pots found\n")
			} else {
				for _, pot := range pots {
					if !pot.Deleted {
						fmt.Printf("  Pot: %s\n", pot.Name)
						fmt.Printf("    Balance: £%.2f\n", float64(pot.Balance)/100)
						fmt.Printf("    Style: %s\n\n", pot.Style)
					}
				}
			}
		}

		// List recent transactions
		fmt.Println("  === Recent Transactions ===")
		txOpts := &monzo.ListOptions{
			AccountID: account.ID,
			Limit:     5,
		}
		transactions, err := client.Transactions.List(ctx, txOpts)
		if err != nil {
			log.Printf("Failed to list transactions: %v", err)
		} else {
			if len(transactions) == 0 {
				fmt.Println("  No transactions found\n")
			} else {
				for _, tx := range transactions {
					amount := float64(tx.Amount) / 100
					amountStr := fmt.Sprintf("£%.2f", amount)
					if amount < 0 {
						amountStr = fmt.Sprintf("-£%.2f", -amount)
					}

					fmt.Printf("  %s: %s (%s)\n", tx.Created[:10], tx.Description, amountStr)
				}
				fmt.Println()
			}
		}

		// List webhooks for this account
		fmt.Println("  === Webhooks ===")
		webhooks, err := client.Webhooks.List(ctx, account.ID)
		if err != nil {
			log.Printf("Failed to list webhooks: %v", err)
		} else {
			if len(webhooks) == 0 {
				fmt.Println("  No webhooks registered\n")
			} else {
				for _, webhook := range webhooks {
					fmt.Printf("  Webhook ID: %s\n", webhook.ID)
					fmt.Printf("    URL: %s\n\n", webhook.URL)
				}
			}
		}
	}
}
