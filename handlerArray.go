// Copyright 2015 Piotr Galar. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package gocelot

import (
	"net/http"
)

// handlerArray holds an array of handlerNodes to represent method/handler
// relationship.
type handlerArray struct {
	nodes []*handlerNode
}

// newHandlerArray returns an empty handlerArray
func newHandlerArray() *handlerArray {
	return &handlerArray{}
}

// get method returns a handler for the specified method if one exists.
// It returns nil otherwise.
func (ha *handlerArray) get(method string) http.Handler {
	for _, node := range ha.nodes {
		if node.method == method {
			return node.handler
		}
	}
	return nil
}

// add method adds the handler for the specified method if one doesn't exist
// yet.
func (ha *handlerArray) add(method string, handler http.Handler) {
	if ha.get(method) == nil {
		ha.nodes = append(ha.nodes, newHandlerNode(method, handler))
	}
}