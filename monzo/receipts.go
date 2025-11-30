package monzo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// ReceiptsService handles communication with the receipts related methods of the Monzo API
type ReceiptsService struct {
	client *Client
}

// Receipt represents a transaction receipt
type Receipt struct {
	ID            string           `json:"id,omitempty"`
	TransactionID string           `json:"transaction_id"`
	ExternalID    string           `json:"external_id"`
	Total         int64            `json:"total"`
	Currency      string           `json:"currency"`
	Items         []ReceiptItem    `json:"items"`
	Taxes         []ReceiptTax     `json:"taxes,omitempty"`
	Payments      []ReceiptPayment `json:"payments,omitempty"`
	Merchant      *ReceiptMerchant `json:"merchant,omitempty"`
}

// ReceiptItem represents an item in a receipt
type ReceiptItem struct {
	Description string        `json:"description"`
	Amount      int64         `json:"amount"`
	Currency    string        `json:"currency"`
	Quantity    float64       `json:"quantity,omitempty"`
	Unit        string        `json:"unit,omitempty"`
	Tax         int64         `json:"tax,omitempty"`
	SubItems    []ReceiptItem `json:"sub_items,omitempty"`
}

// ReceiptTax represents tax information in a receipt
type ReceiptTax struct {
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	TaxNumber   string `json:"tax_number,omitempty"`
}

// ReceiptPayment represents a payment method in a receipt
type ReceiptPayment struct {
	Type         string `json:"type"` // card, cash, gift_card
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Bin          string `json:"bin,omitempty"`
	LastFour     string `json:"last_four,omitempty"`
	AuthCode     string `json:"auth_code,omitempty"`
	AID          string `json:"aid,omitempty"`
	MID          string `json:"mid,omitempty"`
	TID          string `json:"tid,omitempty"`
	GiftCardType string `json:"gift_card_type,omitempty"`
}

// ReceiptMerchant represents merchant information in a receipt
type ReceiptMerchant struct {
	Name          string `json:"name,omitempty"`
	Online        bool   `json:"online,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Email         string `json:"email,omitempty"`
	StoreName     string `json:"store_name,omitempty"`
	StoreAddress  string `json:"store_address,omitempty"`
	StorePostcode string `json:"store_postcode,omitempty"`
}

// CreateReceiptResponse represents the response from creating a receipt
type CreateReceiptResponse struct {
	ReceiptID string `json:"receipt_id"`
}

// GetReceiptResponse represents the response from getting a receipt
type GetReceiptResponse struct {
	Receipt Receipt `json:"receipt"`
}

// Create creates or updates a receipt for a transaction
func (s *ReceiptsService) Create(ctx context.Context, receipt *Receipt) (string, error) {
	if receipt == nil || receipt.TransactionID == "" || receipt.ExternalID == "" {
		return "", fmt.Errorf("transaction_id and external_id are required")
	}

	body, err := json.Marshal(receipt)
	if err != nil {
		return "", fmt.Errorf("failed to marshal receipt: %w", err)
	}

	req, err := s.client.newRequest(ctx, "PUT", "/transaction-receipts", nil)
	if err != nil {
		return "", err
	}

	req.Body = io.NopCloser(bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var resp CreateReceiptResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return "", fmt.Errorf("failed to create receipt: %w", err)
	}

	return resp.ReceiptID, nil
}

// Get retrieves a receipt by its external ID
func (s *ReceiptsService) Get(ctx context.Context, externalID string) (*Receipt, error) {
	params := url.Values{}
	params.Set("external_id", externalID)

	path := "/transaction-receipts?" + params.Encode()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetReceiptResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}

	return &resp.Receipt, nil
}

// Delete deletes a receipt by its external ID
func (s *ReceiptsService) Delete(ctx context.Context, externalID string) error {
	params := url.Values{}
	params.Set("external_id", externalID)

	path := "/transaction-receipts?" + params.Encode()

	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	if _, err := s.client.do(req, nil); err != nil {
		return fmt.Errorf("failed to delete receipt: %w", err)
	}

	return nil
}
