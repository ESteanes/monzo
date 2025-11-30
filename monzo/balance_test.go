package monzo

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestBalanceService_Get(t *testing.T) {
	t.Run("gets balance successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET request, got %s", r.Method)
			}

			if r.URL.Path != "/balance" {
				t.Errorf("expected path /balance, got %s", r.URL.Path)
			}

			if !strings.Contains(r.URL.RawQuery, "account_id=acc_123") {
				t.Error("expected account_id parameter in query")
			}

			resp := Balance{
				Balance:      5000,
				TotalBalance: 6000,
				Currency:     "GBP",
				SpendToday:   -150,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		balance, err := client.Balance.Get(context.Background(), "acc_123")

		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if balance.Balance != 5000 {
			t.Errorf("expected balance 5000, got %d", balance.Balance)
		}

		if balance.TotalBalance != 6000 {
			t.Errorf("expected total_balance 6000, got %d", balance.TotalBalance)
		}

		if balance.Currency != "GBP" {
			t.Errorf("expected currency GBP, got %s", balance.Currency)
		}

		if balance.SpendToday != -150 {
			t.Errorf("expected spend_today -150, got %d", balance.SpendToday)
		}
	})

	t.Run("handles zero balance", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			resp := Balance{
				Balance:      0,
				TotalBalance: 0,
				Currency:     "GBP",
				SpendToday:   0,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		balance, err := client.Balance.Get(context.Background(), "acc_123")

		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if balance.Balance != 0 {
			t.Errorf("expected balance 0, got %d", balance.Balance)
		}
	})

	t.Run("handles error response", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Account not found"}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		_, err := client.Balance.Get(context.Background(), "invalid_account")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
