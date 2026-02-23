package client

import (
	"os"
	"path/filepath"
)

const tokenDir = ".config/chat"
const tokenFile = "token"

func tokenPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, tokenDir, tokenFile)
}

// SaveToken writes the JWT token to disk.
func SaveToken(token string) error {
	p := tokenPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(token), 0o600)
}

// LoadToken reads the stored JWT token from disk.
func LoadToken() (string, error) {
	data, err := os.ReadFile(tokenPath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeleteToken removes the stored token file.
func DeleteToken() error {
	return os.Remove(tokenPath())
}
