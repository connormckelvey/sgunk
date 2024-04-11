package tree

type BaseNode struct {
	path     string
	isDir    bool
	children []Node
}

func NewBaseNode(path string, isDir bool) BaseNode {
	return BaseNode{path: path, isDir: isDir}
}

func (s *BaseNode) IsDir() bool {
	return s.isDir
}

func (s *BaseNode) Path() string {
	return s.path
}

func (s *BaseNode) Children() []Node {
	return s.children
}

func (s *BaseNode) AppendChild(n Node) {
	s.children = append(s.children, n)
}
