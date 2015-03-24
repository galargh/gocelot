package gocelot

type tree struct {
	root *node
}

func newTree() *tree {
	return &tree{}
}

func (t *tree) addRoot() *node {
	if t.root == nil {
		t.root = newNode()
		t.root.path = "/"
	}
	return t.root
}

func (t *tree) getRoot() *node {
	return t.root
}