package monzo

import (
	"context"
	"net/http"
	"testing"
)

func TestFeedItemsService_Create(t *testing.T) {
	t.Run("creates basic feed item successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}

			if r.URL.Path != "/feed" {
				t.Errorf("expected path /feed, got %s", r.URL.Path)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.FormValue("account_id") != "acc_123" {
				t.Error("expected account_id parameter")
			}

			if r.FormValue("type") != "basic" {
				t.Error("expected type parameter to be basic")
			}

			if r.FormValue("params[title]") != "Test Title" {
				t.Error("expected params[title] parameter")
			}

			if r.FormValue("params[image_url]") != "https://example.com/image.png" {
				t.Error("expected params[image_url] parameter")
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		opts := &CreateFeedItemOptions{
			AccountID: "acc_123",
			Type:      FeedItemTypeBasic,
			URL:       "https://example.com/page",
			Params: BasicFeedItemParams{
				Title:           "Test Title",
				ImageURL:        "https://example.com/image.png",
				Body:            "Test body",
				BackgroundColor: "#FCF1EE",
				TitleColor:      "#333333",
			},
		}

		err := client.FeedItems.Create(context.Background(), opts)

		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	})

	t.Run("requires account_id", func(t *testing.T) {
		client := NewClient()

		opts := &CreateFeedItemOptions{
			Params: BasicFeedItemParams{
				Title:    "Test",
				ImageURL: "https://example.com/image.png",
			},
		}

		err := client.FeedItems.Create(context.Background(), opts)

		if err == nil {
			t.Fatal("expected error when account_id is missing")
		}
	})

	t.Run("requires title and image_url", func(t *testing.T) {
		client := NewClient()

		opts := &CreateFeedItemOptions{
			AccountID: "acc_123",
			Params:    BasicFeedItemParams{},
		}

		err := client.FeedItems.Create(context.Background(), opts)

		if err == nil {
			t.Fatal("expected error when required params are missing")
		}
	})

	t.Run("creates feed item with minimal params", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.FormValue("params[body]") != "" {
				t.Error("expected params[body] to be empty")
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		opts := &CreateFeedItemOptions{
			AccountID: "acc_123",
			Params: BasicFeedItemParams{
				Title:    "Test Title",
				ImageURL: "https://example.com/image.png",
			},
		}

		err := client.FeedItems.Create(context.Background(), opts)

		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	})
}
