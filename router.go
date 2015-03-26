// TODO:
//  merging trees
//  panic handler
//  testing
//  readme

// Package gocelot provides a simple url router.
// It supports url parameters and places them in http.request.Form field.
package gocelot

import (
	"net/http"
)

// Router conforms to http.Handler interface.
// Router hold a current tree of urls served being served.
// NotFound handler is used if a handler for a given path/method doesn't exist.
// By default http.NotFound func is used.
// MethodNotAllowed handler is used if there are handlers for a given path,
// but no handler for the given method is found.
type Router struct {
	tree *node
	NotFound http.Handler
	MethodNotAllowed http.Handler
}

// New function creates a new router with an empty tree(just tree root at '/')
// and handlers set to nil.
func New() *Router {
	root := newNode()
	root.path = "/"
	return &Router{root, nil, nil}
}

// Handle method adds the path to the tree and the handler for the method.
// It accepts http.Handler as a handler
func (r *Router) Handle(method, path string, handler http.Handler) {
	r.tree.add(path).handle(method, handler)
}

// HandleFunc method adds the path to the tree and the handler for the method.
// It accepts func(http.ResponseWriter, *http.Request) as a handler
func (r *Router) HandleFunc(method, path string,
	handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.Handle(method, path, http.HandlerFunc(handlerFunc))
}

// Router implements http.Handler ServeHTTP method.
// It routes the path/method to the correct handler or returns an error.
func (r *Router) ServeHTTP(response http.ResponseWriter,
	request *http.Request) {

	path, method := request.URL.Path, request.Method
	handler, pathFound := r.tree.get(path, method, request)
	if handler != nil {
		handler.ServeHTTP(response, request)
		return
	}
	if pathFound && r.MethodNotAllowed != nil {
		r.MethodNotAllowed.ServeHTTP(response, request)
		return
	}
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(response, request)
		return
	}
	http.NotFound(response, request)
}
