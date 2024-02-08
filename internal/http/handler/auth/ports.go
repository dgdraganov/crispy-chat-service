package auth

type Authenticator interface {
	Authenticate(clientID string) (string, error)
}
