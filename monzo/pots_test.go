package monzo

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPotsService_List(t *testing.T) {
	t.Run("lists pots successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET request, got %s", r.Method)
			}

			if r.URL.Path != "/pots" {
				t.Errorf("expected path /pots, got %s", r.URL.Path)
			}

			if !strings.Contains(r.URL.RawQuery, "current_account_id=acc_123") {
				t.Error("expected current_account_id parameter in query")
			}

			resp := ListPotsResponse{
				Pots: []Pot{
					{
						ID:       "pot_0000778xxfgh4iu8z83nWb",
						Name:     "Savings",
						Style:    "beach_ball",
						Balance:  133700,
						Currency: "GBP",
						Created:  "2017-11-09T12:30:53.695Z",
						Updated:  "2017-11-09T12:30:53.695Z",
						Deleted:  false,
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		pots, err := client.Pots.List(context.Background(), "acc_123")

		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(pots) != 1 {
			t.Errorf("expected 1 pot, got %d", len(pots))
		}

		if pots[0].Name != "Savings" {
			t.Errorf("expected pot name Savings, got %s", pots[0].Name)
		}

		if pots[0].Balance != 133700 {
			t.Errorf("expected balance 133700, got %d", pots[0].Balance)
		}
	})
}

func TestPotsService_Deposit(t *testing.T) {
	t.Run("deposits into pot successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("expected PUT request, got %s", r.Method)
			}

			if r.URL.Path != "/pots/pot_123/deposit" {
				t.Errorf("expected path /pots/pot_123/deposit, got %s", r.URL.Path)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.FormValue("source_account_id") != "acc_123" {
				t.Error("expected source_account_id parameter")
			}

			if r.FormValue("amount") != "10000" {
				t.Error("expected amount parameter")
			}

			if r.FormValue("dedupe_id") != "dedupe_123" {
				t.Error("expected dedupe_id parameter")
			}

			resp := Pot{
				ID:       "pot_123",
				Name:     "Savings",
				Balance:  143700,
				Currency: "GBP",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		pot, err := client.Pots.Deposit(context.Background(), "pot_123", "acc_123", "dedupe_123", 10000)

		if err != nil {
			t.Fatalf("Deposit() error = %v", err)
		}

		if pot.Balance != 143700 {
			t.Errorf("expected balance 143700, got %d", pot.Balance)
		}
	})
}

func TestPotsService_Withdraw(t *testing.T) {
	t.Run("withdraws from pot successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("expected PUT request, got %s", r.Method)
			}

			if r.URL.Path != "/pots/pot_123/withdraw" {
				t.Errorf("expected path /pots/pot_123/withdraw, got %s", r.URL.Path)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.FormValue("destination_account_id") != "acc_123" {
				t.Error("expected destination_account_id parameter")
			}

			if r.FormValue("amount") != "5000" {
				t.Error("expected amount parameter")
			}

			if r.FormValue("dedupe_id") != "dedupe_456" {
				t.Error("expected dedupe_id parameter")
			}

			resp := Pot{
				ID:       "pot_123",
				Name:     "Savings",
				Balance:  128700,
				Currency: "GBP",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		pot, err := client.Pots.Withdraw(context.Background(), "pot_123", "acc_123", "dedupe_456", 5000)

		if err != nil {
			t.Fatalf("Withdraw() error = %v", err)
		}

		if pot.Balance != 128700 {
			t.Errorf("expected balance 128700, got %d", pot.Balance)
		}
	})
}
