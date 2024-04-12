package tree

type BaseNode struct {
	path     string
	isDir    bool
	children []Node
	attr     *NodeAttributes
}

func NewBaseNode(path string, isDir bool) BaseNode {
	return BaseNode{path: path, isDir: isDir, attr: newNodeAttributes()}
}

func (s *BaseNode) AddAttrs(key string, attrs any) error {
	return s.attr.Add(key, attrs)
}

func (s *BaseNode) GetAttrs(key string) (map[string]any, bool) {
	return s.attr.Get(key)
}

func (s *BaseNode) Attributes() (map[string]map[string]any, error) {
	return s.attr.Map()
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
