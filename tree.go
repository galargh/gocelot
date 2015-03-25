package gocelot

type tree struct {
	root *node
}

func newTree() *tree {
	t := &tree{newNode()}
	t.root.path = "/"
	return t
}