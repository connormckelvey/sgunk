package renderer

import (
	"bytes"
	"io"

	"github.com/adrg/frontmatter"
	"github.com/connormckelvey/ssg/tree"
	"github.com/spf13/afero"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type Renderer struct {
	options   []RendererOption
	siteFS    afero.Fs
	themeFS   afero.Fs
	buildFS   afero.Fs
	renderers []EntryRenderer
	templater *Evaluator
	markdown  goldmark.Markdown
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

func WithSiteFS(siteFS afero.Fs) RendererOptionFunc {
	return func(r *Renderer) error {
		r.siteFS = siteFS
		r.templater = NewEvaluator(afero.NewIOFS(siteFS))
		return nil
	}
}

func WithThemeFS(themeFS afero.Fs) RendererOptionFunc {
	return func(r *Renderer) error {
		r.themeFS = themeFS
		return nil
	}
}

func WithBuildFS(buildFS afero.Fs) RendererOptionFunc {
	return func(r *Renderer) error {
		r.buildFS = buildFS
		return nil
	}
}

func WithFS(siteFS afero.Fs, themeFS afero.Fs, buildFS afero.Fs) RendererOptionFunc {
	opts := []RendererOptionFunc{
		WithSiteFS(siteFS),
		WithThemeFS(themeFS),
		WithBuildFS(buildFS),
	}
	return func(r *Renderer) error {
		for _, opt := range opts {
			if err := opt.Apply(r); err != nil {
				return err
			}
		}
		return nil
	}
}

func New(opts ...RendererOption) *Renderer {
	return &Renderer{
		options: opts,
		markdown: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithRendererOptions(html.WithUnsafe()),
		),
	}
}

func (r *Renderer) Render(site *tree.Site) error {
	for _, opt := range r.options {
		err := opt.Apply(r)
		if err != nil {
			return err
		}
	}

	return r.render(site, &RenderContext{
		siteFS:  r.siteFS,
		buildFS: r.buildFS,
	})
}

func (r *Renderer) render(root tree.Node, context *RenderContext) error {
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
			if err := r.render(child, context); err != nil {
				return err
			}
		}

		if currentFile := context.CurrentFile(); !root.IsDir() && currentFile != nil {
			if err := r.renderCurrentFile(root, context); err != nil {
				return err
			}
		}

		if err := renderer.Close(root, context); err != nil {
			return err
		}
		break
	}
	return nil
}

func (r *Renderer) renderCurrentFile(root tree.Node, context *RenderContext) error {
	// TODO .Props method on renderer makes no sense
	// It shouldn't be a method at all. Just something
	// done during parsing and attached to the node.
	// I like the term "Attributes"
	// https://github.com/darccio/mergo
	// Want to find a way to make parsers, or some other type composable
	// so that multiple things can attach their own props.
	// Page props, Post props
	source, err := context.Source(root)
	if err != nil {
		return err
	}

	var fm struct {
		Page tree.PageFrontMatter `yaml:"page" mapstructure:"page"`
	}
	content, err := frontmatter.Parse(bytes.NewReader(source), &fm)
	if err != nil {
		return err
	}

	nodeAttrs, err := root.Attributes()
	if err != nil {
		return err
	}

	props := make(map[string]any)
	for k, v := range nodeAttrs {
		props[k] = v
	}

	var templated bytes.Buffer
	if err := r.templater.Render(bytes.NewReader(content), root.Path(), props, &templated); err != nil {
		return err
	}
	var compiledMarkdown bytes.Buffer
	if err := r.markdown.Convert(templated.Bytes(), &compiledMarkdown); err != nil {
		return err
	}
	if fm.Page.Template == "" {
		_, err := io.Copy(context.CurrentFile(), &compiledMarkdown)
		return err
	}
	b, err := WrapTheme(r.themeFS, fm.Page.Template, compiledMarkdown.Bytes(), props)
	if err != nil {
		return err
	}
	if _, err := context.CurrentFile().Write(b); err != nil {
		return err
	}
	return nil
}
