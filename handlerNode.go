package gocelot

import (
	"net/http"
)

// handlerNode represents the method/handler relationship.
type handlerNode struct {
	method string
	handler http.Handler
}

// newHandlerNode returns a new handlerNode for the given method and handler.
func newHandlerNode(method string, handler http.Handler) *handlerNode {
	return &handlerNode{method, handler}
}
