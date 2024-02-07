package sign

type Hasher interface {
	Hash([]byte) []byte
}

type Encoder interface {
	Encode([]byte) string
	Decode(string) ([]byte, error)
}
