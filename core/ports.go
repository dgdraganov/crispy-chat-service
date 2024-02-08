package core

type DSigner interface {
	Sign(message string) (string, error)
	Verify(signature, message string) (bool, error)
}
