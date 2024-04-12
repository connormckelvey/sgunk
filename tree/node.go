package tree

type NodeKind string

func (nk NodeKind) String() string {
	return string(nk)
}

type Node interface {
	Kind() NodeKind
	Path() string
	IsDir() bool
	Children() []Node
	AppendChild(Node)
	AddAttrs(key string, attrs any) error
	GetAttrs(key string) (map[string]any, bool)
	Attributes() (map[string]map[string]any, error)
}
