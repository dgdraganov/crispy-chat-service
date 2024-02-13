package sign

import "crypto/sha256"

type sha256Hasher struct {
}

// NewSHA256Hasher is a constructor function for the sha256Hasher type
func NewSHA256Hasher() *sha256Hasher {
	return &sha256Hasher{}
}

// Hash generates a hash code using Sha256 algo
func (hasher *sha256Hasher) Hash(data []byte) []byte {
	hashCode := sha256.Sum256(data)
	return hashCode[:]
}
