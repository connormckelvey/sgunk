package tree

type Node interface {
	Path() string
	IsDir() bool
	Children() []Node
	AppendChild(Node)
}
