package gocelot

import (
	"testing"
)

func TestNewHandlerNodeCreatesHandlerNodeWithMethodAndHandler(t *testing.T) {
	method := "GET"
	node := newHandlerNode(method, emptyHandler)
	if node == nil || node.handler != emptyHandler || node.method != method {
		t.Fail()
	}
}
