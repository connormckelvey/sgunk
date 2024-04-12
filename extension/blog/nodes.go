package blog

import (
	"time"

	"github.com/connormckelvey/sgunk/tree"
)

const BlogKind = tree.NodeKind("blog")

type BlogPostFrontMatter struct {
	Title string   `yaml:"title"`
	Tags  []string `yaml:"tags"`
}

type BlogNode struct {
	tree.BaseNode
	Root string
}

func NewBlogNode(path string, root string) *BlogNode {
	return &BlogNode{
		BaseNode: tree.NewBaseNode(path, true),
		Root:     root,
	}
}

func (*BlogNode) Kind() tree.NodeKind {
	return BlogKind
}

type BlogCollectionNode struct {
	tree.BaseNode
}

func (*BlogCollectionNode) Kind() tree.NodeKind {
	return BlogKind
}

type BlogPostNode struct {
	tree.BaseNode
	Parts     tree.PageNameParts
	CreatedAt time.Time
}

func NewBlogPostNode(path string, parts tree.PageNameParts, createdAt time.Time) *BlogPostNode {
	return &BlogPostNode{
		BaseNode:  tree.NewBaseNode(path, false),
		Parts:     parts,
		CreatedAt: createdAt,
	}
}

func (*BlogPostNode) Kind() tree.NodeKind {
	return BlogKind
}
