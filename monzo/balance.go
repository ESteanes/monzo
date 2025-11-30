package monzo

import (
	"context"
	"fmt"
	"net/url"
)

// BalanceService handles communication with the balance related methods of the Monzo API
type BalanceService struct {
	client *Client
}

// Balance represents the balance information for an account
type Balance struct {
	Balance      int64  `json:"balance"`
	TotalBalance int64  `json:"total_balance"`
	Currency     string `json:"currency"`
	SpendToday   int64  `json:"spend_today"`
}

// Get returns balance information for a specific account
func (s *BalanceService) Get(ctx context.Context, accountID string) (*Balance, error) {
	params := url.Values{}
	params.Set("account_id", accountID)

	path := "/balance?" + params.Encode()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var balance Balance
	if _, err := s.client.do(req, &balance); err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &balance, nil
}
