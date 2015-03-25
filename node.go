package gocelot

import (
	"net/http"
)

type node struct {
	path string
	next []*node
	handlers *handlerArray
}

func newNode() *node {
	return &node{}
}

// nodeSeqFromPath returns the first and the last node of the sequence.
// If the path doesn't contain params(':'), the first and the last nodes are
// the same.
// All the params(':') are stored in seperate nodes.
// Eg.
// nodeSeqFromPath('/path/:param/end/')
// returns node('/path/'), node('end/')
// and the sequence is node('/path') -> node(':param/') -> node('end/')
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
	}
	for end, letter := range path {
		switch isParam {
		case false:
			if letter == ':' {
				extendSeq(end)
				isParam = true
			}
		case true:
			if letter == '/' {
				extendSeq(end + 1)
				isParam = false
			}
		}
	}
	if start != len(path) {
		extendSeq(len(path))
	}
	return first, last
}

// min returns the minimum of two intergers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// lcp returns the longest common prefix of 2 words
func lcp(a, b string) int {
	lcp, minLen := 0, min(len(a), len(b))
	for lcp < minLen && a[lcp] == b[lcp] {
		lcp++
	}
	return lcp
}

// slice returns a slice of given length starting at start of the given string
func slice(path string, start, length int) string {
	return path[start:min(start+length, len(path))]
}

// addPath adds the path to the node and returns the last node which belongs
// to the path being added.
// The path in the node and the path which is added should have a common prefix.
// Otherwise after addition the current node will contain an empty path.
// Eg.
// node('/path/').AddPath('/paths/')
// returns node('s/')
// and the result sequence is
// node('/path') -> node('/')
// node('/path') -> node('s/')
// Note that different params are stored in different nodes.
// Eg.
// For starting sequence node('/path/') -> node(':param1/') -> node('end/')
// node('/path/').Add('/path/param2/finish/')
// returns node('finish/')
// and the result sequence is
// node('/path/') -> node(':param1/') -> node('end/')
// node('/path/') -> node(':param2/') -> node('finish/')
// NOT
// node('/path/') -> node(':param') -> node('1/end/')
// node('/path/') -> node(':param') -> node('2/finish/')
func (n *node) add(path string) *node {
	diff := lcp(n.path, path)
	// The given path already exists
	if len(n.path) == diff && len(path) == diff {
		return n
	}
	switch diff {
	case len(n.path):
		// diff < len(path)
		// split the path and add to next
		// check the current next for match first
		// params have to match entirely
		if path[diff] == ':' {
			for _, next := range n.next {
				if next.path[0] == ':' {
					if next.path == slice(path, diff, len(next.path)) {
						if next.path[len(next.path) - 1] == '/' || 
							len(path) == diff + len(next.path) {
							return next.add(path[diff:])
						}
					}
				}
			}
		} else {
			for _, next := range n.next {
				if next.path[0] == path[diff] {
					return next.add(path[diff:])
				}
			}
		}
		// add new node if no match was found
		first, last := nodeSeq(path[diff:])

		n.next = append(n.next, first)

		return last
	case len(path):
		// diff < len(n.path)
		// split the n.path and set the node to path
		remainderNode := newNode()
		remainderNode.path = n.path[diff:]
		remainderNode.next = n.next
		remainderNode.handlers = n.handlers

		n.path = n.path[:diff]
		n.next = []*node{remainderNode}
		n.handlers = nil

		return n
	default:
		// diff < len(n.path) && diff < len(path)
		// split the n.path into two
		remainderNode := newNode()
		remainderNode.path = n.path[diff:]
		remainderNode.next = n.next
		remainderNode.handlers = n.handlers

		first, last := nodeSeq(path[diff:])

		n.path = n.path[:diff]
		n.next = []*node{remainderNode, first}
		n.handlers = nil

		return last
	}
}

// indexOf returns index of the next c occurence in s
// or len(s) if not found
func indexOf(s string, c rune) int {
	for i, letter := range s {
		if letter == c {
			return i
		}
	}
	return len(s)
}

func (n *node) get(path string, paramsCount int) (*node, []string) {
	switch n.path[0] {
	case ':':
		paramLen := indexOf(path, '/')
		// current node is a param
		if n.path[len(n.path) - 1] != '/' {
			if paramLen == len(path) {
				paramsPos := (paramsCount + 1) * 2
				paramsList := make([]string, paramsPos)
				paramsList[paramsPos - 2] = n.path[1:]
				paramsList[paramsPos - 1] = path
				return n, paramsList
			}
			// path exceeds the pattern
			return nil, nil
		}
		if paramLen == len(path) {
			// pattern exceeds the path
			return nil, nil
		}
		if paramLen + 1 == len(path) {
			paramsPos := (paramsCount + 1) * 2
			paramsList := make([]string, paramsPos)
			paramsList[paramsPos - 2] = n.path[1:len(n.path) - 1]
			paramsList[paramsPos - 1] = path[:paramLen]
			return n, paramsList
		}

		// check the current next for match
		for _, next := range n.next {
			firstLetter := next.path[0]
			if firstLetter == ':' || firstLetter == path[paramLen + 1] {
				node, paramsList := next.get(path[paramLen + 1:],
					paramsCount + 1)
				if node != nil {
					// exhaustive path was found
					// add param
					paramsPos := (paramsCount + 1) * 2
					if (paramsList == nil) {
						paramsList = make([]string, paramsPos)
					}
					paramsList[paramsPos - 2] = n.path[1:len(n.path) - 1]
					paramsList[paramsPos - 1] = path[:paramLen]
					return node, paramsList
				}
			}
		}
		// no exhaustive path
		return nil, nil
	default:
		// current node is regular
		if len(path) < len(n.path) {
			//no match
			return nil, nil
		}
		if len(path) == len(n.path) {
			//possible match
			if path == n.path {
				return n, nil
			}
			return nil, nil
		}
		if path[:len(n.path)] != n.path {
			return nil, nil
		}
		//if it's a begining of a new segment, can match param next
		canParam := path[len(n.path) - 1] == '/'
		for _, next := range n.next {
			firstLetter := next.path[0]
			if firstLetter == path[len(n.path)] || (canParam && firstLetter == ':') {
				node, paramsList := next.get(path[len(n.path):], paramsCount)
				if node != nil {
					return node, paramsList
				}
			}
		}
		return nil, nil
	}
}

func (n *node) handle(method string, handler http.Handler) {
	if n.handlers == nil {
		n.handlers = newHandlerArray()
	}
	n.handlers.add(method, handler)
}