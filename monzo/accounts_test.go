package monzo

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAccountsService_List(t *testing.T) {
	t.Run("lists all accounts", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET request, got %s", r.Method)
			}

			if r.URL.Path != "/accounts" {
				t.Errorf("expected path /accounts, got %s", r.URL.Path)
			}

			resp := ListAccountsResponse{
				Accounts: []Account{
					{
						ID:          "acc_00009237aqC8c5umZmrRdh",
						Description: "Peter Pan's Account",
						Created:     "2015-11-13T12:17:42Z",
					},
					{
						ID:          "acc_00009237aqC8c5umZmrRd2",
						Description: "Joint Account",
						Created:     "2016-01-15T10:20:30Z",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		accounts, err := client.Accounts.List(context.Background(), nil)

		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(accounts) != 2 {
			t.Errorf("expected 2 accounts, got %d", len(accounts))
		}

		if accounts[0].ID != "acc_00009237aqC8c5umZmrRdh" {
			t.Errorf("expected account ID acc_00009237aqC8c5umZmrRdh, got %s", accounts[0].ID)
		}

		if accounts[0].Description != "Peter Pan's Account" {
			t.Errorf("expected description Peter Pan's Account, got %s", accounts[0].Description)
		}
	})

	t.Run("lists accounts with account type filter", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.RawQuery, "account_type=uk_retail") {
				t.Error("expected account_type parameter in query")
			}

			resp := ListAccountsResponse{
				Accounts: []Account{
					{
						ID:          "acc_00009237aqC8c5umZmrRdh",
						Description: "Retail Account",
						Created:     "2015-11-13T12:17:42Z",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		opts := &ListAccountsOptions{
			AccountType: AccountTypeUKRetail,
		}

		accounts, err := client.Accounts.List(context.Background(), opts)

		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(accounts) != 1 {
			t.Errorf("expected 1 account, got %d", len(accounts))
		}
	})

	t.Run("handles empty account list", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			resp := ListAccountsResponse{
				Accounts: []Account{},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		accounts, err := client.Accounts.List(context.Background(), nil)

		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(accounts) != 0 {
			t.Errorf("expected 0 accounts, got %d", len(accounts))
		}
	})

	t.Run("handles error response", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Unauthorized"}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		_, err := client.Accounts.List(context.Background(), nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
