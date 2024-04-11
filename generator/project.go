package generator

import (
	"bytes"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type Project struct {
	options          []ProjectOption
	dir              string
	contentLoaders   []ContentLoader
	pageTransformers []PageTransformer

	Pages []*Page
}

type ProjectOption interface {
	Apply(g *Project) error
}

type ProjectOptionFunc func(g *Project) error

func (apply ProjectOptionFunc) Apply(g *Project) error {
	return apply(g)
}

func NewProject(opts ...ProjectOption) *Project {
	return &Project{
		options: opts,
	}
}

func (p *Project) applyOptions() error {
	// TODO apply defaults

	for _, opt := range p.options {
		if err := opt.Apply(p); err != nil {
			return nil
		}
	}
	return nil
}

func (p *Project) Load() error {
	if err := p.applyOptions(); err != nil {
		return err
	}
	if err := p.WalkSite(); err != nil {
		return err
	}
	return nil
}

func WithProjectDir(dir string) ProjectOptionFunc {
	return func(g *Project) error {
		g.dir = dir
		return nil
	}
}

func (p *Project) ConfigFile() string {
	return filepath.Join(p.dir, "project.json")
}

func (p *Project) SiteDir() string {
	return filepath.Join(p.dir, "site")
}

func WithContentLoader(loader ContentLoader) ProjectOptionFunc {
	return func(p *Project) error {
		p.contentLoaders = append(p.contentLoaders, loader)
		return nil
	}
}

func (p *Project) WalkSite() error {
	siteDir := os.DirFS(p.SiteDir())
	return fs.WalkDir(siteDir, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			log.Println("err walking:", err)
			return nil
		}
		// skip site root dir entry
		rel := filepath.Join(p.SiteDir(), path)
		if filepath.Clean(rel) == filepath.Clean(p.SiteDir()) {
			log.Println("skipping", path)
			return nil
		}

		log.Println("walking", path)

		for _, loader := range p.contentLoaders {
			abs := filepath.Join(p.SiteDir(), path)
			log.Println("loading", abs)

			if err := loader.Load(p, abs, entry); err != nil {
				return err
			}
			log.Println("loaded", abs)
		}
		return nil
	})
}

func WithPageTransformers(transformers ...PageTransformer) ProjectOptionFunc {
	return func(p *Project) error {
		p.pageTransformers = append(p.pageTransformers, transformers...)
		return nil
	}
}

func (p *Project) build() error {
	for _, page := range p.Pages {
		source, err := os.ReadFile(page.SourcePath)
		if err != nil {
			return err
		}

		var src *bytes.Buffer = bytes.NewBuffer(source)
		var dst *bytes.Buffer = bytes.NewBuffer([]byte{})
		for _, transformer := range p.pageTransformers {
			log.Println("transforming", page.SourcePath)
			// log.Println("src", src.String())
			if err := transformer.Transform(p, page, src, dst); err != nil {
				return err
			}
			// log.Println("dst", dst.String())
			src = dst
			dst = bytes.NewBuffer([]byte{})
		}
		// log.Println("final src", src.String())
		// log.Println("final dst", dst.String())

		page.Content = src

	}
	return nil
}

func (p *Project) Build() error {

	if err := p.build(); err != nil {
		return err
	}

	return nil
}
