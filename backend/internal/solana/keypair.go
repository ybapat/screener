package solana

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	solanago "github.com/gagliardetto/solana-go"
)

// LoadOrGenerateKeypair loads a Solana keypair from a JSON file.
// If the file doesn't exist, it generates a new keypair and saves it.
func LoadOrGenerateKeypair(path string) (solanago.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		// File exists — parse the JSON byte array
		var keyBytes []byte
		if err := json.Unmarshal(data, &keyBytes); err != nil {
			return solanago.PrivateKey{}, fmt.Errorf("parse keypair file: %w", err)
		}
		if len(keyBytes) != ed25519.PrivateKeySize {
			return solanago.PrivateKey{}, fmt.Errorf("invalid keypair size: %d", len(keyBytes))
		}
		return solanago.PrivateKey(keyBytes), nil
	}

	if !os.IsNotExist(err) {
		return solanago.PrivateKey{}, fmt.Errorf("read keypair file: %w", err)
	}

	// Generate new keypair
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return solanago.PrivateKey{}, fmt.Errorf("generate keypair: %w", err)
	}

	// Ensure directory exists
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return solanago.PrivateKey{}, fmt.Errorf("create keypair dir: %w", err)
		}
	}

	// Save as JSON byte array (Solana CLI format)
	jsonBytes, err := json.Marshal([]byte(privKey))
	if err != nil {
		return solanago.PrivateKey{}, fmt.Errorf("marshal keypair: %w", err)
	}
	if err := os.WriteFile(path, jsonBytes, 0600); err != nil {
		return solanago.PrivateKey{}, fmt.Errorf("write keypair file: %w", err)
	}

	return solanago.PrivateKey(privKey), nil
}
