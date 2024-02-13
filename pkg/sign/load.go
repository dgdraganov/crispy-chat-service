package sign

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadPrivateKeyFromFile loads a EC private key from file
func LoadPrivateKeyFromFile(filePath string) (*ecdsa.PrivateKey, error) {
	pkBytes, err := os.ReadFile("./private.pem")
	if err != nil {
		return nil, fmt.Errorf("os read file: %w", err)
	}

	pemBlock, _ := pem.Decode(pkBytes)
	privateKey, err := x509.ParseECPrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509 parse private key: %w", err)
	}

	return privateKey, nil
}

func LoadPrivateKey(file string) (*ecdsa.PrivateKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(file)
	if err != nil {
		return nil, fmt.Errorf("base64 url decode: %w", err)
	}

	pemBlock, _ := pem.Decode(decoded)
	privateKey, err := x509.ParseECPrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509 parse private key: %w", err)
	}

	return privateKey, nil
}
