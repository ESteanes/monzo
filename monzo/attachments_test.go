package monzo

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAttachmentsService_Upload(t *testing.T) {
	t.Run("gets upload URL successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}

			if r.URL.Path != "/attachment/upload" {
				t.Errorf("expected path /attachment/upload, got %s", r.URL.Path)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.FormValue("file_name") != "receipt.png" {
				t.Error("expected file_name parameter")
			}

			if r.FormValue("file_type") != "image/png" {
				t.Error("expected file_type parameter")
			}

			if r.FormValue("content_length") != "12345" {
				t.Error("expected content_length parameter")
			}

			resp := UploadAttachmentResponse{
				FileURL:   "https://s3.amazonaws.com/file.png",
				UploadURL: "https://s3.amazonaws.com/upload?signature=xyz",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		uploadResp, err := client.Attachments.Upload(
			context.Background(),
			"receipt.png",
			"image/png",
			12345,
		)

		if err != nil {
			t.Fatalf("Upload() error = %v", err)
		}

		if uploadResp.FileURL != "https://s3.amazonaws.com/file.png" {
			t.Errorf("expected FileURL https://s3.amazonaws.com/file.png, got %s", uploadResp.FileURL)
		}

		if uploadResp.UploadURL == "" {
			t.Error("expected UploadURL to be set")
		}
	})
}

func TestAttachmentsService_Register(t *testing.T) {
	t.Run("registers attachment successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}

			if r.URL.Path != "/attachment/register" {
				t.Errorf("expected path /attachment/register, got %s", r.URL.Path)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.FormValue("external_id") != "tx_123" {
				t.Error("expected external_id parameter")
			}

			if r.FormValue("file_url") != "https://s3.amazonaws.com/file.png" {
				t.Error("expected file_url parameter")
			}

			if r.FormValue("file_type") != "image/png" {
				t.Error("expected file_type parameter")
			}

			resp := RegisterAttachmentResponse{
				Attachment: Attachment{
					ID:         "attach_123",
					UserID:     "user_123",
					ExternalID: "tx_123",
					FileURL:    "https://s3.amazonaws.com/file.png",
					FileType:   "image/png",
					Created:    "2015-11-12T18:37:02Z",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		attachment, err := client.Attachments.Register(
			context.Background(),
			"tx_123",
			"https://s3.amazonaws.com/file.png",
			"image/png",
		)

		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		if attachment.ID != "attach_123" {
			t.Errorf("expected ID attach_123, got %s", attachment.ID)
		}

		if attachment.ExternalID != "tx_123" {
			t.Errorf("expected ExternalID tx_123, got %s", attachment.ExternalID)
		}
	})
}

func TestAttachmentsService_Deregister(t *testing.T) {
	t.Run("deregisters attachment successfully", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}

			if r.URL.Path != "/attachment/deregister" {
				t.Errorf("expected path /attachment/deregister, got %s", r.URL.Path)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.FormValue("id") != "attach_123" {
				t.Error("expected id parameter")
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		err := client.Attachments.Deregister(context.Background(), "attach_123")

		if err != nil {
			t.Fatalf("Deregister() error = %v", err)
		}
	})

	t.Run("handles error when deregistering", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Attachment not found"}`))
		}

		server, client := setupTestServer(handler)
		defer server.Close()

		err := client.Attachments.Deregister(context.Background(), "invalid_attach")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
