package parser

import (
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/connormckelvey/sgunk/tree"
)

type DefaultParser struct {
}

func (pp *DefaultParser) Test(path string, entry fs.FileInfo) (bool, error) {
	return true, nil
}

func (pp *DefaultParser) Parse(path string, entry fs.FileInfo, context *ParserContext) (tree.Node, error) {
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

type PageAttributes struct {
	Title    string                 `mapstructure:"title"`
	Meta     []*tree.PageMetaValue  `mapstructure:"meta"`
	Links    []*tree.PageLinksValue `mapstructure:"links"`
	Template string                 `mapstructure:"template"`
}
