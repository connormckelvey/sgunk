package tree

type Node interface {
	Path() string
	IsDir() bool
	Children() []Node
	AppendChild(Node)
	AddAttrs(key string, attrs any) error
	GetAttrs(key string) (map[string]any, bool)
	Attributes() (map[string]map[string]any, error)
}
