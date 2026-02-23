package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveToken(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	token := "test-jwt-token-123"
	if err := SaveToken(token); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	path := filepath.Join(tmp, tokenDir, tokenFile)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read token file: %v", err)
	}
	if string(data) != token {
		t.Errorf("token file content = %q, want %q", string(data), token)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat token file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("token file permissions = %o, want 600", perm)
	}
}

func TestLoadToken(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	want := "saved-jwt-token"
	if err := SaveToken(want); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	got, err := LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}
	if got != want {
		t.Errorf("LoadToken() = %q, want %q", got, want)
	}
}

func TestDeleteToken(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := SaveToken("to-be-deleted"); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	if err := DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}

	path := filepath.Join(tmp, tokenDir, tokenFile)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("token file still exists after DeleteToken()")
	}
}

func TestLoadToken_NoFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	_, err := LoadToken()
	if err == nil {
		t.Error("LoadToken() expected error when no token file exists, got nil")
	}
}
