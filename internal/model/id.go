package model

type contextKey int

const (
	RequestID contextKey = iota
	Signature
	ClientID
)
