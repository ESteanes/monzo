package monzo

import (
	"context"
	"fmt"
	"net/url"
)

// FeedItemsService handles communication with the feed items related methods of the Monzo API
type FeedItemsService struct {
	client *Client
}

// FeedItemType represents the type of feed item
type FeedItemType string

const (
	// FeedItemTypeBasic represents a basic feed item
	FeedItemTypeBasic FeedItemType = "basic"
)

// BasicFeedItemParams represents parameters for creating a basic feed item
type BasicFeedItemParams struct {
	Title           string
	ImageURL        string
	Body            string
	BackgroundColor string
	TitleColor      string
	BodyColor       string
}

// CreateFeedItemOptions represents options for creating a feed item
type CreateFeedItemOptions struct {
	AccountID string
	Type      FeedItemType
	URL       string
	Params    BasicFeedItemParams
}

// Create creates a new feed item on the user's feed
func (s *FeedItemsService) Create(ctx context.Context, opts *CreateFeedItemOptions) error {
	if opts == nil || opts.AccountID == "" {
		return fmt.Errorf("account_id is required")
	}
	if opts.Type == "" {
		opts.Type = FeedItemTypeBasic
	}
	if opts.Params.Title == "" || opts.Params.ImageURL == "" {
		return fmt.Errorf("title and image_url are required")
	}

	data := url.Values{}
	data.Set("account_id", opts.AccountID)
	data.Set("type", string(opts.Type))

	if opts.URL != "" {
		data.Set("url", opts.URL)
	}

	// Add params based on type
	if opts.Type == FeedItemTypeBasic {
		data.Set("params[title]", opts.Params.Title)
		data.Set("params[image_url]", opts.Params.ImageURL)

		if opts.Params.Body != "" {
			data.Set("params[body]", opts.Params.Body)
		}
		if opts.Params.BackgroundColor != "" {
			data.Set("params[background_color]", opts.Params.BackgroundColor)
		}
		if opts.Params.TitleColor != "" {
			data.Set("params[title_color]", opts.Params.TitleColor)
		}
		if opts.Params.BodyColor != "" {
			data.Set("params[body_color]", opts.Params.BodyColor)
		}
	}

	req, err := s.client.newRequest(ctx, "POST", "/feed", data)
	if err != nil {
		return err
	}

	if _, err := s.client.do(req, nil); err != nil {
		return fmt.Errorf("failed to create feed item: %w", err)
	}

	return nil
}
