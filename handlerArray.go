package gocelot

import (
	"net/http"
)

type handlerArray struct {
	nodes []*handlerNode
}

func newHandlerArray() *handlerArray {
	return &handlerArray{}
}

func (ha *handlerArray) get(method string) http.Handler {
	for _, node := range ha.nodes {
		if node.method == method {
			return node.handler
		}
	}
	return nil
}

func (ha *handlerArray) add(method string, handler http.Handler) {
	if ha.get(method) == nil {
		ha.nodes = append(ha.nodes, newHandlerNode(method, handler))
	}
}