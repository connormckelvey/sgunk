package project

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
)

type PageKind string

type Page interface {
	Kind() PageKind
	Path() string
}

type BasePage struct {
	kind     PageKind
	entry    *FileEntry
	path     string
	meta     []*PageMetaValue
	links    []*PageLinksValue
	template string
	content  []byte
	metadata map[any]any
}

func (bp *BasePage) Kind() PageKind {
	return bp.kind
}
func (bp *BasePage) setKind(kind PageKind) {
	bp.kind = kind
}

func (bp *BasePage) Dir() string {
	return filepath.Dir(bp.Path())
}

func (bp *BasePage) Name() string {
	return filepath.Base(bp.Path())
}

func (bp *BasePage) Entry() *FileEntry {
	return bp.entry
}
func (bp *BasePage) Path() string {
	return bp.path
}
func (bp *BasePage) Meta() []*PageMetaValue {
	return bp.meta
}
func (bp *BasePage) Links() []*PageLinksValue {
	return bp.links
}
func (bp *BasePage) Template() string {
	return bp.template
}
func (bp *BasePage) Content() []byte {
	return bp.content
}

func parsePage(path string, entry *FileEntry) (*BasePage, error) {
	content, err := entry.Content()
	if err != nil {
		return nil, err
	}

	var fm PageFrontMatter
	if _, err := frontmatter.Parse(bytes.NewReader(content), &fm); err != nil {
		return nil, err
	}

	return &BasePage{
		entry:    entry,
		path:     path,
		meta:     fm.Meta,
		links:    fm.Links,
		template: fm.Template,
		content:  content,
		metadata: make(map[any]any),
	}, nil
}

type PageFrontMatter struct {
	Meta     []*PageMetaValue  `yaml:"meta"`
	Links    []*PageLinksValue `yaml:"links"`
	Template string            `yaml:"template"`
}

type PageNameParts struct {
	Raw   string
	Ext   string
	Kind  string
	Slug  string
	Extra []string
}

func GetEntryNameParts(name string) (PageNameParts, bool) {
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return PageNameParts{}, false
	}

	p := PageNameParts{
		Raw:  name,
		Ext:  parts[len(parts)-1],
		Kind: parts[0],
		Slug: parts[0],
	}

	if len(parts) > 2 {
		p.Slug = parts[len(parts)-2]
	}

	if len(parts) > 3 {
		p.Extra = parts[1 : len(parts)-2]
	}

	return p, true
}
