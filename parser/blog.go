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

func (pp *BlogEntryParser) Parse(path string, entry fs.FileInfo, context *ParserContext) (tree.Node, error) {
	if path == pp.root && entry.IsDir() {
		return tree.NewBlogNode(path, pp.root), nil
	}
	if entry.IsDir() {
		return &tree.BlogCollectionNode{
			BaseNode: tree.NewBaseNode(path, true),
		}, nil
	}

	var fm struct {
		Post tree.BlogPostFrontMatter `yaml:"post"`
	}
	if err := context.FrontMatter(path, &fm); err != nil {
		return nil, err
	}

	name := filepath.Base(path)
	parts, _ := tree.GetEntryNameParts(name)

	var createdAt time.Time
	if len(parts.Extra) > 0 {
		ms, err := strconv.ParseInt(parts.Extra[0], 10, 64)
		if err != nil {
			return nil, err
		}
		createdAt = time.UnixMilli(ms)
	}

	node := tree.NewBlogPostNode(path, parts, createdAt)
	err := node.AddAttrs("post", BlogPostAttributes{
		Title:     fm.Post.Title,
		Tags:      fm.Post.Tags,
		CreatedAt: createdAt.Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	return node, nil
}

type BlogPostAttributes struct {
	Title     string   `mapstructure:"title"`
	Tags      []string `mapstructure:"tags"`
	CreatedAt string   `mapstructure:"createdAt"`
}
