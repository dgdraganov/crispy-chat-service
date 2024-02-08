package sign

import (
	"encoding/base64"
	"fmt"
)

type base64Encoder struct {
}

func NewBase64Encoder() *base64Encoder {
	return &base64Encoder{}
}

func (enc *base64Encoder) Encode(byteData []byte) string {
	return base64.URLEncoding.EncodeToString(byteData)
}

func (enc *base64Encoder) Decode(encodedData string) ([]byte, error) {
	decoded, err := base64.URLEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("base64 url decode: %w", err)
	}
	return decoded, err
}
