package renderer

import (
	"github.com/connormckelvey/ssg/tree"
)

type DefaultRenderer struct {
}

func (r *DefaultRenderer) Test(node tree.Node) (bool, error) {
	switch node.(type) {
	case *tree.DefaultDir, *tree.DefaultPage, *tree.Site:
		return true, nil
	}
	return false, nil
}

func (r *DefaultRenderer) openDefaultDir(node *tree.DefaultDir, context *RenderContext) error {
	if err := context.MkdirAll(node.Path(), 0755); err != nil {
		return err
	}
	context.PushDir(node.Path())
	return nil
}

func (r *DefaultRenderer) closeDefaultDir(_ *tree.DefaultDir, context *RenderContext) error {
	context.PopDir()
	return nil
}

func (r *DefaultRenderer) openDefaultPage(node *tree.DefaultPage, context *RenderContext) error {
	_, err := context.CreateFile(node.Parts.Slug + ".html")
	if err != nil {
		return err
	}
	return nil
}

func (r *DefaultRenderer) closeDefaultPage(_ *tree.DefaultPage, context *RenderContext) error {
	popped := context.PopFile()
	return popped.Close()
}

func (r *DefaultRenderer) Open(node tree.Node, context *RenderContext) error {
	switch n := node.(type) {
	case *tree.DefaultDir:
		return r.openDefaultDir(n, context)
	case *tree.DefaultPage:
		return r.openDefaultPage(n, context)
	}
	return nil
}

func (r *DefaultRenderer) Close(node tree.Node, context *RenderContext) error {
	switch n := node.(type) {
	case *tree.DefaultDir:
		return r.closeDefaultDir(n, context)
	case *tree.DefaultPage:
		return r.closeDefaultPage(n, context)
	}
	return nil
}
