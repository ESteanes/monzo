package monzo

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileTokenStore_SaveAndLoad(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	store := NewFileTokenStore(tokenPath)

	// Create test token
	token := &Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	// Save token
	err := store.Save(token)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check file exists
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Error("expected token file to exist")
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
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "nonexistent.json")

	store := NewFileTokenStore(tokenPath)

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
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	store := NewFileTokenStore(tokenPath)

	// Save a token
	token := &Token{
		AccessToken:  "test",
		RefreshToken: "test",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	err := store.Save(token)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Delete token
	err = store.Delete()
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Check file doesn't exist
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Error("expected token file to be deleted")
	}

	// Deleting again should not error
	err = store.Delete()
	if err != nil {
		t.Fatalf("Delete() on nonexistent file error = %v", err)
	}
}

func TestFileTokenStore_Permissions(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	store := NewFileTokenStore(tokenPath)

	token := &Token{
		AccessToken:  "test",
		RefreshToken: "test",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	err := store.Save(token)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check file permissions (should be 0600)
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	mode := info.Mode().Perm()
	expectedMode := os.FileMode(0600)

	if mode != expectedMode {
		t.Errorf("expected file mode %v, got %v", expectedMode, mode)
	}
}

func TestDefaultTokenStore(t *testing.T) {
	store, err := DefaultTokenStore()
	if err != nil {
		t.Fatalf("DefaultTokenStore() error = %v", err)
	}

	if store == nil {
		t.Fatal("expected store to be created")
	}

	if store.filePath == "" {
		t.Error("expected filePath to be set")
	}

	// Should contain .monzo directory
	if !filepath.IsAbs(store.filePath) {
		t.Error("expected absolute path")
	}

	// Clean up test directory if it was created
	dir := filepath.Dir(store.filePath)
	if filepath.Base(dir) == ".monzo" {
		os.RemoveAll(dir)
	}
}

func TestNewFileTokenStore(t *testing.T) {
	path := "/tmp/test_token.json"
	store := NewFileTokenStore(path)

	if store == nil {
		t.Fatal("expected store to be created")
	}

	if store.filePath != path {
		t.Errorf("expected filePath %s, got %s", path, store.filePath)
	}
}
