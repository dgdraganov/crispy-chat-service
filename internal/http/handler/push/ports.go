package push

import "context"

type Publisher interface {
	Verify(signature, clientID string) (bool, error)
	Publish(ctx context.Context, clientID, message string) error
}
