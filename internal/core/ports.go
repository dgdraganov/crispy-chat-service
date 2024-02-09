package core

import "context"

type DigitalSigner interface {
	Sign(message string) (string, error)
	Verify(signature, message string) (bool, error)
}
type DBClient interface {
	PublishMessage(room string, sortKey float64, message any) (int64, error)
	ReadMessages(ctx context.Context, room string) (<-chan string, <-chan error)
}
