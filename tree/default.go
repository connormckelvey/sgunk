package tree

const DefaultNodeKind = NodeKind("site")

type DefaultPage struct {
	BaseNode
	Parts PageNameParts
}

func (*DefaultPage) Kind() NodeKind {
	return DefaultNodeKind
}

func NewDefaultPage(path string, parts PageNameParts) *DefaultPage {
	return &DefaultPage{
		BaseNode: NewBaseNode(path, false),
		Parts:    parts,
	}
}

type DefaultDir struct {
	BaseNode
}

func (*DefaultDir) Kind() NodeKind {
	return DefaultNodeKind
}
