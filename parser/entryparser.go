package parser

import (
	"io/fs"

	"github.com/connormckelvey/sgunk/tree"
)

type EntryParser interface {
	Test(path string, entry fs.FileInfo) (bool, error)
	Parse(path string, entry fs.FileInfo, context *ParserContext) (tree.Node, error)
}
