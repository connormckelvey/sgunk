package ssg

import (
	"log"
	"os"
	"path/filepath"

	"github.com/connormckelvey/ssg/parser"
	"github.com/connormckelvey/ssg/renderer"
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
}

const (
	defaultSiteDir  = "site"
	defaultThemeDir = "theme"
	defaultBuildDir = "_build"
)

func (p *Project) getConfigDir(c DirConfig, defaultDir string) (string, afero.Fs) {
	dir := defaultDir
	if d := c.GetDir(); d != "" {
		dir = d
	}
	fsys := afero.NewBasePathFs(afero.NewOsFs(), filepath.Join(p.workDir, dir))
	return dir, fsys
}

func (p *Project) Generate() error {
	var success bool

	for _, opt := range p.options {
		if err := opt.Apply(p); err != nil {
			return err
		}
	}

	_, siteFS := p.getConfigDir(&p.config.Site, defaultSiteDir)
	_, themeFS := p.getConfigDir(&p.config.Theme, defaultThemeDir)
	buildDir, buildFS := p.getConfigDir(&p.config.Build, defaultBuildDir)
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
		} else {
			if err := os.RemoveAll(buildDir + ".bk"); err != nil {
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

	if err := renderer.WithFS(siteFS, themeFS, buildFS)(p.renderer); err != nil {
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
		&parser.DefaultParser{},
	)(p.parser); err != nil {
		return err
	}

	if err := renderer.WithEntryRenderers(
		&renderer.DefaultRenderer{},
	)(p.renderer); err != nil {
		return err
	}

	site, err := p.parser.Parse()
	if err != nil {
		return err
	}

	if err := p.renderer.Render(site); err != nil {
		return err
	}

	success = true
	return nil
}
