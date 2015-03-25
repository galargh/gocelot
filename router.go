package gocelot

import (
	"net/http"
	"net/url"
)

type Router struct {
	tree *tree
	paramsPath string
	NotFound http.Handler
	MethodNotAllowed http.Handler
}

func New() *Router {
	return &Router{newTree(), "urlParams", nil, nil}
}

func (r *Router) Handle(method, path string, handler http.Handler) {
	r.tree.root.add(path).handle(method, handler)
}

func (r *Router) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	path, method := request.URL.Path, request.Method
	node, params := r.tree.root.get(path, 0)
	if node != nil{
		if node.handlers != nil {
			handler := node.handlers.get(method)
			if handler != nil {
				if params != nil {
					if request.Form == nil {
						request.Form = url.Values{}
					}
					request.Form[r.paramsPath] = params
				}
				handler.ServeHTTP(response, request)
				return
			}
			if r.MethodNotAllowed != nil {
				r.MethodNotAllowed.ServeHTTP(response, request)
				return
			}
		}
	}
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(response, request)
	} else {
		http.NotFound(response, request)
	}
}
