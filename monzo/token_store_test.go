package monzo

import (
	"testing"
	"time"
)

func TestFileTokenStore_SaveAndLoad(t *testing.T) {
	store, err := DefaultTokenStore()
	if err != nil {
		t.Fatalf("Unable to create default token store.")
	}
	// Create test token
	token := &Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	// Save token
	err = store.Save(token)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load token
	loadedToken, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loadedToken == nil {
		t.Fatal("expected token to be loaded")
	}

	if loadedToken.AccessToken != token.AccessToken {
		t.Errorf("expected access_token %s, got %s", token.AccessToken, loadedToken.AccessToken)
	}

	if loadedToken.RefreshToken != token.RefreshToken {
		t.Errorf("expected refresh_token %s, got %s", token.RefreshToken, loadedToken.RefreshToken)
	}

	// Check expiry time (allow 1 second difference due to serialization)
	diff := loadedToken.ExpiresAt.Sub(token.ExpiresAt)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("expected expires_at to match within 1 second, diff = %v", diff)
	}
}

func TestFileTokenStore_LoadNonexistent(t *testing.T) {
	store, err := DefaultTokenStore()

	// Load should return nil, nil for nonexistent file
	token, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if token != nil {
		t.Error("expected nil token for nonexistent file")
	}
}

func TestFileTokenStore_Delete(t *testing.T) {
	store, err := DefaultTokenStore()

	// Save a token
	token := &Token{
		AccessToken:  "test",
		RefreshToken: "test",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	err = store.Save(token)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Delete token
	err = store.Delete()
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Deleting again should not error
	err = store.Delete()
	if err != nil {
		t.Fatalf("Delete() on nonexistent file error = %v", err)
	}

	// Check that the value was actually deleted
	expectedAccessToken := ""
	expectedRefreshToken := ""
	expectedExpiresAtTime := time.Unix(0, 0)
	finalToken, err := store.Load()
	if err != nil {
		t.Fatalf("unable to load token, error = %v", err)
	}
	if expectedAccessToken != finalToken.AccessToken {
		t.Fatalf("access token not wiped")
	}
	if expectedRefreshToken != finalToken.RefreshToken {
		t.Fatalf("refresh token not wiped")
	}
	if expectedExpiresAtTime != finalToken.ExpiresAt {
		t.Fatalf("expires at time not wiped")
	}
}
