package renderer

import (
	"bytes"
	"io"
	"log"
	"maps"

	"github.com/adrg/frontmatter"
	"github.com/connormckelvey/website/tree"
	"github.com/connormckelvey/website/util"
	"github.com/spf13/afero"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type Renderer interface {
	Test(tree.Node) (bool, error)
	Props(node tree.Node, context *RenderContext) (map[string]any, error)
	Open(node tree.Node, context *RenderContext) error
	Close(node tree.Node, context *RenderContext) error
}

func Render(siteFS afero.Fs, buildFS afero.Fs, root tree.Node, renderers []Renderer, context *RenderContext) error {

	for _, renderer := range renderers {
		ok, err := renderer.Test(root)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		if err := renderer.Open(root, context); err != nil {
			return err
		}
		for _, child := range root.Children() {
			if err := Render(siteFS, buildFS, child, renderers, context); err != nil {
				return err
			}
		}

		if currentFile := context.CurrentFile(); !root.IsDir() && currentFile != nil {
			source, err := context.Source(root)
			if err != nil {
				return err
			}

			var fm struct {
				Page tree.PageFrontMatter `yaml:"page"`
			}
			content, err := frontmatter.Parse(bytes.NewReader(source), &fm)
			if err != nil {
				return err
			}

			var metaProps []map[string]any
			for _, meta := range fm.Page.Meta {
				m, err := util.MarshalMap(meta)
				if err != nil {
					return err
				}
				metaProps = append(metaProps, m)
			}

			var linkProps []map[string]any
			for _, link := range fm.Page.Links {
				l, err := util.MarshalMap(link)
				if err != nil {
					return err
				}
				linkProps = append(linkProps, l)
			}
			props := map[string]any{
				"page": map[string]any{
					"meta":     metaProps,
					"links":    linkProps,
					"template": fm.Page.Template,
					"title":    fm.Page.Title,
				},
			}
			p, err := renderer.Props(root, context)
			if err != nil {
				return err
			}
			maps.Copy(props, p)

			var templated bytes.Buffer
			ev := NewEvaluator(afero.NewIOFS(siteFS))
			if err := ev.Render(bytes.NewReader(content), root.Path(), props, &templated); err != nil {
				return err
			}

			var compiledMarkdown bytes.Buffer
			md := goldmark.New(
				goldmark.WithExtensions(extension.GFM),
				goldmark.WithRendererOptions(html.WithUnsafe()),
			)
			if err := md.Convert(templated.Bytes(), &compiledMarkdown); err != nil {
				return err
			}
			if fm.Page.Template == "" {
				_, err := io.Copy(currentFile, &compiledMarkdown)
				return err
			}
			log.Println("MD", compiledMarkdown.String())
			b, err := WrapTheme(context.themeFS, fm.Page.Template, compiledMarkdown.Bytes(), props)
			if err != nil {
				return err
			}

			if _, err := currentFile.Write(b); err != nil {
				return err
			}

			// if _, err := currentFile.Write(content); err != nil {
			// 	return err
			// }
		}

		if err := renderer.Close(root, context); err != nil {
			return err
		}
		break
	}
	return nil
}

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
