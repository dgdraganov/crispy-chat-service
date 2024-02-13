package sign

import (
	"encoding/base64"
	"fmt"
)

type base64Encoder struct {
}

// NewBase64Encoder is a constructor function for the base64Encoder type
func NewBase64Encoder() *base64Encoder {
	return &base64Encoder{}
}

// Encode will URL encode a message
func (enc *base64Encoder) Encode(byteData []byte) string {
	return base64.URLEncoding.EncodeToString(byteData)
}

// Decode will URL decode a message
func (enc *base64Encoder) Decode(encodedData string) ([]byte, error) {
	decoded, err := base64.URLEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("base64 url decode: %w", err)
	}
	return decoded, err
}
