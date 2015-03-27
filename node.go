package gocelot

import (
	"net/http"
	"net/url"
)

// node represents a prefix tree node and is used for routing.
// It has a path of the current node, a list of next nodes and a handlerArray
// for the path.
type node struct {
	path string
	next []*node
	handlers *handlerArray
}

// newNode is a function which returns a new empty node.
func newNode() *node {
	return &node{}
}

/*
debugging nodes

var nodes []*node

func newNode() *node {
	node := &node{}
	nodes = append(nodes, node)
	return node
}

func Print() {
	for _, node := range nodes {
		log.Print(node.path)
	}
}
*/

// nodeSeqFromPath is a function which returns the first and the last node of
// the sequence.
// If the path doesn't contain params(':'), the first and the last nodes are
// the same.
// All the params(':') are stored in seperate nodes.
// Eg.
// nodeSeqFromPath('/path/:param/end/')
// returns node('/path/'), node('/end/')
// and the sequence is node('/path') -> node(':param') -> node('/end/')
func nodeSeq(path string) (*node, *node) {
	first := newNode()
	last := first
	start, isParam := 0, false
	extendSeq := func(end int) {
		if start != end {
			if last.path == "" {
				last.path = path[start:end]
			} else {
				next := newNode()
				next.path = path[start:end]
				last.next = []*node{next}
				last = next
			}
		}
		start = end
		isParam = !isParam
	}
	for end, letter := range path {
		switch isParam {
		case false:
			if letter == ':' {
				extendSeq(end)
			}
		case true:
			if letter == '/' {
				extendSeq(end)
			}
		}
	}
	extendSeq(len(path))
	return first, last
}

// min is a helper function which returns the minimum of two intergers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// lcp is a helper function which returns the longest common prefix of a and b.
func lcp(a, b string) int {
	lcp, minLen := 0, min(len(a), len(b))
	for lcp < minLen && a[lcp] == b[lcp] {
		lcp++
	}
	return lcp
}

// add is a method which adds the path to the node and returns the last node
// which belongs to the added path.
// The path in the node and the path which is added should have a common prefix.
// Otherwise after addition the current node will contain an empty path.
// Eg.
// node('/path/').add('/paths/')
// returns node('s/')
// and the result sequence is
// node('/path') -> node('/')
// node('/path') -> node('s/')
// Note that different params are stored in different nodes.
// Eg.
// For starting sequence node('/path/') -> node(':param1/') -> node('end/')
// node('/path/').add('/path/:param2/finish/')
// returns node('/finish/')
// and the result sequence is
// node('/path/') -> node(':param1') -> node('/end/')
// node('/path/') -> node(':param2') -> node('/finish/')
// NOT
// node('/path/') -> node(':param') -> node('1/end/')
// node('/path/') -> node(':param') -> node('2/finish/')
func (n *node) add(path string) *node {
	diff := lcp(n.path, path)
	if diff == len(n.path) && diff == len(path) {
		// the paths are equal, no new nodes were created
		return n
	}
	if n.path[0] == ':' {
		// the node.path is a param
		if diff == len(path) || path[diff] != '/' {
			// the params are different, have to rollback
			return nil
		}
	}
	if diff == len(n.path) {
		// n.path matches exactly, have to search next nodes
		for _, next := range n.next {
			if next.path[0] == path[diff] {
				node := next.add(path[diff:])
				if node != nil {
					return node
				}
			}
		}
	} else {
		// n.path doesn't match exactly, have to split it
		remainderNode := newNode()
		remainderNode.path = n.path[diff:]
		remainderNode.next = n.next
		remainderNode.handlers = n.handlers

		n.path = n.path[:diff]
		n.next = []*node{remainderNode}
		n.handlers = nil

		if diff == len(path) {
			// path matches the new n.path exactly
			return n
		}
	}
	// path has to be split at diff
	first, last := nodeSeq(path[diff:])
	n.next = append(n.next, first)
	return last
}

// indexOf function returns index of the next occurence of c in s
// or len(s) if c not found
func indexOf(s string, c rune) int {
	for i, letter := range s {
		if letter == c {
			return i
		}
	}
	return len(s)
}

/*
alternative params return pattern which uses one alloc per request with params

func addParam(paramList []string, paramCount int, key, value string) []string {
	paramPos := (paramCount + 1) * 2
	if paramList == nil {
		paramList = make([]string, paramPos)
	}
	paramList[paramPos - 2] = key
	paramList[paramPos - 1] = value
	return paramList
}
*/

// addParam function adds key/value to request.Form if handler exists.
// It creates request.Form if necessary.
// request.Form is map[string][]string.
// addParam puts value at the end of request.Form[key].
func addParam(request *http.Request, handler http.Handler, key, value string) {
	if handler != nil {
		if request.Form == nil {
			request.Form = url.Values{}
		}
		request.Form.Add(key, value)
	}
}

// get is a method which returns a http.Handler for the specified path/method
// if one exists.
// It also returns a boolean which is true if the specified path exists.
func (n *node) get(path, method string,
	request *http.Request) (http.Handler, bool) {

	if n.path[0] == ':' {
		// n.path is a param, try matching a param in the path
		paramLen := indexOf(path, '/')
		if paramLen == len(path) {
			// param is the last segment of the path
			if n.handlers == nil {
				return nil, false
			}
			handler := n.handlers.get(method)
			addParam(request, handler, n.path[1:], path)
			// maybe return handler, n.handlers != nil
			return handler, true
		}
		// param is not the last segment, have to check next for next segments
		for _, next := range n.next {
			if next.path[0] == '/' {
				handler, pathFound := next.get(path[paramLen:], method, request)
				if pathFound {
					// if path was found, try adding param to the request
					addParam(request, handler, n.path[1:], path[:paramLen])
					return handler, pathFound
				}
			}
		}
	} else if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
		// n.path matches path exactly to n.paths length
		if len(path) == len(n.path) {
			// n.path matched path exactly
			if n.handlers == nil {
				return nil, false
			}
			return n.handlers.get(method), true
		}
		// path is longer than n.path, have to check next for next segments
		for _, next := range n.next {
			if next.path[0] == path[len(n.path)] || next.path[0] == ':' {
				// next segment matches path or is a param
				handler, pathFound := next.get(path[len(n.path):], method,
					request)
				if pathFound {
					return handler, pathFound
				}
			}
		}
	}
	// no path was found at this branch
	return nil, false
}

// handle is a method which adds methodHandler to handlers of the node.
// It creates new handlerArray if necessary.
func (n *node) handle(method string, handler http.Handler) {
	if n.handlers == nil {
		n.handlers = newHandlerArray()
	}
	n.handlers.add(method, handler)
}
