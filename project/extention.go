package project

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/mitchellh/mapstructure"
)

type Extension interface {
	Register(p *Project) error
	// ProvideContext(context BuildContext, page Page) BuildContext
}

type ExtensionLoader interface {
	Name() string
	LoadExtension(m map[string]any) (Extension, error)
}

type BlogExtensionConfig struct {
	Path string `mapstructure:"path"`
}

type BlogExtension struct {
	config BlogExtensionConfig
}

type BlogExtensionLoader struct{}

func (bel *BlogExtensionLoader) Name() string {
	return "blog"
}

func (bel *BlogExtensionLoader) LoadExtension(m map[string]any) (Extension, error) {
	var config BlogExtensionConfig
	if err := mapstructure.Decode(m, &config); err != nil {
		return nil, err
	}
	return &BlogExtension{config: config}, nil
}

const blogPostPageKind = PageKind("post")

type BlogPostFrontMatter struct {
	Title string   `yaml:"title"`
	Tags  []string `yaml:"tags"`
}

type BlogPostPageParser struct {
	config BlogExtensionConfig
}

func (ep *BlogPostPageParser) Parse(base *BasePage) (PageMetadata, error) {

	if !strings.HasSuffix(base.Dir(), ep.config.Path) {
		return nil, nil
	}

	parts, ok := GetEntryNameParts(base.Name())
	if !ok || PageKind(parts.Kind) != blogPostPageKind {
		return nil, nil
	}

	var metadata BlogPostMetadata
	if len(parts.Extra) > 0 {
		ms, err := strconv.ParseInt(parts.Extra[0], 10, 64)
		if err != nil {
			return nil, err
		}
		metadata.CreatedAt = time.UnixMilli(ms)
	}

	var fm struct {
		Post BlogPostFrontMatter `yaml:"post"`
	}
	if _, err := frontmatter.Parse(bytes.NewReader(base.Content()), &fm); err != nil {
		return nil, err
	}
	metadata.Title = fm.Post.Title
	metadata.Tags = fm.Post.Tags

	return metadata, nil
}

type BlogPostMetadata struct {
	CreatedAt time.Time
	Title     string
	Tags      []string
}

func (m BlogPostMetadata) Kind() PageKind {
	return blogPostPageKind
}

func (m BlogPostMetadata) Metadata() map[string]any {
	return map[string]any{
		"created_at": m.CreatedAt,
		"title":      m.Title,
		"tags":       m.Tags,
	}
}

func NewBlogPostPageParser(config BlogExtensionConfig) *BlogPostPageParser {
	return &BlogPostPageParser{config: config}
}

// func (be *BlogExtension)
func (be *BlogExtension) Register(p *Project) error {
	p.RegisterPageParser(blogPostPageKind, NewBlogPostPageParser(be.config))
	return nil
}
