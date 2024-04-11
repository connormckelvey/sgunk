package renderer

import (
	"bytes"
	"io"
	"log"
	"maps"
	"sync"

	"github.com/adrg/frontmatter"
	"github.com/connormckelvey/website/tree"
	"github.com/connormckelvey/website/util"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type Renderer struct {
	options []RendererOption
	once    sync.Once

	renderers []EntryRenderer
}

type RendererOption interface {
	Apply(*Renderer) error
}

type RendererOptionFunc func(*Renderer) error

func (apply RendererOptionFunc) Apply(p *Renderer) error {
	return apply(p)
}

func WithEntryRenderers(renderers ...EntryRenderer) RendererOptionFunc {
	return func(r *Renderer) error {
		r.renderers = append(r.renderers, renderers...)
		return nil
	}
}

func New(opts ...RendererOption) *Renderer {
	return &Renderer{
		options: opts,
	}
}

func (r *Renderer) Render(root tree.Node, context *RenderContext) error {
	var err error
	r.once.Do(func() {
		for _, opt := range r.options {
			err = opt.Apply(r)
			if err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	for _, renderer := range r.renderers {
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
			if err := r.Render(child, context); err != nil {
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
			if err := context.templater.Render(bytes.NewReader(content), root.Path(), props, &templated); err != nil {
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
