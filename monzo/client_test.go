package monzo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient(
		WithBaseURL(server.URL),
		WithAccessToken("test_token"),
	)
	return server, client
}

func TestNewClient(t *testing.T) {
	t.Run("creates client with default options", func(t *testing.T) {
		client := NewClient()

		if client.baseURL != DefaultBaseURL {
			t.Errorf("expected baseURL %s, got %s", DefaultBaseURL, client.baseURL)
		}

		if client.httpClient == nil {
			t.Error("expected httpClient to be initialized")
		}

		if client.Auth == nil {
			t.Error("expected Auth service to be initialized")
		}

		if client.Accounts == nil {
			t.Error("expected Accounts service to be initialized")
		}
	})

	t.Run("creates client with custom options", func(t *testing.T) {
		customBaseURL := "https://custom.api.url"
		customToken := "custom_token"

		client := NewClient(
			WithBaseURL(customBaseURL),
			WithAccessToken(customToken),
		)

		if client.baseURL != customBaseURL {
			t.Errorf("expected baseURL %s, got %s", customBaseURL, client.baseURL)
		}

		if client.accessToken != customToken {
			t.Errorf("expected accessToken %s, got %s", customToken, client.accessToken)
		}
	})
}

func TestSetAccessToken(t *testing.T) {
	client := NewClient()
	token := "new_token"

	client.SetAccessToken(token)

	if client.accessToken != token {
		t.Errorf("expected accessToken %s, got %s", token, client.accessToken)
	}
}

func TestCheckResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"200 OK", http.StatusOK, false},
		{"201 Created", http.StatusCreated, false},
		{"204 No Content", http.StatusNoContent, false},
		{"400 Bad Request", http.StatusBadRequest, true},
		{"401 Unauthorized", http.StatusUnauthorized, true},
		{"403 Forbidden", http.StatusForbidden, true},
		{"404 Not Found", http.StatusNotFound, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Status:     http.StatusText(tt.statusCode),
				Body:       http.NoBody,
			}

			err := checkResponse(resp)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	client := NewClient(WithAccessToken("test_token"))

	t.Run("creates request with authorization header", func(t *testing.T) {
		req, err := client.newRequest(context.Background(), "GET", "/test", nil)
		if err != nil {
			t.Fatalf("newRequest() error = %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		expectedAuth := "Bearer test_token"

		if authHeader != expectedAuth {
			t.Errorf("expected Authorization header %s, got %s", expectedAuth, authHeader)
		}
	})

	t.Run("creates request without body", func(t *testing.T) {
		req, err := client.newRequest(context.Background(), "GET", "/test", nil)
		if err != nil {
			t.Fatalf("newRequest() error = %v", err)
		}

		if req.Body != nil {
			t.Error("expected nil body")
		}
	})
}
