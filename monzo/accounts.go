package monzo

import (
	"context"
	"fmt"
	"net/url"
)

// AccountsService handles communication with the accounts related methods of the Monzo API
type AccountsService struct {
	client *Client
}

// AccountType represents the type of Monzo account
type AccountType string

const (
	// AccountTypeUKRetail represents a UK retail account
	AccountTypeUKRetail AccountType = "uk_retail"
	// AccountTypeUKRetailJoint represents a UK retail joint account
	AccountTypeUKRetailJoint AccountType = "uk_retail_joint"
)

// Account represents a Monzo account
type Account struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Created     string `json:"created"`
}

// ListAccountsResponse represents the response from the list accounts endpoint
type ListAccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

// ListAccountsOptions specifies optional parameters for listing accounts
type ListAccountsOptions struct {
	AccountType AccountType
}

// List returns a list of accounts owned by the currently authorized user
func (s *AccountsService) List(ctx context.Context, opts *ListAccountsOptions) ([]Account, error) {
	path := "/accounts"

	if opts != nil && opts.AccountType != "" {
		params := url.Values{}
		params.Set("account_type", string(opts.AccountType))
		path += "?" + params.Encode()
	}

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListAccountsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	return resp.Accounts, nil
}
