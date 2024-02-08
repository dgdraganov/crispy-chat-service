package sign

import "crypto/sha256"

type sha256Hasher struct {
}

func NewSHA256Hasher() *sha256Hasher {
	return &sha256Hasher{}
}
func (hasher *sha256Hasher) Hash(data []byte) []byte {
	hashCode := sha256.Sum256(data)
	return hashCode[:]
}
