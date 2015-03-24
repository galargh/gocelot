package gocelot

import (
	"net/http"
)

type node struct {
	path string
	next []*node
	handler http.Handler
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
		case false && letter == ':':
			extendSeq(end)
			isParam = true
		case true && letter == '/':
			extendSeq(end + 1)
			isParam = false
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
				if next.path[0] == ':' &&
					next.path == slice(path, diff, len(next.path)) {
					return next.add(path[diff:])
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
		remainderNode.handler = n.handler

		n.path = n.path[:diff]
		n.next = []*node{remainderNode}
		n.handler = nil

		return n
	default:
		// diff < len(n.path) && diff < len(path)
		// split the n.path into two
		remainderNode := newNode()
		remainderNode.path = n.path[diff:]
		remainderNode.next = n.next
		remainderNode.handler = n.handler

		first, last := nodeSeq(path[diff:])

		n.path = n.path[:diff]
		n.next = []*node{remainderNode, first}
		n.handler = nil

		return last
	}
}

// nextByte returns index of the next c occurence in s
// or len(s) if not found
func nextByte(s string, c byte) int {
	n := 0
	for n < len(s) && s[n] != c {
		n++
	}
	return n
}

func (n *node) get(path string, paramsCount int) (*node, []string) {
	//log.Print(path, " ", n.path, " ", paramsCount)
	switch n.path[0] {
	case ':':
		// current node is a param
		paramLen := nextByte(path, '/')
		if paramLen + 1 >= len(path) {
			// param is the last argument
			// +1 to account for trailing slash
			paramsPos := (paramsCount + 1) * 2
			paramsList := make([]string, paramsPos)
			paramsList[paramsPos - 2] = n.path[1:len(n.path) - 1]
			paramsList[paramsPos - 1] = path[:paramLen]

			return n, paramsList
		} else {
			// param is not the last argument
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
		}
	default:
		// current node is regular
		diff := lcp(n.path, path)

		if len(path) == diff && len(n.path) == diff {
			// paths are equal
			if path[len(path) - 1] == '/' {
				// paths have trailing slash
				return n, nil
			} else {
				//check for trailing slash
				for _, next := range n.next {
					if next.path[0] == '/' && len(next.path) == 1 {
						return next, nil
					}
				}
				return nil, nil
			}
		}
		switch diff {
		case len(n.path):
			// check for remainder of the path
			canParam := path[diff - 1] == '/'
			for _, next := range n.next {
				firstLetter := next.path[0]
				if firstLetter == path[diff] || (canParam && firstLetter == ':') {
					node, paramsList := next.get(path[diff:], paramsCount)
					if node != nil {
						return node, paramsList
					}
				}
			}
			return nil, nil
		case len(path):
			// path is different
			if n.path[diff] == '/' && len(n.path) == diff + 1 {
				// path in the node differs only by trailing slash
				return n, nil
			} else {
				return nil, nil
			}
		default:
			return nil, nil
		}
	}
}