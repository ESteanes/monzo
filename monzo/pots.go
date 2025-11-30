package monzo

import (
	"context"
	"fmt"
	"net/url"
)

// PotsService handles communication with the pots related methods of the Monzo API
type PotsService struct {
	client *Client
}

// Pot represents a Monzo pot
type Pot struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Style    string `json:"style"`
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
	Deleted  bool   `json:"deleted"`
}

// ListPotsResponse represents the response from the list pots endpoint
type ListPotsResponse struct {
	Pots []Pot `json:"pots"`
}

// List returns a list of pots for the specified account
func (s *PotsService) List(ctx context.Context, accountID string) ([]Pot, error) {
	params := url.Values{}
	params.Set("current_account_id", accountID)

	path := "/pots?" + params.Encode()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListPotsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to list pots: %w", err)
	}

	return resp.Pots, nil
}

// Deposit moves money from an account into a pot
func (s *PotsService) Deposit(ctx context.Context, potID, sourceAccountID, dedupeID string, amount int64) (*Pot, error) {
	data := url.Values{}
	data.Set("source_account_id", sourceAccountID)
	data.Set("amount", fmt.Sprintf("%d", amount))
	data.Set("dedupe_id", dedupeID)

	path := fmt.Sprintf("/pots/%s/deposit", potID)

	req, err := s.client.newRequest(ctx, "PUT", path, data)
	if err != nil {
		return nil, err
	}

	var pot Pot
	if _, err := s.client.do(req, &pot); err != nil {
		return nil, fmt.Errorf("failed to deposit into pot: %w", err)
	}

	return &pot, nil
}

// Withdraw moves money from a pot into an account
func (s *PotsService) Withdraw(ctx context.Context, potID, destinationAccountID, dedupeID string, amount int64) (*Pot, error) {
	data := url.Values{}
	data.Set("destination_account_id", destinationAccountID)
	data.Set("amount", fmt.Sprintf("%d", amount))
	data.Set("dedupe_id", dedupeID)

	path := fmt.Sprintf("/pots/%s/withdraw", potID)

	req, err := s.client.newRequest(ctx, "PUT", path, data)
	if err != nil {
		return nil, err
	}

	var pot Pot
	if _, err := s.client.do(req, &pot); err != nil {
		return nil, fmt.Errorf("failed to withdraw from pot: %w", err)
	}

	return &pot, nil
}
