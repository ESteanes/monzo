package monzo

import (
	"context"
	"fmt"
	"net/url"
)

// WebhooksService handles communication with the webhooks related methods of the Monzo API
type WebhooksService struct {
	client *Client
}

// Webhook represents a registered webhook
type Webhook struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	URL       string `json:"url"`
}

// RegisterWebhookResponse represents the response from registering a webhook
type RegisterWebhookResponse struct {
	Webhook Webhook `json:"webhook"`
}

// ListWebhooksResponse represents the response from listing webhooks
type ListWebhooksResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

// Register registers a new webhook for the specified account
func (s *WebhooksService) Register(ctx context.Context, accountID, webhookURL string) (*Webhook, error) {
	data := url.Values{}
	data.Set("account_id", accountID)
	data.Set("url", webhookURL)

	req, err := s.client.newRequest(ctx, "POST", "/webhooks", data)
	if err != nil {
		return nil, err
	}

	var resp RegisterWebhookResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to register webhook: %w", err)
	}

	return &resp.Webhook, nil
}

// List returns a list of webhooks registered for the specified account
func (s *WebhooksService) List(ctx context.Context, accountID string) ([]Webhook, error) {
	params := url.Values{}
	params.Set("account_id", accountID)

	path := "/webhooks?" + params.Encode()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListWebhooksResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	return resp.Webhooks, nil
}

// Delete removes a webhook
func (s *WebhooksService) Delete(ctx context.Context, webhookID string) error {
	path := fmt.Sprintf("/webhooks/%s", webhookID)

	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	if _, err := s.client.do(req, nil); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	return nil
}
