package parser

import (
	"errors"
	"io/fs"
	"log"
	"path/filepath"
	"sync"

	"github.com/connormckelvey/ssg/tree"
	"github.com/spf13/afero"
)

type Parser struct {
	options []ParserOption
	once    sync.Once
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

	if err := p.parse(".", site); err != nil {
		return nil, err
	}
	return site, nil
}

func (p *Parser) parse(dir string, root tree.Node) error {
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

		n, err := parser.Parse(path, entry)
		if err != nil {
			return err
		}
		if n == nil {
			continue
		}

		root.AppendChild(n)
		if entry.IsDir() {
			if err := p.parse(path, n); err != nil {
				return err
			}
		}
	}
	return nil
}

type DefaultPageParser struct {
}

func (pp *DefaultPageParser) Test(path string, entry fs.FileInfo) (bool, error) {
	return true, nil
}

func (pp *DefaultPageParser) Parse(path string, entry fs.FileInfo) (tree.Node, error) {
	name := filepath.Base(path)

	if entry.IsDir() {
		return &tree.DefaultDir{
			BaseNode: tree.NewBaseNode(path, true),
		}, nil
	}
	parts, ok := tree.GetEntryNameParts(name)
	if !ok {
		return nil, errors.New("shouldnt have passed Test")
	}
	return tree.NewDefaultPage(path, parts), nil
}
