package monzo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReceiptsService_Create(t *testing.T) {
	t.Run("creates receipt successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("expected PUT request, got %s", r.Method)
			}

			if r.URL.Path != "/transaction-receipts" {
				t.Errorf("expected path /transaction-receipts, got %s", r.URL.Path)
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}

			var receipt Receipt
			if err := json.Unmarshal(body, &receipt); err != nil {
				t.Fatalf("failed to unmarshal body: %v", err)
			}

			if receipt.TransactionID != "tx_123" {
				t.Error("expected transaction_id in body")
			}

			if receipt.ExternalID != "ext_123" {
				t.Error("expected external_id in body")
			}

			resp := CreateReceiptResponse{
				ReceiptID: "receipt_123",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		receipt := &Receipt{
			TransactionID: "tx_123",
			ExternalID:    "ext_123",
			Total:         1299,
			Currency:      "GBP",
			Items: []ReceiptItem{
				{
					Description: "Burger",
					Amount:      1299,
					Currency:    "GBP",
					Quantity:    1,
				},
			},
		}

		receiptID, err := client.Receipts.Create(context.Background(), receipt)

		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if receiptID != "receipt_123" {
			t.Errorf("expected receipt_id receipt_123, got %s", receiptID)
		}
	})

	t.Run("requires transaction_id and external_id", func(t *testing.T) {
		client := NewClient()

		receipt := &Receipt{
			Total:    1299,
			Currency: "GBP",
		}

		_, err := client.Receipts.Create(context.Background(), receipt)

		if err == nil {
			t.Fatal("expected error when required fields are missing")
		}
	})

	t.Run("creates receipt with all fields", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}

			var receipt Receipt
			if err := json.Unmarshal(body, &receipt); err != nil {
				t.Fatalf("failed to unmarshal body: %v", err)
			}

			if len(receipt.Items) != 1 {
				t.Error("expected 1 item")
			}

			if len(receipt.Taxes) != 1 {
				t.Error("expected 1 tax")
			}

			if len(receipt.Payments) != 1 {
				t.Error("expected 1 payment")
			}

			resp := CreateReceiptResponse{
				ReceiptID: "receipt_123",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		receipt := &Receipt{
			TransactionID: "tx_123",
			ExternalID:    "ext_123",
			Total:         1299,
			Currency:      "GBP",
			Items: []ReceiptItem{
				{
					Description: "Burger",
					Amount:      1299,
					Currency:    "GBP",
				},
			},
			Taxes: []ReceiptTax{
				{
					Description: "VAT",
					Amount:      10,
					Currency:    "GBP",
				},
			},
			Payments: []ReceiptPayment{
				{
					Type:     "card",
					Amount:   1299,
					Currency: "GBP",
					LastFour: "1234",
				},
			},
		}

		_, err := client.Receipts.Create(context.Background(), receipt)

		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	})
}

func TestReceiptsService_Get(t *testing.T) {
	t.Run("gets receipt successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET request, got %s", r.Method)
			}

			if r.URL.Path != "/transaction-receipts" {
				t.Errorf("expected path /transaction-receipts, got %s", r.URL.Path)
			}

			if !strings.Contains(r.URL.RawQuery, "external_id=ext_123") {
				t.Error("expected external_id parameter in query")
			}

			resp := GetReceiptResponse{
				Receipt: Receipt{
					ID:            "receipt_123",
					TransactionID: "tx_123",
					ExternalID:    "ext_123",
					Total:         1299,
					Currency:      "GBP",
					Items: []ReceiptItem{
						{
							Description: "Burger",
							Amount:      1299,
							Currency:    "GBP",
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		receipt, err := client.Receipts.Get(context.Background(), "ext_123")

		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if receipt.ExternalID != "ext_123" {
			t.Errorf("expected external_id ext_123, got %s", receipt.ExternalID)
		}

		if receipt.Total != 1299 {
			t.Errorf("expected total 1299, got %d", receipt.Total)
		}
	})
}

func TestReceiptsService_Delete(t *testing.T) {
	t.Run("deletes receipt successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE request, got %s", r.Method)
			}

			if r.URL.Path != "/transaction-receipts" {
				t.Errorf("expected path /transaction-receipts, got %s", r.URL.Path)
			}

			if !strings.Contains(r.URL.RawQuery, "external_id=ext_123") {
				t.Error("expected external_id parameter in query")
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		err := client.Receipts.Delete(context.Background(), "ext_123")

		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
	})

	t.Run("handles error when deleting", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Receipt not found"}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		err := client.Receipts.Delete(context.Background(), "invalid_ext")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
