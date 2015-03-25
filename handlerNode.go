package gocelot

import (
	"net/http"
)

type handlerNode struct {
	method string
	handler http.Handler
}

func newHandlerNode(method string, handler http.Handler) *handlerNode {
	return &handlerNode{method, handler}
}