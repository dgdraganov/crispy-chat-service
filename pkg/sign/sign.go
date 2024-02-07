package sign

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
)

type ECDSA struct {
	privateKey *ecdsa.PrivateKey
	hasher     Hasher
	encoder    Encoder
}

func NewECDSA(privKey *ecdsa.PrivateKey) *ECDSA {
	return &ECDSA{
		privateKey: privKey,
	}
}

func (ec *ECDSA) Sign(message string) (string, error) {
	hash := ec.hasher.Hash([]byte(message))
	signatureBytes, err := ecdsa.SignASN1(rand.Reader, ec.privateKey, hash)
	if err != nil {
		return "", fmt.Errorf("ecdsa sign: %w", err)
	}
	signature := ec.encoder.Encode(signatureBytes)
	return signature, nil
}

func (ec *ECDSA) Verify(signature, message string) (bool, error) {
	signatureBytes, err := ec.encoder.Decode(signature)
	if err != nil {
		return false, fmt.Errorf("signature string decode: %w", err)
	}

	hash := ec.hasher.Hash([]byte(message))

	valid := ecdsa.VerifyASN1(&ec.privateKey.PublicKey, hash, signatureBytes)
	return valid, nil
}
