package sign

import "crypto/sha256"

type SHA256Hasher struct {
}

func (hasher *SHA256Hasher) Hash(data []byte) []byte {
	hashCode := sha256.Sum256(data)
	return hashCode[:]
}
