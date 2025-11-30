package monzo

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestTransactionsService_Get(t *testing.T) {
	t.Run("gets transaction successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET request, got %s", r.Method)
			}

			if r.URL.Path != "/transactions/tx_123" {
				t.Errorf("expected path /transactions/tx_123, got %s", r.URL.Path)
			}

			resp := GetTransactionResponse{
				Transaction: Transaction{
					ID:          "tx_123",
					Description: "Test Transaction",
					Amount:      -510,
					Currency:    "GBP",
					Created:     "2015-08-22T12:20:18Z",
					Category:    "eating_out",
					IsLoad:      false,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		transaction, err := client.Transactions.Get(context.Background(), "tx_123", nil)

		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if transaction.ID != "tx_123" {
			t.Errorf("expected ID tx_123, got %s", transaction.ID)
		}

		if transaction.Amount != -510 {
			t.Errorf("expected amount -510, got %d", transaction.Amount)
		}
	})

	t.Run("gets transaction with merchant expansion", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.RawQuery, "expand%5B%5D=merchant") {
				t.Error("expected expand[] parameter in query")
			}

			resp := GetTransactionResponse{
				Transaction: Transaction{
					ID:          "tx_123",
					Description: "Test Transaction",
					Amount:      -510,
					Currency:    "GBP",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		_, err := client.Transactions.Get(context.Background(), "tx_123", []string{"merchant"})

		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
	})
}

func TestTransactionsService_List(t *testing.T) {
	t.Run("lists transactions successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET request, got %s", r.Method)
			}

			if r.URL.Path != "/transactions" {
				t.Errorf("expected path /transactions, got %s", r.URL.Path)
			}

			if !strings.Contains(r.URL.RawQuery, "account_id=acc_123") {
				t.Error("expected account_id parameter in query")
			}

			resp := ListTransactionsResponse{
				Transactions: []Transaction{
					{
						ID:       "tx_1",
						Amount:   -510,
						Currency: "GBP",
					},
					{
						ID:       "tx_2",
						Amount:   -679,
						Currency: "GBP",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		opts := &ListOptions{
			AccountID: "acc_123",
		}

		transactions, err := client.Transactions.List(context.Background(), opts)

		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(transactions) != 2 {
			t.Errorf("expected 2 transactions, got %d", len(transactions))
		}
	})

	t.Run("lists transactions with pagination", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.RawQuery, "limit=50") {
				t.Error("expected limit parameter in query")
			}

			if !strings.Contains(r.URL.RawQuery, "since=2020-01-01T00%3A00%3A00Z") {
				t.Error("expected since parameter in query")
			}

			resp := ListTransactionsResponse{
				Transactions: []Transaction{},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		opts := &ListOptions{
			AccountID: "acc_123",
			Limit:     50,
			Since:     "2020-01-01T00:00:00Z",
		}

		_, err := client.Transactions.List(context.Background(), opts)

		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
	})

	t.Run("requires account_id", func(t *testing.T) {
		client := NewClient()

		opts := &ListOptions{}

		_, err := client.Transactions.List(context.Background(), opts)

		if err == nil {
			t.Fatal("expected error when account_id is missing")
		}
	})
}

func TestTransactionsService_Annotate(t *testing.T) {
	t.Run("annotates transaction successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PATCH" {
				t.Errorf("expected PATCH request, got %s", r.Method)
			}

			if r.URL.Path != "/transactions/tx_123" {
				t.Errorf("expected path /transactions/tx_123, got %s", r.URL.Path)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.FormValue("metadata[key1]") != "value1" {
				t.Error("expected metadata[key1] parameter")
			}

			resp := GetTransactionResponse{
				Transaction: Transaction{
					ID: "tx_123",
					Metadata: map[string]string{
						"key1": "value1",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		metadata := map[string]string{
			"key1": "value1",
		}

		transaction, err := client.Transactions.Annotate(context.Background(), "tx_123", metadata)

		if err != nil {
			t.Fatalf("Annotate() error = %v", err)
		}

		if transaction.Metadata["key1"] != "value1" {
			t.Error("expected metadata to be updated")
		}
	})
}
