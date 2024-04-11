package tree

type DefaultPage struct {
	BaseNode
	Parts PageNameParts
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
