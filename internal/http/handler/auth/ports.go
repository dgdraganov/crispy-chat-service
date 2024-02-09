package auth

type Authenticator interface {
	Sign(clientID string) (string, error)
}
