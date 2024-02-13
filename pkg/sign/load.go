package sign

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadPrivateKey loads a EC private key
func LoadPrivateKey(filePath string) (*ecdsa.PrivateKey, error) {
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
