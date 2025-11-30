package monzo

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestWebhooksService_Register(t *testing.T) {
	t.Run("registers webhook successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}

			if r.URL.Path != "/webhooks" {
				t.Errorf("expected path /webhooks, got %s", r.URL.Path)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.FormValue("account_id") != "acc_123" {
				t.Error("expected account_id parameter")
			}

			if r.FormValue("url") != "https://example.com/webhook" {
				t.Error("expected url parameter")
			}

			resp := RegisterWebhookResponse{
				Webhook: Webhook{
					ID:        "webhook_123",
					AccountID: "acc_123",
					URL:       "https://example.com/webhook",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		webhook, err := client.Webhooks.Register(
			context.Background(),
			"acc_123",
			"https://example.com/webhook",
		)

		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		if webhook.ID != "webhook_123" {
			t.Errorf("expected ID webhook_123, got %s", webhook.ID)
		}

		if webhook.URL != "https://example.com/webhook" {
			t.Errorf("expected URL https://example.com/webhook, got %s", webhook.URL)
		}
	})
}

func TestWebhooksService_List(t *testing.T) {
	t.Run("lists webhooks successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET request, got %s", r.Method)
			}

			if r.URL.Path != "/webhooks" {
				t.Errorf("expected path /webhooks, got %s", r.URL.Path)
			}

			if !strings.Contains(r.URL.RawQuery, "account_id=acc_123") {
				t.Error("expected account_id parameter in query")
			}

			resp := ListWebhooksResponse{
				Webhooks: []Webhook{
					{
						ID:        "webhook_1",
						AccountID: "acc_123",
						URL:       "https://example.com/webhook1",
					},
					{
						ID:        "webhook_2",
						AccountID: "acc_123",
						URL:       "https://example.com/webhook2",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		webhooks, err := client.Webhooks.List(context.Background(), "acc_123")

		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(webhooks) != 2 {
			t.Errorf("expected 2 webhooks, got %d", len(webhooks))
		}

		if webhooks[0].ID != "webhook_1" {
			t.Errorf("expected ID webhook_1, got %s", webhooks[0].ID)
		}
	})
}

func TestWebhooksService_Delete(t *testing.T) {
	t.Run("deletes webhook successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE request, got %s", r.Method)
			}

			if r.URL.Path != "/webhooks/webhook_123" {
				t.Errorf("expected path /webhooks/webhook_123, got %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		err := client.Webhooks.Delete(context.Background(), "webhook_123")

		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
	})

	t.Run("handles error when deleting webhook", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Webhook not found"}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		err := client.Webhooks.Delete(context.Background(), "invalid_webhook")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
