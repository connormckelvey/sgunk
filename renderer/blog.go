package renderer

import (
	"bytes"
	"path/filepath"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/connormckelvey/ssg/tree"
)

type BlogRenderer struct {
}

func (r *BlogRenderer) Test(node tree.Node) (bool, error) {
	switch node.(type) {
	case *tree.BlogNode, *tree.BlogCollectionNode, *tree.BlogPostNode:
		return true, nil
	}
	return false, nil
}

func (r *BlogRenderer) Props(node tree.Node, context *RenderContext) (map[string]any, error) {
	if node.IsDir() {
		return nil, nil
	}
	source, err := context.Source(node)
	if err != nil {
		return nil, err
	}

	var fm struct {
		Post tree.BlogPostFrontMatter `yaml:"post"`
	}
	if _, err := frontmatter.Parse(bytes.NewReader(source), &fm); err != nil {
		return nil, err
	}
	return map[string]any{
		"post": map[string]any{
			"createdAt": node.(*tree.BlogPostNode).CreatedAt,
			"title":     fm.Post.Title,
			"tags":      fm.Post.Tags,
		},
	}, nil
}

func (f *BlogRenderer) openBlogNode(node *tree.BlogNode, context *RenderContext) error {
	if err := context.MkdirAll(node.Root, 0755); err != nil {
		return err
	}
	context.PushDir(node.Root)
	return nil
}

func (*BlogRenderer) closeBlogNode(_ *tree.BlogNode, context *RenderContext) error {
	context.PopDir()
	return nil
}

func (f *BlogRenderer) openBlogPostNode(node *tree.BlogPostNode, context *RenderContext) error {
	datePath := time.Now().Format("2006/01/02")
	if err := context.MkdirAll(datePath, 0755); err != nil {
		return err
	}
	postPath := filepath.Join(datePath, node.Parts.Slug+".html")
	_, err := context.CreateFile(postPath)
	if err != nil {
		return err
	}

	return nil
}

func (f *BlogRenderer) closeBlogPostNode(_ *tree.BlogPostNode, context *RenderContext) error {
	popped := context.PopFile()
	return popped.Close()
}

type PropsBuilder func(props map[string]any)

func (r *BlogRenderer) Open(node tree.Node, context *RenderContext) error {
	switch n := node.(type) {
	case *tree.BlogNode:
		return r.openBlogNode(n, context)
	case *tree.BlogCollectionNode:
		return nil
	case *tree.BlogPostNode:
		return r.openBlogPostNode(n, context)
	}
	return nil
}

func (r *BlogRenderer) Close(node tree.Node, context *RenderContext) error {
	switch n := node.(type) {
	case *tree.BlogNode:
		return r.closeBlogNode(n, context)
	case *tree.BlogCollectionNode:
		return nil
	case *tree.BlogPostNode:
		return r.closeBlogPostNode(n, context)
	}
	return nil
}
