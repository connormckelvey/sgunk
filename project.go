package ssg

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/connormckelvey/website/parser"
	"github.com/connormckelvey/website/renderer"
	"github.com/connormckelvey/website/tree"
	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
)

type Project struct {
	config   *ProjectConfig
	workDir  string
	options  []ProjectOption
	parser   *parser.Parser
	renderer *renderer.Renderer
}

type ProjectOption interface {
	Apply(*Project) error
}

type ProjectOptionFunc func(*Project) error

func (apply ProjectOptionFunc) Apply(p *Project) error {
	return apply(p)
}

func WithWorkDir(dir string) ProjectOptionFunc {
	return func(p *Project) error {
		p.workDir = dir
		return nil
	}
}

func WithConfig(config *ProjectConfig) ProjectOptionFunc {
	return func(p *Project) error {
		p.config = config
		return nil
	}
}

func WithParserOptions(opts ...parser.ParserOption) ProjectOptionFunc {
	return func(p *Project) error {
		for _, opt := range opts {
			if err := opt.Apply(p.parser); err != nil {
				return err
			}
		}
		return nil
	}
}

func WithRendererOptions(opts ...renderer.RendererOption) ProjectOptionFunc {
	return func(p *Project) error {
		for _, opt := range opts {
			if err := opt.Apply(p.renderer); err != nil {
				return err
			}
		}
		return nil
	}
}

func New(opts ...ProjectOption) *Project {

	return &Project{
		options:  opts,
		parser:   parser.New(),
		renderer: renderer.New(),
	}

	// p := parser.New(
	// // parser.WithSiteFS(siteDir),
	// // parser.WithEntryParsers(
	// // 	parser.NewBlogEntryParser("blog"),
	// // 	&parser.DefaultPageParser{},
	// // ),
	// )

	// r := renderer.New(
	// 	renderer.WithEntryRenderers(
	// 		&renderer.BlogRenderer{},
	// 		&renderer.DefaultRenderer{},
	// 	),
	// )
	// context := renderer.NewRenderContext(
	// 	renderer.WithFS(siteDir, themeDir, buildDir),
	// )
	// err = r.Render(&site, context)
}

func loadConfigFile(projectFS afero.Fs) (*ProjectConfig, error) {
	l, err := afero.ReadDir(projectFS, ".")
	if err != nil {
		panic(err)
	}
	spew.Dump(l)
	for name, unmarshal := range configFiles {
		b, err := afero.ReadFile(projectFS, name)
		if err != nil && os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var c ProjectConfig
		if err := unmarshal(b, &c); err != nil {
			return nil, err
		}
		return &c, nil
	}

	return nil, errors.New("config file not found")
}

func (p *Project) Generate() error {
	var success bool

	for _, opt := range p.options {
		if err := opt.Apply(p); err != nil {
			return err
		}
	}

	siteDir := "site"
	if d := p.config.Site.Dir; d != "" {
		siteDir = d
	}
	siteDir = filepath.Join(p.workDir, siteDir)
	siteFS := afero.NewBasePathFs(afero.NewOsFs(), siteDir)

	themeDir := "theme"
	if d := p.config.Theme.Dir; d != "" {
		themeDir = d
	}
	themeDir = filepath.Join(p.workDir, themeDir)
	themeFS := afero.NewBasePathFs(afero.NewOsFs(), themeDir)

	buildDir := "_build"
	if d := p.config.Build.Dir; d != "" {
		buildDir = d
	}
	buildDir = filepath.Join(p.workDir, buildDir)
	buildFS := afero.NewBasePathFs(afero.NewOsFs(), buildDir)

	err := os.Rename(buildDir, buildDir+".bk")
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	defer func() {
		if !success {
			if err := os.Rename(buildDir, buildDir+".failed"); err != nil {
				log.Println(err)
			}
			if err := os.Rename(buildDir+".bk", buildDir); err != nil {
				log.Println(err)
			}
		}
	}()

	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return err
	}

	if err := parser.WithSiteFS(siteFS)(p.parser); err != nil {
		return err
	}

	for _, use := range p.config.Uses {
		switch use.Name {
		case "blog":
			var config struct {
				Path string `mapstructure:"path"`
			}
			err := mapstructure.Decode(use.Config, &config)
			if err != nil {
				return err
			}
			if err := parser.WithEntryParsers(
				parser.NewBlogEntryParser(config.Path),
			)(p.parser); err != nil {
				return err
			}

			// TODO no need to add more than one blog renderer
			if err := renderer.WithEntryRenderers(
				&renderer.BlogRenderer{},
			)(p.renderer); err != nil {
				return err
			}
		}
	}

	if err := parser.WithEntryParsers(
		&parser.DefaultPageParser{},
	)(p.parser); err != nil {
		return err
	}

	if err := renderer.WithEntryRenderers(
		&renderer.DefaultRenderer{},
	)(p.renderer); err != nil {
		return err
	}

	site := tree.Site{
		BaseNode: tree.NewBaseNode("", true),
	}
	if err := p.parser.Parse(".", &site); err != nil {
		return err
	}
	context := renderer.NewRenderContext(
		renderer.WithFS(siteFS, themeFS, buildFS),
	)
	if err := p.renderer.Render(&site, context); err != nil {
		return err
	}

	success = true
	return nil
}

// func NewGenerator(config *project.ProjectConfig)
