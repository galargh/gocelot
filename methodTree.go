package gocelot

type methodTree struct {
	method string
	t *tree
}

func newMethodTree(method string) *methodTree {
	return &methodTree{method, newTree()}
}

func (mt *methodTree) getTree() *tree {
	return mt.t
}