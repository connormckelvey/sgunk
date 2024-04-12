package parser

import (
	"log"
	"path/filepath"

	"github.com/connormckelvey/sgunk/tree"
	"github.com/spf13/afero"
)

type Parser struct {
	options []ParserOption
	siteFS  afero.Fs
	parsers []EntryParser
}

type ParserOption interface {
	Apply(*Parser) error
}

type ParserOptionFunc func(*Parser) error

func (apply ParserOptionFunc) Apply(p *Parser) error {
	return apply(p)
}

func WithSiteFS(siteFS afero.Fs) ParserOptionFunc {
	return func(p *Parser) error {
		p.siteFS = siteFS
		return nil
	}
}

func WithEntryParsers(parsers ...EntryParser) ParserOptionFunc {
	return func(p *Parser) error {
		p.parsers = append(p.parsers, parsers...)
		return nil
	}
}

func New(opts ...ParserOption) *Parser {
	return &Parser{
		options: opts,
	}
}

func (p *Parser) Parse() (*tree.Site, error) {
	for _, opt := range p.options {
		err := opt.Apply(p)
		if err != nil {
			return nil, err
		}
	}

	site := &tree.Site{
		BaseNode: tree.NewBaseNode("", true),
	}
	context := &ParserContext{
		siteFS:  p.siteFS,
		sources: make(map[string][]byte),
	}
	if err := p.parse(".", site, context); err != nil {
		return nil, err
	}
	return site, nil
}

func (p *Parser) parse(dir string, root tree.Node, context *ParserContext) error {
	entries, err := afero.ReadDir(p.siteFS, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		// find parser
		var parser EntryParser
		for _, pp := range p.parsers {
			ok, err := pp.Test(path, entry)
			if err != nil {
				return err
			}
			if ok {
				parser = pp
				break
			}
		}

		if parser == nil {
			log.Printf("no parser for '%s', skipping...", path)
			continue
		}

		n, err := parser.Parse(path, entry, context)
		if err != nil {
			return err
		}
		if n == nil {
			continue
		}

		root.AppendChild(n)

		if entry.IsDir() {
			if err := p.parse(path, n, context); err != nil {
				return err
			}
			continue
		}

		var fm struct {
			Page tree.PageFrontMatter `yaml:"page"`
		}
		if err := context.FrontMatter(path, &fm); err != nil {
			return err
		}

		err = n.AddAttrs("page", PageAttributes{
			Title:    fm.Page.Title,
			Meta:     fm.Page.Meta,
			Links:    fm.Page.Links,
			Template: fm.Page.Template,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
