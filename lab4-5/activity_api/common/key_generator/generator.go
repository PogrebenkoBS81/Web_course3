package key_generator

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// GenerateKey - generates RSA public and private keys that will be used in auth.
func GenerateKey() (string, string, error) {
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
