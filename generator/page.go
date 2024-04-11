package generator

import (
	"bytes"
	"path/filepath"
)

type PageKind int

const (
	PageKindUnknown PageKind = iota
	PageKindMarkdown
	PageKindHTML
)

var PageKindExtensions = map[PageKind][]string{
	PageKindMarkdown: {".md", ".markdown"},
	PageKindHTML:     {".html", ".htm"},
}

func NewPageKindFromPath(path string) (PageKind, bool) {
	for pageKind, exts := range PageKindExtensions {
		for _, ext := range exts {
			if filepath.Ext(path) == ext {
				return pageKind, true
			}
		}
	}
	return PageKindUnknown, false
}

func (pk PageKind) PathIs(path string) bool {
	for _, ext := range PageKindExtensions[pk] {
		if filepath.Ext(path) == ext {
			return true
		}
	}
	return false
}

type Page struct {
	SourcePath string
	Kind       PageKind
	Metadata   map[string]any
	Content    *bytes.Buffer
}
