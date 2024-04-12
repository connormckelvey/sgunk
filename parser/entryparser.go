package parser

import (
	"io/fs"

	"github.com/connormckelvey/ssg/tree"
)

type EntryParser interface {
	Test(path string, entry fs.FileInfo) (bool, error)
	Parse(path string, entry fs.FileInfo) (tree.Node, error)
}
