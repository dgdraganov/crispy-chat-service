package router

import "net/http"

type chatRouter struct {
	mux *http.ServeMux
}

// New is a constructor function for the chatRouter type
func New() *chatRouter {
	serveMux := http.NewServeMux()
	return &chatRouter{
		mux: serveMux,
	}
}

// Register is used to register the given handler to the underlying serve mux
func (router *chatRouter) Register(pattern string, handler http.Handler) {
	router.mux.Handle(pattern, handler)
}

// ServeMux is used to return the underlying serve mux
func (router *chatRouter) ServeMux() *http.ServeMux {
	return router.mux
}
