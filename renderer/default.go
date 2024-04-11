package renderer

import (
	"bytes"
	"log"

	"github.com/adrg/frontmatter"
	"github.com/connormckelvey/website/tree"
)

type DefaultRenderer struct {
}

func (r *DefaultRenderer) Props(node tree.Node, context *RenderContext) (map[string]any, error) {
	if node.IsDir() {
		return nil, nil
	}
	source, err := context.Source(node)
	if err != nil {
		return nil, err
	}

	var fm tree.PageFrontMatter
	if _, err := frontmatter.Parse(bytes.NewReader(source), &fm); err != nil {
		return nil, err
	}
	return map[string]any{
		"title":    fm.Title,
		"meta":     fm.Meta,
		"links":    fm.Links,
		"template": fm.Template,
	}, nil
}

func (r *DefaultRenderer) Test(node tree.Node) (bool, error) {
	switch node.(type) {
	case *tree.DefaultDir, *tree.DefaultPage, *tree.Site:
		return true, nil
	}
	return false, nil
}

func (r *DefaultRenderer) openDefaultDir(node *tree.DefaultDir, context *RenderContext) error {
	log.Println("render blog", node.Path())
	if err := context.MkdirAll(node.Path(), 0755); err != nil {
		return err
	}
	context.PushDir(node.Path())
	return nil
}

func (r *DefaultRenderer) closeDefaultDir(node *tree.DefaultDir, context *RenderContext) error {
	context.PopDir()
	return nil
}

func (r *DefaultRenderer) openDefaultPage(node *tree.DefaultPage, context *RenderContext) error {
	log.Println("render post", node.Path())

	_, err := context.CreateFile(node.Parts.Slug + ".html")
	if err != nil {
		return err
	}
	return nil
}

func (r *DefaultRenderer) closeDefaultPage(node *tree.DefaultPage, context *RenderContext) error {
	popped := context.PopFile()
	return popped.Close()
}

func (r *DefaultRenderer) Open(node tree.Node, context *RenderContext) error {
	switch n := node.(type) {
	case *tree.DefaultDir:
		log.Println("open dir", n.Path())
		return r.openDefaultDir(n, context)
	case *tree.DefaultPage:
		log.Println("open page", n.Path())
		return r.openDefaultPage(n, context)
	}
	return nil
}

func (r *DefaultRenderer) Close(node tree.Node, context *RenderContext) error {
	switch n := node.(type) {
	case *tree.DefaultDir:
		log.Println("close dir", n.Path())
		return r.closeDefaultDir(n, context)
	case *tree.DefaultPage:
		log.Println("close page", n.Path())
		return r.closeDefaultPage(n, context)
	}
	return nil
}
