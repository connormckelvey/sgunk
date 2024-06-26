package blog

import (
	"path/filepath"
	"time"

	"github.com/connormckelvey/sgunk/renderer"
	"github.com/connormckelvey/sgunk/tree"
)

type BlogRenderer struct {
}

func (r *BlogRenderer) Kind() tree.NodeKind {
	return BlogKind
}

func (f *BlogRenderer) openBlogNode(node *BlogNode, context *renderer.RenderContext) error {
	if err := context.MkdirAll(node.Root, 0755); err != nil {
		return err
	}
	context.PushDir(node.Root)
	return nil
}

func (*BlogRenderer) closeBlogNode(_ *BlogNode, context *renderer.RenderContext) error {
	context.PopDir()
	return nil
}

func (f *BlogRenderer) openBlogPostNode(node *BlogPostNode, context *renderer.RenderContext) error {
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

func (f *BlogRenderer) closeBlogPostNode(_ *BlogPostNode, context *renderer.RenderContext) error {
	popped := context.PopFile()
	return popped.Close()
}

type PropsBuilder func(props map[string]any)

func (r *BlogRenderer) Open(node tree.Node, context *renderer.RenderContext) error {
	switch n := node.(type) {
	case *BlogNode:
		return r.openBlogNode(n, context)
	case *BlogCollectionNode:
		return nil
	case *BlogPostNode:
		return r.openBlogPostNode(n, context)
	}
	return nil
}

func (r *BlogRenderer) Close(node tree.Node, context *renderer.RenderContext) error {
	switch n := node.(type) {
	case *BlogNode:
		return r.closeBlogNode(n, context)
	case *BlogCollectionNode:
		return nil
	case *BlogPostNode:
		return r.closeBlogPostNode(n, context)
	}
	return nil
}
