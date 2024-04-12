package tree

// Root node
type Site struct {
	BaseNode
}

const SiteNodeKind = NodeKind("site")

func (*Site) Kind() NodeKind {
	return SiteNodeKind
}
