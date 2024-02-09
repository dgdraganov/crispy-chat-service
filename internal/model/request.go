package model

type AuthRequest struct {
	ClientID string `json:"client_id"`
}
type PushRequest struct {
	ClientID string `json:"client_id"`
	Message  string `json:"message"`
}

type ListenRequest struct {
	ClientID string `json:"client_id"`
}
