package parser

import (
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/connormckelvey/ssg/tree"
)

type BlogEntryParser struct {
	root string
}

func NewBlogEntryParser(root string) *BlogEntryParser {
	return &BlogEntryParser{
		root: root,
	}
}

func (pp *BlogEntryParser) Test(path string, entry fs.FileInfo) (bool, error) {
	hasPrefix := strings.HasPrefix(path, pp.root)
	if entry.IsDir() {
		return hasPrefix, nil
	}
	name := filepath.Base(path)
	parts, ok := tree.GetEntryNameParts(name)
	if !ok {
		return false, nil
	}
	return hasPrefix && parts.Kind == "post", nil
}

func (pp *BlogEntryParser) Parse(path string, entry fs.FileInfo) (tree.Node, error) {
	if path == pp.root && entry.IsDir() {
		return tree.NewBlogNode(path, pp.root), nil
	}
	if entry.IsDir() {
		return &tree.BlogCollectionNode{
			BaseNode: tree.NewBaseNode(path, true),
		}, nil
	}
	name := filepath.Base(path)
	parts, _ := tree.GetEntryNameParts(name)

	if len(parts.Extra) > 0 {
		ms, err := strconv.ParseInt(parts.Extra[0], 10, 64)
		if err != nil {
			return nil, err
		}
		return tree.NewBlogPostNode(path, parts, time.UnixMilli(ms)), nil
	}

	return tree.NewBlogPostNode(path, parts, time.UnixMilli(0)), nil

}
