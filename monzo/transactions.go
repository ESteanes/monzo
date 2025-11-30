package monzo

import (
	"context"
	"fmt"
	"net/url"
)

// TransactionsService handles communication with the transactions related methods of the Monzo API
type TransactionsService struct {
	client *Client
}

// Transaction represents a Monzo transaction
type Transaction struct {
	ID            string                 `json:"id"`
	Created       string                 `json:"created"`
	Description   string                 `json:"description"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	Merchant      interface{}            `json:"merchant"` // Can be string (ID) or Merchant object if expanded
	Notes         string                 `json:"notes"`
	Metadata      map[string]string      `json:"metadata"`
	IsLoad        bool                   `json:"is_load"`
	Settled       string                 `json:"settled"`
	Category      string                 `json:"category"`
	DeclineReason string                 `json:"decline_reason,omitempty"`
}

// Merchant represents detailed merchant information
type Merchant struct {
	ID       string          `json:"id"`
	GroupID  string          `json:"group_id"`
	Created  string          `json:"created"`
	Name     string          `json:"name"`
	Logo     string          `json:"logo"`
	Emoji    string          `json:"emoji"`
	Category string          `json:"category"`
	Address  MerchantAddress `json:"address"`
}

// MerchantAddress represents a merchant's address
type MerchantAddress struct {
	Address   string  `json:"address"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Postcode  string  `json:"postcode"`
	Region    string  `json:"region"`
}

// GetTransactionResponse represents the response from the get transaction endpoint
type GetTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

// ListTransactionsResponse represents the response from the list transactions endpoint
type ListTransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

// ListOptions specifies optional parameters for listing transactions
type ListOptions struct {
	AccountID string
	Since     string // RFC3339 encoded timestamp
	Before    string // RFC3339 encoded timestamp
	Limit     int
	Expand    []string // e.g., "merchant"
}

// Get retrieves a single transaction by ID
func (s *TransactionsService) Get(ctx context.Context, transactionID string, expand []string) (*Transaction, error) {
	path := fmt.Sprintf("/transactions/%s", transactionID)

	if len(expand) > 0 {
		params := url.Values{}
		for _, e := range expand {
			params.Add("expand[]", e)
		}
		path += "?" + params.Encode()
	}

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetTransactionResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &resp.Transaction, nil
}

// List returns a list of transactions for the specified account
func (s *TransactionsService) List(ctx context.Context, opts *ListOptions) ([]Transaction, error) {
	if opts == nil || opts.AccountID == "" {
		return nil, fmt.Errorf("account_id is required")
	}

	params := url.Values{}
	params.Set("account_id", opts.AccountID)

	if opts.Since != "" {
		params.Set("since", opts.Since)
	}
	if opts.Before != "" {
		params.Set("before", opts.Before)
	}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	for _, e := range opts.Expand {
		params.Add("expand[]", e)
	}

	path := "/transactions?" + params.Encode()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListTransactionsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	return resp.Transactions, nil
}

// Annotate updates the metadata for a transaction
func (s *TransactionsService) Annotate(ctx context.Context, transactionID string, metadata map[string]string) (*Transaction, error) {
	data := url.Values{}
	for key, value := range metadata {
		data.Set(fmt.Sprintf("metadata[%s]", key), value)
	}

	path := fmt.Sprintf("/transactions/%s", transactionID)

	req, err := s.client.newRequest(ctx, "PATCH", path, data)
	if err != nil {
		return nil, err
	}

	var resp GetTransactionResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to annotate transaction: %w", err)
	}

	return &resp.Transaction, nil
}
