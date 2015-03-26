// Copyright 2015 Piotr Galar. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

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