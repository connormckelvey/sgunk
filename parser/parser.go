package parser

import (
	"errors"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/connormckelvey/website/tree"
	"github.com/spf13/afero"
)

func Parse(fsys afero.Fs, dir string, root tree.Node, parsers []EntryParser) error {
	entries, err := afero.ReadDir(fsys, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		// find parser
		var parser EntryParser
		for _, p := range parsers {
			ok, err := p.Test(path, entry)
			if err != nil {
				return err
			}
			if ok {
				parser = p
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
			if err := Parse(fsys, path, n, parsers); err != nil {
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
