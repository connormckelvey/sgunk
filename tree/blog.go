package tree

import "time"

type BlogPostFrontMatter struct {
	Title string   `yaml:"title"`
	Tags  []string `yaml:"tags"`
}

type BlogNode struct {
	BaseNode
	Root string
}

func NewBlogNode(path string, root string) *BlogNode {
	return &BlogNode{
		BaseNode: NewBaseNode(path, true),
		Root:     root,
	}
}

type BlogCollectionNode struct {
	BaseNode
}

type BlogPostNode struct {
	BaseNode
	Parts     PageNameParts
	CreatedAt time.Time
}

func NewBlogPostNode(path string, parts PageNameParts, createdAt time.Time) *BlogPostNode {
	return &BlogPostNode{
		BaseNode:  NewBaseNode(path, false),
		Parts:     parts,
		CreatedAt: createdAt,
	}
}
