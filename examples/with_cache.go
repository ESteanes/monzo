package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/esteanes/monzo/monzo"
)

func main() {
	// Get access token from environment variable
	accessToken := os.Getenv("MONZO_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("MONZO_ACCESS_TOKEN environment variable is required")
	}

	// Example 1: Create a client with default cache (5 minute expiration)
	fmt.Println("=== Example 1: Default Cache ===")
	client := monzo.NewClient(
		monzo.WithAccessToken(accessToken),
	)

	ctx := context.Background()

	// List all accounts
	accounts, err := client.Accounts.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list accounts: %v", err)
	}

	if len(accounts) == 0 {
		log.Fatal("No accounts found")
	}

	accountID := accounts[0].ID
	fmt.Printf("Using account: %s\n\n", accounts[0].Description)

	// Example: Using cache to store balance data
	cacheKey := fmt.Sprintf("balance_%s", accountID)
	if cachedBalance, found := client.CacheGet(cacheKey); found {
		fmt.Println("Using cached balance:")
		balance := cachedBalance.(*monzo.Balance)
		fmt.Printf("  Balance: £%.2f (from cache)\n", float64(balance.Balance)/100)
	} else {
		fmt.Println("Fetching fresh balance from API...")
		balance, err := client.Balance.Get(ctx, accountID)
		if err != nil {
			log.Fatalf("Failed to get balance: %v", err)
		}

		// Store balance in cache (will use default 5 minute expiration)
		client.CacheSet(cacheKey, balance)
		fmt.Printf("  Balance: £%.2f (cached for 5 minutes)\n", float64(balance.Balance)/100)
	}

	fmt.Println()

	// Example: Using cache for transactions with custom expiration
	fmt.Println("=== Example 2: Custom Expiration ===")
	txOpts := &monzo.ListOptions{
		AccountID: accountID,
		Limit:     10,
	}

	transactions, err := client.Transactions.List(ctx, txOpts)
	if err != nil {
		log.Fatalf("Failed to list transactions: %v", err)
	}

	// Store transactions in cache with custom 2 minute expiration
	txCacheKey := fmt.Sprintf("transactions_%s_recent", accountID)
	client.CacheSetWithExpiration(txCacheKey, transactions, 2*time.Minute)
	fmt.Printf("Cached %d transactions (expires in 2 minutes)\n", len(transactions))

	// Verify cache
	if cachedTx, found := client.CacheGet(txCacheKey); found {
		cachedTransactions := cachedTx.([]monzo.Transaction)
		fmt.Printf("Retrieved %d transactions from cache\n", len(cachedTransactions))
		if len(cachedTransactions) > 0 {
			fmt.Printf("  Most recent: %s - %s\n",
				cachedTransactions[0].Description,
				cachedTransactions[0].Created[:10])
		}
	}

	fmt.Println()

	// Example 3: Custom cache configuration
	fmt.Println("=== Example 3: Custom Cache Configuration ===")
	_ = monzo.NewClient(
		monzo.WithAccessToken(accessToken),
		monzo.WithCacheTTL(10*time.Minute, 15*time.Minute), // 10 min expiration, 15 min cleanup
	)

	fmt.Println("Client created with custom cache:")
	fmt.Println("  - Default expiration: 10 minutes")
	fmt.Println("  - Cleanup interval: 15 minutes")

	// Example 4: Direct cache access for advanced operations
	fmt.Println("\n=== Example 4: Direct Cache Access ===")
	cache := client.GetCache()

	// Store multiple values with different expirations
	cache.Set("short_lived", "expires in 10 seconds", 10*time.Second)
	cache.Set("medium_lived", "expires in 1 minute", 1*time.Minute)
	cache.Set("never_expires", "never expires", gocache.NoExpiration)

	fmt.Printf("Cache item count: %d\n", client.CacheItemCount())

	// Get all items from cache
	fmt.Println("\nAll cached items:")
	for key, item := range cache.Items() {
		fmt.Printf("  %s: %v (expires: %v)\n", key, item.Object, item.Expiration)
	}

	// Example 5: Cache statistics
	fmt.Println("\n=== Example 5: Cache Statistics ===")
	fmt.Printf("Total items in cache: %d\n", client.CacheItemCount())

	// Delete specific item
	client.CacheDelete("short_lived")
	fmt.Printf("After deleting 'short_lived': %d items\n", client.CacheItemCount())

	// Clear all cache
	fmt.Println("\nClearing all cached data...")
	client.ClearCache()
	fmt.Printf("After clearing: %d items\n", client.CacheItemCount())

	// Example 6: Using cache with custom client
	fmt.Println("\n=== Example 6: Bring Your Own Cache ===")
	customCache := gocache.New(30*time.Second, 1*time.Minute)
	customCache.Set("predefined_key", "predefined_value", gocache.DefaultExpiration)

	clientWithOwnCache := monzo.NewClient(
		monzo.WithAccessToken(accessToken),
		monzo.WithCache(customCache),
	)

	fmt.Printf("Custom cache already has %d items\n", clientWithOwnCache.CacheItemCount())
	if val, found := clientWithOwnCache.CacheGet("predefined_key"); found {
		fmt.Printf("Found predefined value: %s\n", val)
	}

	fmt.Println("\n=== Cache Usage Examples Completed! ===")
}
