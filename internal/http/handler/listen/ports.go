package listen

import "context"

type Listener interface {
	Verify(signature, clientID string) (bool, error)
	ReadMessages(ctx context.Context, clientID string) <-chan string
}
