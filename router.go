package gocelot

import (
	"net/http"
	"net/url"
)

type Router struct {
	trees []*methodTree
	paramsPath string
}

func New() *Router {
	return &Router{nil, "urlParams"}
}

func (r *Router) addTree(method string) *tree {
	for _, mt := range r.trees {
		if mt.method == method {
			return mt.getTree()
		}
	}
	mt := newMethodTree(method)
	r.trees = append(r.trees, mt)
	return mt.t
}

func (r *Router) getTree(method string) *tree {
	for _, mt := range r.trees {
		if mt.method == method {
			return mt.getTree()
		}
	}
	return nil
}

func (r *Router) Handle(method, path string, handler http.Handler) {
	r.addTree(method).addRoot().add(path).handler = handler
}

func (r *Router) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	path, method := request.URL.Path, request.Method
	tree := r.getTree(method)
	if tree != nil {
		root := tree.getRoot()
		if root != nil {
			node, params := root.get(path, 0)
			if node != nil && node.handler != nil {
				if params != nil {
					if request.Form == nil {
						request.Form = url.Values{}
					}
					request.Form[r.paramsPath] = params
				}
				node.handler.ServeHTTP(response, request)
			}
		}
	}
}
