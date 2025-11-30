package monzo

import (
	"context"
	"fmt"
	"net/url"
)

// AttachmentsService handles communication with the attachments related methods of the Monzo API
type AttachmentsService struct {
	client *Client
}

// Attachment represents a file attachment on a transaction
type Attachment struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	ExternalID string `json:"external_id"`
	FileURL    string `json:"file_url"`
	FileType   string `json:"file_type"`
	Created    string `json:"created"`
}

// UploadAttachmentResponse represents the response from requesting an upload URL
type UploadAttachmentResponse struct {
	FileURL   string `json:"file_url"`
	UploadURL string `json:"upload_url"`
}

// RegisterAttachmentResponse represents the response from registering an attachment
type RegisterAttachmentResponse struct {
	Attachment Attachment `json:"attachment"`
}

// Upload requests a temporary URL to upload an attachment
func (s *AttachmentsService) Upload(ctx context.Context, fileName, fileType string, contentLength int64) (*UploadAttachmentResponse, error) {
	data := url.Values{}
	data.Set("file_name", fileName)
	data.Set("file_type", fileType)
	data.Set("content_length", fmt.Sprintf("%d", contentLength))

	req, err := s.client.newRequest(ctx, "POST", "/attachment/upload", data)
	if err != nil {
		return nil, err
	}

	var resp UploadAttachmentResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get upload URL: %w", err)
	}

	return &resp, nil
}

// Register registers an attachment against a transaction
func (s *AttachmentsService) Register(ctx context.Context, externalID, fileURL, fileType string) (*Attachment, error) {
	data := url.Values{}
	data.Set("external_id", externalID)
	data.Set("file_url", fileURL)
	data.Set("file_type", fileType)

	req, err := s.client.newRequest(ctx, "POST", "/attachment/register", data)
	if err != nil {
		return nil, err
	}

	var resp RegisterAttachmentResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to register attachment: %w", err)
	}

	return &resp.Attachment, nil
}

// Deregister removes an attachment from a transaction
func (s *AttachmentsService) Deregister(ctx context.Context, attachmentID string) error {
	data := url.Values{}
	data.Set("id", attachmentID)

	req, err := s.client.newRequest(ctx, "POST", "/attachment/deregister", data)
	if err != nil {
		return err
	}

	if _, err := s.client.do(req, nil); err != nil {
		return fmt.Errorf("failed to deregister attachment: %w", err)
	}

	return nil
}
