package control

import (
	"activity_api/data_manager/cache"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const pingersNum = 2

// iManageable - interface for service control.
// If service is unavailable, pingers will try to recover service via Open\Close functions.
type iManageable interface {
	Describe() string
	Restart() error
	OK() error
}

// AAServiceConfig - config for AAService
type AAServiceConfig struct {
	DbType     int    // db type, 0 - SQLite
	ConnString string // Conn string to DB
	Addr       string // Addr of service to listen

	LogLevel uint32 // Log level for logrus
	LogFile  string // File to log in // temporary unused

	Cache *cache.ICacheConfig // Config for cache manager
}

// generateKey - generates RSA public and private keys that will be used in auth.
func generateKey() (string, string, error) {
	private, err := rsa.GenerateKey(rand.Reader, 2048) // Generate rsa private key

	if err != nil {
		return "", "", err
	}
	// Create buffer for public key
	publicOut := bytes.NewBuffer(nil)
	// Write public pem key from private key with PKCS1
	if err := pem.Encode(
		publicOut,
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&private.PublicKey),
		},
	); err != nil {
		return "", "", err
	}

	// Create buffer for private key
	privateOut := bytes.NewBuffer(nil)
	// Write private pem key from private key with PKCS1
	if err = pem.Encode(
		privateOut,
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(private),
		},
	); err != nil {
		return "", "", err
	}
	// Return keys as strings
	return privateOut.String(), publicOut.String(), nil
}

// Set keys to env of the project
func setKeys() error {
	private, public, err := generateKey()

	if err != nil {
		return fmt.Errorf("generateKey(): %w", err)
	}
	// Keys sets only for service, and they exists only while service is alive.
	if err := os.Setenv("TOKEN_PRIVATE", private); err != nil {
		return fmt.Errorf("private os.Setenv(): %w", err)
	}

	if err := os.Setenv("TOKEN_PUBLIC", public); err != nil {
		return fmt.Errorf("public os.Setenv(): %w", err)
	}

	return nil
}

func init() {
	if err := setKeys(); err != nil {
		panic(err)
	}
}
